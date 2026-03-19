package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/entity"
)

// MentionRepo — PostgreSQL реализация хранилища упоминаний.

// Данные собираются из таблиц crawler_items + sentiment_results + sources.

type MentionRepo struct {
	pool DB

	sb squirrel.StatementBuilderType
}

// NewMentionRepo создаёт новый экземпляр репозитория упоминаний.

func NewMentionRepo(pool DB) *MentionRepo {
	return &MentionRepo{
		pool: pool,

		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create сохраняет упоминание: вставляет запись в crawler_items и sentiment_results.

func (r *MentionRepo) Create(ctx context.Context, mention *entity.Mention) error {
	itemID := uuid.New()

	ciSQL, ciArgs, err := r.sb.
		Insert("crawler_items").
		Columns("id", "text", "link", "source_id", "published_at", "created_at").
		Values(itemID, mention.Text, mention.URL, mention.SourceID, mention.PublishedAt, mention.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("MentionRepo.Create: build crawler_items query: %w", err)
	}

	if _, err = r.pool.Exec(ctx, ciSQL, ciArgs...); err != nil {

		if isPgFKViolation(err) {
			return fmt.Errorf("MentionRepo.Create: source_id not found: %w", entity.ErrValidation)
		}

		return fmt.Errorf("MentionRepo.Create: insert crawler_item: %w", err)

	}

	srSQL, srArgs, err := r.sb.
		Insert("sentiment_results").
		Columns("id", "item_id", "brand_id", "sentiment", "confidence", "created_at").
		Values(mention.ID, itemID, mention.BrandID, string(mention.Sentiment), mention.ML.Score, mention.CreatedAt).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("MentionRepo.Create: build sentiment_results query: %w", err)
	}

	if err = r.pool.QueryRow(ctx, srSQL, srArgs...).Scan(&mention.CreatedAt); err != nil {

		if isPgFKViolation(err) {
			return fmt.Errorf("MentionRepo.Create: brand_id not found: %w", entity.ErrValidation)
		}

		return fmt.Errorf("MentionRepo.Create: insert sentiment_result: %w", err)

	}

	return nil
}

// GetByID возвращает упоминание по ID sentiment_results.

func (r *MentionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Mention, error) {
	sql, args, err := r.sb.
		Select(
			"sr.id", "sr.brand_id", "ci.source_id",
			"COALESCE(s.name,'')", "COALESCE(s.type,'')",
			"ci.text", "ci.link",
			"sr.sentiment", "ci.published_at", "sr.created_at",
			"sr.cluster_label",
		).
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		LeftJoin("sources s ON s.id = ci.source_id").
		Where(squirrel.Eq{"sr.id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("MentionRepo.GetByID: build query: %w", err)
	}

	mention, err := scanMention(r.pool.QueryRow(ctx, sql, args...))
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("MentionRepo.GetByID: %w", entity.ErrNotFound)
		}

		return nil, fmt.Errorf("MentionRepo.GetByID: %w", err)

	}

	return mention, nil
}

// List возвращает упоминания с фильтрацией и пагинацией.

func (r *MentionRepo) List(ctx context.Context, filter entity.MentionFilter) ([]entity.Mention, int64, error) {
	countBuilder := r.sb.Select("COUNT(*)").
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id")

	countBuilder, err := applyMentionFilters(countBuilder, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("MentionRepo.List: apply filters for count: %w", err)
	}

	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("MentionRepo.List: build count query: %w", err)
	}

	var total int64

	if scanErr := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); scanErr != nil {
		return nil, 0, fmt.Errorf("MentionRepo.List: count: %w", scanErr)
	}

	dataBuilder := r.sb.
		Select(
			"sr.id", "sr.brand_id", "ci.source_id",
			"COALESCE(s.name,'')", "COALESCE(s.type,'')",
			"ci.text", "ci.link",
			"sr.sentiment", "ci.published_at", "sr.created_at",
			"sr.cluster_label",
		).
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		LeftJoin("sources s ON s.id = ci.source_id").
		OrderBy("ci.published_at DESC")

	dataBuilder, err = applyMentionFilters(dataBuilder, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("MentionRepo.List: apply filters for data: %w", err)
	}

	if filter.Limit > 0 {
		dataBuilder = dataBuilder.Limit(uint64(filter.Limit))
	}

	if filter.Offset > 0 {
		dataBuilder = dataBuilder.Offset(uint64(filter.Offset))
	}

	dataSQL, dataArgs, err := dataBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("MentionRepo.List: build data query: %w", err)
	}

	rows, err := r.pool.Query(ctx, dataSQL, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("MentionRepo.List: query: %w", err)
	}

	defer rows.Close()

	mentions := make([]entity.Mention, 0)

	for rows.Next() {

		mention, scanErr := scanMention(rows)

		if scanErr != nil {
			return nil, 0, fmt.Errorf("MentionRepo.List: scan: %w", scanErr)
		}

		mentions = append(mentions, *mention)

	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("MentionRepo.List: rows: %w", rows.Err())
	}

	return mentions, total, nil
}

