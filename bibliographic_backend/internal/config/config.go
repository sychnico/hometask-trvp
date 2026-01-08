package config

import (
	errs "bibliographic_litriture_gigachat/internal/errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	MaxFindingEnvInParentDirDepth = 10

	GIGACHAT_CLIENT_ID     = "GIGACHAT_CLIENT_ID"
	GIGACHAT_CLIENT_SECRET = "GIGACHAT_CLIENT_SECRET"
)

const (
	FrontendHostEnv     = "FRONTEND_HOST"
	FrontendHostWSEnv   = "FRONTEND_HOST_WS"
	BACKEND_HOST_ADRESS = "BACKEND_HOST_ADRESS"
	BACKEND_HOST_PORT   = "BACKEND_HOST_PORT"
	AllowedMethodsEnv   = "server.access_control_allow_methods"
	AllowCredentialsEnv = "server.access_control_allow_credentials"
	AllowedHeadersEnv   = "server.access_control_allow_headers"
)

type ConfigVariablesStruct struct {
	Server Server `yaml:"server" mapstructure:"server"`
	Cookie Cookie `yaml:"cookie" mapstructure:"cookie"`
}

type Server struct {
	Address                       string        `yaml:"address" mapstructure:"address"`
	Port                          int           `yaml:"port" mapstructure:"port"`
	ReadTimeout                   time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout                  time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	ShutdownTimeout               time.Duration `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout"`
	IdleTimeout                   time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	AccessControlAllowHeaders     string        `yaml:"idle_timeout" mapstructure:"access_control_allow_headers"`
	AccessControlAllowMethods     string        `yaml:"idle_timeout" mapstructure:"access_control_allow_methods"`
	AccessControlAllowCredentials bool          `yaml:"idle_timeout" mapstructure:"access_control_allow_credentials"`
}

type Cookie struct {
	SessionName   string        `yaml:"session_name" mapstructure:"session_name"`
	SessionLength int           `yaml:"session_length" mapstructure:"session_length"`
	HTTPOnly      bool          `yaml:"http_only" mapstructure:"http_only"`
	Secure        bool          `yaml:"secure" mapstructure:"secure"`
	SameSite      http.SameSite `yaml:"same_site" mapstructure:"same_site"`
	Path          string        `yaml:"path" mapstructure:"path"`
	ExpirationAge int           `yaml:"expiration_age" mapstructure:"expiration_age"`
}

func SetupNewConfig() (*ConfigVariablesStruct, error) {
	log.Info().Msg("Initializing config")

	if err := setupViper(); err != nil {
		log.Error().Err(fmt.Errorf("%w: %s", err, errs.ErrInitializeConfig)).Msg(fmt.Errorf("%w: %s", err, errs.ErrInitializeConfig).Error())
		return nil, fmt.Errorf("%w: %s", err, errs.ErrInitializeConfig)
	}

	var config ConfigVariablesStruct
	if err := viper.Unmarshal(&config); err != nil {
		log.Error().Err(fmt.Errorf("%w: %s", err, errs.ErrUnmarshalConfig)).Msg(fmt.Errorf("%w: %s", err, errs.ErrUnmarshalConfig).Error())
		return nil, fmt.Errorf("%w: %s", err, errs.ErrUnmarshalConfig)
	}

	// preoritize .env config
	if viper.GetString(BACKEND_HOST_ADRESS) != "" {
		config.Server.Address = viper.GetString(BACKEND_HOST_ADRESS)
	}

	// preoritize .env config
	if viper.GetString(BACKEND_HOST_PORT) != "" {
		config.Server.Port = viper.GetInt(BACKEND_HOST_PORT)
	}

	log.Info().Msg("Config initialized")

	return &config, nil
}

