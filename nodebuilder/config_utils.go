package nodebuilder

import (
	"io"

	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"

	"github.com/celestiaorg/celestia-node/libs/fslock"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
)

// ConfigLoader defines a function that loads a config from any source.
type ConfigLoader func() (*Config, error)

// SaveConfig saves Config 'cfg' under the given 'path'.
func SaveConfig(path string, cfg *Config) error {
	f, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return cfg.Encode(f)
}

// LoadConfig loads Config from the given 'path'.
func LoadConfig(path string) (*Config, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	return &cfg, cfg.Decode(f)
}

// RemoveConfig removes the Config from the given store path.
func RemoveConfig(path string) (err error) {
	path, err = storePath(path)
	if err != nil {
		return
	}

	flock, err := fslock.Lock(lockPath(path))
	if err != nil {
		if err == fslock.ErrLocked {
			err = ErrOpened
		}
		return
	}
	defer flock.Unlock() //nolint: errcheck

	return removeConfig(configPath(path))
}

// removeConfig removes Config from the given 'path'.
func removeConfig(path string) error {
	return fs.Remove(path)
}

// UpdateConfig loads the node's config and applies new values
// from the default config of the given node type, saving the
// newly updated config into the node's config path.
func UpdateConfig(tp node.Type, path string) (err error) {
	path, err = storePath(path)
	if err != nil {
		return err
	}

	flock, err := fslock.Lock(lockPath(path))
	if err != nil {
		if err == fslock.ErrLocked {
			err = ErrOpened
		}
		return err
	}
	defer flock.Unlock() //nolint: errcheck

	newCfg := DefaultConfig(tp)

	cfgPath := configPath(path)
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	cfg, err = MergeConfig(cfg, newCfg)
	if err != nil {
		return err
	}

	// save the updated config
	err = removeConfig(cfgPath)
	if err != nil {
		return err
	}
	return SaveConfig(cfgPath, cfg)
}

// MergeConfig merges new values from the new config into the old
// config, returning the updated old config.
func MergeConfig(oldCfg *Config, newCfg *Config) (*Config, error) {
	err := mergo.Merge(oldCfg, newCfg, mergo.WithOverrideEmptySlice)
	return oldCfg, err
}

// TODO(@Wondertan): We should have a description for each field written into w,
// 	so users can instantly understand purpose of each field. Ideally, we should have a utility
// program to parse comments 	from actual sources(*.go files) and generate docs from comments.

// Hint: use 'ast' package.
// Encode encodes a given Config into w.
func (cfg *Config) Encode(w io.Writer) error {
	return toml.NewEncoder(w).Encode(cfg)
}

// Decode decodes a Config from a given reader r.
func (cfg *Config) Decode(r io.Reader) error {
	_, err := toml.NewDecoder(r).Decode(cfg)
	return err
}
