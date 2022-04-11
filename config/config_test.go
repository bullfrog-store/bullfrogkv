package config

import (
	"log"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	LoadConfigFile("./config_example.toml")
	log.Printf("%+v", GlobalConfig)
}
