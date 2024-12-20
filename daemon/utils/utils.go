package utils

import (
	"errors"
	"os"
)

func ExtractFilePath(inputs []interface{}) string {
	filePath := ""
	for _, input := range inputs {
		if input == nil {
			continue
		}
		path, ok := input.(string)
		if !ok {
			continue
		}
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}
		filePath = path
		break
	}

	return filePath
}
