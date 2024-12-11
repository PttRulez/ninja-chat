package config

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/pttrulez/ninja-chat/internal/validator"
)

func ParseAndValidate(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	b, err := io.ReadAll(file)
	if err != nil {
		return Config{}, err
	}

	_, err = toml.Decode(string(b), &cfg)
	if err != nil {
		return Config{}, err
	}

	err = validator.Validator.Struct(cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
