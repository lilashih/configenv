package config

type DBConfig struct {
	Path string `env:"DB_PATH" envDefault:"./database.sqlite"`
}
