// Package config centralizes environment-derived configuration so it is
// read once at startup and injected into dependent packages via their
// constructors, rather than each package reading os.Getenv itself.
package config

import (
	"log"
	"os"
)

const (
	defaultAddr       = ":8080"
	defaultCORSOrigin = "*"
	devAuthSecret     = "playard-dev-secret-do-not-use-in-prod"
)

// Config holds all environment-derived settings for the server.
type Config struct {
	Addr       string
	CORSOrigin string
	AuthSecret []byte
}

// Load reads configuration from the environment, applying defaults for
// anything unset. AUTH_SECRET falls back to an insecure dev value with a
// loud warning rather than failing startup, since this is a small
// no-accounts project where a leaked dev secret has low blast radius.
func Load() Config {
	authSecret := os.Getenv("AUTH_SECRET")
	if authSecret == "" {
		log.Println("WARNING: AUTH_SECRET not set, using insecure dev secret — do not run this in production")
		authSecret = devAuthSecret
	}

	return Config{
		Addr:       getEnv("ADDR", defaultAddr),
		CORSOrigin: getEnv("CORS_ORIGIN", defaultCORSOrigin),
		AuthSecret: []byte(authSecret),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
