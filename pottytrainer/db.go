package pottytrainer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"twos.dev/mainframe/ddb"
)

func LogPoop(ctx context.Context, db *ddb.Client, userID string, poopedAt time.Time, bad bool) error {
	db.Poops.Put(&ddb.Poop{
		ID:        uuid.New().String(),
		UserID:    "test",
		PoopedAt:  poopedAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Run()
	return nil
}
