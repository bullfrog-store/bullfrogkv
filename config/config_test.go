package config

import (
	"log"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	config := LoadConfigFile("./config_example.toml")
	log.Printf("%+v", config)
}
