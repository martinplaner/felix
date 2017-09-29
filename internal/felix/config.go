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

const (
	DefaultFetchInterval   = 65 * time.Minute
	DefaultUserAgent       = "felix"
	DefaultCleanupInterval = 1 * time.Hour
	DefaultCleanupMaxAge   = 24 * time.Hour
)

type Config struct {
	FetchInterval   time.Duration  `yaml:"fetchInterval"`
	UserAgent       string         `yaml:"userAgent"`
	CleanupInterval time.Duration  `yaml:"cleanupInterval"`
	CleanupMaxAge   time.Duration  `yaml:"cleanupMaxAge"`
	Feeds           []FeedConfig   `yaml:"feeds"`
	ItemFilters     []FilterConfig `yaml:"itemFilters"`
	LinkFilters     []FilterConfig `yaml:"linkFilters"`
}

type FeedConfig struct {
	Type          string
	URL           string
	FetchInterval time.Duration
}

type FilterConfig struct {
	Type string
	raw  map[string]interface{}
}

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

type TitleItemFilterConfig struct {
	Type   string
	Titles []string
	// TODO: allow grouping of title and alternative title? -> [][]string
}

type LinkDomainFilterConfig struct {
	Domains []string
}

type LinkURLRegexFilterConfig struct {
	Exprs []string
}

func NewConfig() *Config {
	return &Config{
		UserAgent:       DefaultUserAgent,
		FetchInterval:   DefaultFetchInterval,
		CleanupInterval: DefaultCleanupInterval,
		CleanupMaxAge:   DefaultCleanupMaxAge,
	}
}

func ConfigFromFile(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := NewConfig()
	yaml.Unmarshal(b, config)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal config")
	}

	return config, nil
}
