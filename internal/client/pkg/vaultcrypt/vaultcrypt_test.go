package vaultcrypt

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestVaultCrypt_Encrypt(t *testing.T) {
	type fields struct {
		login    string
		password string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Success encrypted",
			fields: fields{
				login:    "Alex",
				password: "123",
			},
			args: args{
				data: []byte("123"),
			},
			want:    "wVba8uysA12MosMFCcVfVn1SnQ==",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			err := c.SetMasterPassword(tt.fields.login, tt.fields.password)

			wantBytes, _ := base64.StdEncoding.DecodeString(tt.want)

			got, err := c.Encrypt(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, wantBytes) {
				t.Errorf("Encrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVaultCrypt_Decrypt(t *testing.T) {
	type fields struct {
		login    string
		password string
	}
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Success decrypt",
			fields: fields{
				login:    "Alex",
				password: "123",
			},
			args: args{
				data: "wVba8uysA12MosMFCcVfVn1SnQ==",
			},
			want:    []byte("123"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			err := c.SetMasterPassword(tt.fields.login, tt.fields.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetMasterPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			entryData, _ := base64.StdEncoding.DecodeString(tt.args.data)

			got, err := c.Decrypt(entryData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
