// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"io/ioutil"

	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Default values for the configuration
const (
	DefaultFeedFetchInterval = 65 * time.Minute
	DefaultUserAgent         = "felix"
	DefaultCleanupInterval   = 1 * time.Hour
	DefaultCleanupMaxAge     = 24 * time.Hour
)

// Config contains the configuration
type Config struct {
	FetchInterval   time.Duration  `yaml:"fetchInterval"`
	UserAgent       string         `yaml:"userAgent"`
	CleanupInterval time.Duration  `yaml:"cleanupInterval"`
	CleanupMaxAge   time.Duration  `yaml:"cleanupMaxAge"`
	Feeds           []FeedConfig   `yaml:"feeds"`
	ItemFilters     []FilterConfig `yaml:"itemFilters"`
	LinkFilters     []FilterConfig `yaml:"linkFilters"`
}

var emptyConfig = Config{}

// FeedConfig contains the configuration of a single feed.
type FeedConfig struct {
	Type          string
	URL           string
	FetchInterval time.Duration
}

// FilterConfig is the common configuration of all filter types.
// See FilterConfig.Unmarshal for unmarshaling of the raw config value for more specific types.
type FilterConfig struct {
	Type string
	raw  map[string]interface{}
}

// UnmarshalYAML is a custom YAML unmarshal handler to handle the common filter config elements.
// See https://godoc.org/gopkg.in/yaml.v2#Unmarshaler.
func (fc *FilterConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	aux := &struct {
		Type string
	}{}
	if err := unmarshal(aux); err != nil {
		return err
	}

	raw := map[string]interface{}{}
	if err := unmarshal(raw); err != nil {
		return err
	}

	fc.Type = aux.Type
	fc.raw = raw
	return nil
}

// Unmarshal decodes the raw config values for a more specific config type.
func (fc *FilterConfig) Unmarshal(v interface{}) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "yaml", // Re-use 'yaml' field tags instead of 'mapstructure'
		Result:  v,
	})

	if err != nil {
		return err
	}

	return dec.Decode(fc.raw)
}

// ItemTitleFilterConfig contains the configuration of a ItemTitleFilter.
type ItemTitleFilterConfig struct {
	Type   string
	Titles []string
	// TODO: allow grouping of title and alternative title? -> [][]string
}

// LinkDomainFilterConfig contains the configuration of a LinkDomainFilter.
type LinkDomainFilterConfig struct {
	Domains []string
}

// LinkURLRegexFilterConfig contains the configuration of a LinkURLRegexFilter.
type LinkURLRegexFilterConfig struct {
	Exprs []string
}

// NewConfig returns a new configuration with default values.
func NewConfig() Config {
	// TODO: return value or pointer?
	return Config{
		UserAgent:       DefaultUserAgent,
		FetchInterval:   DefaultFeedFetchInterval,
		CleanupInterval: DefaultCleanupInterval,
		CleanupMaxAge:   DefaultCleanupMaxAge,
	}
}

// ConfigFromFile returns a configuration parsed from the given file.
func ConfigFromFile(filename string) (Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return emptyConfig, err
	}

	config := NewConfig()
	if err := yaml.Unmarshal(b, &config); err != nil {
		return emptyConfig, errors.Wrap(err, "could not unmarshal config")
	}

	return config, nil
}
