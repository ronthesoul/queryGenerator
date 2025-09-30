package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	// Get path to input file
	txtPath := readLine(in, "Enter path to .txt file (vendors + keywords): ")
	txtPath = strings.TrimSpace(txtPath)
	if txtPath == "" {
		fmt.Fprintln(os.Stderr, "Error: file path is required.")
		os.Exit(1)
	}
	if _, err := os.Stat(txtPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access %q: %v\n", txtPath, err)
		os.Exit(1)
	}

	// Parse vendors/keywords from file
	vendors, keywords, err := parseVendorsKeywords(txtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse file: %v\n", err)
		os.Exit(1)
	}
	if len(vendors) == 0 || len(keywords) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no vendors or no keywords found in file.")
		os.Exit(1)
	}

	// Get remaining inputs
	proxStr := readLine(in, "Enter proximity number (e.g., 10): ")
	if strings.TrimSpace(proxStr) == "" {
		proxStr = "10"
	}
	outFile := readLine(in, "Enter output filename (e.g., query.txt): ")
	if strings.TrimSpace(outFile) == "" {
		outFile = "query.txt"
	}
	tail := strings.TrimSpace(readLine(in, "Optional tail filter (e.g., NOT crawler: paste*). Leave empty for none: "))

	// Build query
	parts := make([]string, 0, len(vendors)*len(keywords))
	for _, v := range vendors {
		for _, k := range keywords {
			parts = append(parts, fmt.Sprintf("\"%s %s\" ~%s", v, k, proxStr))
		}
	}
	query := strings.Join(parts, " OR ")
	if tail != "" {
		if !strings.HasPrefix(tail, " ") {
			query += " "
		}
		query += tail
	}

	// Write file
	if err := os.WriteFile(outFile, []byte(query), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write %q: %v\n", outFile, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Query built from %s and saved to %s\n", filepath.Base(txtPath), outFile)
	fmt.Println("Preview:")
	fmt.Println(query)
}

func readLine(in *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	text, _ := in.ReadString('\n')
	return strings.TrimSpace(text)
}

// parseVendorsKeywords supports any of these formats in the .txt:
//  1. Block format:
//     vendors:
//     Zoom, Slack, Okta
//     keywords:
//     hacked, breached, 0day, cve
//
//  2. CSV lines (order-agnostic):
//     vendors: Zoom, Slack, Okta
//     keywords: hacked, breached, 0day, cve
//
//  3. Plain lists separated by a blank line (vendors first, keywords second):
//     Zoom, Slack, Okta
//
//     hacked, breached, 0day, cve
//
// Lines starting with '#' are comments and ignored.
func parseVendorsKeywords(path string) ([]string, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var rawLines []string
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rawLines = append(rawLines, line)
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	if len(rawLines) == 0 {
		return nil, nil, errors.New("file is empty after removing comments/blank lines")
	}

	// Try labeled parsing first
	var vendors, keywords []string
	var section string
	for _, line := range rawLines {
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "vendors:"):
			section = "vendors"
			vendors = append(vendors, splitCSV(afterColon(line))...)
		case strings.HasPrefix(lower, "keywords:"):
			section = "keywords"
			keywords = append(keywords, splitCSV(afterColon(line))...)
		default:
			// if we're in a section, keep appending; else we’ll handle plain format later
			if section == "vendors" {
				vendors = append(vendors, splitCSV(line)...)
			} else if section == "keywords" {
				keywords = append(keywords, splitCSV(line)...)
			}
		}
	}
	vendors = normalizeList(vendors)
	keywords = normalizeList(keywords)

	// If labeled parsing worked, return
	if len(vendors) > 0 && len(keywords) > 0 {
		return vendors, keywords, nil
	}

	// Fallback: plain two-block format separated by blank line(s)
	// Re-read file to preserve blank lines for block split
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	blocks := splitOnBlankBlocks(string(contentBytes))
	if len(blocks) >= 2 && len(vendors) == 0 && len(keywords) == 0 {
		vendors = normalizeList(splitCSV(blocks[0]))
		keywords = normalizeList(splitCSV(blocks[1]))
	}

	if len(vendors) == 0 || len(keywords) == 0 {
		return nil, nil, errors.New("could not locate both vendors and keywords in file")
	}
	return vendors, keywords, nil
}

func afterColon(s string) string {
	if i := strings.Index(s, ":"); i >= 0 {
		return strings.TrimSpace(s[i+1:])
	}
	return s
}

func splitCSV(s string) []string {
	// Allow commas or semicolons; also split on whitespace if no commas/semicolons
	s = strings.ReplaceAll(s, ";", ",")
	parts := strings.Split(s, ",")
	if len(parts) == 1 {
		parts = strings.Fields(s)
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitOnBlankBlocks(s string) []string {
	lines := strings.Split(s, "\n")
	var blocks []string
	var cur []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			// blank/comment => block boundary
			if len(cur) > 0 {
				blocks = append(blocks, strings.Join(cur, "\n"))
				cur = nil
			}
			continue
		}
		cur = append(cur, trim)
	}
	if len(cur) > 0 {
		blocks = append(blocks, strings.Join(cur, "\n"))
	}
	return blocks
}

func normalizeList(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, it := range items {
		k := strings.TrimSpace(it)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}
