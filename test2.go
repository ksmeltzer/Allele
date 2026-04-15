package main

import (
	"encoding/json"
	"fmt"
)

type ConfigField struct {
	Key      string `json:"key"`
	Required bool   `json:"required"`
}

type Manifest struct {
	Name   string        `json:"name"`
	Config []ConfigField `json:"config"`
}

func main() {
	type ConfigFieldWithValue struct {
		ConfigField
		Value string `json:"value"`
	}
	type ManifestWithValues struct {
		Manifest
		Config []ConfigFieldWithValue `json:"config"`
	}

	var m ManifestWithValues
	m.Name = "Test"
	m.Config = []ConfigFieldWithValue{
		{ConfigField: ConfigField{Key: "POLY", Required: true}, Value: ""},
	}

	b, _ := json.Marshal(m)
	fmt.Println(string(b))
}
