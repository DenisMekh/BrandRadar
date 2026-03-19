package pgrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"prod-pobeda-2026/internal/entity"
)

// BrandRepo — PostgreSQL реализация хранилища брендов.
type BrandRepo struct {
	pool DB
	sb   squirrel.StatementBuilderType
}

// NewBrandRepo создаёт новый экземпляр репозитория брендов.
func NewBrandRepo(pool DB) *BrandRepo {
	return &BrandRepo{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *BrandRepo) Create(ctx context.Context, brand *entity.Brand) error {
	sql, args, err := r.sb.
		Insert("brands").
		Columns("id", "name", "keywords", "exclusions", "risk_words").
		Values(brand.ID, brand.Name, brand.Keywords, brand.Exclusions, brand.RiskWords).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("BrandRepo.Create: build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("BrandRepo.Create: %w", entity.ErrDuplicate)
		}
		return fmt.Errorf("BrandRepo.Create: %w", err)
	}
	return nil
}

func (r *BrandRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Brand, error) {
	sql, args, err := r.sb.
		Select("id", "name", "keywords", "exclusions", "risk_words", "created_at", "updated_at").
		From("brands").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("BrandRepo.GetByID: build query: %w", err)
	}

	var brand entity.Brand
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Keywords,
		&brand.Exclusions,
		&brand.RiskWords,
		&brand.CreatedAt,
		&brand.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("BrandRepo.GetByID: %w", entity.ErrNotFound)
		}
		return nil, fmt.Errorf("BrandRepo.GetByID: %w", err)
	}
	return &brand, nil
}

func (r *BrandRepo) List(ctx context.Context, limit, offset int) ([]entity.Brand, int64, error) {
	countSQL, countArgs, err := r.sb.
		Select("COUNT(*)").
		From("brands").
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("BrandRepo.List: build count query: %w", err)
	}

	var total int64
	if scanErr := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); scanErr != nil {
		return nil, 0, fmt.Errorf("BrandRepo.List: count: %w", scanErr)
	}

	dataBuilder := r.sb.
		Select("id", "name", "keywords", "exclusions", "risk_words", "created_at", "updated_at").
		From("brands").
		OrderBy("created_at DESC")
	if limit > 0 {
		dataBuilder = dataBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		dataBuilder = dataBuilder.Offset(uint64(offset))
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("BrandRepo.List: build data query: %w", err)
	}

	rows, err := r.pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("BrandRepo.List: %w", err)
	}
	defer rows.Close()

	brands := make([]entity.Brand, 0)
	for rows.Next() {
		var brand entity.Brand
		if scanErr := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Keywords,
			&brand.Exclusions,
			&brand.RiskWords,
			&brand.CreatedAt,
			&brand.UpdatedAt,
		); scanErr != nil {
			return nil, 0, fmt.Errorf("BrandRepo.List: scan: %w", scanErr)
		}
		brands = append(brands, brand)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("BrandRepo.List: rows: %w", rows.Err())
	}

	return brands, total, nil
}

func (r *BrandRepo) Count(ctx context.Context) (int64, error) {
	sql, args, err := r.sb.
		Select("COUNT(*)").
		From("brands").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("BrandRepo.Count: build query: %w", err)
	}

	var total int64
	if scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&total); scanErr != nil {
		return 0, fmt.Errorf("BrandRepo.Count: %w", scanErr)
	}

	return total, nil
}

func (r *BrandRepo) Update(ctx context.Context, brand *entity.Brand) error {
	sql, args, err := r.sb.
		Update("brands").
		Set("name", brand.Name).
		Set("keywords", brand.Keywords).
		Set("exclusions", brand.Exclusions).
		Set("risk_words", brand.RiskWords).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": brand.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("BrandRepo.Update: build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&brand.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("BrandRepo.Update: %w", entity.ErrNotFound)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("BrandRepo.Update: %w", entity.ErrDuplicate)
		}
		return fmt.Errorf("BrandRepo.Update: %w", err)
	}
	return nil
}

func (r *BrandRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.sb.
		Delete("brands").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("BrandRepo.Delete: build query: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("BrandRepo.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("BrandRepo.Delete: %w", entity.ErrNotFound)
	}
	return nil
}
