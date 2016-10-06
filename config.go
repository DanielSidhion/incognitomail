package incognitomail

import (
	"errors"
	"io"
	"os"

	"gopkg.in/gcfg.v1"
)

type generalConfig struct {
	MailSystem    string
	UnixSockPath  string
	LockFilePath  string
	ListenPath    string
	ListenAddress string
	TLSCertFile   string
	TLSKeyFile    string
}

type persistenceConfig struct {
	Type         string
	DatabasePath string
}

type postfixConfig struct {
	Domain      string
	MapFilePath string
}

type config struct {
	General       generalConfig
	Persistence   persistenceConfig
	PostfixConfig postfixConfig
}

var (
	defaultConfig = config{
		General: generalConfig{
			MailSystem:    "postfix",
			UnixSockPath:  "/tmp/incognitomail.sock",
			LockFilePath:  "/var/lock/incognitomail.lock",
			ListenPath:    "/incognitomail",
			ListenAddress: ":8080",
			TLSCertFile:   "",
			TLSKeyFile:    "",
		},
		Persistence: persistenceConfig{
			Type:         "boltdb",
			DatabasePath: "incognitomail.db",
		},
		PostfixConfig: postfixConfig{
			Domain: "",
			MapFilePath: "",
		},
	}

	// Config holds all global configuration.
	Config = defaultConfig

	// ErrInvalidConfig is used when loading a configuration with invalid values.
	ErrInvalidConfig = errors.New("invalid configuration values")
)

// ResetConfig switches all values back to the default.
func ResetConfig() {
	Config = defaultConfig
}

// ReadConfigFromFile reads the file in the given path and parses all config data from it. Any value not defined in this configuration file will be kept as its default value.
func ReadConfigFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = ReadConfigFromReader(f)
	return err
}

// ReadConfigFromReader parses all config data from the given reader. Any value not defined in the read string will be kept as its default value.
func ReadConfigFromReader(reader io.Reader) error {
	err := gcfg.ReadInto(&Config, reader)
	if err != nil {
		return err
	}

	if !ValidConfig() {
		return ErrInvalidConfig
	}

	return nil
}

// ValidConfig returns true if the current Config is valid, i.e. not likely to crash the server.
func ValidConfig() bool {
	invalid := false

	invalid = invalid || Config.General.MailSystem != "postfix"
	invalid = invalid || Config.General.UnixSockPath == ""
	invalid = invalid || Config.General.LockFilePath == ""
	invalid = invalid || Config.General.ListenPath == ""
	invalid = invalid || Config.General.ListenAddress == ""
	invalid = invalid || Config.Persistence.Type != "boltdb"
	invalid = invalid || Config.Persistence.DatabasePath == ""

	if Config.General.MailSystem == "postfix" {
		invalid = invalid || Config.PostfixConfig.Domain == ""
		invalid = invalid || Config.PostfixConfig.MapFilePath == ""
	}

	return !invalid
}
