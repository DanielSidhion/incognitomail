package incognitomail_test

import (
	"strings"
	"testing"

	"github.com/danielsidhion/incognitomail"
)

// Ensure an empty config returns an error.
func TestConfig_empty(t *testing.T) {
	incognitomail.ResetConfig()

	// Representing an empty config file as well.
	reader := strings.NewReader("")

	err := incognitomail.ReadConfigFromReader(reader)
	if err == nil {
		t.Fatal("expected error")
	}

	if err != incognitomail.ErrInvalidConfig {
		t.Fatal("expected ErrInvalidConfig")
	}
}

// Ensure that resetting the config actually changes all values to the default.
func TestConfig_reset(t *testing.T) {
	incognitomail.Config.General.MailSystem = "c0mpl3t3g4rb4g3"
	incognitomail.Config.General.UnixSockPath = "c0mpl3t3g4rb4g3"
	incognitomail.Config.General.LockFilePath = "c0mpl3t3g4rb4g3"
	incognitomail.Config.General.ListenPath = "c0mpl3t3g4rb4g3"
	incognitomail.Config.General.ListenAddress = "c0mpl3t3g4rb4g3"
	incognitomail.Config.General.TLSCertFile = "c0mpl3t3g4rb4g3"
	incognitomail.Config.General.TLSKeyFile = "c0mpl3t3g4rb4g3"
	incognitomail.Config.Persistence.Type = "c0mpl3t3g4rb4g3"
	incognitomail.Config.Persistence.DatabasePath = "c0mpl3t3g4rb4g3"
	incognitomail.Config.PostfixConfig.Domain = "c0mpl3t3g4rb4g3"
	incognitomail.Config.PostfixConfig.MapFilePath = "c0mpl3t3g4rb4g3"

	incognitomail.ResetConfig()

	if incognitomail.Config.General.MailSystem != "postfix" {
		t.Errorf("Config.General.MailSystem != \"%s\"", "postfix")
	}

	if incognitomail.Config.General.UnixSockPath != "/tmp/incognitomail.sock" {
		t.Errorf("Config.General.UnixSockPath != \"%s\"", "/tmp/incognitomail.sock")
	}

	if incognitomail.Config.General.LockFilePath != "/var/lock/incognitomail.lock" {
		t.Errorf("Config.General.LockFilePath != \"%s\"", "/var/lock/incognitomail.lock")
	}

	if incognitomail.Config.General.ListenPath != "/incognitomail" {
		t.Errorf("Config.General.ListenPath != \"%s\"", "/incognitomail")
	}

	if incognitomail.Config.General.ListenAddress != ":8080" {
		t.Errorf("Config.General.ListenAddress != \"%s\"", ":8080")
	}

	if incognitomail.Config.General.TLSCertFile != "" {
		t.Errorf("Config.General.TLSCertFile != \"%s\"", "")
	}

	if incognitomail.Config.General.TLSKeyFile != "" {
		t.Errorf("Config.General.TLSKeyFile != \"%s\"", "")
	}

	if incognitomail.Config.Persistence.Type != "boltdb" {
		t.Errorf("Config.Persistence.Type != \"%s\"", "boltdb")
	}

	if incognitomail.Config.Persistence.DatabasePath != "incognitomail.db" {
		t.Errorf("Config.Persistence.DatabasePath != \"%s\"", "incognitomail.db")
	}

	if incognitomail.Config.PostfixConfig.Domain != "" {
		t.Errorf("Config.PostfixConfig.Domain != \"%s\"", "")
	}

	if incognitomail.Config.PostfixConfig.MapFilePath != "" {
		t.Errorf("Config.PostfixConfig.MapFilePath != \"%s\"", "")
	}
}

// Ensures that a minimal config (one with only required values) doesn't return any errors.
func TestConfig_minimal(t *testing.T) {
	incognitomail.ResetConfig()

	err := incognitomail.ReadConfigFromFile("testdata/minimal_config")
	if err != nil {
		t.Fatal(err)
	}

	if incognitomail.Config.PostfixConfig.Domain != "@sidhion.com" {
		t.Errorf("Config.PostfixConfig.Domain != \"%s\"", "@sidhion.com")
	}

	if incognitomail.Config.PostfixConfig.MapFilePath != "/tmp/postfix/canonical" {
		t.Errorf("Config.PostfixConfig.MapFilePath != \"%s\"", "/tmp/postfix/canonical")
	}
}

// Ensures that a config with all specified values doesn't return any errors.
func TestConfig_full(t *testing.T) {
	incognitomail.ResetConfig()

	err := incognitomail.ReadConfigFromFile("testdata/full_config")
	if err != nil {
		t.Fatal(err)
	}

	if incognitomail.Config.General.MailSystem != "postfix" {
		t.Errorf("Config.General.MailSystem != \"%s\"", "postfix")
	}

	if incognitomail.Config.General.UnixSockPath != "/tmp/incognitomail/incognito.sock" {
		t.Errorf("Config.General.UnixSockPath != \"%s\"", "/tmp/incognitomail/incognito.sock")
	}

	if incognitomail.Config.General.LockFilePath != "/var/lock/incognitomail/incognito.lock" {
		t.Errorf("Config.General.LockFilePath != \"%s\"", "/var/lock/incognitomail/incognito.lock")
	}

	if incognitomail.Config.General.ListenPath != "/" {
		t.Errorf("Config.General.ListenPath != \"%s\"", "/")
	}

	if incognitomail.Config.General.ListenAddress != ":9090" {
		t.Errorf("Config.General.ListenAddress != \"%s\"", ":9090")
	}

	if incognitomail.Config.General.TLSCertFile != "server.pem" {
		t.Errorf("Config.General.TLSCertFile != \"%s\"", "server.pem")
	}

	if incognitomail.Config.General.TLSKeyFile != "server.key" {
		t.Errorf("Config.General.TLSKeyFile != \"%s\"", "server.key")
	}

	if incognitomail.Config.Persistence.Type != "boltdb" {
		t.Errorf("Config.Persistence.Type != \"%s\"", "boltdb")
	}

	if incognitomail.Config.Persistence.DatabasePath != "incognito.db" {
		t.Errorf("Config.Persistence.DatabasePath != \"%s\"", "incognito.db")
	}

	if incognitomail.Config.PostfixConfig.Domain != "@sidhion.com" {
		t.Errorf("Config.PostfixConfig.Domain != \"%s\"", "@sidhion.com")
	}

	if incognitomail.Config.PostfixConfig.MapFilePath != "/tmp/postfix/canonical" {
		t.Errorf("Config.PostfixConfig.MapFilePath != \"%s\"", "/tmp/postfix/canonical")
	}
}
