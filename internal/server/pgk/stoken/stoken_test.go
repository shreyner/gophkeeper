package stoken

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

var UUID1, _ = uuid.Parse("1DB1B358-A87D-407E-A8A2-2C761D75CFFC")

func TestService_CreateToken(t *testing.T) {
	type fields struct {
		signKey []byte
	}
	type args struct {
		data *Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Success create token",
			fields:  fields{signKey: []byte("123")},
			args:    args{data: &Data{ID: UUID1}},
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjFkYjFiMzU4LWE4N2QtNDA3ZS1hOGEyLTJjNzYxZDc1Y2ZmYyJ9._ECDYNRbnUk_1QtIGHXiumnOwKSHuFk7IaHRQYUwnQ8",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				signKey: tt.fields.signKey,
			}
			got, err := s.CreateToken(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_ParseToken(t *testing.T) {
	type fields struct {
		signKey []byte
	}
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Data
		wantErr bool
	}{
		{
			name:   "Success parse token",
			fields: fields{signKey: []byte("123")},
			args:   args{tokenString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjFkYjFiMzU4LWE4N2QtNDA3ZS1hOGEyLTJjNzYxZDc1Y2ZmYyJ9._ECDYNRbnUk_1QtIGHXiumnOwKSHuFk7IaHRQYUwnQ8"},
			want: &Data{
				ID: UUID1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				signKey: tt.fields.signKey,
			}
			got, err := s.ParseToken(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
