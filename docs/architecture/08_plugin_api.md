# Allele WASM Plugin API Reference

The Allele Trading Engine uses a **Tri-Plugin Microkernel Architecture**. To satisfy the "No Black Boxes" rule, trading strategies must be compiled to WebAssembly (WASM) as pure, deterministic mathematical functions. They cannot access the network, read files, or maintain hidden state.

This document details the Application Binary Interface (ABI) used to pass data between the Go host (Allele Engine) and the WASM guest (your strategy), along with language-specific examples.

---

## 1. The Memory ABI

Because WebAssembly only understands numeric types (`i32`, `i64`, `f32`, `f64`), passing complex JSON objects requires shared memory serialization.

### Host to Guest (Invocation)
When the engine evaluates your strategy, it calls the exported `Evaluate` function:
`Evaluate(statePtr i32, stateLen i32) -> (outPtr i32, outLen i32)`

1. **`statePtr`**: Memory offset where the JSON `MarketState` begins.
2. **`stateLen`**: Length of the JSON string in bytes.

### Guest to Host (Return)
The plugin must parse the state, compute its signals, serialize a JSON array of `Signal` objects, write it to its own memory, and return:
1. **`outPtr`**: Memory offset where the returned JSON string begins.
2. **`outLen`**: Length of the returned JSON string.

*(Note: In some WASM targets that do not support multiple return values, these are packed into a single `i64` where the high 32 bits are the pointer and the low 32 bits are the length).*

---

## 2. Language Examples

### Golang (compiled via TinyGo)
Go uses `//export` to expose functions to the host. TinyGo is highly recommended for building lightweight WASM modules.

```go
//go:build wasm
package main

import (
	"encoding/json"
	"unsafe"
)

//export Evaluate
func Evaluate(statePtr uint32, stateLength uint32) uint64 {
	// 1. Read input from memory
	stateBytes := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(statePtr))), stateLength)

	// 2. Parse MarketState
	var state map[string]interface{}
	json.Unmarshal(stateBytes, &state)

	// 3. Compute logic...
	signals := []map[string]interface{}{
		{"asset": "TOKEN_A", "action": "BUY", "confidence": 0.95},
	}

	// 4. Serialize output
	outBytes, _ := json.Marshal(signals)
	outPtr := uint32(uintptr(unsafe.Pointer(&outBytes[0])))
	outLen := uint32(len(outBytes))

	// Pack 2 uint32s into 1 uint64 for the return
	return (uint64(outPtr) << 32) | uint64(outLen)
}

func main() {} // Required for TinyGo WASM compilation
```
*Build with:* `tinygo build -o strategy.wasm -target=wasi main.go`

### TypeScript / AssemblyScript
AssemblyScript compiles a strict subset of TypeScript directly to WebAssembly. It is excellent for Allele plugins due to its zero-overhead memory access.

```typescript
// AssemblyScript Plugin
export function Evaluate(statePtr: i32, stateLen: i32): u64 {
  // 1. Read input from memory
  let jsonString = String.UTF8.decodeUnsafe(statePtr, stateLen);
  
  // 2. Parse JSON (requires 'assemblyscript-json' library)
  // let state = JSON.parse(jsonString);
  
  // 3. Compute logic...
  let outString = '[{"asset":"TOKEN_A", "action":"BUY", "confidence": 0.95}]';
  
  // 4. Write back to memory
  let outBuffer = String.UTF8.encode(outString);
  let outPtr = changetype<usize>(outBuffer);
  let outLen = outBuffer.byteLength;
  
  // Pack return values
  return (u64(outPtr) << 32) | u64(outLen);
}
```
*Build with:* `asc index.ts -b strategy.wasm --target release`

### Python (via Extism PDK or Pyodide)
Python cannot compile natively to raw WASM easily without embedding a runtime. For Allele plugins, using the **Extism Python PDK** is the standard approach to wrap your logic in a WASM module.

```python
import extism_pdk
import json

@extism_pdk.plugin_fn
def evaluate():
    # 1. Extism automatically handles memory reading
    input_bytes = extism_pdk.input()
    state = json.loads(input_bytes)
    
    # 2. Compute logic...
    signals = [{
        "asset": "TOKEN_A",
        "action": "BUY",
        "confidence": 0.95
    }]
    
    # 3. Write output back to the host
    extism_pdk.output(json.dumps(signals))
```
*Build with:* Requires the Extism Pyodide compiler toolchain: `extism-py plugin.py -o strategy.wasm`
