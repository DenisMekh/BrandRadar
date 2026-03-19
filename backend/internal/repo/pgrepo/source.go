package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"prod-pobeda-2026/internal/entity"
)

// SourceRepo — PostgreSQL реализация хранилища источников.
type SourceRepo struct {
	pool DB
	sb   squirrel.StatementBuilderType
}

// NewSourceRepo создаёт новый экземпляр репозитория источников.
func NewSourceRepo(pool DB) *SourceRepo {
	return &SourceRepo{
		pool: pool,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *SourceRepo) Create(ctx context.Context, source *entity.Source) error {
	sql, args, err := r.sb.
		Insert("sources").
		Columns("id", "type", "name", "url", "status").
		Values(source.ID, source.Type, source.Name, source.URL, source.Status).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("SourceRepo.Create: build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&source.CreatedAt, &source.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("SourceRepo.Create: %w", entity.ErrDuplicate)
		}
		return fmt.Errorf("SourceRepo.Create: %w", err)
	}
	return nil
}

func (r *SourceRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Source, error) {
	sql, args, err := r.sb.
		Select("id", "type", "name", "url", "status", "created_at", "updated_at").
		From("sources").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("SourceRepo.GetByID: build query: %w", err)
	}

	source, err := scanSource(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("SourceRepo.GetByID: %w", entity.ErrNotFound)
		}
		return nil, fmt.Errorf("SourceRepo.GetByID: %w", err)
	}
	return source, nil
}

func (r *SourceRepo) List(ctx context.Context, limit, offset int) ([]entity.Source, int64, error) {
	countSQL, countArgs, err := r.sb.
		Select("COUNT(*)").
		From("sources").
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("SourceRepo.List: build count query: %w", err)
	}

	var total int64
	if scanErr := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); scanErr != nil {
		return nil, 0, fmt.Errorf("SourceRepo.List: count: %w", scanErr)
	}

	dataBuilder := r.sb.
		Select("id", "type", "name", "url", "status", "created_at", "updated_at").
		From("sources").
		OrderBy("created_at DESC")
	if limit > 0 {
		dataBuilder = dataBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		dataBuilder = dataBuilder.Offset(uint64(offset))
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("SourceRepo.List: build data query: %w", err)
	}

	rows, err := r.pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("SourceRepo.List: %w", err)
	}
	defer rows.Close()

	sources := make([]entity.Source, 0)
	for rows.Next() {
		source, scanErr := scanSource(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("SourceRepo.List: scan: %w", scanErr)
		}
		sources = append(sources, *source)
	}
	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("SourceRepo.List: rows: %w", rows.Err())
	}

	return sources, total, nil
}

func (r *SourceRepo) CountActiveByType(ctx context.Context) (map[string]int64, error) {
	sql, args, err := r.sb.
		Select("type", "COUNT(*)").
		From("sources").
		Where(squirrel.Eq{"status": entity.SourceStatusActive}).
		GroupBy("type").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("SourceRepo.CountActiveByType: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("SourceRepo.CountActiveByType: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var sourceType string
		var count int64
		if scanErr := rows.Scan(&sourceType, &count); scanErr != nil {
			return nil, fmt.Errorf("SourceRepo.CountActiveByType: scan: %w", scanErr)
		}
		counts[strings.TrimSpace(sourceType)] = count
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("SourceRepo.CountActiveByType: rows: %w", rows.Err())
	}

	return counts, nil
}

func (r *SourceRepo) ListActive(ctx context.Context) ([]entity.Source, error) {
	sql, args, err := r.sb.
		Select("id", "type", "name", "url", "status", "created_at", "updated_at").
		From("sources").
		Where(squirrel.Eq{"status": entity.SourceStatusActive}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("SourceRepo.ListActive: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("SourceRepo.ListActive: %w", err)
	}
	defer rows.Close()

	sources := make([]entity.Source, 0)
	for rows.Next() {
		source, scanErr := scanSource(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("SourceRepo.ListActive: scan: %w", scanErr)
		}
		sources = append(sources, *source)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("SourceRepo.ListActive: rows: %w", rows.Err())
	}

	return sources, nil
}

func (r *SourceRepo) Update(ctx context.Context, source *entity.Source) error {
	sql, args, err := r.sb.
		Update("sources").
		Set("type", source.Type).
		Set("name", source.Name).
		Set("url", source.URL).
		Set("status", source.Status).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": source.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("SourceRepo.Update: build query: %w", err)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&source.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("SourceRepo.Update: %w", entity.ErrNotFound)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("SourceRepo.Update: %w", entity.ErrDuplicate)
		}
		return fmt.Errorf("SourceRepo.Update: %w", err)
	}
	return nil
}

func (r *SourceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.sb.
		Delete("sources").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("SourceRepo.Delete: build query: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SourceRepo.Delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("SourceRepo.Delete: %w", entity.ErrNotFound)
	}
	return nil
}

type sourceScanner interface {
	Scan(dest ...any) error
}

func scanSource(row sourceScanner) (*entity.Source, error) {
	var source entity.Source

	err := row.Scan(
		&source.ID,
		&source.Type,
		&source.Name,
		&source.URL,
		&source.Status,
		&source.CreatedAt,
		&source.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &source, nil
}
