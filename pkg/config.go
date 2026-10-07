package pkg

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ExecveFilter []string `toml:"ExecveFilter"`
	Filepaths    []string `toml:"Filepaths"`
}

// Returns list of filepaths to monitor
func GetFilepaths(filePath string) ([]string, error) {
	var cfg Config
	_, err := toml.DecodeFile(filePath, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg.Filepaths, nil
}

// Return a single string of filepaths
func GetFilePathsSingle(filePaths []string) string {
	fPs := strings.Join(filePaths, "\n")

	return fPs
}

// Returns list of filtered commands
func GetExecveFilter(filePath string) ([]string, error) {
	var conf Config

	_, err := toml.DecodeFile(filePath, &conf)
	if err != nil {
		return nil, err
	}

	return conf.ExecveFilter, nil
}

func GetConfigPath() string {
	// Read User's Home Path
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		// Return error if home directory not found
		return "Error finding user's home directory; Please set $HOME" + err.Error()
	}

	if _, err := os.Stat(userHomeDir + "/.config/blooview/blooview.toml"); os.IsNotExist(err) {
		return userHomeDir + "/.config/blooview/blooview.toml does not exist; Please create that config file;" + err.Error()
	}

	configPath := userHomeDir + "/.config/blooview/blooview.toml"

	return configPath
}