// CountByBrandSince считает упоминания бренда с момента since (для spike detection).

func (r *MentionRepo) CountByBrandSince(ctx context.Context, brandID uuid.UUID, since time.Time, sentimentFilter string) (int64, error) {
	builder := r.sb.
		Select("COUNT(*)").
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		Where(squirrel.Eq{"sr.brand_id": brandID}).
		Where(squirrel.GtOrEq{"ci.published_at": since})

	if sentimentFilter != "" {
		builder = builder.Where(squirrel.Eq{"sr.sentiment": sentimentFilter})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("MentionRepo.CountByBrandSince: build query: %w", err)
	}

	var total int64

	if scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&total); scanErr != nil {
		return 0, fmt.Errorf("MentionRepo.CountByBrandSince: %w", scanErr)
	}

	return total, nil
}

func (r *MentionRepo) CountByBrandInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, sentimentFilter string) (int64, error) {
	builder := r.sb.
		Select("COUNT(*)").
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		Where(squirrel.Eq{"sr.brand_id": brandID}).
		Where(squirrel.And{
			squirrel.GtOrEq{"ci.published_at": from},
			squirrel.Lt{"ci.published_at": to},
		})

	if sentimentFilter != "" {
		builder = builder.Where(squirrel.Eq{"sr.sentiment": sentimentFilter})
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("MentionRepo.CountByBrandInRange: build query: %w", err)
	}

	var total int64

	if scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&total); scanErr != nil {
		return 0, fmt.Errorf("MentionRepo.CountByBrandInRange: %w", scanErr)
	}

	return total, nil
}

func (r *MentionRepo) GetAverageSentimentInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time) (float64, error) {
	sql, args, err := r.sb.
		Select(`COALESCE(AVG(CASE WHEN sr.sentiment = 'positive' THEN sr.confidence * 1.0 WHEN sr.sentiment = 'negative' THEN sr.confidence * (-1.0) ELSE 0.0 END), 0)`).
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		Where(squirrel.Eq{"sr.brand_id": brandID}).
		Where(squirrel.And{
			squirrel.GtOrEq{"ci.published_at": from},
			squirrel.Lt{"ci.published_at": to},
		}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("MentionRepo.GetAverageSentimentInRange: build query: %w", err)
	}

	var avgSentiment float64

	if scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&avgSentiment); scanErr != nil {
		return 0, fmt.Errorf("MentionRepo.GetAverageSentimentInRange: %w", scanErr)
	}

	return avgSentiment, nil
}

