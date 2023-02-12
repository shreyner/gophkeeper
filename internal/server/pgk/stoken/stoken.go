// Package stoken - generate session toke for authentication user
package stoken

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	_ JWTService = (*Service)(nil)
)

type Data struct {
	ID uuid.UUID // User uuid
}

type Service struct {
	signKey []byte
}

func NewService(signKey []byte) *Service {
	service := Service{
		signKey: signKey,
	}

	return &service
}

func (s *Service) ParseToken(tokenString string) (*Data, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.signKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	claim, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, ErrTokenInvalidClaims
	}

	id := claim["id"].(string)

	userID, err := uuid.Parse(id)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", "invalid id", ErrParsingData)
	}

	data := Data{
		ID: userID,
	}

	return &data, nil

}

func (s *Service) CreateToken(data *Data) (string, error) {
	mapClaims := jwt.MapClaims{
		"id": data.ID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)

	return token.SignedString(s.signKey)
}
