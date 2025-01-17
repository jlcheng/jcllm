package defaultconfig

import (
	"github.com/go-errors/errors"
	"github.com/jlcheng/jcllm/configuration"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

// New returns a Configuration which looks for configurations with the following precedence:
//
//  1. Command-line arguments.
//  2. Environment variables.
//  3. Entries from a configuration file.
func New(stringConfigs []configuration.Metadata, boolConfigs []configuration.Metadata) (configuration.Configuration, error) {
	k := koanf.New(".")

	configFile := findConfigFile()
	if configFile != "" {
		if err := k.Load(file.Provider(configFile), toml.Parser()); err != nil {
			return nil, errors.Errorf("config file error (%s): %s", configFile, err)
		}
	}

	if err := k.Load(env.Provider("JCLLM_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "JCLLM_")), "_", "-", -1)
	}), nil); err != nil {
		return nil, err
	}

	f := flag.NewFlagSet("config", flag.ContinueOnError)
	for _, meta := range stringConfigs {
		f.String(meta.Name, meta.DefaultValue, meta.Usage)
	}
	for _, meta := range boolConfigs {
		f.Bool(meta.Name, false, meta.Usage)
	}

	if err := f.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, configuration.ErrHelp
		}
	}

	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	return k, nil
}

// findConfigFile finds the config file by recursively searching up from the cwd
// and if not found checks in ~/.jcllm.d/jcllm.toml
func findConfigFile() string {
	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	var configPath string
	for {
		configPath = filepath.Join(currentDir, ".jcllm.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break // We have reached the root
		}
		currentDir = parentDir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath = filepath.Join(homeDir, ".jcllm.d", "jcllm.toml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}
	return ""
}

var _ configuration.ConfigProvider = New
