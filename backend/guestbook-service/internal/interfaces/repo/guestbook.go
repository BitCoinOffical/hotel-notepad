package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GuestBookRepo struct {
	pool *pgxpool.Pool
}

func NewGuestBookRepo(pool *pgxpool.Pool) *GuestBookRepo {
	return &GuestBookRepo{pool: pool}
}

func (r *GuestBookRepo) CreateEntry(ctx context.Context, guest *models.Guest) error {
	sql := `INSER INTO guests 
	(user_id, first_name, mid_name, last_name, note, phone, current_date, departure_date)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, sql, guest.UserID, guest.FirstName, guest.MidName, guest.LastName, guest.Note, guest.Phone, guest.CurrentDate, guest.DepartureDate)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
func (r *GuestBookRepo) GetEntries(ctx context.Context) {

}
func (r *GuestBookRepo) GetEntryByID(ctx context.Context) {

}
func (r *GuestBookRepo) UpdateEntry(ctx context.Context, id, userID string, guest *models.UpdateGuest) error {
	builder := squirrel.Update("guests").
		Where(squirrel.Eq{"id": id}, squirrel.Eq{"user_id": userID}).PlaceholderFormat(squirrel.Dollar)
	if guest.FirstName != nil {
		builder.Set("first_name", guest.FirstName)
	}
	if guest.MidName != nil {
		builder.Set("mid_name", guest.MidName)
	}
	if guest.LastName != nil {
		builder.Set("last_name", guest.LastName)
	}
	if guest.Note != nil {
		builder.Set("note", guest.Note)
	}
	if guest.Phone != nil {
		builder.Set("phone", guest.Phone)
	}
	if guest.CurrentDate != nil {
		builder.Set("current_date", guest.CurrentDate)
	}
	if guest.DepartureDate != nil {
		builder.Set("departure_date", guest.DepartureDate)
	}
	builder.Set("updated_at", time.Now())
	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("builder.ToSql: %w", err)
	}
	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}

	return nil

}
func (r *GuestBookRepo) DeleteEntry(ctx context.Context, id, userID string) error {
	sql := `UPDATE guests SET is_delete = true WHERE id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, sql, id, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
