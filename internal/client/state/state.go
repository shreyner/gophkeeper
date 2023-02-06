package state

import (
	"sync"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultdata"
)

var (
	_ vaultdata.State = (*State)(nil)
)

type State struct {
	IsAuth    bool
	userToken string

	mux sync.RWMutex
}

func New() *State {
	state := State{}

	return &state
}

func (s *State) SetUserToken(token string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.userToken = token
	s.IsAuth = true
}

func (s *State) GetUserToken() string {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.userToken
}
