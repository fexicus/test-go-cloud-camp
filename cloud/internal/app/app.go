package app

import (
	"cloud/internal/backend"
	"cloud/internal/config"
	"cloud/internal/db"
	"cloud/internal/loadbalancer"
	"cloud/internal/pkg/logger"
	"cloud/internal/rate-limiting"
	"cloud/internal/server"
	"context"
	"os"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()

	log := logger.NewThreadLogger(ctx)
	defer log.Stop()

	dsn, err := cfg.GetDSN()
	if err != nil {
		log.Error("Ошибка при получении DSN: %v", err)
		os.Exit(1)
	}
	dbPool, err := db.NewDb(ctx, dsn)
	if err != nil {
		log.Error("Не удалось подключиться к базе: %v", err)
		os.Exit(1)
	}

	store := rate_limiting.NewDBStore(dbPool)
	rl := rate_limiting.NewRateLimiter(store)
	go rl.Run(ctx)

	var backends []*backend.Backend
	for _, u := range cfg.Backends {
		b, err := backend.NewBackend(u, log)
		if err != nil {
			log.Warn("Некорректный бэкенд %s: %v", u, err)
			continue
		}
		backends = append(backends, b)
	}
	
	for _, b := range backends {
		b.HealthCheck()
	}
	go backend.StartHealthCheck(backends, cfg.HealthCheckInterval)

	lb := loadbalancer.NewRoundRobin(backends, log)

	log.Info("Запускаем сервер на %s", cfg.Address)
	server.StartServer(cfg, lb, rl, log)
}
