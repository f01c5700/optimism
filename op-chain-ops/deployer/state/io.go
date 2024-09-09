package state

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func ReadJSONFile(path string, data any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open JSON file for reading: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(data); err != nil {
		return fmt.Errorf("failed to decode JSON data: %w", err)
	}
	return nil
}

func WriteJSONFile(path string, data any) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open JSON file for writing: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON data: %w", err)
	}
	return nil
}

func ReadTOMLFile(path string, data any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open TOML file for reading: %w", err)
	}
	defer f.Close()
	if _, err := toml.NewDecoder(f).Decode(data); err != nil {
		return fmt.Errorf("failed to decode TOML data: %w", err)
	}
	return nil
}

func WriteTOMLFile(path string, data any) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open TOML file for writing: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(data); err != nil {
		return fmt.Errorf("failed to encode TOML data: %w", err)
	}
	return nil
}
