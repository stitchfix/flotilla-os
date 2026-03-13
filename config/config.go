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
	GetFloat64(key string) float64
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

// GetString returns the value associated with the key as a string.
func (c *conf) GetString(key string) string {
	return c.v.GetString(key)
}

// GetFloat returns the value associated with the key as a float.
func (c *conf) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

// GetInt returns the value associated with the key as an integer.
func (c *conf) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool returns the value associated with the key as a boolean.
func (c *conf) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *conf) GetStringMapString(key string) map[string]string {
	return c.v.GetStringMapString(key)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *conf) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
func (c *conf) IsSet(key string) bool {
	return c.v.IsSet(key)
}

// GetPriorityClassesEnabled returns whether priority class feature is enabled
func (c *conf) GetPriorityClassesEnabled() bool {
	return c.GetBool("eks_priority_classes_enabled")
}

// GetPriorityClassValidationEnabled returns whether to validate priority classes at startup
func (c *conf) GetPriorityClassValidationEnabled() bool {
	return c.GetBool("eks_priority_class_validation_enabled")
}

// GetPriorityClassForTier returns the priority class name for a given tier
func (c *conf) GetPriorityClassForTier(tier string) string {
	mapping := c.GetStringMapString("eks_priority_classes")
	if priorityClass, ok := mapping[tier]; ok {
		return priorityClass
	}
	return c.GetString("eks_priority_class_default")
}

// GetDefaultPriorityClass returns the default priority class
func (c *conf) GetDefaultPriorityClass() string {
	return c.GetString("eks_priority_class_default")
}

// GetAllConfiguredPriorityClasses returns all configured priority class names
func (c *conf) GetAllConfiguredPriorityClasses() []string {
	mapping := c.GetStringMapString("eks_priority_classes")
	classes := make([]string, 0, len(mapping))
	for _, class := range mapping {
		classes = append(classes, class)
	}
	defaultClass := c.GetDefaultPriorityClass()
	if defaultClass != "" {
		classes = append(classes, defaultClass)
	}
	return classes
}
