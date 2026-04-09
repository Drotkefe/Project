package service

import (
	"math"
	"tripshare/internal/models"
	"tripshare/internal/repository"
)

type BalanceService struct {
	memberRepo *repository.MemberRepository
	tripRepo   *repository.TripRepository
}

type Graph map[string]map[string]int

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
	trips, err := s.tripRepo.FindAll()
	if err != nil {
		return nil, err
	}

	balances := make([][]models.MemberBalance, len(trips))
	for idx, trip := range trips {
		tripMembers := make([]models.MemberBalance, len(trip.Members))
		for i := 0; i < len(trip.Members); i++ {
			m := trip.Members[i]
			tripMembers[i] = models.MemberBalance{
				Member:  m,
				Balance: m.Balance,
				TripID:  trip.ID,
			}
		}
		balances[idx] = tripMembers
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
func ComputeSettlements(balances [][]models.MemberBalance) []models.Settlement {
	graph := Graph{}

	for _, trip := range balances {
		payerName, _ := findPayer(trip)
		for _, member := range trip {
			if member.Balance < -0.01 {
				if _, exists := graph[member.Member.Name]; !exists {
					graph[member.Member.Name] = make(map[string]int)
				}
				graph[member.Member.Name][payerName] = int(math.Abs(member.Balance))
			}
		}
	}

	simplifyGraph(graph)

	var settlements []models.Settlement
	for i, node := range graph {
		for name, amount := range node {
			settlements = append(settlements, models.Settlement{
				FromName: i,
				ToName:   name,
				Amount:   float64(amount),
			})
		}
	}

	return settlements
}

func findPayer(trip []models.MemberBalance) (string, uint) {
	for _, m := range trip {
		if m.Balance > 0 {
			return m.Member.Name, m.Member.ID
		}
	}
	return "", 0
}

func simplifyGraph(graph map[string]map[string]int) {
	visited := make(map[string]map[string]bool)

	for u, neighbors := range graph {
		if _, ok := visited[u]; !ok {
			visited[u] = make(map[string]bool)
		}

		for v, weightUV := range neighbors {
			if visited[u][v] || visited[v][u] {
				continue
			}

			if reverseNeighbors, ok := graph[v]; ok {
				if weightVU, exists := reverseNeighbors[u]; exists {

					if weightUV > weightVU {
						graph[u][v] = weightUV - weightVU
						delete(graph[v], u)
					} else if weightVU > weightUV {
						graph[v][u] = weightVU - weightUV
						delete(graph[u], v)
					} else {
						delete(graph[u], v)
						delete(graph[v], u)
					}
				}
			}

			if _, ok := visited[v]; !ok {
				visited[v] = make(map[string]bool)
			}
			visited[u][v] = true
			visited[v][u] = true
		}
	}
}
