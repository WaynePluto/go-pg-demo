package pkgs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new zap logger based on the server mode configuration.
// In debug mode, it uses a human-friendly console encoder.
// In release mode, it uses a JSON encoder for production environments.
func NewLogger(config *Config) (*zap.Logger, error) {
	var loggerConfig zap.Config
	
	if config.Server.Mode == "debug" {
		loggerConfig = zap.NewDevelopmentConfig()
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		loggerConfig = zap.NewProductionConfig()
	}
	
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}
	
	return logger, nil
}