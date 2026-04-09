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

type tripInfo struct {
	name   string
	amount float64
}

type graphEdge struct {
	amount float64
	trips  map[uint]*tripInfo // tripID -> per-trip contribution
}

type Graph map[string]map[string]*graphEdge

type nettingEvent struct {
	debtorName   string
	creditorName string
	amount       float64
	tripID       uint
}

type settlementResult struct {
	Settlements     []models.Settlement
	TripAdjustments map[uint]map[uint]float64
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
	trips, err := s.tripRepo.FindAll()
	if err != nil {
		return nil, err
	}

	perTripBalances := make([][]models.MemberBalance, 0, len(trips))

	for _, trip := range trips {
		if len(trip.Members) == 0 {
			continue
		}
		equalShare := trip.TotalCost / float64(len(trip.Members))

		paidByMember := make(map[uint]float64)
		for _, p := range trip.Payments {
			paidByMember[p.MemberID] += p.Amount
		}

		tripBalances := make([]models.MemberBalance, 0, len(trip.Members))
		for _, m := range trip.Members {
			bal := paidByMember[m.ID] - equalShare
			tripBalances = append(tripBalances, models.MemberBalance{
				Member:   m,
				Balance:  bal,
				TripID:   trip.ID,
				TripName: trip.Name,
			})
		}
		perTripBalances = append(perTripBalances, tripBalances)
	}

	result := ComputeSettlements(perTripBalances)

	nameToID := make(map[string]uint)
	for _, m := range members {
		nameToID[m.Name] = m.ID
	}
	netPosition := make(map[uint]float64)
	for _, s := range result.Settlements {
		netPosition[nameToID[s.ToName]] += s.Amount
		netPosition[nameToID[s.FromName]] -= s.Amount
	}

	balances := make([]models.MemberBalance, 0, len(members))
	for _, m := range members {
		balances = append(balances, models.MemberBalance{
			Member:  m,
			Balance: math.Round(netPosition[m.ID]*100) / 100,
		})
	}

	return &models.BalanceResponse{
		Balances:        balances,
		Settlements:     result.Settlements,
		TripAdjustments: result.TripAdjustments,
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

// ComputeSettlements builds a directed graph of who-owes-whom across all trips,
// then simplifies bidirectional edges to net amounts.
func ComputeSettlements(balances [][]models.MemberBalance) settlementResult {
	graph := Graph{}
	nameToID := make(map[string]uint)

	for _, trip := range balances {
		var tripID uint
		var tripName string
		var creditors []models.MemberBalance
		totalPositive := 0.0
		for _, m := range trip {
			nameToID[m.Member.Name] = m.Member.ID
			if tripID == 0 {
				tripID = m.TripID
				tripName = m.TripName
			}
			if m.Balance > 0.01 {
				creditors = append(creditors, m)
				totalPositive += m.Balance
			}
		}
		if len(creditors) == 0 {
			continue
		}

		for _, debtor := range trip {
			if debtor.Balance >= -0.01 {
				continue
			}
			debt := math.Abs(debtor.Balance)
			for _, creditor := range creditors {
				share := debt * creditor.Balance / totalPositive
				if share < 0.01 {
					continue
				}
				from := debtor.Member.Name
				to := creditor.Member.Name
				if graph[from] == nil {
					graph[from] = make(map[string]*graphEdge)
				}
				edge := graph[from][to]
				if edge == nil {
					edge = &graphEdge{trips: make(map[uint]*tripInfo)}
					graph[from][to] = edge
				}
				edge.amount += share
				ti := edge.trips[tripID]
				if ti == nil {
					ti = &tripInfo{name: tripName}
					edge.trips[tripID] = ti
				}
				ti.amount += share
			}
		}
	}

	events := simplifyGraph(graph)

	adjustments := make(map[uint]map[uint]float64)
	for _, ev := range events {
		if adjustments[ev.tripID] == nil {
			adjustments[ev.tripID] = make(map[uint]float64)
		}
		adjustments[ev.tripID][nameToID[ev.debtorName]] += ev.amount
		adjustments[ev.tripID][nameToID[ev.creditorName]] -= ev.amount
	}
	for tid, members := range adjustments {
		for mid, v := range members {
			adjustments[tid][mid] = math.Round(v*100) / 100
		}
	}

	var settlements []models.Settlement
	for from, neighbors := range graph {
		for to, edge := range neighbors {
			if edge.amount < 0.01 {
				continue
			}
			var tripNames []string
			for _, ti := range edge.trips {
				if ti.amount > 0.01 {
					tripNames = append(tripNames, ti.name)
				}
			}
			settlements = append(settlements, models.Settlement{
				FromName:  from,
				ToName:    to,
				Amount:    math.Round(edge.amount*100) / 100,
				TripNames: tripNames,
			})
		}
	}

	return settlementResult{
		Settlements:     settlements,
		TripAdjustments: adjustments,
	}
}

func simplifyGraph(graph Graph) []nettingEvent {
	var events []nettingEvent
	visited := make(map[string]map[string]bool)

	for u := range graph {
		if _, ok := visited[u]; !ok {
			visited[u] = make(map[string]bool)
		}

		for v, edgeUV := range graph[u] {
			if visited[u][v] || visited[v][u] {
				continue
			}

			if reverseNeighbors, ok := graph[v]; ok {
				if edgeVU, exists := reverseNeighbors[u]; exists {
					if edgeUV.amount > edgeVU.amount {
						events = append(events, netEdgePair(edgeVU, v, u, edgeUV, u, v)...)
						edgeUV.amount -= edgeVU.amount
						delete(graph[v], u)
					} else if edgeVU.amount > edgeUV.amount {
						events = append(events, netEdgePair(edgeUV, u, v, edgeVU, v, u)...)
						edgeVU.amount -= edgeUV.amount
						delete(graph[u], v)
					} else {
						for tripID, ti := range edgeUV.trips {
							events = append(events, nettingEvent{u, v, ti.amount, tripID})
						}
						for tripID, ti := range edgeVU.trips {
							events = append(events, nettingEvent{v, u, ti.amount, tripID})
						}
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

	return events
}

// netEdgePair emits netting events when fully cancelling `cancelled` against the
// surviving `survivor` edge, and proportionally deducts from survivor's trip contribs.
func netEdgePair(
	cancelled *graphEdge, cFrom, cTo string,
	survivor *graphEdge, sFrom, sTo string,
) []nettingEvent {
	var events []nettingEvent
	amount := cancelled.amount

	for tripID, ti := range cancelled.trips {
		events = append(events, nettingEvent{cFrom, cTo, ti.amount, tripID})
	}

	for tripID, ti := range survivor.trips {
		deducted := amount * ti.amount / survivor.amount
		ti.amount -= deducted
		events = append(events, nettingEvent{sFrom, sTo, deducted, tripID})
	}

	return events
}

