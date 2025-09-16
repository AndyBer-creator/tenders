package handlers

import (
	"context"
	"tenders/db"
)

type StorageInterface interface {
	GetEmployeeByUsername(ctx context.Context, username string) (*db.Employee, error)
	IsUserResponsibleForOrganization(ctx context.Context, userID, organizationID int) (bool, error)

	CreateTender(ctx context.Context, tender *db.Tender) error
	GetTender(ctx context.Context, tenderID int) (*db.Tender, error)
	UpdateTender(ctx context.Context, tender *db.Tender) error
	SaveTenderVersion(ctx context.Context, tender *db.Tender) error
	GetTenderVersion(ctx context.Context, tenderID int, version int) (*db.Tender, error)
	GetTenders(ctx context.Context, serviceTypes []string, limit, offset int) ([]db.Tender, error)
	GetUserTenders(ctx context.Context, username string, limit, offset int) ([]db.Tender, error)

	CreateBid(ctx context.Context, bid *db.Bid) error
	GetBid(ctx context.Context, bidID int) (*db.Bid, error)
	UpdateBid(ctx context.Context, bid *db.Bid) error
	GetUserBids(ctx context.Context, username string, limit, offset int) ([]db.Bid, error)
	GetBidsForTender(ctx context.Context, tenderID int, username string, limit, offset int) ([]db.Bid, error)
	GetBidVersion(ctx context.Context, bidID, version int) (*db.Bid, error)
	SaveBidVersion(ctx context.Context, bid *db.Bid) error

	AddBidDecision(ctx context.Context, bidID, employeeID int, decision string) error
	GetBidDecisionsCount(ctx context.Context, bidID int) (accepts int, rejects int, err error)
	GetResponsibleCount(ctx context.Context, organizationID int) (int, error)

	GetBidReviewsByAuthorForTender(ctx context.Context, authorUsername string, tenderID int) ([]db.BidReview, error)
	CreateBidReview(ctx context.Context, review *db.BidReview) error
}
