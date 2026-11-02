package config

import "github.com/lilashih/configenv"

var App AppConfig
var DB DBConfig

type configs struct {
	*AppConfig
	*DBConfig
}

// Initialize loads the application and database settings.
func Initialize() error {
	return configenv.Load(&configs{&App, &DB})
}

// InitializeWithResult loads settings and reports the .env file path used.
func InitializeWithResult() (configenv.Result, error) {
	return configenv.LoadWithResult(&configs{&App, &DB})
}
