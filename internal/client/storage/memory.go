package storage

type MemoryStorage struct {
	loginMap map[string]string
	notesMap map[string]string
	cardMap  map[string]string
}

func New() *MemoryStorage {
	s := MemoryStorage{}

	return &s
}