func (r *MentionRepo) GetWeightedSentimentPercentileInRange(ctx context.Context, brandID uuid.UUID, from, to time.Time, percentile float64) (float64, error) {
	// Sentiment кодируется: positive=1, neutral=0, negative=-1

	// Умножаем на confidence (ml_score) для каждого упоминания

	// Затем вычисляем перцентиль среди всех упоминаний в диапазоне

	// Получаем все взвешенные сентименты за период

	sql, args, err := r.sb.
		Select(`CASE WHEN sr.sentiment = 'positive' THEN sr.confidence * 1.0 WHEN sr.sentiment = 'negative' THEN sr.confidence * (-1.0) ELSE 0.0 END AS weighted_sentiment`).
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		Where(squirrel.Eq{"sr.brand_id": brandID}).
		Where(squirrel.And{
			squirrel.GtOrEq{"ci.published_at": from},
			squirrel.Lt{"ci.published_at": to},
		}).
		OrderBy("weighted_sentiment ASC").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("MentionRepo.GetWeightedSentimentPercentileInRange: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("MentionRepo.GetWeightedSentimentPercentileInRange: query: %w", err)
	}

	defer rows.Close()

	// Собираем все значения

	var values []float64

	for rows.Next() {

		var val float64

		if scanErr := rows.Scan(&val); scanErr != nil {
			return 0, fmt.Errorf("MentionRepo.GetWeightedSentimentPercentileInRange: scan: %w", scanErr)
		}

		values = append(values, val)

	}

	if rows.Err() != nil {
		return 0, fmt.Errorf("MentionRepo.GetWeightedSentimentPercentileInRange: rows: %w", rows.Err())
	}

	// Дебаг: выводим количество наблюдений и границы окна

	logrus.Debugf("Percentile calculation: window=[%s, %s], observations=%d, percentile=%.2f",

		from.Format(time.RFC3339), to.Format(time.RFC3339), len(values), percentile)

	// Если нет данных, возвращаем 0

	if len(values) == 0 {
		return 0, nil
	}

	// Вычисляем перцентиль

	// percentile - это значение от 0 до 100

	// индекс = (percentile / 100) * (len - 1)

	rank := (percentile / 100.0) * float64(len(values)-1)

	lowerIdx := int(rank)

	upperIdx := lowerIdx + 1

	if upperIdx >= len(values) {
		return values[len(values)-1], nil
	}

	// Интерполяция между значениями

	fraction := rank - float64(lowerIdx)

	return values[lowerIdx] + fraction*(values[upperIdx]-values[lowerIdx]), nil
}

// UpdateML обновляет ML-поля упоминания.

func (r *MentionRepo) UpdateML(ctx context.Context, id uuid.UUID, ml entity.MentionML) error {
	sql, args, err := r.sb.
		Update("mentions").
		Set("ml_label", ml.Label).
		Set("ml_score", ml.Score).
		Set("ml_is_relevant", ml.IsRelevant).
		Set("ml_similar_ids", normalizeSimilarIDs(ml.SimilarIDs)).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("MentionRepo.UpdateML: build query: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("MentionRepo.UpdateML: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("MentionRepo.UpdateML: %w", entity.ErrNotFound)
	}

	return nil
}

// UpdateStatus обновляет статус упоминания.

func (r *MentionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.MentionStatus) error {
	sql, args, err := r.sb.
		Update("mentions").
		Set("status", string(status)).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("MentionRepo.UpdateStatus: build query: %w", err)
	}

	ct, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("MentionRepo.UpdateStatus: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("MentionRepo.UpdateStatus: %w", entity.ErrNotFound)
	}

	return nil
}

// normalizeSimilarIDs возвращает nil вместо пустого массива для корректной работы с PostgreSQL.

func normalizeSimilarIDs(ids []string) interface{} {
	if len(ids) == 0 {
		return nil
	}

	return ids
}

// ListByIDs возвращает упоминания по списку ID.

func (r *MentionRepo) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Mention, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	sql, args, err := r.sb.
		Select(

			"id", "brand_id", "source_id", "external_id",

			"title", "text", "url", "author", "published_at",

			"ml_label", "ml_score", "ml_is_relevant", "ml_similar_ids",

			"status", "deduplicated", "created_at", "updated_at",
		).
		From("mentions").
		Where(squirrel.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("MentionRepo.ListByIDs: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("MentionRepo.ListByIDs: query: %w", err)
	}

	defer rows.Close()

	var mentions []entity.Mention

	for rows.Next() {

		m, scanErr := scanMention(rows)

		if scanErr != nil {
			return nil, fmt.Errorf("MentionRepo.ListByIDs: scan: %w", scanErr)
		}

		mentions = append(mentions, *m)

	}

	return mentions, rows.Err()
}

// ListByStatusWithLimit возвращает упоминания по статусу с лимитом.

func (r *MentionRepo) ListByStatusWithLimit(ctx context.Context, status entity.MentionStatus, limit int) ([]entity.Mention, error) {
	if limit <= 0 {
		return []entity.Mention{}, nil
	}

	sql, args, err := r.sb.
		Select(

			"id", "brand_id", "source_id", "external_id",

			"title", "text", "url", "author", "published_at",

			"ml_label", "ml_score", "ml_is_relevant", "ml_similar_ids",

			"status", "deduplicated", "created_at", "updated_at",
		).
		From("mentions").
		Where(squirrel.Eq{"status": status}).
		OrderBy("created_at ASC").
		Suffix("LIMIT ?", limit).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("MentionRepo.ListByStatusWithLimit: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("MentionRepo.ListByStatusWithLimit: query: %w", err)
	}

	defer rows.Close()

	var mentions []entity.Mention

	for rows.Next() {

		m, scanErr := scanMention(rows)

		if scanErr != nil {
			return nil, fmt.Errorf("MentionRepo.ListByStatusWithLimit: scan: %w", scanErr)
		}

		mentions = append(mentions, *m)

	}

	return mentions, rows.Err()
}

