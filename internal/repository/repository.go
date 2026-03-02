package repository

import (
	"tripshare/internal/models"

	"gorm.io/gorm"
)

// --- Member Repository ---

type MemberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(member *models.Member) error {
	return r.db.Create(member).Error
}

func (r *MemberRepository) FindAll() ([]models.Member, error) {
	var members []models.Member
	err := r.db.Order("name ASC").Find(&members).Error
	return members, err
}

func (r *MemberRepository) FindByID(id uint) (*models.Member, error) {
	var member models.Member
	err := r.db.First(&member, id).Error
	return &member, err
}

func (r *MemberRepository) Update(member *models.Member) error {
	return r.db.Model(member).Updates(map[string]interface{}{
		"name": member.Name,
	}).Error
}

func (r *MemberRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM trip_members WHERE member_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Where("member_id = ?", id).Delete(&models.Payment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Member{}, id).Error
	})
}

func (r *MemberRepository) FindByIDs(ids []uint) ([]models.Member, error) {
	var members []models.Member
	err := r.db.Where("id IN ?", ids).Find(&members).Error
	return members, err
}

func (r *MemberRepository) UpdateBalance(id uint, balance float64) error {
	return r.db.Model(&models.Member{}).Where("id = ?", id).Update("balance", balance).Error
}

func (r *MemberRepository) ResetAllBalances() error {
	return r.db.Model(&models.Member{}).Where("1 = 1").Update("balance", 0).Error
}

// --- Trip Repository ---

type TripRepository struct {
	db *gorm.DB
}

func NewTripRepository(db *gorm.DB) *TripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) Create(trip *models.Trip) error {
	return r.db.Create(trip).Error
}

func (r *TripRepository) FindAll() ([]models.Trip, error) {
	var trips []models.Trip
	err := r.db.Preload("Members").Preload("Payments").Preload("Payments.Member").
		Order("date DESC").Find(&trips).Error
	return trips, err
}

func (r *TripRepository) FindByID(id uint) (*models.Trip, error) {
	var trip models.Trip
	err := r.db.Preload("Members").Preload("Payments").Preload("Payments.Member").
		First(&trip, id).Error
	return &trip, err
}

func (r *TripRepository) Update(trip *models.Trip) error {
	return r.db.Model(trip).Select("name", "total_cost", "date").Updates(trip).Error
}

func (r *TripRepository) ReplaceMembers(trip *models.Trip, members []models.Member) error {
	return r.db.Model(trip).Association("Members").Replace(members)
}

func (r *TripRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM trip_members WHERE trip_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Where("trip_id = ?", id).Delete(&models.Payment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Trip{}, id).Error
	})
}

// --- Payment Repository ---

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *PaymentRepository) FindByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("Member").First(&payment, id).Error
	return &payment, err
}

func (r *PaymentRepository) FindByTripID(tripID uint) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Preload("Member").Where("trip_id = ?", tripID).Find(&payments).Error
	return payments, err
}

func (r *PaymentRepository) Update(payment *models.Payment) error {
	return r.db.Model(payment).Update("amount", payment.Amount).Error
}

func (r *PaymentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Payment{}, id).Error
}
