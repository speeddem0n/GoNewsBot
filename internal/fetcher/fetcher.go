package fetcher

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/speeddem0n/GoNewsBot/internal/models"
	"github.com/speeddem0n/GoNewsBot/internal/source"
)

type ArticleStorage interface { // interface Article для работы со слоем Article бд
	Store(ctx context.Context, article models.Article) error
}

type SourceProvider interface { // interface SourceProvider для работы со слоем Source бд
	Sources(ctx context.Context) ([]models.Source, error)
}

type Source interface { // interface для связи со слоем fetcher/fetch
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]models.Item, error)
}

type Fetcher struct {
	articles ArticleStorage // interface Article для работы со слоем Article бд
	sources  SourceProvider // interface SourceProvider для работы со слоем Source бд

	fetchInterval  time.Duration // Интервал с которым мы будем проходить по источникам и собирать статьи
	filterKeywords []string      // Ключевые слова по которым мы будем фильтровать статьи
}

func NewFetcher( // Конструктор для структуры Fetcher
	articleStorage ArticleStorage,
	sources SourceProvider,
	fetchInterval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:       articleStorage,
		sources:        sources,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Start(ctx context.Context) error { // Метод для запуска Fetch'a в отдельной горутине
	tiker := time.NewTicker(f.fetchInterval) // Создаем тикер с заданым интервалом
	defer tiker.Stop()                       // Откладываем завершение тикера

	if err := f.Fetch(ctx); err != nil { // Первый Fetch запуск без ожидания интервала
		return err
	}

	for { // Бесконечный цикл
		select {
		case <-ctx.Done(): // Контекст завершен
			return ctx.Err() // Возвращаем ошибку контекста
		case <-tiker.C: // Сработал тикер
			if err := f.Fetch(ctx); err != nil { // Вызываем Fetch
				return err
			}
		}
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error { // Метод Fetch проходится по источникам, достает из них статьи, и сохраняем их в базу данных
	sources, err := f.sources.Sources(ctx) // Получаем список источников
	if err != nil {
		return err
	}

	var wg sync.WaitGroup // Создаем WaitGroup

	for _, src := range sources { // в отдельных горутинах проходимся по источникам
		wg.Add(1)

		rssSource := source.NewRSSSourceFromModel(src) // преобразуем модель source в rssSource

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx) // Достаем статью из источника методом Fetch()
			if err != nil {
				logrus.Errorf("An error occured while fetching items from source %q: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil { // Сохраням статью в БД методом processItems
				logrus.Errorf("An error occured processing items from source %q: %v", source.Name(), err)
				return
			}

		}(rssSource)
	}

	wg.Wait() // Ждем завершения всех горутин

	return nil // Возвращаем нил в слуае успеха
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []models.Item) error { // Метод для добавления статьи в БД
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldbeSkipped(item) { // Методом itemShouldbeSkipped проверям нужно ли пропустить статью
			continue
		}

		if err := f.articles.Store(ctx, models.Article{ // Методом articles.Store сохраняем статью в БД
			SourceID:  source.ID(),
			Title:     item.Title,
			Link:      item.Link,
			Summary:   item.Summary,
			Published: item.Date,
		}); err != nil {
			return err
		}
	}

	return nil // Возвращаем нил в случае успеха
}

func (f *Fetcher) itemShouldbeSkipped(item models.Item) bool { // Метод для проверки, нужно ли пропускать статью

	for _, keyword := range f.filterKeywords {
		if titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword); titleContainsKeyword { // Если Keyword содержится в названии статьи возвращаем true
			return true
		}

		for _, category := range item.Categories {
			if keyword == category { // Если Keyword содержится в категориях статьи возвращаем true
				return true
			}
		}
	}

	return false // Возврашаем false если статья проходит фильтр
}
