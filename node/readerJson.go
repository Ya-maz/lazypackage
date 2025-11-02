package node

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadScriptsFromPackageJSON() map[string]string {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return map[string]string{"error": "❌ package.json not found"}
	}

	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return map[string]string{"error": "❌ invalid package.json"}
	}

	scripts := map[string]string{}
	if s, ok := pkg["scripts"].(map[string]interface{}); ok {
		for key, value := range s {
			scripts[key] = fmt.Sprintf("%v", value)
		}
	} else {
		scripts["info"] = "❌ no scripts found"
	}
	return scripts
}

func ReadSymbolsFromJSON() map[string]string {
	data, err := os.ReadFile("simbol.json")
	if err != nil {
		fmt.Println("Warning: symbol.json not found, using empty map.")
		return map[string]string{} // Лучше вернуть пустую мапу
	}

	var jsn map[string]string
	if err := json.Unmarshal(data, &jsn); err != nil {
		fmt.Println("Warning: invalid symbol.json, using empty map.")
		return map[string]string{} // И здесь тоже
	}

	return jsn
}


