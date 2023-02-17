package config

import "github.com/caarlos0/env/v7"

type Config struct {
	HostGRPC   string `env:"HOST_GRPC" envDefault:":3200"`
	ServerName string `env:"SERVER_NAME" envDefault:"example.com"`
	HostREST   string `env:"HOST_REST" envDefault:"https://localhost:3280"`
	Insecure   bool   `env:"INSECURE" envDefault:"true"`

	CertFile string `env:"CERT_FILE,file" envDefault:"./cert/server-cert.pem"`

	DataFolder string `env:"DATA_FOLDER" envDefault:"./data"`
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
