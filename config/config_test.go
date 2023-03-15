package config

import (
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	require.Nil(t, LoadConfigFile("./config_example.toml"))
	log.Printf("%+v", GlobalConfig)
}
