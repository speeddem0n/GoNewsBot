package source

import (
	"context"

	"github.com/SlyMarbo/rss"
	"github.com/samber/lo"

	"github.com/speeddem0n/GoNewsBot/internal/models"
)

type RSSSource struct { // структура для источника rss
	URL        string // Ссылка на источник
	SourceID   int64  // ID источника
	SourceName string // Название источника
}

func NewRSSSourceFromModel(m models.Source) RSSSource { // Конструктор для RSSSource
	return RSSSource{
		URL:        m.FeedURL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

func (s RSSSource) Fetch(ctx context.Context) ([]models.Item, error) {
	feed, err := s.loadFeed(ctx, s.URL) // Загрузаем методом loadFeed фид из источников
	if err != nil {
		return nil, err
	}

	return lo.Map(feed.Items, func(item *rss.Item, _ int) models.Item { // lo.Map Запускает цикл на переданом слайсе (feed.Items) и записывает все в слайс структур []models.Item
		return models.Item{
			Title:      item.Title,
			Categories: item.Categories,
			Link:       item.Link,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: s.SourceName,
		}
	}), nil

}

func (s RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) { // Метод для загрузки данных из источников
	var (
		feedCh = make(chan *rss.Feed) // Канал для передачи данных
		errCh  = make(chan error)     // Канал для передачи ошибок
	)

	go func() {
		feed, err := rss.Fetch(url) // Fetch downloads and parses the RSS feed at the given URL
		if err != nil {
			errCh <- err // Перадаем ошибку в канал в случае ее возникновения
			return
		}

		feedCh <- feed // Передаем rss фид в канал
	}()

	select {
	case <-ctx.Done(): // Кейс если контекст отменен или дедлайн наступил
		return nil, ctx.Err()
	case err := <-errCh: // Кейс если не получилось распарсить данные
		return nil, err
	case feed := <-feedCh: // Успешный кейс
		return feed, nil
	}
}

func (s RSSSource) ID() int64 { // Метод ID() для получения ID источника
	return s.SourceID
}

func (s RSSSource) Name() string { // Метод Name() для получения названия источника
	return s.SourceName
}
