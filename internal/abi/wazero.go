package abi

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

// ReadFromMemory reads length bytes from WASM memory starting at ptr.
func ReadFromMemory(ctx context.Context, mod api.Module, ptr uint32, length uint32) ([]byte, error) {
	bytes, ok := mod.Memory().Read(ptr, length)
	if !ok {
		return nil, fmt.Errorf("out of range reading memory: ptr=%d, len=%d", ptr, length)
	}
	// Copy to avoid retaining a slice to wazero's internal memory
	res := make([]byte, length)
	copy(res, bytes)
	return res, nil
}

// WriteToMemory allocates memory via malloc (if exported) or expects the host to manage ptr,
// and writes data to WASM memory. For simplicity, we assume the module exports a malloc function,
// or we use a pre-allocated buffer pointer provided by the caller.
func WriteToMemory(ctx context.Context, mod api.Module, data []byte) (uint32, uint32, error) {
	malloc := mod.ExportedFunction("malloc")
	if malloc == nil {
		return 0, 0, fmt.Errorf("module does not export 'malloc'")
	}
	
	length := uint32(len(data))
	res, err := malloc.Call(ctx, uint64(length))
	if err != nil {
		return 0, 0, fmt.Errorf("calling malloc: %w", err)
	}
	if len(res) == 0 {
		return 0, 0, fmt.Errorf("malloc returned no value")
	}
	
	ptr := uint32(res[0])
	ok := mod.Memory().Write(ptr, data)
	if !ok {
		return 0, 0, fmt.Errorf("out of range writing memory: ptr=%d, len=%d", ptr, length)
	}
	
	return ptr, length, nil
}
