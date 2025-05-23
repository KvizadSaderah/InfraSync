package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	tfjson "github.com/hashicorp/terraform-json"
)

// printAttributeDiff выводит детальные изменения атрибутов.
// indent используется для форматирования вложенных атрибутов.
func printAttributeDiff(before, after, afterUnknown, beforeSensitive, afterSensitive interface{}, indent string) {
	beforeMap, okBefore := before.(map[string]interface{})
	afterMap, okAfter := after.(map[string]interface{})
	afterUnknownMap, _ := afterUnknown.(map[string]interface{}) // Может быть nil
	// beforeSensitiveMap, _ := beforeSensitive.(map[string]interface{}) // Пока не используем детально
	// afterSensitiveMap, _ := afterSensitive.(map[string]interface{})   // Пока не используем детально

	if !okBefore && !okAfter {
		// Если оба не карты, и before != after, то это простое изменение значения
		if fmt.Sprintf("%v", before) != fmt.Sprintf("%v", after) {
			isSensitive := false // Упрощенная проверка на чувствительность
			if bs, ok := beforeSensitive.(bool); ok && bs {
				isSensitive = true
			} else if as, ok := afterSensitive.(bool); ok && as {
				isSensitive = true
			}

			beforeStr := fmt.Sprintf("%v", before)
			afterStr := fmt.Sprintf("%v", after)
			if isSensitive {
				beforeStr = "(sensitive)"
				afterStr = "(sensitive)"
			}
			color.Yellow("%s  ~ %s -> %s", indent, beforeStr, afterStr)
		}
		return
	}

	// Собираем все ключи из before и after
	allKeys := make(map[string]bool)
	if okBefore {
		for k := range beforeMap {
			allKeys[k] = true
		}
	}
	if okAfter {
		for k := range afterMap {
			allKeys[k] = true
		}
	}
	if afterUnknownMap != nil {
		for k := range afterUnknownMap {
			allKeys[k] = true
		}
	}

	for key := range allKeys {
		valBefore, existsBefore := getMapValue(beforeMap, key) // getMapValue - helper to get value if map exists
		valAfter, existsAfter := getMapValue(afterMap, key)
		_, isUnknown := getMapValue(afterUnknownMap, key)

		// Проверка чувствительности для текущего ключа
		// Это упрощенная проверка, в реальности beforeSensitive/afterSensitive имеют ту же структуру, что и before/after
		isKeySensitive := false
		if bs, ok := beforeSensitive.(map[string]interface{}); ok && bs[key] == true {
			isKeySensitive = true
		} else if as, ok := afterSensitive.(map[string]interface{}); ok && as[key] == true {
			isKeySensitive = true
		}

		if isUnknown {
			color.Cyan("%s  • %s: (known after apply)", indent, key)
		} else if existsBefore && existsAfter {
			if fmt.Sprintf("%v", valBefore) != fmt.Sprintf("%v", valAfter) {
				// Если это вложенные структуры, рекурсивно вызываем diff
				if _, bIsMap := valBefore.(map[string]interface{}); bIsMap {
					// _, aIsMap := valAfter.(map[string]interface{}); aIsMap &&
					color.Yellow("%s~ %s:", indent, key)
					// Передаем соответствующие части sensitive-структур
					var subBeforeSensitive, subAfterSensitive interface{}
					if bsMap, ok := beforeSensitive.(map[string]interface{}); ok {
						subBeforeSensitive = bsMap[key]
					}
					if asMap, ok := afterSensitive.(map[string]interface{}); ok {
						subAfterSensitive = asMap[key]
					}
					printAttributeDiff(valBefore, valAfter, nil, subBeforeSensitive, subAfterSensitive, indent+"    ")
				} else {
					afterValStr := fmt.Sprintf("%v", valAfter)
					beforeValStr := fmt.Sprintf("%v", valBefore)
					if isKeySensitive {
						afterValStr = "(sensitive)"
						beforeValStr = "(sensitive)"
					}
					color.Yellow("%s  ~ %s: %s -> %s", indent, key, beforeValStr, afterValStr)
				}
			} else {
				// Значения не изменились, можно не выводить или выводить серым
				// fmt.Printf("%s    %s: %v (unchanged)\n", indent, key, valBefore)
			}
		} else if existsAfter {
			afterValStr := fmt.Sprintf("%v", valAfter)
			if isKeySensitive {
				afterValStr = "(sensitive)"
			}
			color.Green("%s  + %s: %s", indent, key, afterValStr)
		} else if existsBefore { // Существует только в before, значит удален
			beforeValStr := fmt.Sprintf("%v", valBefore)
			if isKeySensitive {
				beforeValStr = "(sensitive)"
			}
			color.Red("%s  - %s: %s", indent, key, beforeValStr)
		}
	}
}

// Вспомогательная функция для безопасного получения значения из карты
func getMapValue(m map[string]interface{}, key string) (interface{}, bool) {
	if m == nil {
		return nil, false
	}
	val, exists := m[key]
	return val, exists
}

func printDiff(rc *tfjson.ResourceChange) {
	action := rc.Change.Actions[0] // For simplicity, consider only the first action

	switch action {
	case tfjson.ActionCreate:
		color.Green("+ %s (%s)", rc.Address, rc.Type)
	case tfjson.ActionDelete:
		color.Red("- %s (%s)", rc.Address, rc.Type)
	case tfjson.ActionUpdate:
		color.Yellow("~ %s (%s)", rc.Address, rc.Type)
		printAttributeDiff(rc.Change.Before, rc.Change.After, rc.Change.AfterUnknown, rc.Change.BeforeSensitive, rc.Change.AfterSensitive, "  ")
	default:
		fmt.Printf("  %s (%s) - Actions: %s\n", rc.Address, rc.Type, rc.Change.Actions)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: infrasyncli <path_to_tfplan.json>")
		os.Exit(1)
	}

	planFile := os.Args[1]

	jsonData, err := ioutil.ReadFile(planFile)
	if err != nil {
		fmt.Printf("Error reading plan file: %v\n", err)
		os.Exit(1)
	}

	var plan tfjson.Plan
	err = json.Unmarshal(jsonData, &plan)
	if err != nil {
		fmt.Printf("Error unmarshalling plan JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully parsed plan file: %s\n", planFile)
	fmt.Printf("Terraform Version: %s\n", plan.TerraformVersion)
	fmt.Printf("Format Version: %s\n", plan.FormatVersion)

	if plan.ResourceChanges == nil {
		fmt.Println("No resource changes found in the plan.")
		return
	}

	fmt.Printf("Found %d resource changes:\n", len(plan.ResourceChanges))
	for _, rc := range plan.ResourceChanges {
		printDiff(rc)
	}
}
