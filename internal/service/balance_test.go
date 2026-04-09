package service

import (
	"math"
	"sort"
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
	balances := [][]models.MemberBalance{{
		{Member: models.Member{ID: 1, Name: "Alice"}, Balance: 750},
		{Member: models.Member{ID: 2, Name: "Bob"}, Balance: -250},
		{Member: models.Member{ID: 3, Name: "Carol"}, Balance: -250},
		{Member: models.Member{ID: 4, Name: "Dave"}, Balance: -250},
	}}

	result := ComputeSettlements(balances)

	if len(result.Settlements) != 3 {
		t.Fatalf("expected 3 settlements, got %d", len(result.Settlements))
	}

	for _, s := range result.Settlements {
		if s.ToName != "Alice" {
			t.Errorf("expected all payments to go to Alice, got %s", s.ToName)
		}
		if math.Abs(s.Amount-250) > 0.01 {
			t.Errorf("expected each debtor to pay 250, got %.2f", s.Amount)
		}
	}
}

func TestComputeSettlements_AllEven(t *testing.T) {
	balances := [][]models.MemberBalance{{
		{Member: models.Member{ID: 1, Name: "Alice"}, Balance: 0},
		{Member: models.Member{ID: 2, Name: "Bob"}, Balance: 0},
	}}

	result := ComputeSettlements(balances)
	if len(result.Settlements) != 0 {
		t.Errorf("expected 0 settlements when balanced, got %d", len(result.Settlements))
	}
}

func TestComputeSettlements_TwoCreditors(t *testing.T) {
	balances := [][]models.MemberBalance{{
		{Member: models.Member{ID: 1, Name: "Alice"}, Balance: 300},
		{Member: models.Member{ID: 2, Name: "Bob"}, Balance: -100},
		{Member: models.Member{ID: 3, Name: "Carol"}, Balance: 100},
		{Member: models.Member{ID: 4, Name: "Dave"}, Balance: -300},
	}}

	result := ComputeSettlements(balances)

	if len(result.Settlements) != 4 {
		t.Fatalf("expected 4 settlements, got %d", len(result.Settlements))
	}

	aliceTotal := 0.0
	carolTotal := 0.0
	for _, s := range result.Settlements {
		if s.ToName == "Alice" {
			aliceTotal += s.Amount
		} else if s.ToName == "Carol" {
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

func TestComputeSettlements_MultiTrip_GraphSimplification(t *testing.T) {
	kaja := []models.MemberBalance{
		{Member: models.Member{ID: 1, Name: "Patrik"}, Balance: -2000, TripID: 1, TripName: "Kaja"},
		{Member: models.Member{ID: 2, Name: "Balint"}, Balance: -2000, TripID: 1, TripName: "Kaja"},
		{Member: models.Member{ID: 3, Name: "Balazs"}, Balance: -2000, TripID: 1, TripName: "Kaja"},
		{Member: models.Member{ID: 4, Name: "Peti"}, Balance: 8000, TripID: 1, TripName: "Kaja"},
	}
	pia := []models.MemberBalance{
		{Member: models.Member{ID: 1, Name: "Patrik"}, Balance: 15000, TripID: 2, TripName: "Pia"},
		{Member: models.Member{ID: 2, Name: "Balint"}, Balance: -3750, TripID: 2, TripName: "Pia"},
		{Member: models.Member{ID: 3, Name: "Balazs"}, Balance: -3750, TripID: 2, TripName: "Pia"},
		{Member: models.Member{ID: 4, Name: "Peti"}, Balance: -3750, TripID: 2, TripName: "Pia"},
	}
	sajtburi := []models.MemberBalance{
		{Member: models.Member{ID: 1, Name: "Patrik"}, Balance: -1200, TripID: 3, TripName: "Sajtburi"},
		{Member: models.Member{ID: 2, Name: "Balint"}, Balance: 3600, TripID: 3, TripName: "Sajtburi"},
		{Member: models.Member{ID: 3, Name: "Balazs"}, Balance: -1200, TripID: 3, TripName: "Sajtburi"},
	}

	result := ComputeSettlements([][]models.MemberBalance{kaja, pia, sajtburi})

	if len(result.Settlements) != 6 {
		t.Fatalf("expected 6 settlements, got %d", len(result.Settlements))
	}

	assertSettlement(t, result.Settlements, "Peti", "Patrik", 1750, "Pia")
	assertSettlement(t, result.Settlements, "Balint", "Patrik", 2550, "Pia")
	assertSettlement(t, result.Settlements, "Balint", "Peti", 2000, "Kaja")
	assertSettlement(t, result.Settlements, "Balazs", "Patrik", 3750, "Pia")
	assertSettlement(t, result.Settlements, "Balazs", "Peti", 2000, "Kaja")
	assertSettlement(t, result.Settlements, "Balazs", "Balint", 1200, "Sajtburi")

	adj := result.TripAdjustments

	// Kaja (trip 1): Patrik's -2000 debt to Peti netted away
	assertAdjustment(t, adj, 1, 1, 2000)   // Patrik freed
	assertAdjustment(t, adj, 1, 4, -2000)  // Peti loses that credit

	// Pia (trip 2): Peti's debt reduced by 2000, Balint's by 1200, Patrik's credit reduced by both
	assertAdjustment(t, adj, 2, 4, 2000)   // Peti's debt reduced
	assertAdjustment(t, adj, 2, 2, 1200)   // Balint's debt reduced
	assertAdjustment(t, adj, 2, 1, -3200)  // Patrik loses 2000+1200 credit

	// Sajtburi (trip 3): Patrik's -1200 debt to Balint netted away
	assertAdjustment(t, adj, 3, 1, 1200)   // Patrik freed
	assertAdjustment(t, adj, 3, 2, -1200)  // Balint loses that credit
}

func assertBalance(t *testing.T, balances map[uint]float64, id uint, expected float64) {
	t.Helper()
	got := balances[id]
	if math.Abs(got-expected) > 0.01 {
		t.Errorf("member %d: expected balance %.2f, got %.2f", id, expected, got)
	}
}

func assertAdjustment(t *testing.T, adj map[uint]map[uint]float64, tripID, memberID uint, expected float64) {
	t.Helper()
	got := 0.0
	if m, ok := adj[tripID]; ok {
		got = m[memberID]
	}
	if math.Abs(got-expected) > 0.01 {
		t.Errorf("trip %d, member %d: expected adjustment %.2f, got %.2f", tripID, memberID, expected, got)
	}
}

func assertSettlement(t *testing.T, settlements []models.Settlement, from, to string, expectedAmount float64, expectedTrips ...string) {
	t.Helper()
	for _, s := range settlements {
		if s.FromName == from && s.ToName == to {
			if math.Abs(s.Amount-expectedAmount) > 0.01 {
				t.Errorf("%s -> %s: expected %.2f, got %.2f", from, to, expectedAmount, s.Amount)
			}
			if len(expectedTrips) > 0 {
				got := make([]string, len(s.TripNames))
				copy(got, s.TripNames)
				want := make([]string, len(expectedTrips))
				copy(want, expectedTrips)
				sort.Strings(got)
				sort.Strings(want)
				if len(got) != len(want) {
					t.Errorf("%s -> %s: expected trips %v, got %v", from, to, want, got)
				} else {
					for i := range got {
						if got[i] != want[i] {
							t.Errorf("%s -> %s: expected trips %v, got %v", from, to, want, got)
							break
						}
					}
				}
			}
			return
		}
	}
	t.Errorf("settlement %s -> %s: %.2f not found", from, to, expectedAmount)
}
