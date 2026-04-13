package config

import (
	"path/filepath"

	"github.com/tetratelabs/wazero"
)

// MountPluginFS creates a wazero.FSConfig that mounts the plugin's directory
// so that the wasm guest can access its configuration files (e.g. config.yaml).
func MountPluginFS(pluginPath string, mountPath string) wazero.FSConfig {
	pluginDir := filepath.Dir(pluginPath)
	
	// Create the wazero FSConfig and mount the plugin's directory
	return wazero.NewFSConfig().WithDirMount(pluginDir, mountPath)
}
