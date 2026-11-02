package configenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testSettings struct {
	Name string `env:"CONFIG_TEST_NAME" envDefault:"預設名稱"`
	Port int    `env:"CONFIG_TEST_PORT" envDefault:"8000"`
}

type TestSettings = testSettings

func writeEnv(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(workingDir(t), ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func workingDir(t *testing.T) string {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	return dir
}

func TestLoadProcessEnvironmentFirstAndNoProcessMutation(t *testing.T) {
	t.Setenv("CONFIG_TEST_NAME", "程序名稱")
	path := writeEnv(t, "CONFIG_TEST_NAME=檔案名稱\nCONFIG_TEST_PORT=9000\n")
	var got testSettings
	result, err := LoadWithResult(&got)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "程序名稱" || got.Port != 9000 || result.EnvFile != path {
		t.Fatalf("設定載入結果錯誤：%+v，來源：%+v", got, result)
	}
	if value := os.Getenv("CONFIG_TEST_NAME"); value != "程序名稱" {
		t.Fatalf("程序環境遭修改：%q", value)
	}
}

func TestLoadMissingFileAndDefaults(t *testing.T) {
	workingDir(t)
	t.Setenv("CONFIG_TEST_PORT", "9000")
	var got testSettings
	result, err := LoadWithResult(&got)
	if err != nil {
		t.Fatal(err)
	}
	if result.EnvFile != "" || got.Name != "預設名稱" || got.Port != 9000 {
		t.Fatalf("預設值錯誤：%+v", got)
	}
}

func TestLoadBOM(t *testing.T) {
	path := writeEnv(t, "\ufeffCONFIG_TEST_NAME=第一個檔案\n")
	var got testSettings
	result, err := LoadWithResult(&got)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "第一個檔案" || result.EnvFile != path {
		t.Fatalf("BOM 處理錯誤：%+v，%+v", got, result)
	}
}

func TestLoadInvalidInputAndFile(t *testing.T) {
	for _, target := range []any{nil, testSettings{}, (*testSettings)(nil)} {
		if err := Load(target); err == nil {
			t.Fatalf("目標 %T 應回報錯誤", target)
		}
	}
	_ = writeEnv(t, "CONFIG_TEST_PORT=abc\n")
	var got testSettings
	if err := Load(&got); err == nil || !strings.Contains(err.Error(), "Port") {
		t.Fatalf("欄位錯誤訊息不足：%v", err)
	}
	_ = writeEnv(t, "bad line without equal sign\n")
	if err := Load(&got); err == nil {
		t.Fatal("格式錯誤應回報")
	}
}

func TestLoadEmbeddedSettings(t *testing.T) {
	type AppSettings struct {
		Mode string `env:"CONFIG_TEST_MODE" envDefault:"release"`
	}
	type combined struct {
		*AppSettings
		*TestSettings
	}
	app := &AppSettings{}
	values := &TestSettings{}
	workingDir(t)
	if err := Load(&combined{app, values}); err != nil {
		t.Fatal(err)
	}
	if app.Mode != "release" || values.Port != 8000 {
		t.Fatalf("巢狀設定未載入：%+v，%+v", app, values)
	}
}

func TestLoadNamedNestedSettings(t *testing.T) {
	type database struct {
		Main struct {
			Path string `env:"CONFIG_TEST_DB_PATH" envDefault:"./database.sqlite"`
		}
	}
	type settings struct {
		DB database
	}
	workingDir(t)
	var got settings
	if err := Load(&got); err != nil {
		t.Fatal(err)
	}
	if got.DB.Main.Path != "./database.sqlite" {
		t.Fatalf("nested default was not loaded: %+v", got)
	}
}

func TestLoadUnreadableCandidateDoesNotBecomeMissing(t *testing.T) {
	dir := workingDir(t)
	if err := os.Mkdir(filepath.Join(dir, ".env"), 0700); err != nil {
		t.Fatal(err)
	}
	var got testSettings
	if err := Load(&got); err == nil {
		t.Fatalf("無法讀取的候選路徑應回傳錯誤：%v", err)
	}
}

func TestCandidatePaths(t *testing.T) {
	cwd := filepath.Join(t.TempDir(), "workspace")
	executable := filepath.Join(cwd, "bin", "app.exe")
	paths := candidatePaths(cwd, executable, "")
	if len(paths) != 1 || paths[0] != filepath.Join(cwd, ".env") {
		t.Fatalf("工作目錄內的執行檔不應產生後備路徑：%v", paths)
	}
	other := filepath.Join(t.TempDir(), "release")
	paths = candidatePaths(cwd, filepath.Join(other, "app.exe"), filepath.Join(other, "app.AppImage"))
	if len(paths) != 3 || paths[1] != filepath.Join(other, ".env") || paths[2] != paths[1] {
		t.Fatalf("外部執行位置應加入後備路徑：%v", paths)
	}
}
