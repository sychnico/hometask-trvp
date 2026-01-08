package config

import (
	"net/http"
	"time"
)

// default config constants
var (
	defaultConfigVariables = ConfigVariablesStruct{
		Server: Server{
			Address:                       "localhost",
			Port:                          8080,
			ReadTimeout:                   time.Second * 5,
			WriteTimeout:                  time.Second * 5,
			ShutdownTimeout:               time.Second * 30,
			IdleTimeout:                   time.Second * 60,
			AccessControlAllowHeaders:     "Content-Type",
			AccessControlAllowMethods:     "POST, OPTIONS",
			AccessControlAllowCredentials: true,
		},
		Cookie: Cookie{
			SessionName:   "session_id",
			SessionLength: 32,
			HTTPOnly:      true,
			Secure:        false,
			SameSite:      http.SameSiteStrictMode,
			Path:          "/",
			ExpirationAge: -1,
		},
	}
)
