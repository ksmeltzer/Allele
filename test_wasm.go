package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func main() {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	b, _ := os.ReadFile(".allele/plugins/allele-strategy-completeness-go.wasm")
	compiled, _ := r.CompileModule(ctx, b)

	fmt.Println("Exports:")
	for _, exp := range compiled.ExportedFunctions() {
		fmt.Println(" -", exp.Name())
	}

	modConfig := wazero.NewModuleConfig().WithStartFunctions()
	mod, err := r.InstantiateModule(ctx, compiled, modConfig)
	if err != nil {
		panic(err)
	}

	initFunc := mod.ExportedFunction("_initialize")
	if initFunc != nil {
		fmt.Println("Calling _initialize...")
		_, err = initFunc.Call(ctx)
		fmt.Println("err:", err)
	}

	manFunc := mod.ExportedFunction("Manifest")
	if manFunc != nil {
		fmt.Println("Calling Manifest...")
		_, err = manFunc.Call(ctx)
		fmt.Println("err:", err)
	}
}
