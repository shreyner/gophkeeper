package config

import "github.com/caarlos0/env/v7"

type Config struct {
	JWTSign        string `env:"JWT_SIGN,required" envDefault:"123"`
	GRPCServerPort int    `env:"GRPC_PORT" envDefault:"3200"`
	Port           int    `env:"PORT" envDefault:"3280"`

	S3MinioEndpoint        string `env:"S3_MINIO_ENDPOINT" envDefault:"localhost:9000"`
	S3MinioAccessKeyID     string `env:"S3_MINIO_ACCESS_KEY_ID" envDefault:"minio_access_key"`
	S3MinioSecretAccessKey string `env:"S3_MINIO_SECRET_ACCESS_KEY" envDefault:"minio_secret_key"`

	DBDSN string `env:"DB_DNS" envDefault:"postgres://postgres:postgres@localhost:5432/develop?sslmode=disable"`
}

func New() *Config {
	c := Config{}

	return &c
}

func (c *Config) Parse() error {
	err := env.Parse(c)

	if err != nil {
		return err
	}

	return nil
}
