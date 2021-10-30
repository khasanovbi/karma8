package storageholder

import (
	"karma8"
	"karma8/internal/storage"
	"sync"
)

type storageHolder struct {
	hostToStorage map[string]karma8.Storage
	lock          sync.Mutex
}

func (m *storageHolder) GetStorage(host string) karma8.Storage {
	m.lock.Lock()
	defer m.lock.Unlock()

	s, ok := m.hostToStorage[host]
	if !ok {
		s = storage.NewInMemory()
		m.hostToStorage[host] = s
	}
	return s
}

func New() karma8.StorageHolder {
	return &storageHolder{
		hostToStorage: map[string]karma8.Storage{},
	}
}
