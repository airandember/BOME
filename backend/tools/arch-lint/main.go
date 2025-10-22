package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ArchRule represents an architecture rule from the JSON config
type ArchRule struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Severity    string          `json:"severity"`
	Description string          `json:"description"`
	Pattern     json.RawMessage `json:"pattern"`
	Rationale   string          `json:"rationale"`
}

// ArchRules represents the full architecture rules configuration
type ArchRules struct {
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Rules       []ArchRule `json:"rules"`
}

// Violation represents an architecture rule violation
type Violation struct {
	RuleID    string `json:"rule_id"`
	Severity  string `json:"severity"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Message   string `json:"message"`
	Rationale string `json:"rationale"`
}

// Report represents the final violation report
type Report struct {
	TotalFiles      int         `json:"total_files"`
	TotalViolations int         `json:"total_violations"`
	Errors          int         `json:"errors"`
	Warnings        int         `json:"warnings"`
	Info            int         `json:"info"`
	Violations      []Violation `json:"violations"`
}

var (
	rulesFile  = flag.String("rules", "architecture-rules.json", "Path to architecture rules JSON file")
	outputFile = flag.String("output", "", "Output file for violations (default: stdout)")
	verbose    = flag.Bool("verbose", false, "Verbose output")
	failOnWarn = flag.Bool("fail-on-warning", false, "Fail on warnings (default: only errors)")
)

func main() {
	flag.Parse()

	// Load rules
	rules, err := loadRules(*rulesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading rules: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Loaded %d architecture rules\n", len(rules.Rules))
	}

	// Scan all Go files
	violations := []Violation{}
	filesScanned := 0

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, node_modules, and hidden directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		filesScanned++

		// Check file against rules
		fileViolations := checkFile(path, rules)
		violations = append(violations, fileViolations...)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning files: %v\n", err)
		os.Exit(1)
	}

	// Generate report
	report := Report{
		TotalFiles:      filesScanned,
		TotalViolations: len(violations),
		Violations:      violations,
	}

	for _, v := range violations {
		switch v.Severity {
		case "error":
			report.Errors++
		case "warning":
			report.Warnings++
		case "info":
			report.Info++
		}
	}

	// Output report
	if *outputFile != "" {
		outputJSON, _ := json.MarshalIndent(report, "", "  ")
		os.WriteFile(*outputFile, outputJSON, 0644)
	}

	// Print summary
	printReport(report)

	// Exit code
	if report.Errors > 0 {
		os.Exit(1)
	}
	if *failOnWarn && report.Warnings > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

func loadRules(path string) (*ArchRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rules ArchRules
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}

func checkFile(path string, rules *ArchRules) []Violation {
	violations := []Violation{}

	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		// Skip files that can't be parsed
		return violations
	}

	// Extract imports
	imports := []string{}
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		imports = append(imports, importPath)
	}

	// Check each rule
	for _, rule := range rules.Rules {
		ruleViolations := checkRule(path, imports, rule)
		violations = append(violations, ruleViolations...)
	}

	return violations
}

func checkRule(path string, imports []string, rule ArchRule) []Violation {
	violations := []Violation{}

	// Simple pattern matching for key rules
	switch rule.ID {
	case "BRAID-001":
		// No cross-braid imports in handlers
		if strings.Contains(path, "/handlers/") && !strings.Contains(path, "_test.go") {
			currentBraid := extractBraid(path)
			for _, imp := range imports {
				if strings.Contains(imp, "bome-backend/") {
					// Check if importing from another braid
					if isCrossBraidImport(imp, currentBraid) {
						violations = append(violations, Violation{
							RuleID:    rule.ID,
							Severity:  rule.Severity,
							File:      path,
							Line:      1, // Would need more parsing for exact line
							Message:   fmt.Sprintf("Handler imports from another braid: %s", imp),
							Rationale: rule.Rationale,
						})
					}
				}
			}
		}

	case "BRAID-003":
		// Use cases cannot import HTTP packages
		if strings.Contains(path, "/usecases/") {
			for _, imp := range imports {
				if strings.Contains(imp, "gin-gonic") || strings.Contains(imp, "net/http") {
					violations = append(violations, Violation{
						RuleID:    rule.ID,
						Severity:  rule.Severity,
						File:      path,
						Line:      1,
						Message:   fmt.Sprintf("Use case imports HTTP package: %s", imp),
						Rationale: rule.Rationale,
					})
				}
			}
		}

	case "BRAID-004":
		// Models cannot import services
		if strings.Contains(path, "/models/") {
			for _, imp := range imports {
				if strings.Contains(imp, "/services/") || strings.Contains(imp, "/usecases/") || strings.Contains(imp, "/handlers/") {
					violations = append(violations, Violation{
						RuleID:    rule.ID,
						Severity:  rule.Severity,
						File:      path,
						Line:      1,
						Message:   fmt.Sprintf("Model imports from higher layer: %s", imp),
						Rationale: rule.Rationale,
					})
				}
			}
		}

	case "NAMING-001":
		// Use case naming convention
		if strings.Contains(path, "/usecases/") && !strings.Contains(path, "_test.go") {
			filename := filepath.Base(path)
			matched, _ := regexp.MatchString(`^[a-z]+_[a-z]+\.go$`, filename)
			if !matched && filename != "usecases.go" {
				violations = append(violations, Violation{
					RuleID:    rule.ID,
					Severity:  rule.Severity,
					File:      path,
					Line:      1,
					Message:   fmt.Sprintf("Use case file doesn't follow naming convention: %s", filename),
					Rationale: rule.Rationale,
				})
			}
		}
	}

	return violations
}

func extractBraid(path string) string {
	// Extract braid name from path like: authentication/handlers/auth.go -> authentication
	parts := strings.Split(path, string(os.PathSeparator))
	for i, part := range parts {
		if part == "handlers" && i > 0 {
			return parts[i-1]
		}
		if part == "models" && i > 0 {
			return parts[i-1]
		}
		if part == "usecases" && i > 0 {
			return parts[i-1]
		}
	}
	return ""
}

func isCrossBraidImport(imp string, currentBraid string) bool {
	// List of braids
	braids := []string{
		"authentication",
		"subscription",
		"video-streaming",
		"user-management",
		"content",
		"admin",
		"analytics",
		"communication",
	}

	for _, braid := range braids {
		if braid != currentBraid && strings.Contains(imp, "/"+braid+"/") {
			return true
		}
	}

	return false
}

func printReport(report Report) {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              ARCHITECTURE LINT REPORT                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n📊 Scanned: %d files\n", report.TotalFiles)
	fmt.Printf("🔍 Total Violations: %d\n", report.TotalViolations)
	fmt.Printf("   ❌ Errors: %d\n", report.Errors)
	fmt.Printf("   ⚠️  Warnings: %d\n", report.Warnings)
	fmt.Printf("   ℹ️  Info: %d\n\n", report.Info)

	if len(report.Violations) > 0 {
		fmt.Println("📋 Violations:")
		fmt.Println("─────────────────────────────────────────────────────────────")
		for _, v := range report.Violations {
			icon := "ℹ️"
			if v.Severity == "error" {
				icon = "❌"
			} else if v.Severity == "warning" {
				icon = "⚠️"
			}
			fmt.Printf("\n%s [%s] %s\n", icon, v.RuleID, v.File)
			fmt.Printf("   %s\n", v.Message)
			fmt.Printf("   💡 %s\n", v.Rationale)
		}
		fmt.Println("\n─────────────────────────────────────────────────────────────")
	} else {
		fmt.Println("✅ No violations found! Architecture is clean! 🎉")
	}

	fmt.Println()
}
