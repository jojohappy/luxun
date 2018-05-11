package storage

import (
	"fmt"
	"sync"
)

type Storage struct {
	lock         sync.RWMutex
	podStatusMap map[string]float64
}

func (m *Storage) Get(name string) (float64, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	count, ok := m.podStatusMap[name]
	if !ok {
		return -1, fmt.Errorf("unable to find data for pod status %v", name)
	}
	return count, nil
}

func (m *Storage) GetAll() map[string]float64 {
	m.lock.RLock()
	defer m.lock.RUnlock()
	status := make(map[string]float64, len(m.podStatusMap))

	for name, count := range m.podStatusMap {
		status[name] = count
	}
	return status
}

func (m *Storage) AddTick(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.podStatusMap[name]; !ok {
		m.podStatusMap[name] = 1
	} else {
		m.podStatusMap[name]++
	}
	return nil
}

func (m *Storage) Remove(name string) error {
	m.lock.Lock()
	delete(m.podStatusMap, name)
	m.lock.Unlock()
	return nil
}

func New() *Storage {
	return &Storage{
		podStatusMap: make(map[string]float64, 0),
	}
}

var storage *Storage
var once sync.Once

func StorageInst() *Storage {
	once.Do(func() {
		storage = New()
	})
	return storage
}
