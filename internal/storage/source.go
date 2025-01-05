package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/speeddem0n/GoNewsBot/internal/models"
)

type SourcePostgresStorage struct {
	db *sqlx.DB
}

type dbSource struct { // Внутренний тип для работы с базой данных
	ID      int64     `db:"id"`
	Name    string    `db:"name"`
	FeedURL string    `db:"feed_url"`
	Created time.Time `db:"created"`
}

func (s *SourcePostgresStorage) Sources(ctx context.Context) ([]models.Source, error) { // Метод для получения списка источников
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var sources []dbSource
	if err := conn.SelectContext(ctx, &sources, `SELECT * FROM source`); err != nil { // Выполняем sql запрос для получения списка источников
		return nil, err
	}

	return lo.Map(sources, func(dbSource dbSource, _ int) models.Source { return models.Source(dbSource) }), nil // Мапим структуру dbSource в models.Source
}

func (s *SourcePostgresStorage) SourceByID(ctx context.Context, id int64) (*models.Source, error) { // Метод для получения источника по его ID
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var source dbSource
	if err := conn.GetContext(ctx, &source, `SELECT * FROM source WHERE id = $1`, id); err != nil { // Выполняем sql запрос для получения источника по его id
		return nil, err
	}

	return (*models.Source)(&source), nil
}

func (s *SourcePostgresStorage) Add(ctx context.Context, source models.Source) (int64, error) { // Метод для добавления источника
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var id int64

	row := conn.QueryRowxContext( // Выполняем sql запрос для добавления источника
		ctx,
		`INSERT INTO source (name, feed_url, created) VALUES ($1, $2, $3) RETURNING id`,
		source.Name,
		source.FeedURL,
		source.Created,
	)

	if err := row.Err(); err != nil {
		return 0, err
	}

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *SourcePostgresStorage) Delete(ctx context.Context, id int64) error { // Метод для удаления источника
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DELETE FROM source WHERE id = $1`, id); err != nil { // Выполняем sql запрос для удаления источника
		return err
	}

	return nil
}
