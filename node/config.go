package node

import (
	"encoding/json"
	"fmt"
	"os"
)

// ScriptInfo represents a script with its command and icon
type ScriptInfo struct {
	Name    string
	Command string
	Icon    string
}

// ConfigManager handles loading configuration from package.json and symbols.json
type ConfigManager struct {
	PackagePath string
	SymbolsPath string
}

// NewConfigManager creates a new ConfigManager with default paths
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		PackagePath: "package.json",
		SymbolsPath: "symbols.json",
	}
}

// LoadScripts reads scripts from package.json and matches them with symbols
func (cm *ConfigManager) LoadScripts() ([]ScriptInfo, error) {
	pkgData, err := os.ReadFile(cm.PackagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", cm.PackagePath, err)
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(pkgData, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", cm.PackagePath, err)
	}

	symbols := cm.loadSymbols()
	scripts := make([]ScriptInfo, 0, len(pkg.Scripts))

	for name, cmd := range pkg.Scripts {
		scripts = append(scripts, ScriptInfo{
			Name:    name,
			Command: cmd,
			Icon:    symbols[name],
		})
	}

	return scripts, nil
}

func (cm *ConfigManager) loadSymbols() map[string]string {
	symbols := make(map[string]string)
	data, err := os.ReadFile(cm.SymbolsPath)
	if err != nil {
		return symbols
	}

	_ = json.Unmarshal(data, &symbols)
	return symbols
}
