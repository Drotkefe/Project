package service

import (
	"math"
	"testing"
	"tripshare/internal/models"
)

func TestComputeBalanceMap_SingleTrip(t *testing.T) {
	members := []models.Member{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Carol"},
		{ID: 4, Name: "Dave"},
	}

	trips := []models.Trip{
		{
			ID:        1,
			TotalCost: 1000,
			Members:   members,
			Payments: []models.Payment{
				{ID: 1, TripID: 1, MemberID: 1, Amount: 1000},
			},
		},
	}

	balances := ComputeBalanceMap(members, trips)

	assertBalance(t, balances, 1, 750)
	assertBalance(t, balances, 2, -250)
	assertBalance(t, balances, 3, -250)
	assertBalance(t, balances, 4, -250)
}

func TestComputeBalanceMap_MultipleTrips(t *testing.T) {
	members := []models.Member{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Carol"},
		{ID: 4, Name: "Dave"},
	}

	trips := []models.Trip{
		{
			ID:        1,
			TotalCost: 1000,
			Members:   members,
			Payments: []models.Payment{
				{ID: 1, TripID: 1, MemberID: 1, Amount: 1000},
			},
		},
		{
			ID:        2,
			TotalCost: 400,
			Members:   members,
			Payments: []models.Payment{
				{ID: 2, TripID: 2, MemberID: 2, Amount: 400},
			},
		},
	}

	balances := ComputeBalanceMap(members, trips)

	// Trip 1: Alice +750, Bob -250, Carol -250, Dave -250
	// Trip 2: Alice -100, Bob +300, Carol -100, Dave -100
	// Total:  Alice +650, Bob +50,  Carol -350, Dave -350
	assertBalance(t, balances, 1, 650)
	assertBalance(t, balances, 2, 50)
	assertBalance(t, balances, 3, -350)
	assertBalance(t, balances, 4, -350)
}

func TestComputeBalanceMap_EvenSplit(t *testing.T) {
	members := []models.Member{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}

	trips := []models.Trip{
		{
			ID:        1,
			TotalCost: 100,
			Members:   members,
			Payments: []models.Payment{
				{ID: 1, TripID: 1, MemberID: 1, Amount: 50},
				{ID: 2, TripID: 1, MemberID: 2, Amount: 50},
			},
		},
	}

	balances := ComputeBalanceMap(members, trips)

	assertBalance(t, balances, 1, 0)
	assertBalance(t, balances, 2, 0)
}

func TestComputeBalanceMap_PartialParticipants(t *testing.T) {
	allMembers := []models.Member{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Carol"},
	}

	tripMembers := []models.Member{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}

	trips := []models.Trip{
		{
			ID:        1,
			TotalCost: 200,
			Members:   tripMembers,
			Payments: []models.Payment{
				{ID: 1, TripID: 1, MemberID: 1, Amount: 200},
			},
		},
	}

	balances := ComputeBalanceMap(allMembers, trips)

	assertBalance(t, balances, 1, 100)  // paid 200, share 100 → +100
	assertBalance(t, balances, 2, -100) // paid 0, share 100 → -100
	assertBalance(t, balances, 3, 0)    // not a participant
}

func TestComputeBalanceMap_NoTrips(t *testing.T) {
	members := []models.Member{
		{ID: 1, Name: "Alice"},
	}

	balances := ComputeBalanceMap(members, nil)
	assertBalance(t, balances, 1, 0)
}

func TestComputeSettlements_Basic(t *testing.T) {
	balances := []models.MemberBalance{
		{Member: models.Member{ID: 1, Name: "Alice"}, Balance: 750},
		{Member: models.Member{ID: 2, Name: "Bob"}, Balance: -250},
		{Member: models.Member{ID: 3, Name: "Carol"}, Balance: -250},
		{Member: models.Member{ID: 4, Name: "Dave"}, Balance: -250},
	}

	settlements := ComputeSettlements(balances)

	// 1 creditor (Alice +750), 3 debtors → each debtor pays 750/3 = 250
	if len(settlements) != 3 {
		t.Fatalf("expected 3 settlements, got %d", len(settlements))
	}

	for _, s := range settlements {
		if s.ToID != 1 {
			t.Errorf("expected all payments to go to Alice (ID 1), got %d", s.ToID)
		}
		if math.Abs(s.Amount-250) > 0.01 {
			t.Errorf("expected each debtor to pay 250, got %.2f", s.Amount)
		}
	}
}

func TestComputeSettlements_AllEven(t *testing.T) {
	balances := []models.MemberBalance{
		{Member: models.Member{ID: 1, Name: "Alice"}, Balance: 0},
		{Member: models.Member{ID: 2, Name: "Bob"}, Balance: 0},
	}

	settlements := ComputeSettlements(balances)
	if len(settlements) != 0 {
		t.Errorf("expected 0 settlements when balanced, got %d", len(settlements))
	}
}

func TestComputeSettlements_TwoCreditors(t *testing.T) {
	balances := []models.MemberBalance{
		{Member: models.Member{ID: 1, Name: "Alice"}, Balance: 300},
		{Member: models.Member{ID: 2, Name: "Bob"}, Balance: -100},
		{Member: models.Member{ID: 3, Name: "Carol"}, Balance: 100},
		{Member: models.Member{ID: 4, Name: "Dave"}, Balance: -300},
	}

	settlements := ComputeSettlements(balances)

	// 2 creditors (Alice +300, Carol +100), 2 debtors (Bob, Dave)
	// Alice: each debtor pays 300/2 = 150 → 2 settlements
	// Carol: each debtor pays 100/2 = 50  → 2 settlements
	if len(settlements) != 4 {
		t.Fatalf("expected 4 settlements, got %d", len(settlements))
	}

	aliceTotal := 0.0
	carolTotal := 0.0
	for _, s := range settlements {
		if s.ToID == 1 {
			aliceTotal += s.Amount
		} else if s.ToID == 3 {
			carolTotal += s.Amount
		}
	}
	if math.Abs(aliceTotal-300) > 0.01 {
		t.Errorf("expected Alice to receive 300, got %.2f", aliceTotal)
	}
	if math.Abs(carolTotal-100) > 0.01 {
		t.Errorf("expected Carol to receive 100, got %.2f", carolTotal)
	}
}

func assertBalance(t *testing.T, balances map[uint]float64, id uint, expected float64) {
	t.Helper()
	got := balances[id]
	if math.Abs(got-expected) > 0.01 {
		t.Errorf("member %d: expected balance %.2f, got %.2f", id, expected, got)
	}
}
