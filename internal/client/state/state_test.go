package state

import (
	"testing"
)

func TestState_SetUserToken(t *testing.T) {
	t.Run("Success set and check", func(t *testing.T) {
		var token = "some token"
		s := New()

		s.SetUserToken(token)

		if got := s.GetUserToken(); got != token {
			t.Errorf("GetUserToken() = %v, want %v", got, token)
		}
	})

	t.Run("Check parallel read", func(t *testing.T) {
		var token = "some token"
		s := New()

		s.SetUserToken(token)

		if got := s.GetUserToken(); got != token {
			t.Errorf("GetUserToken() = %v, want %v", got, token)
		}
	})
}

func TestState_GetUserToken(t *testing.T) {
	t.Run("Check get data without set auth", func(t *testing.T) {
		s := New()

		if got := s.GetUserToken(); got != "" {
			t.Errorf("GetUserToken() = %v, want empty string", got)
		}
	})
}
