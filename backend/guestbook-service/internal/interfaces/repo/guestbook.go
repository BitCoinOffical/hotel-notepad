package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GuestBookRepo struct {
	pool *pgxpool.Pool
}

func NewGuestBookRepo(pool *pgxpool.Pool) *GuestBookRepo {
	return &GuestBookRepo{pool: pool}
}

func (r *GuestBookRepo) CreateEntry(ctx context.Context, guest *models.Guest) error {
	sql := `INSERT INTO guests 
	(user_id, first_name, mid_name, last_name, note, phone, arrival_date, departure_date)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, sql, guest.UserID, guest.FirstName, guest.MidName, guest.LastName, guest.Note, guest.Phone, guest.ArrivalDate, guest.DepartureDate)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	return nil
}
func (r *GuestBookRepo) GetEntries(ctx context.Context, userID string) ([]models.Guest, error) {
	sql := `SELECT 
	id, 
	user_id, 
	first_name, 
	mid_name, 
	last_name, 
	note, 
	phone, 
	arrival_date, 
	departure_date, 
	created_at, 
	updated_at
	FROM guests WHERE user_id = $1 AND is_delete = false ORDER BY updated_at DESC`
	var guests []models.Guest
	rows, err := r.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var guest models.Guest
		if err := rows.Scan(
			&guest.GuestID,
			&guest.UserID,
			&guest.FirstName,
			&guest.MidName,
			&guest.LastName,
			&guest.Note,
			&guest.Phone,
			&guest.ArrivalDate,
			&guest.DepartureDate,
			&guest.CreatedAt,
			&guest.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		guests = append(guests, guest)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return guests, nil

}
func (r *GuestBookRepo) GetEntryByID(ctx context.Context, id, userID string) (*models.Guest, error) {
	sql := `SELECT 
	id, 
	user_id,
	first_name,
	mid_name, 
	last_name, 
	note, phone, 
	arrival_date, 
	departure_date, 
	created_at, 
	updated_at
	FROM guests WHERE id = $1 AND user_id = $2 AND is_delete = false`
	var guest models.Guest
	err := r.pool.QueryRow(ctx, sql, id, userID).Scan(
		&guest.GuestID,
		&guest.UserID,
		&guest.FirstName,
		&guest.MidName,
		&guest.LastName,
		&guest.Note,
		&guest.Phone,
		&guest.ArrivalDate,
		&guest.DepartureDate,
		&guest.CreatedAt,
		&guest.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &guest, nil
}
func (r *GuestBookRepo) UpdateEntry(ctx context.Context, id, userID string, guest *models.UpdateGuest) (*models.Guest, error) {
	builder := squirrel.Update("guests").
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.Eq{"is_delete": false}).
		PlaceholderFormat(squirrel.Dollar).
		Suffix("RETURNING id, user_id, first_name, mid_name, last_name, note, phone, arrival_date, departure_date, created_at, updated_at")
	if guest.FirstName != nil {
		builder = builder.Set("first_name", *guest.FirstName)
	}
	if guest.MidName != nil {
		builder = builder.Set("mid_name", *guest.MidName)
	}
	if guest.LastName != nil {
		builder = builder.Set("last_name", *guest.LastName)
	}
	if guest.Note != nil {
		builder = builder.Set("note", *guest.Note)
	}
	if guest.Phone != nil {
		builder = builder.Set("phone", *guest.Phone)
	}
	if guest.ArrivalDate != nil {
		builder = builder.Set("arrival_date", *guest.ArrivalDate)
	}
	if guest.DepartureDate != nil {
		builder = builder.Set("departure_date", *guest.DepartureDate)
	}
	builder = builder.Set("updated_at", time.Now())
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder.ToSql: %w", err)
	}

	var updatedGuest models.Guest
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&updatedGuest.GuestID,
		&updatedGuest.UserID,
		&updatedGuest.FirstName,
		&updatedGuest.MidName,
		&updatedGuest.LastName,
		&updatedGuest.Note,
		&updatedGuest.Phone,
		&updatedGuest.ArrivalDate,
		&updatedGuest.DepartureDate,
		&updatedGuest.CreatedAt,
		&updatedGuest.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &updatedGuest, nil

}
func (r *GuestBookRepo) DeleteEntry(ctx context.Context, id, userID string) error {
	sql := `UPDATE guests SET is_delete = true WHERE id = $1 AND user_id = $2 AND is_delete = false`
	tag, err := r.pool.Exec(ctx, sql, id, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *GuestBookRepo) SearchEntry(ctx context.Context, userID string, firstName, midName, lastName string) (*models.Guest, error) {
	builder := squirrel.Select(
		"id, user_id, first_name, mid_name, last_name, note, phone, arrival_date, departure_date, created_at, updated_at",
	).
		From("guests").
		Where(squirrel.Eq{"user_id": userID, "is_delete": false}).
		PlaceholderFormat(squirrel.Dollar)

	if strings.TrimSpace(firstName) != "" {
		builder = builder.Where(squirrel.ILike{
			"first_name": "%" + strings.TrimSpace(firstName) + "%",
		})
	}
	if strings.TrimSpace(midName) != "" {
		builder = builder.Where(squirrel.ILike{
			"mid_name": "%" + strings.TrimSpace(midName) + "%",
		})
	}
	if strings.TrimSpace(lastName) != "" {
		builder = builder.Where(squirrel.ILike{
			"last_name": "%" + strings.TrimSpace(lastName) + "%",
		})
	}

	builder = builder.OrderBy("updated_at DESC").Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("builder.ToSql: %w", err)
	}

	var guest models.Guest

	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&guest.GuestID,
		&guest.UserID,
		&guest.FirstName,
		&guest.MidName,
		&guest.LastName,
		&guest.Note,
		&guest.Phone,
		&guest.ArrivalDate,
		&guest.DepartureDate,
		&guest.CreatedAt,
		&guest.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}
	return &guest, nil
}
