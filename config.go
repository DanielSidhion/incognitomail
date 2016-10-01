package incognitomail

import (
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
	// Config holds all global configuration. It is initially started with default values, and then overridden by reading values from a file.
	Config = config{
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
	}
)

// ReadConfigFromFile reads the file in the given path and parses all config data from it. Any value not defined in this configuration file will be kept as its default value.
func ReadConfigFromFile(path string) error {
	err := gcfg.ReadFileInto(&Config, path)
	if err != nil {
		return err
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
