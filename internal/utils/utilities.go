package utils

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/roshanlc/soe-backend/internal/data"
)

// ReadConfig reads toml config files
// It returns *data.Config and an error
func ReadConfig(filepath string) (*data.Config, error) {

	_, err := os.Stat(filepath)

	if err != nil {
		log.Fatal("Config file is missing: ", filepath)
		return nil, err
	}

	cfg := data.Config{}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Unable to read the config file: ", filepath)
		return nil, err
	}

	if err := toml.Unmarshal(content, &cfg); err != nil {
		log.Fatal(err)
		return nil, err

	}

	return &cfg, nil
}
