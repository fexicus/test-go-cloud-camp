package backend

import (
	"time"
)

func StartHealthCheck(backends []*Backend, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		for _, b := range backends {
			go b.HealthCheck()
		}
	}
}
