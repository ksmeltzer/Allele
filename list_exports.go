package main
import (
	"context"
	"fmt"
	"os"
	"github.com/tetratelabs/wazero"
)
func main() {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	b, _ := os.ReadFile(".allele/plugins/allele-exchange-polymarket.wasm")
	mod, _ := r.CompileModule(ctx, b)
	for _, exp := range mod.ExportedFunctions() {
		fmt.Println("Exported:", exp)
	}
}
