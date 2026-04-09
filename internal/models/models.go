package models

import "time"

type Member struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex"`
	Balance   float64   `json:"balance" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Trip struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	TotalCost float64   `json:"total_cost" gorm:"not null"`
	Date      time.Time `json:"date"`
	Members   []Member  `json:"members" gorm:"many2many:trip_members;"`
	Payments  []Payment `json:"payments,omitempty" gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Payment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TripID    uint      `json:"trip_id" gorm:"not null;index"`
	MemberID  uint      `json:"member_id" gorm:"not null;index"`
	Amount    float64   `json:"amount" gorm:"not null"`
	Member    Member    `json:"member" gorm:"foreignKey:MemberID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- Request types ---

type CreateMemberRequest struct {
	Name string `json:"name"`
}

type UpdateMemberRequest struct {
	Name string `json:"name"`
}

type CreateTripRequest struct {
	Name      string  `json:"name"`
	TotalCost float64 `json:"total_cost"`
	Date      string  `json:"date"`
	MemberIDs []uint  `json:"member_ids"`
}

type UpdateTripRequest struct {
	Name      string  `json:"name"`
	TotalCost float64 `json:"total_cost"`
	Date      string  `json:"date"`
	MemberIDs []uint  `json:"member_ids"`
}

type CreatePaymentRequest struct {
	MemberID uint    `json:"member_id"`
	Amount   float64 `json:"amount"`
}

type UpdatePaymentRequest struct {
	Amount float64 `json:"amount"`
}

// --- Response types ---

type MemberBalance struct {
	Member   Member  `json:"member"`
	Balance  float64 `json:"balance"`
	TripID   uint    `json:"trip_id"`
	TripName string  `json:"trip_name,omitempty"`
}

type Settlement struct {
	FromName  string   `json:"from_name"`
	ToName    string   `json:"to_name"`
	Amount    float64  `json:"amount"`
	TripNames []string `json:"trip_names"`
}

type BalanceResponse struct {
	Balances        []MemberBalance          `json:"balances"`
	Settlements     []Settlement             `json:"settlements"`
	TripAdjustments map[uint]map[uint]float64 `json:"trip_adjustments,omitempty"`
}

type TripDetail struct {
	Trip       Trip            `json:"trip"`
	EqualShare float64         `json:"equal_share"`
	Breakdowns []TripBreakdown `json:"breakdowns"`
}

type TripBreakdown struct {
	MemberID   uint    `json:"member_id"`
	MemberName string  `json:"member_name"`
	Paid       float64 `json:"paid"`
	EqualShare float64 `json:"equal_share"`
	Delta      float64 `json:"delta"`
}
