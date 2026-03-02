package service

import (
	"math"
	"sort"
	"tripshare/internal/models"
	"tripshare/internal/repository"
)

type BalanceService struct {
	memberRepo *repository.MemberRepository
	tripRepo   *repository.TripRepository
}

func NewBalanceService(
	memberRepo *repository.MemberRepository,
	tripRepo *repository.TripRepository,
) *BalanceService {
	return &BalanceService{
		memberRepo: memberRepo,
		tripRepo:   tripRepo,
	}
}

func (s *BalanceService) GetBalances() (*models.BalanceResponse, error) {
	members, err := s.memberRepo.FindAll()
	if err != nil {
		return nil, err
	}

	balances := make([]models.MemberBalance, len(members))
	for i, m := range members {
		balances[i] = models.MemberBalance{
			Member:  m,
			Balance: m.Balance,
		}
	}

	settlements := ComputeSettlements(balances)

	return &models.BalanceResponse{
		Balances:    balances,
		Settlements: settlements,
	}, nil
}

// RecalculateAll recomputes every member's balance from all trips and payments.
func (s *BalanceService) RecalculateAll() error {
	members, err := s.memberRepo.FindAll()
	if err != nil {
		return err
	}
	trips, err := s.tripRepo.FindAll()
	if err != nil {
		return err
	}

	balanceMap := ComputeBalanceMap(members, trips)

	for _, m := range members {
		bal := math.Round(balanceMap[m.ID]*100) / 100
		if err := s.memberRepo.UpdateBalance(m.ID, bal); err != nil {
			return err
		}
	}
	return nil
}

// ComputeBalanceMap calculates the balance for each member across all trips.
// Positive = overpaid (is owed money), negative = underpaid (owes money).
func ComputeBalanceMap(members []models.Member, trips []models.Trip) map[uint]float64 {
	balanceMap := make(map[uint]float64)
	for _, m := range members {
		balanceMap[m.ID] = 0
	}

	for _, trip := range trips {
		if len(trip.Members) == 0 {
			continue
		}
		equalShare := trip.TotalCost / float64(len(trip.Members))

		participantSet := make(map[uint]bool)
		for _, m := range trip.Members {
			participantSet[m.ID] = true
		}

		paidByMember := make(map[uint]float64)
		for _, m := range trip.Members {
			paidByMember[m.ID] = 0
		}
		for _, p := range trip.Payments {
			if participantSet[p.MemberID] {
				paidByMember[p.MemberID] += p.Amount
			}
		}

		for memberID, paid := range paidByMember {
			balanceMap[memberID] += paid - equalShare
		}
	}

	return balanceMap
}

// ComputeSettlements builds a settlement list: for each creditor (overpayer),
// every debtor owes them overpayment / number_of_debtors.
func ComputeSettlements(balances []models.MemberBalance) []models.Settlement {
	type entry struct {
		id      uint
		name    string
		balance float64
	}

	var debtors []entry
	var creditors []entry

	for _, b := range balances {
		if b.Balance < -0.01 {
			debtors = append(debtors, entry{b.Member.ID, b.Member.Name, b.Balance})
		} else if b.Balance > 0.01 {
			creditors = append(creditors, entry{b.Member.ID, b.Member.Name, b.Balance})
		}
	}

	if len(debtors) == 0 || len(creditors) == 0 {
		return nil
	}

	sort.Slice(creditors, func(i, j int) bool { return creditors[i].balance > creditors[j].balance })
	sort.Slice(debtors, func(i, j int) bool { return debtors[i].name < debtors[j].name })

	var settlements []models.Settlement
	for _, c := range creditors {
		perDebtor := math.Round(c.balance/float64(len(debtors))*100) / 100
		if perDebtor < 0.01 {
			continue
		}
		for _, d := range debtors {
			settlements = append(settlements, models.Settlement{
				FromID:   d.id,
				FromName: d.name,
				ToID:     c.id,
				ToName:   c.name,
				Amount:   perDebtor,
			})
		}
	}

	return settlements
}
