package state

import "sync"

type State struct {
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
}

func (s *State) GetUserToken() string {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.userToken
}
