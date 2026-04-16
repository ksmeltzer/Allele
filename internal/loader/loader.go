package loader

import (
	"context"
	"fmt"
	"os"

	"allele/internal/config"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WasmModule wraps the loaded wazero components.
type WasmModule struct {
	Runtime wazero.Runtime
	Module  api.Module
}

// Load loads a WebAssembly module from the given file path,
// instantiates a WASI environment, and returns the initialized api.Module.
func Load(ctx context.Context, path string) (*WasmModule, error) {
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading wasm file: %w", err)
	}

	r := wazero.NewRuntime(ctx)

	// Instantiate WASI environment
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Export env.read_memory and env.write_memory for the WASM ABI
	_, err = r.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, ptr uint32, length uint32) uint32 {
			// read_memory implementation (dummy/passthrough for now, actual logic depends on host)
			return ptr
		}).Export("read_memory").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, ptr uint32, length uint32) uint32 {
			// write_memory implementation
			return ptr
		}).Export("write_memory").
		Instantiate(ctx)
	if err != nil && err.Error() != "module env has already been instantiated" {
		r.Close(ctx)
		return nil, fmt.Errorf("instantiating env: %w", err)
	}

	// Compile the module
	compiled, err := r.CompileModule(ctx, wasmBytes)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("compiling wasm: %w", err)
	}

	// Prepare wazero configuration
	modConfig := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStartFunctions() // Call nothing automatically

	// Mount the plugin's directory so it can access config.yaml at /config
	modConfig = modConfig.WithFSConfig(config.MountPluginFS(path, "/config"))

	// Instantiate the compiled module
	mod, err := r.InstantiateModule(ctx, compiled, modConfig)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("instantiating module: %w", err)
	}

	// Try to initialize as a reactor module first (_initialize)
	// Go 1.24+ c-shared WASI modules export _initialize which returns immediately after setup.
	initFunc := mod.ExportedFunction("_initialize")
	if initFunc != nil {
		_, err := initFunc.Call(context.Background())
		if err != nil {
			r.Close(ctx)
			return nil, fmt.Errorf("calling _initialize: %w", err)
		}
	} else {
		// Fallback for older Go WASM binaries (which use _start and run main())
		// Start the Go WASM runtime in a background goroutine so it doesn't block the host
		startFunc := mod.ExportedFunction("_start")
		if startFunc != nil {
			go func() {
				// This will block forever if main() has a select{} or similar
				startFunc.Call(context.Background())
			}()
		}
	}

	return &WasmModule{
		Runtime: r,
		Module:  mod,
	}, nil
}

// Close cleans up the runtime and module.
func (wm *WasmModule) Close(ctx context.Context) error {
	if wm.Runtime != nil {
		return wm.Runtime.Close(ctx)
	}
	return nil
}
