//go:generate ./bin/mockgen -source=./interfaces.go -destination=./mock/storage.go -package=stoken
package stoken

type JWTService interface {
	ParseToken(tokenString string) (*Data, error)
	CreateToken(data *Data) (string, error)
}
