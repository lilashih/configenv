# configenv

`configenv` is a Go package for loading environment settings. It helps keep application settings in one place instead of reading environment variables or repeating default values throughout the codebase. You can create a `config` directory in your project, define settings in separate files by purpose, and load them from a single entry point. The package reads values from `.env` and process environment variables, then populates struct fields using the `env` and `envDefault` tags. It does not modify process environment variables.

## Installation

```sh
go get github.com/lilashih/configenv@latest
```

## Quick start

Create a `.env` file in the program's working directory:

```dotenv
APP_PORT=9345
APP_MODE=debug
```

Define a settings struct and call `Load` at the program entry point:

```go
package main

import (
	"log"

	"github.com/lilashih/configenv"
)

type AppConfig struct {
	Port int    `env:"APP_PORT" envDefault:"8000"`
	Mode string `env:"APP_MODE" envDefault:"release"`
}

func main() {
	var app AppConfig
	if err := configenv.Load(&app); err != nil {
		log.Fatal(err)
	}
	log.Printf("mode: %s, port: %d", app.Mode, app.Port)
}
```

To also get the path of the `.env` file that was read, use `LoadWithResult`:

```go
result, err := configenv.LoadWithResult(&app)
if err != nil {
	log.Fatal(err)
}
log.Printf("config file: %s", result.EnvFile)
```

If no `.env` file was read, `result.EnvFile` is an empty string.

## Complete example

[`example/`](example/) shows how to keep settings in a `config` directory, with separate files for different purposes:

```text
example/
├── main.go          # Program entry point: loads and uses settings
└── config/
    ├── app.go       # AppConfig: application settings
    ├── db.go        # DBConfig: database settings
    └── main.go      # Shared loading entry point and settings objects
```

[`example/config/main.go`](example/config/main.go) defines `App` and `DB` and loads both groups of settings together. Other code can access settings through `config.App` and `config.DB` without parsing environment variables separately. This file also demonstrates both loading methods:

- [`Initialize`](example/config/main.go) uses `Load` to get settings only.
- [`InitializeWithResult`](example/config/main.go) uses `LoadWithResult` to also get the `.env` file path.
- [`example/main.go`](example/main.go) calls both methods in sequence and prints the results.

Run this from the project root:

```sh
go run ./example
```

You can create a `.env` file in the project root to try it. Without one, the example uses process environment variables and defaults.

## Loading rules

### Setting sources

The priority is fixed: **process environment variables → `.env` → `envDefault`**. For example, if you run `$env:APP_PORT = "9000"` in PowerShell before starting the program, it uses `9000` even if `.env` contains `APP_PORT=9345`. It uses the value `9345` from `.env` only when the process environment does not provide `APP_PORT`. The package does not write values from `.env` back to the process environment.

If neither source provides a variable for a field, the field uses its `envDefault` value. If there is no `envDefault`, the existing value in the struct field is retained; an unset field keeps its type's zero value.

An absent `.env` file is normal. Process environment variables and field defaults are still loaded.

### `.env` search locations

The package searches these locations in order and uses the first file it finds. It does not merge multiple `.env` files:

1. The current working directory.
2. The directory containing the AppImage pointed to by `APPIMAGE`, if that directory is outside the working directory.
3. The directory containing the executable, if that directory is outside the working directory.

If a file is found but cannot be read or parsed, the package returns an error immediately. `.env` files with a UTF-8 BOM are supported. File syntax is parsed by `godotenv`.

Settings are loaded **when the Go program runs**. If Task, the Wails CLI, or Vite needs an environment variable before the Go program starts, it must be loaded separately as part of the startup process.

## Multiple groups of settings

You can embed initialized struct pointers in a single loading target:

```go
type DBConfig struct {
	Path string `env:"DB_PATH" envDefault:"./database.sqlite"`
}

type Settings struct {
	*AppConfig
	*DBConfig
}

app := &AppConfig{}
db := &DBConfig{}
err := configenv.Load(&Settings{AppConfig: app, DBConfig: db})
```

The loading target must be a non-nil pointer to a struct, and embedded pointers must be initialized first. Field types and tags are parsed by `github.com/caarlos0/env/v6`.

## Error handling

`Load` returns an `error`; `LoadWithResult` additionally returns a `Result` containing `EnvFile`. An absent `.env` file is not an error, and `EnvFile` is an empty string in that case. An invalid target, file permission or read errors, invalid `.env` syntax, and field parsing errors return an `error`.

If loading fails, some fields in the target struct may already have been populated. Use the struct only after loading succeeds. Callers must also synchronize concurrent loading into the same settings object.

## Tests

```sh
go test ./...
```
