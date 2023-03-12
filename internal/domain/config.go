package domain

import (
	"net/url"
	"testing"
)

type Config struct {
	BaseURL *url.URL `env:"BASE_URL" envDefault:"http://localhost:3000/"`
	Bind    string   `end:"BIND" envDefault:":3000"`
	Name    string   `env:"NAME" envDefault:"WebSub"`
}

func TestConfig(tb testing.TB) *Config {
	tb.Helper()

	return &Config{
		BaseURL: &url.URL{
			Scheme: "https",
			Host:   "hub.example.com",
			Path:   "/",
		},
		Bind: ":3000",
		Name: "WebSub",
	}
}
