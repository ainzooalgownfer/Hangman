package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"io/fs"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Address     string        `yaml:"address"`
	MaxLineSize int           `yaml:"max_line_size"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	MaxAttempts int           `yaml:"max_attempts"`
	SessionTTL  time.Duration `yaml:"session_time_to_live"`
	WordsPath   string        `yaml:"words_path"`
	Words       []string      `yaml:"words"`
}

func Default() Config {
	return Config{
		Address:     ":8080",
		MaxLineSize: 4096,
		IdleTimeout: 5 * time.Minute,
		MaxAttempts: 10,
		SessionTTL:  5 * time.Minute,
		WordsPath:   "words.json",
		Words: []string{
			"apple", "brave", "crane", "delta", "eagle",
			"flame", "grape", "honey", "igloo", "joker",
			"kneel", "lemon", "mango", "noble", "ocean",
			"piano", "queen", "river", "stone", "tiger",
			"ultra", "vivid", "whale", "xenon", "yacht",
			"zebra",
		},
	}
}

const fileName = "config.yaml"

func Load() (Config, error) {

	cfg := Default()
	data, err := os.ReadFile(fileName)

	if errors.Is(err, fs.ErrNotExist) {
		//default
	} else if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", fileName, err)
	} else {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parsing %s: %w", fileName, err)
		}
	}

	if cfg.WordsPath != "" {
		data, err := os.ReadFile(cfg.WordsPath)

		if errors.Is(err, fs.ErrNotExist) {
			//keep current list
		} else if err != nil {
			return Config{}, fmt.Errorf("reading %s: %w", cfg.WordsPath, err)
		} else {
			var words []string
			if err := json.Unmarshal(data, &words); err != nil {
				return Config{}, fmt.Errorf("parsing %s: %w", cfg.WordsPath, err)
			}
			cfg.Words = words
		}
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil

}

func (c Config) validate() error {
	if c.Address == "" {
		return errors.New("address must not be empty")
	}
	if c.MaxLineSize <= 0 {
		return fmt.Errorf("max_line_size must be > 0, got %d", c.MaxLineSize)
	}
	if c.IdleTimeout <= 0 {
		return fmt.Errorf("idle_timeout must be > 0, got %s", c.IdleTimeout)
	}
	if c.MaxAttempts <= 0 {
		return fmt.Errorf("max_attempts must be > 0, got %d", c.MaxAttempts)
	}
	if c.SessionTTL <= 0 {
		return fmt.Errorf("session_time_to_live must be > 0, got %s", c.SessionTTL)
	}
	if len(c.Words) == 0 {
		return errors.New("word list must not be empty")
	}
	for i, w := range c.Words {
		if w == "" {
			return fmt.Errorf("word %d is empty", i)
		}
	}
	return nil
}
