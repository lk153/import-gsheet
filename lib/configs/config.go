//go:generate mockery --name=Config --output=../mocks/config --with-expecter

package configs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var c *viperConfig

type Config interface {
	GetString(key string) string
}

type viperConfig struct {
	filename string
	v        *viper.Viper
}

// New creates and returns a new config instance
// If there are any yaml file configurations, these files will be loaded in as environment variables
func New() Config {
	c = &viperConfig{v: viper.New()}

	args := os.Args[1:]
	if len(args) > 0 {
		configFile := args[0]
		if strings.HasSuffix(configFile, ".yaml") {
			c.filename = configFile
			if _, err := os.Stat(configFile); err == nil {
				configType := filepath.Ext(configFile)
				configName := strings.TrimSuffix(configFile, configType)

				c.v.SetConfigName(configName)
				c.v.SetConfigType(configType[1:])
				c.v.AddConfigPath(".")

				err := c.v.ReadInConfig()
				if err != nil {
					log.Error().Err(err).Msg("Error reading config file")
				}

				log.Info().Msgf("Reading config file: %s", configFile)
			}
		}
	}

	// environment variables have priority and will
	// override the config file variables loaded above
	c.v.AutomaticEnv()

	return c
}

func GetConfig() Config {
	return c
}

// GetString returns the env var value of the key parameter
// If no value exists, an empty string will be returned
func (c *viperConfig) GetString(key string) string {
	return c.v.GetString(key)
}

// IsSet returns true if a env var with the key parameter is set, false if it is not
func (c *viperConfig) IsSet(key string) bool {
	return c.v.IsSet(key)
}
