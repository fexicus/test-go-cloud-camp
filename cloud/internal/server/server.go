package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud/internal/config"
	"cloud/internal/loadbalancer"
	"cloud/internal/pkg/logger"
	"cloud/internal/rate-limiting"
)

func StartServer(cfg *config.Config, lb loadbalancer.LoadBalancer, rl *rate_limiting.RateLimiter, log *logger.ThreadLogger) {
	// создаём HTTP multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// определяем IP клиента
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = strings.Split(r.RemoteAddr, ":")[0]
		}
		// проверяем rate limit
		if !rl.Allow(r.Context(), clientIP) {
			log.Error("[ERROR]: Клиент %s превысил лимит", clientIP)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// выбираем бэкенд и проксируем
		start := time.Now()
		b, err := lb.NextBackend()
		if err != nil {
			log.Error("[ERROR]: Все бэкенды недоступны: %v", err)
			http.Error(w, "Service Unavailable", http.StatusBadGateway)
			return
		}
		b.ServeProxy(w, r)
		// логируем информацию о запросе и времени обработки
		log.Info("[INFO]: Запрос %s %s -> %s (%.3fms)",
			r.Method, r.URL.Path, b.URL.Host,
			float64(time.Since(start).Microseconds())/1000,
		)
	})

	// настраиваем HTTP сервер
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// graceful shutdown по сигналам
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info("[INFO]: Останавливаем сервер")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		os.Exit(0)
	}()

	// старт сервера
	log.Info("[INFO]: Запускаем сервер на %s", cfg.Address)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("[FATAL]: Фатальная ошибка сервера: %v", err)
	}
}
