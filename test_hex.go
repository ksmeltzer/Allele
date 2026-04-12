package main
import (
    "encoding/json"
    "fmt"
    "math/big"
    "github.com/ethereum/go-ethereum/common/math"
)
type X struct {
    Val *math.HexOrDecimal256 `json:"val"`
}
func main() {
    v := (*math.HexOrDecimal256)(big.NewInt(1234))
    b, err := json.Marshal(X{v})
    if err != nil { fmt.Println(err) }
    fmt.Println(string(b))
}
