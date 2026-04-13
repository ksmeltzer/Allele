package loader_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"allele/internal/loader"
)

func TestLoadWasm(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config.yaml in the plugin directory
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test_key: test_value\n"), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	// Create a dummy Go file to compile to WebAssembly.
	srcPath := filepath.Join(tmpDir, "main.go")
	srcCode := `
package main

import (
	"fmt"
	"os"
)

//export test_func
func test_func() int32 {
	return 42
}

func main() {
	// Attempt to read the mounted config file
	data, err := os.ReadFile("/config/config.yaml")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CONFIG_LOADED: %s", string(data))
}
`
	if err := os.WriteFile(srcPath, []byte(srcCode), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	wasmPath := filepath.Join(tmpDir, "hello.wasm")

	// Compile the dummy Go file to WASM.
	cmd := exec.Command("go", "build", "-o", wasmPath, srcPath)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to compile to wasm: %v\noutput: %s", err, string(out))
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test the Load function.
	ctx := context.Background()
	mod, err := loader.Load(ctx, wasmPath)
	
	// Restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	if err != nil {
		t.Fatalf("failed to load wasm: %v", err)
	}
	defer mod.Close(ctx)

	if mod.Module == nil {
		t.Fatal("expected module to not be nil")
	}

	// Read stdout output to verify config.yaml was successfully read
	var buf strings.Builder
	buf.Grow(1024)
	pipeBytes := make([]byte, 1024)
	for {
		n, err := r.Read(pipeBytes)
		if n > 0 {
			buf.Write(pipeBytes[:n])
		}
		if err != nil {
			break
		}
	}
	output := buf.String()
	
	if !strings.Contains(output, "CONFIG_LOADED: test_key: test_value") {
		t.Fatalf("WASM failed to read config.yaml. Output: %s", output)
	}
}
