package url

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"urlshortener/internal/config"
	"urlshortener/internal/models"
	"urlshortener/internal/services/userinfo"
)

const (
	urlEventsTopic = "url_events"

	eventCreated = "created"
	eventVisited = "visited"
	eventDeleted = "deleted"
)

//go:generate mockery --name=URLRepository --output=mocks --case=underscore
type URLRepository interface {
	SaveURL(ctx context.Context, urlToSave, short_url string) error
	GetURL(ctx context.Context, short_url string) (*models.URL, error)
	FetchAll(ctx context.Context) ([]*models.URL, error)
	DeleteURL(ctx context.Context, short_url string) error
}

//go:generate mockery --name=CacheRepository --output=mocks --case=underscore
type CacheRepository interface {
	Get(ctx context.Context, key string, target any) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type MessageBroker interface {
	Produce(msg any, topic, key string) error
}

type URLService struct {
	cfg        *config.Config
	logger     *slog.Logger
	repository URLRepository
	kafka      MessageBroker
	cache      CacheRepository
	userInfo   *userinfo.Service
}

func New(
	cfg *config.Config,
	l *slog.Logger,
	r URLRepository,
	k MessageBroker,
	c CacheRepository,
	us *userinfo.Service,
) *URLService {
	return &URLService{
		cfg:        cfg,
		logger:     l,
		repository: r,
		kafka:      k,
		cache:      c,
		userInfo:   us,
	}
}

func (s *URLService) SaveURL(ctx context.Context, original_url, short_url string) error {
	const op = "service.url.SaveURL"

	if err := s.repository.SaveURL(context.Background(), original_url, short_url); err != nil {
		s.logger.Error("failed to save url in db:", "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	urlEvent := models.UrlEvent{
		EventType:  eventCreated,
		ShortURL:   short_url,
		OriginaUrl: original_url,
		// TODO get from Auth service
		UserID:    "",
		EventTime: time.Now().UTC(),
	}
	go func() {
		if err := s.kafka.Produce(urlEvent, urlEventsTopic, short_url); err != nil {
			s.logger.Error("failed to produce message to broker", "msg", short_url, "error", err)
		}
	}()

	return nil
}

func (s *URLService) GetURL(ctx context.Context, short_url string) (*models.URL, error) {
	const op = "services.url.GetURL"

	// TODO get data from cache first
	url, err := s.repository.GetURL(ctx, short_url)
	if err != nil {
		s.logger.Error("url cannot be found in db", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *URLService) Visit(ctx context.Context, url *models.URL, r *http.Request) error {
	const op = "services.url.Visit"

	rm, err := s.userInfo.ExtractRequestMeta(ctx, r)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	urlEvent := models.UrlEvent{
		EventType:   "visited",
		ShortURL:    url.ShortURL,
		OriginaUrl:  url.OriginalURL,
		UserID:      "",
		EventTime:   time.Now().UTC(),
		RequestMeta: rm,
	}
	go func() {
		if err := s.kafka.Produce(urlEvent, urlEventsTopic, url.ShortURL); err != nil {
			s.logger.Error("failed to produce message to broker", "msg", url.ShortURL, "error", err)
		}
	}()

	return nil
}

func (s *URLService) DeleteURL(ctx context.Context, short_url string) error {
	const op = "services.url.DeleteURL"

	err := s.repository.DeleteURL(ctx, short_url)
	if err != nil {
		s.logger.Error("failed to delete url", "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	urlEvent := models.UrlEvent{
		EventType: "deleted",
		ShortURL:  short_url,
		UserID:    "",
		EventTime: time.Now().UTC(),
	}
	go func() {
		if err := s.kafka.Produce(urlEvent, urlEventsTopic, short_url); err != nil {
			s.logger.Error("failed to produce message to broker", "msg", short_url, "error", err)
		}
	}()

	return nil
}

// TODO add Pagination Query
func (s *URLService) GetAll(ctx context.Context) ([]*models.URL, error) {
	const op = "services.url.GetAll"

	urls, err := s.repository.FetchAll(ctx)
	if err != nil {
		s.logger.Error("error getting url from db", "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return urls, nil
}
