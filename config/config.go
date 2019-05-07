package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/user"

	yaml "gopkg.in/yaml.v2"

	"github.com/spf13/viper"
)

// ClusterName is the name of an ECS Cluster (example: "mountain")
type ClusterName string

// FilePath is the path to a yaml file on disk
type FilePath string

// Keys is a map of cluster names
type Keys map[ClusterName]FilePath

// Config represents global application configuration
type Config struct {
	Keys Keys `yaml:"keys"`
}

// Storage is a mechanism for storing ECS Commander Config!
type Storage interface {
	ReadKeys() (Keys, error)
	SaveKeys(Keys) error
	IsModified() bool
}

// ReadKeys returns you the existing keys if they exist,
// if not, will return a new, blank, Keys struct
func ReadKeys(adapter Storage) (Keys, error) {
	return adapter.ReadKeys()
}

// YAMLFile stores config in a YAML file on disk,
// this is the default.  It implements the storage
// adapter interface.
type YAMLFile struct {
	path     string
	content  []byte
	modified bool
}

// GetYAMLConfig gets, or creates, a yaml configuration
// file from disk
func GetYAMLConfig() *YAMLFile {
	file := viper.ConfigFileUsed()
	configFile := NewYAMLFile()
	if file != "" {
		configFile = ReadYAMLFile(file)
	}
	return configFile
}

// NewYAMLFile creates a new and empty YAML file
func NewYAMLFile() *YAMLFile {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return &YAMLFile{path: fmt.Sprintf("%s/.ecsy.yaml", usr.HomeDir)}
}

// ReadYAMLFile Sets the file to be used for storing config
func ReadYAMLFile(path string) *YAMLFile {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("Tried to read config file from %s, but failed. %v", content, err))
	}
	yamlFile := &YAMLFile{path: path, content: content}
	return yamlFile
}

// ReadKeys implements a key reader for yaml files
func (yamlFile *YAMLFile) ReadKeys() (Keys, error) {
	config := &Config{}
	err := yaml.Unmarshal(yamlFile.content, &config)
	if err != nil {
		return nil, err
	}
	if config.Keys == nil {
		config.Keys = make(Keys)
	}
	return config.Keys, nil
}

// SaveKeys implements a key writer for yaml files
func (yamlFile *YAMLFile) SaveKeys(keys Keys) error {
	config := &Config{}
	err := yaml.Unmarshal(yamlFile.content, &config)
	if err != nil {
		return err
	}
	config.Keys = keys
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(yamlFile.path, bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// IsModified implements a modified
func (yamlFile *YAMLFile) IsModified() bool {
	return true
}

// GetClusterKey returns the registered key path for a cluster
func GetClusterKey(cluster string) string {
	out := fmt.Sprintf("~/%s.pem", cluster)
	allKeys, err := GetYAMLConfig().ReadKeys()
	if err != nil {
		return out
	}
	if key, ok := allKeys[ClusterName(cluster)]; ok {
		out = string(key)
	}
	return out
}
