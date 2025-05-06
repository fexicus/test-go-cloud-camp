package loadbalancer

import (
	"errors"
	"sync/atomic"

	"cloud/internal/backend"
	"cloud/internal/pkg/logger"
)

var ErrNoAliveBackends = errors.New("no alive backends available")

// RoundRobin распределяет HTTP-запросы по бэкендам в порядке round-robin
type RoundRobin struct {
	backends []*backend.Backend   // список бэкендов
	counter  uint64               // счётчик для выбора
	log      *logger.ThreadLogger // логгер
}

// NewRoundRobin создаёт RoundRobin с переданными бэкендами и логгером
func NewRoundRobin(backends []*backend.Backend, log *logger.ThreadLogger) *RoundRobin {
	rr := &RoundRobin{backends: backends, log: log}
	atomic.StoreUint64(&rr.counter, uint64(len(backends)-1))
	rr.log.Info("[INFO]: Инициализирован RoundRobin с %d бэкендами", len(backends))
	return rr
}

// NextBackend возвращает следующий доступный бэкенд или ошибку, если их нет
func (rr *RoundRobin) NextBackend() (*backend.Backend, error) {
	total := uint64(len(rr.backends))
	for i := uint64(0); i < total; i++ {
		idx := atomic.AddUint64(&rr.counter, 1) % total
		b := rr.backends[idx]
		if b.IsAlive() {
			rr.log.Info("[INFO]: Выбран бэкенд %s", b.URL.Host)
			return b, nil
		}
	}
	rr.log.Warn("[WARN]: Нет доступных бэкендов")
	return nil, ErrNoAliveBackends
}
