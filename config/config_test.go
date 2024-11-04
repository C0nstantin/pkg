package config

import (
	"os"
	"testing"
)

type testConfigStruct struct {
	AuthKey1    string `env-required:"true" yaml:"auth_key" env:"AUTH_KEY"`
	AuthRID     string `env-required:"true" yaml:"auth_rid" env:"AUTH_RID"`
	SecondValue string `env-required:"true" yaml:"second-value" env:"SECOND"`
	User        string `yaml:"userr" env:"USERR" env-default:"user"`
	LocalSecret string `yaml:"local_secret" env:"LOCAL_SECRET" env-default:"local_secret"`
	ImapHost    string `yaml:"imap_host" env:"IMAP_HOST" env-default:"imap_host"`
}

// 1. test default values
func TestLoadConfig(t *testing.T) {
	os.Setenv("AUTH_KEY", "auth_key")
	os.Setenv("AUTH_RID", "auth_rid")
	os.Setenv("SECOND", "second-value")

	cfg := testConfigStruct{}
	err := LoadConfig(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.AuthKey1 != "auth_key" {
		t.Errorf("expected auth_key, got %s", cfg.AuthKey1)
	}
	if cfg.AuthRID != "auth_rid" {
		t.Errorf("expected auth_rid, got %s", cfg.AuthRID)
	}
	if cfg.SecondValue != "second-value" {
		t.Errorf("expected second-value, got %s", cfg.SecondValue)
	}
	if cfg.User != "user" {
		t.Errorf("expected user, got %s", cfg.User)
	}
	if cfg.LocalSecret != "local_secret" {
		t.Errorf("expected local_secret, got %s", cfg.LocalSecret)
	}
	if cfg.ImapHost != "imap_host" {
		t.Errorf("expected imap_host, got %s", cfg.ImapHost)
	}
}

// 2. test load yaml from file
// 3. test load env variables
// 4. test override configuration env over yaml
