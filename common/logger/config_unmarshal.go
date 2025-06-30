package logger

import (
	"encoding/json"
)

type UnmarshalConfig struct {
	Level             string `mapstructure:"level" json:"level"`
	Format            string `mapstructure:"format" json:"format"`
	OutputPath        string `mapstructure:"output_path" json:"output_path"`
	DisableCaller     bool   `mapstructure:"disable_caller" json:"disable_caller"`
	DisableStacktrace bool   `mapstructure:"disable_stacktrace" json:"disable_stacktrace"`
	FilePath          string `mapstructure:"file_path" json:"file_path"`
	MaxSize           int    `mapstructure:"max_size" json:"max_size"`
	MaxBackups        int    `mapstructure:"max_backups" json:"max_backups"`
	MaxAge            int    `mapstructure:"max_age" json:"max_age"`
}

// ToConfig converts UnmarshalConfig to Config
func (uc *UnmarshalConfig) ToConfig() *Config {
	return &Config{
		Level:             ParseLogLevel(uc.Level),
		Format:            uc.Format,
		OutputPath:        uc.OutputPath,
		DisableCaller:     uc.DisableCaller,
		DisableStacktrace: uc.DisableStacktrace,
		FilePath:          uc.FilePath,
		MaxSize:           uc.MaxSize,
		MaxBackups:        uc.MaxBackups,
		MaxAge:            uc.MaxAge,
	}
}

// ConfigWrapper wraps Config for custom unmarshaling
type ConfigWrapper struct {
	*Config
}

// UnmarshalJSON implements custom JSON unmarshaling
func (cw *ConfigWrapper) UnmarshalJSON(data []byte) error {
	var uc UnmarshalConfig
	if err := json.Unmarshal(data, &uc); err != nil {
		return err
	}
	cw.Config = uc.ToConfig()
	return nil
}

// UnmarshalText implements custom text unmarshaling for LogLevel
func (l *LogLevel) UnmarshalText(text []byte) error {
	*l = ParseLogLevel(string(text))
	return nil
}

// MarshalText implements custom text marshaling for LogLevel
func (l LogLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}
