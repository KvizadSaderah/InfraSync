package formatter

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/kvizadsaderah/infrasync/pkg/parser"
)

// CLIFormatter formats plan output for terminal display
type CLIFormatter struct {
	ShowUnchanged bool
	Verbose       bool
}

// NewCLIFormatter creates a new CLI formatter
func NewCLIFormatter(showUnchanged, verbose bool) *CLIFormatter {
	return &CLIFormatter{
		ShowUnchanged: showUnchanged,
		Verbose:       verbose,
	}
}

// Format outputs the plan summary to terminal with colors
func (f *CLIFormatter) Format(summary *parser.PlanSummary) {
	// Print header
	fmt.Printf("\n")
	color.Cyan("═══════════════════════════════════════════════════════")
	color.Cyan("  Terraform Plan Summary")
	color.Cyan("═══════════════════════════════════════════════════════")
	fmt.Printf("\n")

	fmt.Printf("Terraform Version: %s\n", summary.TerraformVersion)
	fmt.Printf("Format Version: %s\n\n", summary.FormatVersion)

	// Print statistics
	f.printStatistics(summary)
	fmt.Printf("\n")

	// Print changes
	if len(summary.Changes) == 0 {
		color.Green("✓ No changes detected. Infrastructure is up-to-date.")
		return
	}

	// Group and print changes by type
	f.printChangesByType(summary)
}

func (f *CLIFormatter) printStatistics(summary *parser.PlanSummary) {
	color.Cyan("Changes Overview:")
	color.Cyan("─────────────────")

	if summary.ToCreate > 0 {
		color.Green("  ✓ %d to create", summary.ToCreate)
	}
	if summary.ToUpdate > 0 {
		color.Yellow("  ~ %d to update", summary.ToUpdate)
	}
	if summary.ToReplace > 0 {
		color.Magenta("  ⟳ %d to replace", summary.ToReplace)
	}
	if summary.ToDelete > 0 {
		color.Red("  ✗ %d to destroy", summary.ToDelete)
	}
	if summary.NoChanges > 0 && f.ShowUnchanged {
		color.White("  • %d unchanged", summary.NoChanges)
	}

	total := summary.ToCreate + summary.ToUpdate + summary.ToReplace + summary.ToDelete
	if total == 0 && summary.NoChanges > 0 {
		color.Green("\n  All resources are up-to-date!")
	}
}

func (f *CLIFormatter) printChangesByType(summary *parser.PlanSummary) {
	// Print creates
	creates := f.filterChanges(summary.Changes, func(c parser.ResourceChange) bool { return c.IsCreate })
	if len(creates) > 0 {
		color.Green("\n✓ Resources to CREATE (%d):", len(creates))
		color.Green("────────────────────────────")
		for _, c := range creates {
			color.Green("  + %s", c.Address)
			color.HiBlack("    Type: %s", c.Type)
			if f.Verbose {
				f.printAttributes(c.After, "    ", color.HiGreen)
			}
		}
	}

	// Print updates
	updates := f.filterChanges(summary.Changes, func(c parser.ResourceChange) bool { return c.IsUpdate })
	if len(updates) > 0 {
		color.Yellow("\n~ Resources to UPDATE (%d):", len(updates))
		color.Yellow("────────────────────────────")
		for _, c := range updates {
			color.Yellow("  ~ %s", c.Address)
			color.HiBlack("    Type: %s", c.Type)
			f.printAttributeDiff(c.Before, c.After, c.AfterUnknown, c.BeforeSensitive, c.AfterSensitive, "    ")
		}
	}

	// Print replaces
	replaces := f.filterChanges(summary.Changes, func(c parser.ResourceChange) bool { return c.IsReplace })
	if len(replaces) > 0 {
		color.Magenta("\n⟳ Resources to REPLACE (%d):", len(replaces))
		color.Magenta("────────────────────────────")
		for _, c := range replaces {
			color.Magenta("  ⟳ %s", c.Address)
			color.HiBlack("    Type: %s", c.Type)
			color.HiYellow("    ⚠ This resource will be destroyed and recreated")
			if f.Verbose {
				f.printAttributeDiff(c.Before, c.After, c.AfterUnknown, c.BeforeSensitive, c.AfterSensitive, "    ")
			}
		}
	}

	// Print deletes
	deletes := f.filterChanges(summary.Changes, func(c parser.ResourceChange) bool { return c.IsDelete })
	if len(deletes) > 0 {
		color.Red("\n✗ Resources to DESTROY (%d):", len(deletes))
		color.Red("────────────────────────────")
		for _, c := range deletes {
			color.Red("  - %s", c.Address)
			color.HiBlack("    Type: %s", c.Type)
			if f.Verbose {
				f.printAttributes(c.Before, "    ", color.HiRed)
			}
		}
	}

	fmt.Printf("\n")
}

