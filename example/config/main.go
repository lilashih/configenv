package config

import "github.com/lilashih/configenv"

var App AppConfig
var DB DBConfig

type configs struct {
	*AppConfig
	*DBConfig
}

// Initialize 載入應用與資料庫設定。
// Initialize loads the application and database settings.
func Initialize() error {
	return configenv.Load(&configs{&App, &DB})
}

// InitializeWithResult 載入設定，並回報實際使用的 .env 路徑。
// InitializeWithResult loads settings and reports the .env file path used.
func InitializeWithResult() (configenv.Result, error) {
	return configenv.LoadWithResult(&configs{&App, &DB})
}
