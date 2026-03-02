package service

import (
	"errors"
	"tripshare/internal/models"
	"tripshare/internal/repository"
)

type PaymentService struct {
	paymentRepo *repository.PaymentRepository
	tripRepo    *repository.TripRepository
	balance     *BalanceService
}

func NewPaymentService(
	paymentRepo *repository.PaymentRepository,
	tripRepo *repository.TripRepository,
	balance *BalanceService,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		tripRepo:    tripRepo,
		balance:     balance,
	}
}

func (s *PaymentService) Create(tripID uint, req models.CreatePaymentRequest) (*models.Payment, error) {
	trip, err := s.tripRepo.FindByID(tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}

	if req.Amount < 0 {
		return nil, errors.New("payment amount cannot be negative")
	}

	isMember := false
	for _, m := range trip.Members {
		if m.ID == req.MemberID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.New("member is not a participant of this trip")
	}

	payment := &models.Payment{
		TripID:   tripID,
		MemberID: req.MemberID,
		Amount:   req.Amount,
	}
	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, err
	}

	if err := s.balance.RecalculateAll(); err != nil {
		return nil, err
	}

	return s.paymentRepo.FindByID(payment.ID)
}

func (s *PaymentService) Update(tripID, paymentID uint, req models.UpdatePaymentRequest) (*models.Payment, error) {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return nil, errors.New("payment not found")
	}
	if payment.TripID != tripID {
		return nil, errors.New("payment does not belong to this trip")
	}
	if req.Amount < 0 {
		return nil, errors.New("payment amount cannot be negative")
	}

	payment.Amount = req.Amount
	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, err
	}

	if err := s.balance.RecalculateAll(); err != nil {
		return nil, err
	}

	return s.paymentRepo.FindByID(paymentID)
}

func (s *PaymentService) Delete(tripID, paymentID uint) error {
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return errors.New("payment not found")
	}
	if payment.TripID != tripID {
		return errors.New("payment does not belong to this trip")
	}
	if err := s.paymentRepo.Delete(paymentID); err != nil {
		return err
	}
	return s.balance.RecalculateAll()
}
