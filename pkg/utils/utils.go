package utils

import (
	"encoding/json"
	"os"
)

func LoadConfig(path string, target interface{}) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, target)
}

func ShowDedSecArt() {
	if art, err := os.ReadFile("./assets/dedsec.txt"); err != nil {
		println(string(art))
	}
}
