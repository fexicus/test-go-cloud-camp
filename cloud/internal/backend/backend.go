package backend

import (
	"cloud/internal/pkg/logger"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	URL         *url.URL
	Proxy       *httputil.ReverseProxy
	LastChecked time.Time
	alive       bool
	mux         sync.RWMutex
	log         *logger.ThreadLogger
}

func NewBackend(rawURL string, log *logger.ThreadLogger) (*Backend, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		ts := time.Now().Format("2006-01-02 15:04:05")
		log.Error("[ERROR] %s Invalid backend URL %s: %v", ts, rawURL, err)
		return nil, err
	}

	b := &Backend{
		URL:         parsed,
		alive:       true,
		LastChecked: time.Now(),
		log:         log,
	}

	proxy := httputil.NewSingleHostReverseProxy(parsed)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		ts := time.Now().Format("2006-01-02 15:04:05")
		b.log.Error("[ERROR] %s Proxy error to %s: %v", ts, parsed.Host, err)
		go b.SetAlive(false)
		http.Error(rw, "Bad Gateway", http.StatusBadGateway)
	}
	b.Proxy = proxy

	return b, nil
}

func (b *Backend) ServeProxy(w http.ResponseWriter, r *http.Request) {
	b.Proxy.ServeHTTP(w, r)
}

func (b *Backend) SetAlive(up bool) {
	b.mux.Lock()
	b.alive = up
	b.LastChecked = time.Now()
	b.mux.Unlock()
}

// IsAlive возвращает состояние живости
func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.alive
}

// HealthCheck получает /health, обновляет alive и логирует переходы с префиксом и временем
func (b *Backend) HealthCheck() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(b.URL.String() + "/health")
	up := err == nil && resp.StatusCode == http.StatusOK

	ts := time.Now().Format("2006-01-02 15:04:05")
	if err != nil {
		b.log.Warn("[WARN] %s Health check failed on %s: %v", ts, b.URL.Host, err)
	}

	prev := b.IsAlive()
	b.SetAlive(up)

	if up && !prev {
		b.log.Check("[CHECK] %s Backend %s is healthy", ts, b.URL.Host)
	} else if !up && prev {
		b.log.Warn("[WARN] %s Backend %s is down", ts, b.URL.Host)
	}
}
