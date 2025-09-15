package models

import "time"

// Сущность Тендера
type Tender struct {
	ID             int       `db:"id" json:"id"`
	Name           string    `db:"name" json:"name" validate:"required,max=100"`
	Description    string    `db:"description" json:"description" validate:"required,max=500"`
	ServiceType    string    `db:"service_type" json:"serviceType" validate:"required,oneof=Construction Delivery Manufacture"`
	Status         string    `db:"status" json:"status" validate:"required,oneof=Created Published Closed"`
	OrganizationID int       `db:"organization_id" json:"organizationId" validate:"required"`
	Version        int       `db:"version" json:"version"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at" json:"-"`
}

// Сущность Предложения
type Bid struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name" validate:"required,max=100"`
	Description string    `db:"description" json:"description" validate:"required,max=500"`
	Status      string    `db:"status" json:"status" validate:"required,oneof=Created Published Canceled Approved Rejected"`
	TenderID    int       `db:"tender_id" json:"tenderId" validate:"required"`
	AuthorType  string    `db:"author_type" json:"authorType" validate:"required,oneof=Organization User"`
	AuthorID    int       `db:"author_id" json:"authorId" validate:"required"`
	Version     int       `db:"version" json:"version"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}

// Сущность Отзыва
type BidReview struct {
	ID          int       `db:"id" json:"id"`
	Description string    `db:"description" json:"description" validate:"required,max=1000"`
	BidID       int       `db:"bid_id" json:"bidId"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

// Сущность Пользователя (из БД, для связи)
type Employee struct {
	ID        int       `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	FirstName string    `db:"first_name" json:"firstName"`
	LastName  string    `db:"last_name" json:"lastName"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"-"`
}

// Сущность Организации (из БД, для связи)
type Organization struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Type        string    `db:"type" json:"type"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"-"`
}
