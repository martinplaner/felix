package felix

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestFilterConfig_UnmarshalYAML(t *testing.T) {
	y := []byte(`
type: testtype
otherfield: 123
`)

	var fc FilterConfig
	err := yaml.Unmarshal(y, &fc)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	if fc.Type != "testtype" || fc.raw == nil {
		t.Error("did not unmarshal yaml properly")
	}
}

func TestFilterConfig_Unmarshal(t *testing.T) {
	fc := FilterConfig{
		Type: "testtype",
		raw: map[string]interface{}{
			"Type":       "testtype",
			"OtherField": 123,
			"NamedField": "value",
		},
	}

	aux := struct {
		Type         string
		OtherField   int
		RenamedField string `yaml:"namedField"`
	}{}

	err := fc.Unmarshal(&aux)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	if aux.Type != "testtype" || aux.OtherField != 123 || aux.RenamedField != "value" {
		t.Error("did not unmarshal FilterConfig properly")
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config.Port != DefaultPort {
		t.Error("invalid default config")
	}
}

func TestConfigFromFile(t *testing.T) {
	config, err := ConfigFromFile("../../config.example.yml")
	if err != nil {
		t.Error("unexpected error:", err)
	}

	if config.Port != DefaultPort {
		t.Error("invalid default config")
	}
}
