package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluetuith-org/bluetooth-classic/api/bluetooth"
	"github.com/knadh/koanf/parsers/hjson"
	"github.com/knadh/koanf/providers/cliflagv2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v2"
)

const (
	configFile    = "bluetuith.conf"
	oldConfigFile = "config"
)

// Config describes the configuration for the app.
type Config struct {
	path string

	Values Values
}

// NewConfig returns a new configuration.
func NewConfig() *Config {
	return &Config{}
}

// Load loads the configuration from the configuration file and the command-line flags.
func (c *Config) Load(k *koanf.Koanf, cliCtx *cli.Context) error {
	if err := c.createConfigDir(); err != nil {
		return err
	}

	cfgfile, err := c.FilePath(configFile)
	if err != nil {
		return err
	}

	if err := k.Load(file.Provider(cfgfile), hjson.Parser()); err != nil {
		return err
	}

	if err := k.Load(cliflagv2.Provider(cliCtx, "."), nil); err != nil {
		return err
	}

	return k.UnmarshalWithConf("", &c.Values, koanf.UnmarshalConf{Tag: "koanf"})
}

// ValidateValues validates the configuration values.
func (c *Config) ValidateValues() error {
	return c.Values.validateValues()
}

// ValidateSessionValues validates all configuration values that require a bluetooth session.
func (c *Config) ValidateSessionValues(session bluetooth.Session) error {
	return c.Values.validateSessionValues(session)
}

// createConfigDir checks for and/or creates a configuration directory.
func (c *Config) createConfigDir() error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	type configDir struct {
		path, fullpath               string
		exist, hidden, prefixHomeDir bool
	}

	configPaths := []*configDir{
		{path: os.Getenv("XDG_CONFIG_HOME")},
		{path: ".config", prefixHomeDir: true},
		{path: ".", hidden: true, prefixHomeDir: true},
	}

	for _, dir := range configPaths {
		name := "bluetuith"

		if dir.path == "" {
			continue
		}

		if dir.hidden {
			name = "." + name
		}

		if dir.prefixHomeDir {
			dir.path = filepath.Join(homedir, dir.path)
		}

		if _, err := os.Stat(filepath.Clean(dir.path)); err == nil {
			dir.exist = true
		}

		dir.fullpath = filepath.Join(dir.path, name)
		if _, err := os.Stat(filepath.Clean(dir.fullpath)); err == nil {
			c.path = dir.fullpath
			break
		}
	}

	if c.path == "" {
		var pathErrors []string

		for _, dir := range configPaths {
			if err := os.Mkdir(dir.fullpath, os.ModePerm); err == nil {
				c.path = dir.fullpath
				break
			}

			pathErrors = append(pathErrors, dir.fullpath)
		}

		if len(pathErrors) == len(configPaths) {
			return fmt.Errorf("the configuration directories could not be created at %s%s", "\n", strings.Join(pathErrors, "\n"))
		}
	}

	return nil
}

// FilePath returns the absolute path for the given configuration file.
func (c *Config) FilePath(configFile string) (string, error) {
	confPath := filepath.Join(c.path, configFile)

	if _, err := os.Stat(confPath); err != nil {
		fd, err := os.Create(confPath)
		fd.Close()
		if err != nil {
			return "", fmt.Errorf("Cannot create "+configFile+" file at %s", confPath)
		}
	}

	return confPath, nil
}

// GenerateAndSave generates and updates the configuration.
// Any existing values are appended to it.
func (c *Config) GenerateAndSave(currentCfg *koanf.Koanf) (bool, error) {
	var parsedOldCfg bool

	cfg, err := c.parseOldConfig(currentCfg)
	if err == nil {
		parsedOldCfg = true
	}

	data, err := hjson.Parser().Marshal(cfg.All())
	if err != nil {
		return parsedOldCfg, err
	}

	conf, err := c.FilePath(configFile)
	if err != nil {
		return parsedOldCfg, err
	}

	f, err := os.OpenFile(conf, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return parsedOldCfg, err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return parsedOldCfg, err
	}

	if err := f.Sync(); err != nil {
		return parsedOldCfg, err
	}

	return parsedOldCfg, nil
}

// parseOldConfig parses and stores values from the old configuration.
func (c *Config) parseOldConfig(currentCfg *koanf.Koanf) (*koanf.Koanf, error) {
	f, err := c.FilePath(oldConfigFile)
	if err != nil {
		return currentCfg, nil
	}

	fd, err := os.OpenFile(f, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return currentCfg, errors.New("the old configuration could not be read")
	}

	k := koanf.New(".")
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		values := strings.Split(line, "=")
		if len(values) != 2 {
			continue
		}

		k.Set(values[0], values[1])
	}

	fd.Close()

	if err = scanner.Err(); err != nil && err != io.EOF {
		return currentCfg, errors.New("the old configuration could not be parsed")
	}

	if err := k.Merge(currentCfg); err != nil {
		return currentCfg, err
	}

	return k, nil
}
