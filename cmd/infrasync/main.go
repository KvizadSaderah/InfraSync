package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/kvizadsaderah/infrasync/pkg/analyzer"
	"github.com/kvizadsaderah/infrasync/pkg/formatter"
	"github.com/kvizadsaderah/infrasync/pkg/parser"
)

var (
	version = "0.2.0"
)

func main() {
	// Define flags
	outputFormat := flag.String("format", "cli", "Output format: cli, markdown")
	showVersion := flag.Bool("version", false, "Show version")
	showUnchanged := flag.Bool("show-unchanged", false, "Show unchanged resources")
	verbose := flag.Bool("verbose", false, "Verbose output with all attributes")
	compact := flag.Bool("compact", false, "Compact output")
	showWarnings := flag.Bool("warnings", true, "Show security and risk warnings")
	outputFile := flag.String("output", "", "Write output to file instead of stdout")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "InfraSync - Beautiful Terraform Plan Analysis\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <tfplan.json>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Analyze plan with colored CLI output\n")
		fmt.Fprintf(os.Stderr, "  %s tfplan.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate markdown for GitHub PR comment\n")
		fmt.Fprintf(os.Stderr, "  %s --format markdown tfplan.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Save output to file\n")
		fmt.Fprintf(os.Stderr, "  %s --format markdown --output plan.md tfplan.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Verbose output with all attribute changes\n")
		fmt.Fprintf(os.Stderr, "  %s --verbose tfplan.json\n\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("InfraSync version %s\n", version)
		os.Exit(0)
	}

	// Check for plan file argument
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	planFile := flag.Arg(0)

	// Parse the plan
	summary, err := parser.ParsePlanFile(planFile)
	if err != nil {
		color.Red("Error parsing plan: %v", err)
		os.Exit(1)
	}

	// Analyze for warnings
	var warnings []analyzer.Warning
	if *showWarnings {
		warnings = analyzer.AnalyzeChanges(summary)
	}

	// Format and output
	var output string

	switch *outputFormat {
	case "cli":
		cliFormatter := formatter.NewCLIFormatter(*showUnchanged, *verbose)

		// Print to stdout (this is the visual output)
		cliFormatter.Format(summary)

		// Print warnings
		if len(warnings) > 0 {
			printWarningsCLI(warnings)
		}

		// Print general warnings
		cliFormatter.PrintWarnings(summary)

		// Exit with appropriate code
		exitCode := determineExitCode(summary, warnings)
		os.Exit(exitCode)

	case "markdown":
		mdFormatter := formatter.NewMarkdownFormatter(!*compact, *compact, *showUnchanged)
		output = mdFormatter.Format(summary)

		// Add warnings section
		if len(warnings) > 0 {
			output += "\n" + formatWarningsMarkdown(warnings)
		}

		// Write to file or stdout
		if *outputFile != "" {
			err := os.WriteFile(*outputFile, []byte(output), 0644)
			if err != nil {
				color.Red("Error writing output file: %v", err)
				os.Exit(1)
			}
			color.Green("✓ Output written to %s", *outputFile)
		} else {
			fmt.Print(output)
		}

	default:
		color.Red("Unknown format: %s", *outputFormat)
		color.Yellow("Supported formats: cli, markdown")
		os.Exit(1)
	}
}

func printWarningsCLI(warnings []analyzer.Warning) {
	if len(warnings) == 0 {
		return
	}

	fmt.Printf("\n")
	color.New(color.BgYellow, color.FgBlack, color.Bold).Printf(" 🔍 SECURITY & RISK ANALYSIS ")
	fmt.Printf("\n\n")

	criticals := filterWarnings(warnings, analyzer.RiskCritical)
	highs := filterWarnings(warnings, analyzer.RiskHigh)
	mediums := filterWarnings(warnings, analyzer.RiskMedium)

	if len(criticals) > 0 {
		color.Red("🚨 CRITICAL WARNINGS (%d):", len(criticals))
		color.Red("─────────────────────────")
		for _, w := range criticals {
			color.Red("  • %s", w.Message)
			color.HiBlack("    Resource: %s", w.Resource)
			color.HiBlack("    %s", w.Explanation)
		}
		fmt.Printf("\n")
	}

	if len(highs) > 0 {
		color.Yellow("⚠️  HIGH RISK WARNINGS (%d):", len(highs))
		color.Yellow("──────────────────────────")
		for _, w := range highs {
			color.Yellow("  • %s", w.Message)
			color.HiBlack("    Resource: %s", w.Resource)
			color.HiBlack("    %s", w.Explanation)
		}
		fmt.Printf("\n")
	}

	if len(mediums) > 0 {
		color.Cyan("ℹ️  MEDIUM RISK WARNINGS (%d):", len(mediums))
		color.Cyan("────────────────────────────")
		for _, w := range mediums {
			color.Cyan("  • %s", w.Message)
			color.HiBlack("    Resource: %s", w.Resource)
		}
		fmt.Printf("\n")
	}
}

func formatWarningsMarkdown(warnings []analyzer.Warning) string {
	if len(warnings) == 0 {
		return ""
	}

	output := "\n### 🔍 Security & Risk Analysis\n\n"

	criticals := filterWarnings(warnings, analyzer.RiskCritical)
	highs := filterWarnings(warnings, analyzer.RiskHigh)

	if len(criticals) > 0 {
		output += "#### 🚨 Critical Warnings\n\n"
		for _, w := range criticals {
			output += fmt.Sprintf("- **%s**\n", w.Message)
			output += fmt.Sprintf("  - Resource: `%s`\n", w.Resource)
			output += fmt.Sprintf("  - %s\n", w.Explanation)
		}
		output += "\n"
	}

	if len(highs) > 0 {
		output += "#### ⚠️ High Risk Warnings\n\n"
		for _, w := range highs {
			output += fmt.Sprintf("- **%s**\n", w.Message)
			output += fmt.Sprintf("  - Resource: `%s`\n", w.Resource)
			output += fmt.Sprintf("  - %s\n", w.Explanation)
		}
		output += "\n"
	}

	return output
}

func filterWarnings(warnings []analyzer.Warning, level analyzer.RiskLevel) []analyzer.Warning {
	result := make([]analyzer.Warning, 0)
	for _, w := range warnings {
		if w.Level == level {
			result = append(result, w)
		}
	}
	return result
}

func determineExitCode(summary *parser.PlanSummary, warnings []analyzer.Warning) int {
	// Exit 0 if no changes
	if summary.ToCreate == 0 && summary.ToUpdate == 0 && summary.ToDelete == 0 && summary.ToReplace == 0 {
		return 0
	}

	// Exit 2 if critical warnings
	for _, w := range warnings {
		if w.Level == analyzer.RiskCritical {
			return 2
		}
	}

	// Exit 1 if there are changes
	return 1
}
