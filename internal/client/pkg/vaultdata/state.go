package vaultdata

type State interface {
	SetUserToken(token string)
	GetUserToken() string
}
