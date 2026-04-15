package main
import (
	"encoding/json"
	"fmt"
)
type ConfigField struct { Key string `json:"key"` }
type Manifest struct {
	Name   string        `json:"name"`
	Config []ConfigField `json:"config"`
}
type ConfigFieldWithValue struct {
	ConfigField
	Value string `json:"value"`
}
type ManifestWithValues struct {
	Manifest
	Config []ConfigFieldWithValue `json:"config"`
}
func main() {
	m := ManifestWithValues{
		Manifest: Manifest{Name: "Test"},
		Config: []ConfigFieldWithValue{
			{ConfigField: ConfigField{Key: "foo"}, Value: "bar"},
		},
	}
	b, _ := json.Marshal(m)
	fmt.Println(string(b))
}
