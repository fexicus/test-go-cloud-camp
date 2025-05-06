package rate_limiting

import (
	"context"
	"errors"
	"sync"
	"time"

	"cloud/internal/db"
	"cloud/internal/pkg/logger"
)

// ErrClientNotFound возвращается, когда для clientID нет записи в таблице rate_limits
var ErrClientNotFound = errors.New("настройки для клиента не найдены")

// ClientConfig хранит параметры bucket из БД
type ClientConfig struct {
	Capacity   int // максимальное число токенов
	RefillRate int // скорость пополнения токенов в секунду
}

// ConfigStore позволяет получить ClientConfig по clientID
type ConfigStore interface {
	Get(ctx context.Context, clientID string) (ClientConfig, error)
}

// DBStore читает конфиги из БД и кеширует их в памяти
type DBStore struct {
	db    *db.Database
	cache map[string]ClientConfig
	mux   sync.RWMutex
}

// NewDBStore создаёт новый DBStore с пустым кешем
func NewDBStore(database *db.Database) *DBStore {
	return &DBStore{db: database, cache: make(map[string]ClientConfig)}
}

// Get пытается взять ClientConfig из кеша, иначе из БД
func (s *DBStore) Get(ctx context.Context, clientID string) (ClientConfig, error) {
	s.mux.RLock()
	if cfg, ok := s.cache[clientID]; ok {
		s.mux.RUnlock()
		return cfg, nil
	}
	s.mux.RUnlock()

	var cfg ClientConfig
	err := s.db.Get(ctx, &cfg,
		"SELECT capacity, refill_rate FROM rate_limits WHERE client_id=$1", clientID)
	if err != nil {
		return ClientConfig{}, ErrClientNotFound
	}

	s.mux.Lock()
	s.cache[clientID] = cfg
	s.mux.Unlock()
	return cfg, nil
}

// TokenBucket хранит текущее количество токенов и параметры refill
type TokenBucket struct {
	capacity   int
	tokens     int
	refillRate int
	mux        sync.Mutex
}

// newBucket инициализирует bucket полным количеством токенов
func newBucket(cfg ClientConfig) *TokenBucket {
	return &TokenBucket{capacity: cfg.Capacity, tokens: cfg.Capacity, refillRate: cfg.RefillRate}
}

// refill пополняет токены, не превышая capacity
func (b *TokenBucket) refill() {
	b.mux.Lock()
	b.tokens += b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.mux.Unlock()
}

// allow уменьшает токен и возвращает true, если токен был доступен
func (b *TokenBucket) allow() bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter управляет множеством TokenBucket для разных клиентов
type RateLimiter struct {
	store   ConfigStore
	buckets map[string]*TokenBucket
	mux     sync.Mutex
}

// NewRateLimiter создаёт RateLimiter с пустым набором buckets
func NewRateLimiter(store ConfigStore) *RateLimiter {
	return &RateLimiter{store: store, buckets: make(map[string]*TokenBucket)}
}

// Run запускает периодическое пополнение всех buckets
func (rl *RateLimiter) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mux.Lock()
			for _, b := range rl.buckets {
				b.refill()
			}
			rl.mux.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// Allow проверяет и расходует токен для clientID, логируя результаты
func (rl *RateLimiter) Allow(ctx context.Context, clientID string) bool {
	log := logger.NewThreadLogger(ctx)

	// ищем существующий bucket или создаём новый
	rl.mux.Lock()
	bucket, ok := rl.buckets[clientID]
	if !ok {
		log.Warn("[WARN]: нет конфига для clientID=%q", clientID)

		cfg, err := rl.store.Get(ctx, clientID)
		if err != nil {
			log.Error("[ERROR]: не найден конфиг для clientID=%q", clientID)
			rl.mux.Unlock()
			return false
		}

		bucket = newBucket(cfg)
		rl.buckets[clientID] = bucket
		log.Info("[RATE]: создан Bucket для %q capacity=%d refill_rate=%d",
			clientID, cfg.Capacity, cfg.RefillRate)
	}
	rl.mux.Unlock()

	// пытаемся взять токен
	allowed := bucket.allow()
	if allowed {
		log.Check("[CHECK]: клиент %q получил токен, осталось %d", clientID, bucket.tokens)
	} else {
		log.Warn("[WARN]: клиент %q превысил лимит, осталось %d", clientID, bucket.tokens)
	}

	return allowed
}
