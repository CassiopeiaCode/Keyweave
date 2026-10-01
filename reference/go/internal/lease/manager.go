package lease

import (
	"fmt"
	"sync"
	"time"
)

type Lease struct {
	ID         string
	RequestID  string
	Attempt    int
	GroupID    string
	TemplateID string
	InstanceID string
	AcquiredAt time.Time
}

type Manager struct {
	mu     sync.Mutex
	active map[string]int
	nextID uint64
}

func NewManager() *Manager {
	return &Manager{active: make(map[string]int)}
}

// WithLock demonstrates the required atomicity boundary. Production code
// should hide this primitive behind a routing service so no caller can do
// network I/O while the lock is held.
func (m *Manager) WithLock(
	fn func(
		active func(string) int,
		acquire func(Lease) Lease,
	) error,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	active := func(instanceID string) int {
		return m.active[instanceID]
	}

	acquire := func(l Lease) Lease {
		m.nextID++
		l.ID = fmt.Sprintf("lease-%d", m.nextID)
		l.AcquiredAt = time.Now()
		m.active[l.InstanceID]++
		return l
	}

	return fn(active, acquire)
}

func (m *Manager) Release(instanceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.active[instanceID] <= 1 {
		delete(m.active, instanceID)
		return
	}
	m.active[instanceID]--
}

func (m *Manager) Active(instanceID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active[instanceID]
}
