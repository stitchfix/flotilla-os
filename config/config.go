package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"strings"
)

//
// Config interface to wrap external configuration object
//
type Config interface {
	GetString(key string) string
	GetStringSlice(key string) []string
	GetStringMapString(key string) map[string]string
	GetInt(key string) int
	GetBool(key string) bool
	IsSet(key string) bool
}

//
// NewConfig initializes a configuration object
// - if confDir is non-nil searches there and loads a "config.yml"
// - sets configuration to read from environment variables automatically
//
func NewConfig(confDir *string) (Config, error) {
	v := viper.New()
	if v == nil {
		return &conf{}, errors.New("Error initializing internal config")
	}
	if confDir != nil {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(*confDir)
		if err := v.ReadInConfig(); err != nil {
			return &conf{}, errors.Wrapf(err, "problem reading config from [%s]", *confDir)
		}
	}
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return &conf{v}, nil
}

type conf struct {
	v *viper.Viper
}

func (c *conf) GetString(key string) string {
	return c.v.GetString(key)
}

func (c *conf) GetInt(key string) int {
	return c.v.GetInt(key)
}

func (c *conf) GetBool(key string) bool {
	return c.v.GetBool(key)
}

func (c *conf) GetStringMapString(key string) map[string]string {
	return c.v.GetStringMapString(key)
}

func (c *conf) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

func (c *conf) IsSet(key string) bool {
	return c.v.IsSet(key)
}
