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

type ConfigFieldWithValue struct {
	ConfigField
	Value string `json:"value"`
}

type ManifestWithValues struct {
	Manifest
	Config []ConfigFieldWithValue `json:"config"`
}

func main() {
	m1 := Manifest{
		Name: "Test",
		Config: []ConfigField{
			{Key: "POLY", Required: true},
		},
	}
	
	newConfig := []ConfigFieldWithValue{
		{ConfigField: m1.Config[0], Value: "secret"},
	}

	m2 := ManifestWithValues{
		Manifest: m1,
		Config: newConfig,
	}

	b, _ := json.Marshal(m2)
	fmt.Println(string(b))
}
