package config

type ServerConfig struct {
	Port string `env:"PORT" envDefault:"8080"`
}
