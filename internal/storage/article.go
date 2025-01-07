package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/speeddem0n/GoNewsBot/internal/models"
)

type ArticlePostgresStorage struct { // Структура Хранилища статей принимает подключение к бд
	db *sqlx.DB
}

func NewArticleStorage(db *sqlx.DB) *ArticlePostgresStorage { // Конструктор для структуры ArticlePostgresStorage
	return &ArticlePostgresStorage{db: db}
}

type dbArticle struct { // Внутренний тип для работы с базой данных
	ID        int64        `db:"id"`
	SourceID  int64        `db:"source"`
	Title     string       `db:"title"`
	Link      string       `db:"link"`
	Summary   string       `db:"summary"`
	Published time.Time    `db:"published"`
	Posted    sql.NullTime `db:"posted"`
	Created   time.Time    `db:"created"`
}

func (s *ArticlePostgresStorage) Store(ctx context.Context, article models.Article) error { // Метод Store для сохранения статьи в бд
	conn, err := s.db.Connx(ctx) // Получаем соеденение с БД
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `INTERT INTO article (source_id, title, link, summary, published) 
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT DO NOTHING`, // Выолняем sql запрос для добавления статьи в БД
		article.SourceID,
		article.Title,
		article.Link,
		article.Summary,
		article.Published,
	); err != nil {
		return err
	}

	return nil
}

func (s *ArticlePostgresStorage) AllNotPosted(ctx context.Context, since time.Time, limit uint64) ([]models.Article, error) { // Метод AllNotPosted возвращает все статьи которые не были опубликованы в тг канал начиная с определенного времени
	conn, err := s.db.Connx(ctx) // Получаем соеденение с БД
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var articles []dbArticle
	if err := conn.SelectContext(ctx, &articles, `SELECT * FROM article WHERE posted IS NULL 
	AND published >= $1::timestamp ORDER BY published DESC LIMIT $2`, // Выолняем sql запрос для получения всех неопубликованных статей
		since.UTC().Format(time.RFC3339), // Ворматируем дату в нужный формат
		limit,
	); err != nil {
		return nil, err
	}

	return lo.Map(articles, func(article dbArticle, _ int) models.Article { // Мапим структуру dbArticle в models.Article
		return models.Article{
			ID:        article.ID,
			SourceID:  article.SourceID,
			Title:     article.Title,
			Link:      article.Link,
			Summary:   article.Summary,
			Posted:    article.Posted.Time,
			Published: article.Published,
			Created:   article.Created,
		}
	}), nil
}

func (s *ArticlePostgresStorage) MarkPosted(ctx context.Context, id int64) error { // Метод MarkPosted для отметки статьи как уже опубликованую
	conn, err := s.db.Connx(ctx) // Получаем соеденение с БД
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `UPDATE article SET posted = $1::timestamp WHERE id = $2`, // Выолняем sql запрос UPDATE для добавления даты публикации в бд
		time.Now().UTC().Format(time.RFC3339),
		id,
	); err != nil {
		return err
	}

	return nil
}
