package time

import (
	arrow "github.com/bmuller/arrow/lib"

	"github.com/aisbergg/gonja/pkg/gonja/config"
)

type Config struct {
	// Default format
	DatetimeFormat string
	// If defined, now returns this parsed value
	Now *arrow.Arrow
}

func NewConfig() *Config {
	return &Config{
		DatetimeFormat: "%Y-%m-%d",
		Now:            nil,
	}
}

func (cfg *Config) Inherit() config.Inheritable {
	return &Config{
		DatetimeFormat: cfg.DatetimeFormat,
		Now:            cfg.Now,
	}
}

// DefaultConfig is a configuration with default values
var DefaultConfig = NewConfig()
