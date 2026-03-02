package service

import (
	"errors"
	"strings"
	"time"
	"tripshare/internal/models"
	"tripshare/internal/repository"
)

type TripService struct {
	tripRepo   *repository.TripRepository
	memberRepo *repository.MemberRepository
	balance    *BalanceService
}

func NewTripService(
	tripRepo *repository.TripRepository,
	memberRepo *repository.MemberRepository,
	balance *BalanceService,
) *TripService {
	return &TripService{
		tripRepo:   tripRepo,
		memberRepo: memberRepo,
		balance:    balance,
	}
}

func (s *TripService) Create(req models.CreateTripRequest) (*models.Trip, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("trip name is required")
	}
	if req.TotalCost < 0 {
		return nil, errors.New("total cost cannot be negative")
	}
	if len(req.MemberIDs) == 0 {
		return nil, errors.New("at least one member is required")
	}

	members, err := s.memberRepo.FindByIDs(req.MemberIDs)
	if err != nil {
		return nil, err
	}
	if len(members) != len(req.MemberIDs) {
		return nil, errors.New("one or more member IDs are invalid")
	}

	date := time.Now()
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errors.New("invalid date format, expected YYYY-MM-DD")
		}
		date = parsed
	}

	trip := &models.Trip{
		Name:      name,
		TotalCost: req.TotalCost,
		Date:      date,
		Members:   members,
	}

	if err := s.tripRepo.Create(trip); err != nil {
		return nil, err
	}

	if err := s.balance.RecalculateAll(); err != nil {
		return nil, err
	}

	return s.tripRepo.FindByID(trip.ID)
}

func (s *TripService) List() ([]models.Trip, error) {
	return s.tripRepo.FindAll()
}

func (s *TripService) Get(id uint) (*models.TripDetail, error) {
	trip, err := s.tripRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("trip not found")
	}

	equalShare := 0.0
	if len(trip.Members) > 0 {
		equalShare = trip.TotalCost / float64(len(trip.Members))
	}

	paidByMember := make(map[uint]float64)
	for _, m := range trip.Members {
		paidByMember[m.ID] = 0
	}
	for _, p := range trip.Payments {
		paidByMember[p.MemberID] += p.Amount
	}

	var breakdowns []models.TripBreakdown
	for _, m := range trip.Members {
		paid := paidByMember[m.ID]
		breakdowns = append(breakdowns, models.TripBreakdown{
			MemberID:   m.ID,
			MemberName: m.Name,
			Paid:       paid,
			EqualShare: equalShare,
			Delta:      paid - equalShare,
		})
	}

	return &models.TripDetail{
		Trip:       *trip,
		EqualShare: equalShare,
		Breakdowns: breakdowns,
	}, nil
}

func (s *TripService) Update(id uint, req models.UpdateTripRequest) (*models.Trip, error) {
	trip, err := s.tripRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("trip not found")
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		trip.Name = name
	}
	if req.TotalCost >= 0 {
		trip.TotalCost = req.TotalCost
	}
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, errors.New("invalid date format, expected YYYY-MM-DD")
		}
		trip.Date = parsed
	}

	if err := s.tripRepo.Update(trip); err != nil {
		return nil, err
	}

	if len(req.MemberIDs) > 0 {
		members, err := s.memberRepo.FindByIDs(req.MemberIDs)
		if err != nil {
			return nil, err
		}
		if len(members) != len(req.MemberIDs) {
			return nil, errors.New("one or more member IDs are invalid")
		}
		if err := s.tripRepo.ReplaceMembers(trip, members); err != nil {
			return nil, err
		}
	}

	if err := s.balance.RecalculateAll(); err != nil {
		return nil, err
	}

	return s.tripRepo.FindByID(id)
}

func (s *TripService) Delete(id uint) error {
	if _, err := s.tripRepo.FindByID(id); err != nil {
		return errors.New("trip not found")
	}
	if err := s.tripRepo.Delete(id); err != nil {
		return err
	}
	return s.balance.RecalculateAll()
}