func setupServer() {
	viper.SetDefault("server.address", defaultConfigVariables.Server.Address)
	viper.SetDefault("server.port", defaultConfigVariables.Server.Port)
	viper.SetDefault("server.read_timeout", defaultConfigVariables.Server.ReadTimeout)
	viper.SetDefault("server.write_timeout", defaultConfigVariables.Server.WriteTimeout)
	viper.SetDefault("server.shutdown_timeout", defaultConfigVariables.Server.ShutdownTimeout)
	viper.SetDefault("server.idle_timeout", defaultConfigVariables.Server.IdleTimeout)
	viper.SetDefault("server.access_control_allow_headers", defaultConfigVariables.Server.AccessControlAllowHeaders)
	viper.SetDefault("server.access_control_allow_methods", defaultConfigVariables.Server.AccessControlAllowMethods)
	viper.SetDefault("server.access_control_allow_credentials", defaultConfigVariables.Server.AccessControlAllowCredentials)
}

func setupCookie() {
	viper.SetDefault("cookie.session_name", defaultConfigVariables.Cookie.SessionName)
	viper.SetDefault("cookie.session_length", defaultConfigVariables.Cookie.SessionLength)
	viper.SetDefault("cookie.http_only", defaultConfigVariables.Cookie.HTTPOnly)
	viper.SetDefault("cookie.secure", defaultConfigVariables.Cookie.Secure)
	viper.SetDefault("cookie.same_site", defaultConfigVariables.Cookie.SameSite)
	viper.SetDefault("cookie.path", defaultConfigVariables.Cookie.Path)
	viper.SetDefault("cookie.expiration_age", defaultConfigVariables.Cookie.ExpirationAge)
}

func findEnvDir() (string, error) {
	log.Info().Msg("Finding environment dir")
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, errs.ErrGetDirectory)
	}

	for i := 0; i < MaxFindingEnvInParentDirDepth; i++ {
		path := filepath.Join(currentDir, ".env")
		if _, err := os.Stat(path); err == nil {
			log.Info().Msgf("Found .env file in %s", path)
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("%w: %s", err, errs.ErrDirectoryNotFound)
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("%w: %s", err, errs.ErrDirectoryNotFound)
}

func findServerConfigDir() (string, error) {
	log.Info().Msg("Finding server config dir")
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, errs.ErrGetDirectory)
	}

	for i := 0; i < MaxFindingEnvInParentDirDepth; i++ {
		path := filepath.Join(currentDir, viper.GetString("VIPER_SERVER_CONFIG_PATH")+"config.yml")
		if _, err := os.Stat(path); err == nil {
			log.Info().Msgf("Found server config file in %s", path)
			return filepath.Join(currentDir, viper.GetString("VIPER_SERVER_CONFIG_PATH")), nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("%w: %s", err, errs.ErrDirectoryNotFound)
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("%w: %s", err, errs.ErrDirectoryNotFound)
}

func setupViper() error {
	log.Info().Msg("Initializing viper")

	envDir, err := findEnvDir()
	if err != nil {
		wrapped := fmt.Errorf("%w: %s", err, errs.ErrDirectoryNotFound)
		log.Error().Err(wrapped).Msg(wrapped.Error())
		return wrapped
	}

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(envDir)

	if err := viper.ReadInConfig(); err != nil {
		wrapped := fmt.Errorf("%w: %s", err, errs.ErrReadEnvironment)
		log.Error().Err(wrapped).Msg(wrapped.Error())
		return wrapped
	}

	viperServerConfigDir, err := findServerConfigDir()
	if err != nil {
		wrapped := fmt.Errorf("%w: %s", err, errs.ErrDirectoryNotFound)
		log.Error().Err(wrapped).Msg("Server config not found, using defaults instead!")
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
		viper.AddConfigPath(viperServerConfigDir)
	}

	setupServer()
	setupCookie()

	if err := viper.MergeInConfig(); err != nil {
		wrapped := fmt.Errorf("%w: %s", err, errs.ErrReadConfig)
		log.Error().Err(wrapped).Msg(wrapped.Error())
		return wrapped
	}

	log.Info().Msg("Viper initialized")
	return nil
}
