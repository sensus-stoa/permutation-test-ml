package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const version = "0.1.0"

func main() {
	rootFl := flag.String("root", ".", "repository root (directory containing trust/)")
	outFl := flag.String("out", "", "report output path (default <root>/trust_check_report.json)")
	sgnFl := flag.Bool("signoff", false, "interactive human sign-off (requires a green auto-check)")
	signFl := flag.String("signer", "", "signer name for --signoff")
	flag.Parse()

	root, err := filepath.Abs(*rootFl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	rep, err := runChecks(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	outPath := *outFl
	if outPath == "" {
		outPath = filepath.Join(root, "trust_check_report.json")
	}
	if err := writeReport(outPath, rep); err != nil {
		fmt.Fprintln(os.Stderr, "error writing report:", err)
		os.Exit(2)
	}
	printReport(rep, outPath)

	if rep.Overall != "PASS" {
		os.Exit(1)
	}

	if *sgnFl {
		if err := doSignoff(root, *signFl, outPath, rep); err != nil {
			fmt.Fprintln(os.Stderr, "sign-off aborted:", err)
			os.Exit(1)
		}
	}
}

func printReport(rep *Report, outPath string) {
	for _, c := range rep.Checks {
		fmt.Printf("  [%s] %s: %s\n", c.Status, c.ID, c.Detail)
	}
	fmt.Printf("overall: %s (%d checks) -> %s\n", rep.Overall, len(rep.Checks), outPath)
}
