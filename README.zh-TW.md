# configenv

`configenv` 是 Go 的環境設定載入套件，目的是讓應用程式的設定集中管理，避免各處直接讀取環境變數或重複寫預設值。你可以在專案建立一個 `config` 目錄，依用途分檔定義設定，再由同一個入口載入。套件會從 `.env` 與程序環境變數讀取值，依 struct 欄位的 `env`、`envDefault` 標籤填入設定，套件不會修改程序環境變數。

## 安裝

```sh
go get github.com/lilashih/configenv@latest
```

## 快速開始

在程式的工作目錄建立 `.env`：

```dotenv
APP_PORT=9345
APP_MODE=debug
```

定義設定 struct，並在程式入口呼叫 `Load`：

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
	log.Printf("模式：%s，埠號：%d", app.Mode, app.Port)
}
```

如果還想知道實際讀取的 `.env` 路徑，改用 `LoadWithResult`：

```go
result, err := configenv.LoadWithResult(&app)
if err != nil {
	log.Fatal(err)
}
log.Printf("設定檔：%s", result.EnvFile)
```

沒有讀取 `.env` 時，`result.EnvFile` 是空字串。

## 完整範例

[`example/`](example/) 示範把設定集中放在 `config` 目錄，依用途分檔管理：

```text
example/
├── main.go          # 程式入口：載入後使用設定
└── config/
    ├── app.go       # AppConfig：應用設定
    ├── db.go        # DBConfig：資料庫設定
    └── main.go      # 統一載入入口與設定物件
```

[`example/config/main.go`](example/config/main.go) 定義 `App`、`DB`，並將兩組設定一起載入。其他程式碼可以透過 `config.App`、`config.DB` 取得設定，不必各自解析環境變數。此檔也示範兩種載入方法：

- [`Initialize`](example/config/main.go) 使用 `Load`，只取得設定值。
- [`InitializeWithResult`](example/config/main.go) 使用 `LoadWithResult`，同時取得 `.env` 路徑。
- [`example/main.go`](example/main.go) 依序呼叫兩種方法並印出結果。

在專案根目錄執行：

```sh
go run ./example
```

可先在專案根目錄建立 `.env` 試用，沒有檔案時，會使用程序環境變數與預設值。

## 載入規則

### 設定來源

載入順序固定為：**程序環境變數 → `.env` → `envDefault`**。例如在 PowerShell 執行 `$env:APP_PORT = "9000"` 後啟動程式，即使 `.env` 設了 `APP_PORT=9345`，程式仍會使用 `9000`。只有程序環境變數未提供 `APP_PORT` 時，才會使用 `.env` 的 `9345`。套件不會把 `.env` 的值寫回程序環境變數。

如果兩邊都沒有提供欄位所需的變數，才使用該欄位的 `envDefault`，若也沒有設定 `envDefault`，則保留 struct 欄位原本的值，原本沒有值時就是該型別的零值。

找不到 `.env` 是正常情況，仍會載入程序環境變數與欄位預設值。

### `.env` 搜尋位置

依序搜尋以下位置，使用第一個找到的檔案，不合併多個 `.env`：

1. 目前工作目錄。
2. `APPIMAGE` 指向的 AppImage 所在目錄，且該目錄位於工作目錄之外時。
3. 執行檔所在目錄，且該目錄位於工作目錄之外時。

若找到的檔案無法讀取或解析，會直接回傳錯誤。支援帶 UTF-8 BOM 的 `.env`，語法由 `godotenv` 解析。

這些設定在 **Go 程式執行時**才載入，若 Task、Wails CLI 或 Vite 在 Go 程式啟動前就需要環境變數，仍需在啟動流程中另外載入。

## 多組設定

可將已配置的 struct 指標嵌入同一個載入目標：

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

載入目標必須是非 nil 的 struct 指標，嵌入的指標也需先配置，欄位型別與標籤由 `github.com/caarlos0/env/v6` 解析。

## 錯誤處理

`Load` 回傳 `error`；`LoadWithResult` 額外回傳包含 `EnvFile` 的 `Result`。找不到 `.env` 不會報錯，此時 `EnvFile` 為空字串。非法目標、檔案權限或讀取錯誤、`.env` 格式錯誤及欄位解析錯誤會回傳 `error`。

載入失敗時，目標 struct 可能已有部分欄位被填入，應在成功後才使用。若多個 goroutine 對同一個設定物件載入，呼叫端需自行同步。

## 測試

```sh
go test ./...
```
