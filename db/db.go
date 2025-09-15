package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

// Employee (Пользователь)
type Employee struct {
	ID        int       `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	FirstName string    `db:"first_name" json:"firstName"`
	LastName  string    `db:"last_name" json:"lastName"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

func (s *Storage) CreateEmployee(ctx context.Context, e *Employee) error {
	query := `
        INSERT INTO employee (username, first_name, last_name)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(ctx, query, e.Username, e.FirstName, e.LastName).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (s *Storage) GetEmployeeByUsername(ctx context.Context, username string) (*Employee, error) {
	e := &Employee{}
	query := `SELECT * FROM employee WHERE username=$1`
	err := s.db.GetContext(ctx, e, query, username)
	return e, err
}

func (s *Storage) UpdateEmployee(ctx context.Context, e *Employee) error {
	query := `
        UPDATE employee
        SET first_name = $1, last_name = $2, updated_at = NOW()
        WHERE username = $3`
	_, err := s.db.ExecContext(ctx, query, e.FirstName, e.LastName, e.Username)
	return err
}

func (s *Storage) DeleteEmployee(ctx context.Context, username string) error {
	query := `DELETE FROM employee WHERE username = $1`
	_, err := s.db.ExecContext(ctx, query, username)
	return err
}

// Organization (Организация)
type Organization struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Type        string    `db:"type" json:"type"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

func (s *Storage) CreateOrganization(ctx context.Context, o *Organization) error {
	query := `
        INSERT INTO organization (name, description, type)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`
	return s.db.QueryRowContext(ctx, query, o.Name, o.Description, o.Type).
		Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (s *Storage) GetOrganization(ctx context.Context, id int) (*Organization, error) {
	o := &Organization{}
	query := `SELECT * FROM organization WHERE id=$1`
	err := s.db.GetContext(ctx, o, query, id)
	return o, err
}

func (s *Storage) UpdateOrganization(ctx context.Context, o *Organization) error {
	query := `
        UPDATE organization
        SET name=$1, description=$2, type=$3, updated_at=NOW()
        WHERE id=$4`
	_, err := s.db.ExecContext(ctx, query, o.Name, o.Description, o.Type, o.ID)
	return err
}

func (s *Storage) DeleteOrganization(ctx context.Context, id int) error {
	query := `DELETE FROM organization WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *Storage) IsUserResponsibleForOrganization(ctx context.Context, userID int, orgID int) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM organization_responsible WHERE user_id=$1 AND organization_id=$2`
	err := s.db.GetContext(ctx, &count, query, userID, orgID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Tender (Тендер)
type Tender struct {
	ID             int       `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Description    string    `db:"description" json:"description"`
	ServiceType    string    `db:"service_type" json:"serviceType"`
	Status         string    `db:"status" json:"status"`
	OrganizationID int       `db:"organization_id" json:"organizationId"`
	Version        int       `db:"version" json:"version"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
}

func (s *Storage) CreateTender(ctx context.Context, t *Tender) error {
	query := `
        INSERT INTO tender
            (name, description, service_type, status, organization_id, version)
        VALUES
            ($1, $2, $3, $4, $5, 1)
        RETURNING id, created_at`
	err := s.db.QueryRowContext(ctx, query,
		t.Name, t.Description, t.ServiceType, t.Status, t.OrganizationID).
		Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return err
	}
	// Сохраняем первую версию
	return s.SaveTenderVersion(ctx, t)
}

func (s *Storage) GetTender(ctx context.Context, id int) (*Tender, error) {
	t := &Tender{}
	query := `SELECT * FROM tender WHERE id=$1`
	err := s.db.GetContext(ctx, t, query, id)
	return t, err
}

func (s *Storage) UpdateTender(ctx context.Context, t *Tender) error {
	t.Version++
	query := `
        UPDATE tender
        SET name=$1, description=$2, service_type=$3, status=$4, version=$5
        WHERE id=$6`
	_, err := s.db.ExecContext(ctx, query,
		t.Name, t.Description, t.ServiceType, t.Status, t.Version, t.ID)
	if err != nil {
		return err
	}
	// Сохраняем новую версию
	return s.SaveTenderVersion(ctx, t)
}

func (s *Storage) DeleteTender(ctx context.Context, id int) error {
	query := `DELETE FROM tender WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *Storage) GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]Tender, error) {
	baseQuery := "SELECT id, name, description, service_type, status, organization_id, version, created_at FROM tender"
	var args []interface{}
	filter := ""

	if len(serviceTypes) > 0 {
		placeholders := make([]string, len(serviceTypes))
		for i := range serviceTypes {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		filter = fmt.Sprintf(" WHERE service_type IN (%s)", strings.Join(placeholders, ", "))
		for _, v := range serviceTypes {
			args = append(args, v)
		}
	}

	query := baseQuery + filter + " ORDER BY name ASC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	tenders := []Tender{}
	err := s.db.SelectContext(ctx, &tenders, query, args...)
	if err != nil {
		return nil, err
	}
	return tenders, nil
}

func (s *Storage) GetUserTenders(ctx context.Context, username string, limit, offset int) ([]Tender, error) {
	query := `
        SELECT t.id, t.name, t.description, t.service_type, t.status, t.organization_id, t.version, t.created_at
        FROM tender t
        JOIN organization_responsible orr ON t.organization_id = orr.organization_id
        JOIN employee e ON orr.user_id = e.id
        WHERE e.username = $1
        ORDER BY t.name ASC
        LIMIT $2 OFFSET $3
    `
	tenders := []Tender{}
	err := s.db.SelectContext(ctx, &tenders, query, username, limit, offset)
	if err != nil {
		return nil, err
	}
	return tenders, nil
}

func (s *Storage) SaveTenderVersion(ctx context.Context, t *Tender) error {
	query := `
        INSERT INTO tender_versions
            (tender_id, name, description, service_type, status, organization_id, version, created_at)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, NOW())
    `
	_, err := s.db.ExecContext(ctx, query,
		t.ID, t.Name, t.Description, t.ServiceType, t.Status, t.OrganizationID, t.Version)
	return err
}

func (s *Storage) GetTenderVersion(ctx context.Context, tenderID, version int) (*Tender, error) {
	var t Tender
	query := `
        SELECT tender_id AS id, name, description, service_type, status, organization_id, version, created_at
        FROM tender_versions
        WHERE tender_id = $1 AND version = $2
    `
	err := s.db.GetContext(ctx, &t, query, tenderID, version)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Bid (Предложение)

type Bid struct {
	ID              int       `db:"id" json:"id"`
	Name            string    `db:"name" json:"name" validate:"required,max=100"`
	Description     string    `db:"description" json:"description" validate:"required,max=500"`
	Status          string    `db:"status" json:"status" validate:"required,oneof=Created Published Canceled Approved Rejected"`
	TenderID        int       `db:"tender_id" json:"tenderId" validate:"required"`
	OrganizationID  int       `db:"organization_id" json:"organizationId"`   // используйте это поле вместо AuthorID
	CreatorUsername string    `db:"creator_username" json:"creatorUsername"` // вместо AuthorType
	Version         int       `db:"version" json:"version"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at" json:"-"`
}

func (s *Storage) CreateBid(ctx context.Context, b *Bid) error {
	query := `
        INSERT INTO bid
            (name, description, status, tender_id, organization_id, creator_username, version)
        VALUES
            ($1, $2, $3, $4, $5, $6, 1)
        RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query,
		b.Name, b.Description, b.Status, b.TenderID, b.OrganizationID, b.CreatorUsername).
		Scan(&b.ID, &b.CreatedAt)
}

func (s *Storage) GetBid(ctx context.Context, id int) (*Bid, error) {
	b := &Bid{}
	query := `SELECT * FROM bid WHERE id=$1`
	err := s.db.GetContext(ctx, b, query, id)
	return b, err
}

func (s *Storage) UpdateBid(ctx context.Context, b *Bid) error {
	b.Version++
	query := `
        UPDATE bid
        SET name=$1, description=$2, status=$3, version=$4
        WHERE id=$5`
	_, err := s.db.ExecContext(ctx, query, b.Name, b.Description, b.Status, b.Version, b.ID)
	return err
}

func (s *Storage) DeleteBid(ctx context.Context, id int) error {
	query := `DELETE FROM bid WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// BidReview (Отзыв)
type BidReview struct {
	ID          int       `db:"id" json:"id"`
	BidID       int       `db:"bid_id" json:"bidId"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

func (s *Storage) CreateBidReview(ctx context.Context, r *BidReview) error {
	query := `
        INSERT INTO bid_review (bid_id, description)
        VALUES ($1, $2)
        RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, r.BidID, r.Description).Scan(&r.ID, &r.CreatedAt)
}

func (s *Storage) GetBidReviewsByBidID(ctx context.Context, bidID int) ([]BidReview, error) {
	var reviews []BidReview
	query := `SELECT * FROM bid_review WHERE bid_id=$1`
	err := s.db.SelectContext(ctx, &reviews, query, bidID)
	return reviews, err
}

func (s *Storage) DeleteBidReview(ctx context.Context, id int) error {
	query := `DELETE FROM bid_review WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
func (s *Storage) GetUserBids(ctx context.Context, username string, limit, offset int) ([]Bid, error) {
	query := `
        SELECT * FROM bid
        WHERE creator_username = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3`
	bids := []Bid{}
	err := s.db.SelectContext(ctx, &bids, query, username, limit, offset)
	return bids, err
}

func (s *Storage) GetBidsForTender(ctx context.Context, tenderID int, username string, limit, offset int) ([]Bid, error) {
	query := `
        SELECT b.* FROM bid b
        JOIN employee e ON b.creator_username = e.username
        WHERE b.tender_id = $1
        AND (e.username = $2 OR (SELECT COUNT(1) FROM organization_responsible WHERE organization_id = b.organization_id AND user_id = e.id) > 0)
        ORDER BY b.created_at DESC
        LIMIT $3 OFFSET $4
    `
	bids := []Bid{}
	err := s.db.SelectContext(ctx, &bids, query, tenderID, username, limit, offset)
	return bids, err
}

func (s *Storage) GetBidVersion(ctx context.Context, bidID, version int) (*Bid, error) {
	var b Bid
	query := `
        SELECT bid_id AS id, name, description, status, tender_id, organization_id, creator_username, version, created_at
        FROM bid_versions
        WHERE bid_id = $1 AND version = $2
    `
	err := s.db.GetContext(ctx, &b, query, bidID, version)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Storage) SaveBidVersion(ctx context.Context, b *Bid) error {
	query := `
        INSERT INTO bid_versions
            (bid_id, name, description, status, tender_id, organization_id, creator_username, version, created_at)
        VALUES
            ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
    `
	_, err := s.db.ExecContext(ctx, query,
		b.ID, b.Name, b.Description, b.Status, b.TenderID, b.OrganizationID, b.CreatorUsername, b.Version)
	return err
}

func (s *Storage) GetBidReviewsByAuthorForTender(ctx context.Context, authorUsername string, tenderID int) ([]BidReview, error) {
	var reviews []BidReview
	query := `
        SELECT r.*
        FROM bid_review r
        JOIN bid b ON r.bid_id = b.id
        WHERE b.creator_username = $1 AND b.tender_id = $2
        ORDER BY r.created_at DESC
    `
	err := s.db.SelectContext(ctx, &reviews, query, authorUsername, tenderID)
	return reviews, err
}

func (s *Storage) AddBidDecision(ctx context.Context, bidID, userID int, decision string) error {
	query := `
        INSERT INTO bid_decision (bid_id, user_id, decision, created_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (bid_id, user_id) DO UPDATE SET decision = EXCLUDED.decision, created_at = NOW()
    `
	_, err := s.db.ExecContext(ctx, query, bidID, userID, decision)
	return err
}
func (s *Storage) GetBidDecisionsCount(ctx context.Context, bidID int) (accepts int, rejects int, err error) {
	query := `
        SELECT 
            COUNT(CASE WHEN decision = 'Approved' THEN 1 END),
            COUNT(CASE WHEN decision = 'Rejected' THEN 1 END)
        FROM bid_decision
        WHERE bid_id = $1
    `
	err = s.db.QueryRowContext(ctx, query, bidID).Scan(&accepts, &rejects)
	return
}

func (s *Storage) GetResponsibleCount(ctx context.Context, organizationID int) (int, error) {
	var count int
	query := `
        SELECT COUNT(1) FROM organization_responsible WHERE organization_id = $1
    `
	err := s.db.GetContext(ctx, &count, query, organizationID)
	return count, err
}