func (f *CLIFormatter) filterChanges(changes []parser.ResourceChange, predicate func(parser.ResourceChange) bool) []parser.ResourceChange {
	result := make([]parser.ResourceChange, 0)
	for _, c := range changes {
		if predicate(c) {
			result = append(result, c)
		}
	}
	return result
}

func (f *CLIFormatter) printAttributes(attrs map[string]interface{}, indent string, colorFunc func(format string, a ...interface{})) {
	for key, val := range attrs {
		valStr := formatValue(val)
		if valStr != "" {
			colorFunc("%s  %s: %s", indent, key, valStr)
		}
	}
}

func (f *CLIFormatter) printAttributeDiff(before, after map[string]interface{}, afterUnknown, beforeSensitive, afterSensitive interface{}, indent string) {
	afterUnknownMap := toMap(afterUnknown)

	// Collect all keys
	allKeys := make(map[string]bool)
	for k := range before {
		allKeys[k] = true
	}
	for k := range after {
		allKeys[k] = true
	}

	for key := range allKeys {
		beforeVal, existsBefore := before[key]
		afterVal, existsAfter := after[key]
		_, isUnknown := afterUnknownMap[key]

		isSensitive := f.isSensitiveKey(key, beforeSensitive, afterSensitive)

		if isUnknown {
			color.Cyan("%s  • %s: (known after apply)", indent, key)
		} else if existsBefore && existsAfter {
			beforeStr := formatValue(beforeVal)
			afterStr := formatValue(afterVal)

			if isSensitive {
				beforeStr = "(sensitive)"
				afterStr = "(sensitive)"
			}

			if beforeStr != afterStr {
				color.Yellow("%s  ~ %s: %s → %s", indent, key, beforeStr, afterStr)
			}
		} else if existsAfter {
			afterStr := formatValue(afterVal)
			if isSensitive {
				afterStr = "(sensitive)"
			}
			color.Green("%s  + %s: %s", indent, key, afterStr)
		} else if existsBefore {
			beforeStr := formatValue(beforeVal)
			if isSensitive {
				beforeStr = "(sensitive)"
			}
			color.Red("%s  - %s: %s", indent, key, beforeStr)
		}
	}
}

func (f *CLIFormatter) isSensitiveKey(key string, beforeSensitive, afterSensitive interface{}) bool {
	if bm, ok := beforeSensitive.(map[string]interface{}); ok {
		if val, exists := bm[key]; exists {
			if b, ok := val.(bool); ok && b {
				return true
			}
		}
	}
	if am, ok := afterSensitive.(map[string]interface{}); ok {
		if val, exists := am[key]; exists {
			if b, ok := val.(bool); ok && b {
				return true
			}
		}
	}
	return false
}

func formatValue(v interface{}) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		if len(val) > 60 {
			return fmt.Sprintf("%.60s...", val)
		}
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	case map[string]interface{}:
		return "{...}"
	case []interface{}:
		return fmt.Sprintf("[%d items]", len(val))
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 60 {
			return str[:60] + "..."
		}
		return str
	}
}

func toMap(v interface{}) map[string]interface{} {
	if v == nil {
		return make(map[string]interface{})
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return make(map[string]interface{})
}

// PrintWarnings prints any important warnings about the changes
func (f *CLIFormatter) PrintWarnings(summary *parser.PlanSummary) {
	hasDestructions := summary.ToDelete > 0 || summary.ToReplace > 0

	if hasDestructions {
		fmt.Printf("\n")
		color.New(color.BgRed, color.FgWhite, color.Bold).Printf(" ⚠ WARNING ")
		color.Red(" This plan includes destructive changes!")

		if summary.ToDelete > 0 {
			color.Red("  → %d resource(s) will be DESTROYED", summary.ToDelete)
		}
		if summary.ToReplace > 0 {
			color.Red("  → %d resource(s) will be REPLACED (destroyed and recreated)", summary.ToReplace)
		}

		color.Yellow("\n  Please review carefully before applying.")
		fmt.Printf("\n")
	}
}

// FormatCompact provides a one-line summary
func (f *CLIFormatter) FormatCompact(summary *parser.PlanSummary) string {
	parts := make([]string, 0)

	if summary.ToCreate > 0 {
		parts = append(parts, color.GreenString("+%d", summary.ToCreate))
	}
	if summary.ToUpdate > 0 {
		parts = append(parts, color.YellowString("~%d", summary.ToUpdate))
	}
	if summary.ToReplace > 0 {
		parts = append(parts, color.MagentaString("⟳%d", summary.ToReplace))
	}
	if summary.ToDelete > 0 {
		parts = append(parts, color.RedString("-%d", summary.ToDelete))
	}

	if len(parts) == 0 {
		return color.GreenString("No changes")
	}

	return strings.Join(parts, " ")
}
