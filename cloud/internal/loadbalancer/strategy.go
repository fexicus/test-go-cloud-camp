package loadbalancer

import (
	"cloud/internal/backend"
)

type LoadBalancer interface {
	NextBackend() (*backend.Backend, error)
}