func applyMentionFilters(builder squirrel.SelectBuilder, filter entity.MentionFilter) (squirrel.SelectBuilder, error) {
	if filter.BrandID != uuid.Nil {
		builder = builder.Where(squirrel.Eq{"sr.brand_id": filter.BrandID})
	}

	if filter.SourceID != uuid.Nil {
		builder = builder.Where(squirrel.Eq{"ci.source_id": filter.SourceID})
	}

	if filter.Sentiment != "" {
		builder = builder.Where(squirrel.Eq{"sr.sentiment": filter.Sentiment})
	}

	if filter.Search != "" {

		searchPattern := "%" + filter.Search + "%"

		builder = builder.Where("ci.text ILIKE ?", searchPattern)

	}

	if filter.DateFrom != "" {

		dateFrom, err := time.Parse(time.RFC3339, filter.DateFrom)
		if err != nil {
			return builder, fmt.Errorf("invalid date_from: %w", err)
		}

		builder = builder.Where(squirrel.GtOrEq{"ci.published_at": dateFrom})

	}

	if filter.DateTo != "" {

		dateTo, err := time.Parse(time.RFC3339, filter.DateTo)
		if err != nil {
			return builder, fmt.Errorf("invalid date_to: %w", err)
		}

		builder = builder.Where(squirrel.LtOrEq{"ci.published_at": dateTo})

	}

	return builder, nil
}

type mentionScanner interface {
	Scan(dest ...any) error
}

func scanMention(row mentionScanner) (*entity.Mention, error) {
	var mention entity.Mention

	var sentiment string
	var sourceID pgtype.UUID
	var clusterLabel pgtype.Int4

	err := row.Scan(

		&mention.ID,

		&mention.BrandID,

		&sourceID,

		&mention.SourceName,

		&mention.SourceType,

		&mention.Text,

		&mention.URL,

		&sentiment,

		&mention.PublishedAt,

		&mention.CreatedAt,

		&clusterLabel,
	)
	if err != nil {
		return nil, err
	}

	if sourceID.Valid {
		mention.SourceID = uuid.UUID(sourceID.Bytes)
	}

	if clusterLabel.Valid {
		v := int(clusterLabel.Int32)
		mention.ClusterLabel = &v
	}

	mention.Sentiment = entity.Sentiment(sentiment)

	return &mention, nil
}

// GetSimilarMentions возвращает до limit упоминаний того же кластера (brand_id + cluster_label), исключая excludeID.
func (r *MentionRepo) GetSimilarMentions(ctx context.Context, brandID uuid.UUID, clusterLabel int, excludeID uuid.UUID, limit int) ([]entity.Mention, error) {
	sql, args, err := r.sb.
		Select(
			"sr.id", "sr.brand_id", "ci.source_id",
			"COALESCE(s.name,'')", "COALESCE(s.type,'')",
			"ci.text", "ci.link",
			"sr.sentiment", "ci.published_at", "sr.created_at",
			"sr.cluster_label",
		).
		From("sentiment_results sr").
		Join("crawler_items ci ON ci.id = sr.item_id").
		LeftJoin("sources s ON s.id = ci.source_id").
		Where(squirrel.Eq{"sr.brand_id": brandID}).
		Where(squirrel.Eq{"sr.cluster_label": clusterLabel}).
		Where(squirrel.NotEq{"sr.id": excludeID}).
		OrderBy("ci.published_at DESC").
		Limit(uint64(limit)). //nolint:gosec
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("MentionRepo.GetSimilarMentions: build query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("MentionRepo.GetSimilarMentions: query: %w", err)
	}
	defer rows.Close()

	var mentions []entity.Mention
	for rows.Next() {
		m, scanErr := scanMention(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("MentionRepo.GetSimilarMentions: scan: %w", scanErr)
		}
		mentions = append(mentions, *m)
	}

	return mentions, rows.Err()
}

// isPgFKViolation возвращает true если ошибка — нарушение внешнего ключа (23503).

func isPgFKViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
