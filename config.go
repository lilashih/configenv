// Package configenv loads application settings explicitly and testably without modifying the process environment.
package configenv

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Result reports the .env file used; an empty EnvFile means no file was read.
type Result struct {
	EnvFile string
}

// Load populates the struct pointer provided by the caller.
func Load(target any) error {
	_, err := LoadWithResult(target)
	return err
}

// LoadWithResult populates settings and reports the .env file path used.
func LoadWithResult(target any) (Result, error) {
	var result Result
	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return result, errors.New("config target must be a non-nil pointer to a struct")
	}
	paths, err := envFileCandidates()
	if err != nil {
		return result, err
	}
	fileVars, path, err := readFirstEnvFile(paths)
	if err != nil {
		return result, err
	}
	result.EnvFile = path

	variables := make(map[string]string)
	merge(variables, fileVars)
	merge(variables, processEnvironment())

	if err := env.Parse(target, env.Options{Environment: variables}); err != nil {
		return result, fmt.Errorf("failed to parse config fields: %w", err)
	}
	return result, nil
}

func envFileCandidates() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	return candidatePaths(cwd, executable, os.Getenv("APPIMAGE")), nil
}

func candidatePaths(cwd, executable, appImage string) []string {
	paths := []string{filepath.Join(cwd, ".env")}
	if appImage = strings.TrimSpace(appImage); appImage != "" {
		if dir := filepath.Dir(appImage); shouldFallback(cwd, dir) {
			paths = append(paths, filepath.Join(dir, ".env"))
		}
	}
	if dir := filepath.Dir(executable); shouldFallback(cwd, dir) {
		paths = append(paths, filepath.Join(dir, ".env"))
	}
	return paths
}

func shouldFallback(cwd, dir string) bool {
	base, baseErr := normalizedPath(cwd)
	target, targetErr := normalizedPath(dir)
	if baseErr != nil || targetErr != nil {
		return false
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel)
}

func normalizedPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(abs); resolveErr == nil {
		abs = resolved
	} else if !errors.Is(resolveErr, os.ErrNotExist) {
		return "", resolveErr
	}
	abs = filepath.Clean(abs)
	if runtime.GOOS == "windows" {
		abs = strings.ToLower(abs)
	}
	return abs, nil
}

func readFirstEnvFile(paths []string) (map[string]string, string, error) {
	for _, path := range paths {
		if path == "" {
			return nil, "", errors.New("config file candidate path cannot be empty")
		}
		file, err := os.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to open config file %q: %w", path, err)
		}
		content, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil {
			return nil, "", fmt.Errorf("failed to read config file %q: %w", path, readErr)
		}
		if closeErr != nil {
			return nil, "", fmt.Errorf("failed to close config file %q: %w", path, closeErr)
		}
		vars, parseErr := godotenv.UnmarshalBytes(bytes.TrimPrefix(content, []byte("\xef\xbb\xbf")))
		if parseErr != nil {
			return nil, "", fmt.Errorf("failed to parse config file %q: %w", path, parseErr)
		}
		return vars, path, nil
	}
	return nil, "", nil
}

func processEnvironment() map[string]string {
	variables := make(map[string]string)
	for _, entry := range os.Environ() {
		for i := 0; i < len(entry); i++ {
			if entry[i] == '=' {
				variables[entry[:i]] = entry[i+1:]
				break
			}
		}
	}
	return variables
}

func merge(dst, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}
