package config

type AppConfig struct {
	Version string `envDefault:"1.0.0"`
	Mode    string `env:"APP_MODE" envDefault:"release"`
	Port    string `env:"APP_PORT" envDefault:"8000"`
}
