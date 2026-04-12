package main
import (
    "encoding/json"
    "fmt"
    "math/big"
)
type X struct {
    Val *big.Int `json:"val,string"`
}
func main() {
    v := big.NewInt(1234)
    b, err := json.Marshal(X{v})
    if err != nil { fmt.Println(err) }
    fmt.Println(string(b))
}
