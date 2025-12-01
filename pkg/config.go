package pkg

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	FilePaths []string
}

func ReadConfig() string {
	// Read User's Home Path
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		// Return error if home directory not found
		return "Error finding user's home directory; Please set $HOME" + err.Error()
	}
	// Read Config File
	// If the file does not exist; return error
	if _, err := os.Stat(userHomeDir + "/.config/blooview/blooview.toml"); os.IsNotExist(err) {
		return userHomeDir + "/.config/blooview/blooview.toml does not exist; Please create that config file;" + err.Error()
	}

	filesToMonitor := ReadToml(userHomeDir + "/.config/blooview/blooview.toml")
	fls := strings.Join(filesToMonitor, "\n")

	return fls
}

func ReadToml(path string) []string {
	var fPs Config

	// Decode toml
	_, err := toml.DecodeFile(path, &fPs)
	if err != nil {
		var errorMsg []string
		errorMsg = append(errorMsg, "Error decoding TOML config File"+err.Error())
		return errorMsg
	}

	// Get all file paths
	files := fPs.FilePaths

	return files
}

func GetConfigPath() string {
	// Read User's Home Path
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		// Return error if home directory not found
		return "Error finding user's home directory; Please set $HOME" + err.Error()
	}

	configPath := userHomeDir + "/.config/blooview/blooview.toml"

	return configPath
}
