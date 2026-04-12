package main
import (
    "encoding/json"
    "fmt"
    "math/big"
)
type X struct {
    Val *big.Int `json:"val"`
}
func main() {
    v := big.NewInt(1234)
    b, _ := json.Marshal(X{v})
    fmt.Println(string(b))
}
