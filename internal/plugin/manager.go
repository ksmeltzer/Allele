package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"allele/internal/abi"
	"allele/internal/loader"
)

type Manager struct {
	pluginsDir string
	manifests  map[string]abi.Manifest
	modules    map[string]*loader.WasmModule
}

func NewManager(pluginsDir string) *Manager {
	return &Manager{
		pluginsDir: pluginsDir,
		manifests:  make(map[string]abi.Manifest),
		modules:    make(map[string]*loader.WasmModule),
	}
}

func (m *Manager) LoadAll(ctx context.Context) error {
	if err := os.MkdirAll(m.pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins dir: %w", err)
	}

	files, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		return fmt.Errorf("failed to read plugins dir: %w", err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".wasm") {
			path := filepath.Join(m.pluginsDir, f.Name())
			if err := m.loadAndResolve(ctx, path); err != nil {
				log.Printf("Failed to load plugin %s: %v", f.Name(), err)
			}
		}
	}

	return nil
}

func (m *Manager) loadAndResolve(ctx context.Context, path string) error {
	mod, err := loader.Load(ctx, path)
	if err != nil {
		return fmt.Errorf("loader.Load: %w", err)
	}

	manifestFn := mod.Module.ExportedFunction("Manifest")
	if manifestFn == nil {
		mod.Close(ctx)
		return fmt.Errorf("module does not export 'Manifest' function")
	}

	res, err := manifestFn.Call(ctx)
	if err != nil {
		mod.Close(ctx)
		return fmt.Errorf("calling Manifest: %w", err)
	}

	if len(res) == 0 {
		mod.Close(ctx)
		return fmt.Errorf("Manifest returned no value")
	}

	val := res[0]
	ptr := uint32(val >> 32)
	length := uint32(val & 0xFFFFFFFF)

	manifestBytes, err := abi.ReadFromMemory(ctx, mod.Module, ptr, length)
	if err != nil {
		mod.Close(ctx)
		return fmt.Errorf("ReadFromMemory: %w", err)
	}

	var manifest abi.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		mod.Close(ctx)
		return fmt.Errorf("unmarshaling manifest: %w", err)
	}

	m.manifests[manifest.Name] = manifest
	m.modules[manifest.Name] = mod
	log.Printf("Loaded plugin %s (v%s)", manifest.Name, manifest.Version)

	// Resolve dependencies
	for _, dep := range manifest.Dependencies {
		if _, exists := m.manifests[dep.Name]; !exists {
			log.Printf("Dependency %s missing for %s. Attempting to download from %s...", dep.Name, manifest.Name, dep.Url)
			if err := m.downloadAndLoadDependency(ctx, dep); err != nil {
				log.Printf("Warning: failed to download dependency %s: %v", dep.Name, err)
			}
		}
	}

	return nil
}

func (m *Manager) downloadAndLoadDependency(ctx context.Context, dep abi.Dependency) error {
	if dep.Url == "" {
		return fmt.Errorf("no URL provided for dependency %s", dep.Name)
	}

	resp, err := http.Get(dep.Url)
	if err != nil {
		return fmt.Errorf("http.Get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	filename := fmt.Sprintf("%s.wasm", dep.Name)
	path := filepath.Join(m.pluginsDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	log.Printf("Downloaded dependency %s to %s", dep.Name, path)

	// Recursively load the downloaded dependency
	return m.loadAndResolve(ctx, path)
}

func (m *Manager) GetManifests() []abi.Manifest {
	var list []abi.Manifest
	for _, man := range m.manifests {
		list = append(list, man)
	}
	return list
}
