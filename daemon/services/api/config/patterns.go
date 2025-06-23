package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/validation"
)

// ConfigPattern defines a standardized configuration pattern
type ConfigPattern interface {
	Load() error
	Save() error
	Validate() error
	GetDefaults() interface{}
	Reset() error
}

// BaseConfig provides common configuration functionality
type BaseConfig struct {
	ConfigPath string `json:"-"`
	Version    string `json:"version"`
	UpdatedAt  string `json:"updated_at"`
}

// ConfigManager manages configuration patterns across services
type ConfigManager struct {
	configs map[string]ConfigPattern
	baseDir string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(baseDir string) *ConfigManager {
	return &ConfigManager{
		configs: make(map[string]ConfigPattern),
		baseDir: baseDir,
	}
}

// RegisterConfig registers a configuration pattern
func (cm *ConfigManager) RegisterConfig(name string, config ConfigPattern) error {
	if name == "" {
		return fmt.Errorf("config name cannot be empty")
	}
	
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	cm.configs[name] = config
	logger.Blue("Registered configuration pattern: %s", name)
	
	return nil
}

// LoadConfig loads a specific configuration
func (cm *ConfigManager) LoadConfig(name string) error {
	config, exists := cm.configs[name]
	if !exists {
		return fmt.Errorf("configuration '%s' not found", name)
	}
	
	return config.Load()
}

// SaveConfig saves a specific configuration
func (cm *ConfigManager) SaveConfig(name string) error {
	config, exists := cm.configs[name]
	if !exists {
		return fmt.Errorf("configuration '%s' not found", name)
	}
	
	return config.Save()
}

// LoadAllConfigs loads all registered configurations
func (cm *ConfigManager) LoadAllConfigs() error {
	var errors []string
	
	for name, config := range cm.configs {
		if err := config.Load(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			logger.Red("Failed to load config %s: %v", name, err)
		} else {
			logger.Green("Loaded config: %s", name)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to load configurations: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// ValidateAllConfigs validates all registered configurations
func (cm *ConfigManager) ValidateAllConfigs() error {
	var errors []string
	
	for name, config := range cm.configs {
		if err := config.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// GetConfig returns a specific configuration
func (cm *ConfigManager) GetConfig(name string) (ConfigPattern, error) {
	config, exists := cm.configs[name]
	if !exists {
		return nil, fmt.Errorf("configuration '%s' not found", name)
	}
	
	return config, nil
}

// JSONConfig implements ConfigPattern for JSON-based configurations
type JSONConfig struct {
	BaseConfig
	data interface{}
}

// NewJSONConfig creates a new JSON configuration
func NewJSONConfig(configPath string, data interface{}) *JSONConfig {
	return &JSONConfig{
		BaseConfig: BaseConfig{
			ConfigPath: configPath,
			Version:    "1.0",
		},
		data: data,
	}
}

// Load loads the JSON configuration from file
func (jc *JSONConfig) Load() error {
	if !fileExists(jc.ConfigPath) {
		// Create default configuration if file doesn't exist
		logger.Yellow("Config file not found, creating default: %s", jc.ConfigPath)
		return jc.createDefaultConfig()
	}
	
	file, err := os.Open(jc.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(jc.data); err != nil {
		return fmt.Errorf("failed to decode JSON config: %v", err)
	}
	
	// Update metadata
	jc.UpdatedAt = time.Now().Format(time.RFC3339)
	
	logger.Blue("Loaded JSON config: %s", jc.ConfigPath)
	return nil
}

// Save saves the JSON configuration to file
func (jc *JSONConfig) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(jc.ConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	
	// Update metadata
	jc.UpdatedAt = time.Now().Format(time.RFC3339)
	
	file, err := os.Create(jc.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jc.data); err != nil {
		return fmt.Errorf("failed to encode JSON config: %v", err)
	}
	
	logger.Blue("Saved JSON config: %s", jc.ConfigPath)
	return nil
}

// Validate validates the JSON configuration
func (jc *JSONConfig) Validate() error {
	if jc.data == nil {
		return fmt.Errorf("configuration data is nil")
	}
	
	// Use validation framework if the data implements validation tags
	result := validation.ValidateStruct(jc.data)
	if !result.Valid {
		var errors []string
		for _, err := range result.Errors {
			errors = append(errors, err.Message)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// GetDefaults returns the default configuration
func (jc *JSONConfig) GetDefaults() interface{} {
	// Create a new instance of the same type with default values
	dataType := reflect.TypeOf(jc.data)
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}
	
	defaultValue := reflect.New(dataType).Interface()
	return defaultValue
}

// Reset resets the configuration to defaults
func (jc *JSONConfig) Reset() error {
	jc.data = jc.GetDefaults()
	return jc.Save()
}

// createDefaultConfig creates a default configuration file
func (jc *JSONConfig) createDefaultConfig() error {
	jc.data = jc.GetDefaults()
	return jc.Save()
}

// EnvironmentConfig implements ConfigPattern for environment-based configurations
type EnvironmentConfig struct {
	BaseConfig
	prefix string
	data   interface{}
}

// NewEnvironmentConfig creates a new environment configuration
func NewEnvironmentConfig(prefix string, data interface{}) *EnvironmentConfig {
	return &EnvironmentConfig{
		BaseConfig: BaseConfig{
			Version: "1.0",
		},
		prefix: prefix,
		data:   data,
	}
}

// Load loads configuration from environment variables
func (ec *EnvironmentConfig) Load() error {
	return ec.loadFromEnvironment(ec.data, ec.prefix)
}

// Save is not applicable for environment configurations
func (ec *EnvironmentConfig) Save() error {
	return fmt.Errorf("environment configurations cannot be saved")
}

// Validate validates the environment configuration
func (ec *EnvironmentConfig) Validate() error {
	result := validation.ValidateStruct(ec.data)
	if !result.Valid {
		var errors []string
		for _, err := range result.Errors {
			errors = append(errors, err.Message)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// GetDefaults returns the default configuration
func (ec *EnvironmentConfig) GetDefaults() interface{} {
	dataType := reflect.TypeOf(ec.data)
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}
	
	defaultValue := reflect.New(dataType).Interface()
	return defaultValue
}

// Reset resets the configuration to defaults
func (ec *EnvironmentConfig) Reset() error {
	ec.data = ec.GetDefaults()
	return nil
}

// loadFromEnvironment loads configuration from environment variables using reflection
func (ec *EnvironmentConfig) loadFromEnvironment(data interface{}, prefix string) error {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct or pointer to struct")
	}
	
	dataType := value.Type()
	
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := dataType.Field(i)
		
		// Skip unexported fields
		if !field.CanSet() {
			continue
		}
		
		// Get environment variable name
		envName := ec.getEnvName(fieldType, prefix)
		envValue := os.Getenv(envName)
		
		if envValue == "" {
			continue
		}
		
		// Set field value based on type
		if err := ec.setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("failed to set field %s: %v", fieldType.Name, err)
		}
	}
	
	return nil
}

// getEnvName gets the environment variable name for a field
func (ec *EnvironmentConfig) getEnvName(field reflect.StructField, prefix string) string {
	// Check for env tag
	if envTag := field.Tag.Get("env"); envTag != "" {
		return envTag
	}
	
	// Use field name with prefix
	fieldName := strings.ToUpper(field.Name)
	if prefix != "" {
		return fmt.Sprintf("%s_%s", strings.ToUpper(prefix), fieldName)
	}
	
	return fieldName
}

// setFieldValue sets a field value from a string
func (ec *EnvironmentConfig) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle duration
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(duration))
		} else {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	
	return nil
}

// Helper functions

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Global configuration manager
var globalConfigManager = NewConfigManager("/boot/config/plugins/uma")

// GetGlobalConfigManager returns the global configuration manager
func GetGlobalConfigManager() *ConfigManager {
	return globalConfigManager
}

// Convenience functions

// RegisterConfig registers a configuration with the global manager
func RegisterConfig(name string, config ConfigPattern) error {
	return globalConfigManager.RegisterConfig(name, config)
}

// LoadConfig loads a configuration using the global manager
func LoadConfig(name string) error {
	return globalConfigManager.LoadConfig(name)
}

// SaveConfig saves a configuration using the global manager
func SaveConfig(name string) error {
	return globalConfigManager.SaveConfig(name)
}
