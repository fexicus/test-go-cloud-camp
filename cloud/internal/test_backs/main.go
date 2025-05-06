package main

import (
	"cloud/internal/pkg/logger"
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// парсим флаги
	port := flag.String("port", "8081", "порт для запуска сервера")
	delay := flag.Duration("sleep", 0, "задержка для того, чтобы смотреть разные сценарии")
	flag.Parse()

	// готовим контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// создаём логгер
	log := logger.NewThreadLogger(ctx)
	defer log.Stop()

	// эндпоинт для проверки здоровья
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Check("[CHECK]: Health check received")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("statusOK"))
	})

	// основной обработчик всех остальных запросов
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// пропускаем /health
		if r.URL.Path == "/health" {
			return
		}

		/*		// симулируем задержку при необходимости
				if *delay > 0 {
					log.Info("[INFO]: Симулируем задержку %v", *delay)
					time.Sleep(*delay)
				}*/

		// возвращаем успех
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
		log.Info("[INFO]: Обработан запрос %s", r.URL.Path)
	})

	addr := ":" + *port
	log.Info("[INFO]: Запуск сервера на %s с задержкой %v", addr, *delay)

	// запускаем HTTP-сервер
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("[FATAL]: Ошибка запуска сервера: %v", err)
	}
}
