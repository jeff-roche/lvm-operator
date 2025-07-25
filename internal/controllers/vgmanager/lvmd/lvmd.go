package lvmd

import (
	"context"
	"fmt"
	"io"
	"os"

	lvmdCMD "github.com/topolvm/topolvm/cmd/lvmd/app"
	lvmd "github.com/topolvm/topolvm/pkg/lvmd/types"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

type Config = lvmdCMD.Config

type DeviceClass = lvmd.DeviceClass
type ThinPoolConfig = lvmd.ThinPoolConfig

var (
	TypeThin  = lvmd.TypeThin
	TypeThick = lvmd.TypeThick
)

const (
	DefaultFileConfigDir     = "/etc/topolvm"
	MicroShiftFileConfigDir  = "/var/lib/microshift/lvms"
	DefaultFileConfigPath    = DefaultFileConfigDir + "/lvmd.yaml"
	MicroShiftFileConfigPath = MicroShiftFileConfigDir + "/lvmd.yaml"
	maxReadLength            = 2 * 1 << 20 // 2MB
)

func DeepCopyConfig(c *Config) *Config {
	if c == nil {
		return nil
	}

	conf := &Config{
		SocketName: c.SocketName,
	}

	for _, dc := range c.DeviceClasses {
		newDc := &DeviceClass{
			Name:            dc.Name,
			VolumeGroup:     dc.VolumeGroup,
			Default:         dc.Default,
			StripeSize:      dc.StripeSize,
			LVCreateOptions: dc.LVCreateOptions,
			Type:            dc.Type,
		}
		if dc.SpareGB != nil {
			newDc.SpareGB = ptr.To(*dc.SpareGB)
		}
		if dc.Stripe != nil {
			newDc.Stripe = ptr.To(*dc.Stripe)
		}
		if dc.ThinPoolConfig != nil {
			newDc.ThinPoolConfig = &ThinPoolConfig{
				Name:               dc.ThinPoolConfig.Name,
				OverprovisionRatio: dc.ThinPoolConfig.OverprovisionRatio,
			}
		}
		conf.DeviceClasses = append(conf.DeviceClasses, newDc)
	}

	for _, co := range c.LvcreateOptionClasses {
		opt := &lvmd.LvcreateOptionClass{
			Name:    co.Name,
			Options: co.Options,
		}
		conf.LvcreateOptionClasses = append(conf.LvcreateOptionClasses, opt)
	}

	return conf
}

func DefaultConfigurator() *CachedFileConfig {
	return NewFileConfigurator(DefaultFileConfigPath)
}

func NewFileConfigurator(path string) *CachedFileConfig {
	return &CachedFileConfig{
		FileConfig: FileConfig{path: path},
	}
}

type Configurator interface {
	Load(ctx context.Context) (*Config, error)
	Save(ctx context.Context, config *Config) error
	Delete(ctx context.Context) error
}

type CachedFileConfig struct {
	*Config
	FileConfig
}

func (c *CachedFileConfig) Load(ctx context.Context) (*Config, error) {
	if c.Config != nil {
		return c.Config, nil
	}
	log.FromContext(ctx).Info("lvmd config not found in cache, loading from store")
	conf, err := c.FileConfig.Load(ctx)
	if err != nil {
		return nil, err
	}
	c.Config = conf
	return conf, nil
}

func (c *CachedFileConfig) Save(ctx context.Context, config *Config) error {
	c.Config = config
	log.FromContext(ctx).Info("saving lvmd config to cache and store")
	return c.FileConfig.Save(ctx, config)
}

type FileConfig struct {
	path string
}

func (c FileConfig) Load(ctx context.Context) (*Config, error) {
	file, err := os.Open(c.path)
	if os.IsNotExist(err) {
		// If the file does not exist, return nil for both
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", c.path, err)
	}

	defer func() {
		_ = file.Close()
	}()

	limitedReader := &io.LimitedReader{R: file, N: maxReadLength}
	cfgBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", c.path, err)
	}

	if limitedReader.N <= 0 {
		return nil, fmt.Errorf("the read limit is reached for config file %s", c.path)
	}

	config := &Config{}
	if err = yaml.Unmarshal(cfgBytes, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %s: %w", c.path, err)
	}
	return config, nil
}

func (c FileConfig) Save(ctx context.Context, config *Config) error {
	out, err := yaml.Marshal(config)
	if err == nil {
		err = os.WriteFile(c.path, out, 0600)
	}
	if err != nil {
		return fmt.Errorf("failed to save config file %s: %w", c.path, err)
	}
	return nil
}

func (c FileConfig) Delete(ctx context.Context) error {
	err := os.Remove(c.path)
	if err != nil {
		return fmt.Errorf("failed to delete config file %s: %w", c.path, err)
	}
	return err
}
