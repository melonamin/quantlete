// Adapter parity checker.
//
// Ensures methods annotated with //adapter:wasm also have //adapter:http,
// and vice versa, except for explicit allowlisted methods.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type methodKey struct {
	Service string
	Method  string
}

type adapterFlags struct {
	hasWasm bool
	hasHTTP bool
}

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	servicesDir := filepath.Join(root, "internal", "services")
	methods, err := parseAdapters(servicesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing adapters: %v\n", err)
		os.Exit(1)
	}

	if len(methods) == 0 {
		fmt.Println("No //adapter annotations found.")
		return
	}

	allowedWasmOnly := map[methodKey]bool{
		{Service: "ActivityService", Method: "SaveActivity"}:        true,
		{Service: "ActivityService", Method: "SaveStream"}:          true,
		{Service: "GearService", Method: "SaveGear"}:                true,
		{Service: "PhotosService", Method: "SavePhoto"}:             true,
		{Service: "SegmentsService", Method: "SaveSegment"}:         true,
		{Service: "SegmentsService", Method: "SaveSegmentEffort"}:   true,
		{Service: "StatsService", Method: "SaveBestEfforts"}:        true,
	}
	allowedHTTPOnly := map[methodKey]bool{}

	var missingHTTP []methodKey
	var missingWasm []methodKey

	for key, flags := range methods {
		if flags.hasWasm && !flags.hasHTTP && !allowedWasmOnly[key] {
			missingHTTP = append(missingHTTP, key)
		}
		if flags.hasHTTP && !flags.hasWasm && !allowedHTTPOnly[key] {
			missingWasm = append(missingWasm, key)
		}
	}

	if len(missingHTTP) == 0 && len(missingWasm) == 0 {
		fmt.Println("Adapter parity check passed.")
		return
	}

	sort.Slice(missingHTTP, func(i, j int) bool {
		if missingHTTP[i].Service == missingHTTP[j].Service {
			return missingHTTP[i].Method < missingHTTP[j].Method
		}
		return missingHTTP[i].Service < missingHTTP[j].Service
	})
	sort.Slice(missingWasm, func(i, j int) bool {
		if missingWasm[i].Service == missingWasm[j].Service {
			return missingWasm[i].Method < missingWasm[j].Method
		}
		return missingWasm[i].Service < missingWasm[j].Service
	})

	if len(missingHTTP) > 0 {
		fmt.Fprintln(os.Stderr, "Missing HTTP adapters for:")
		for _, key := range missingHTTP {
			fmt.Fprintf(os.Stderr, "  - %s.%s\n", key.Service, key.Method)
		}
	}
	if len(missingWasm) > 0 {
		fmt.Fprintln(os.Stderr, "Missing WASM adapters for:")
		for _, key := range missingWasm {
			fmt.Fprintf(os.Stderr, "  - %s.%s\n", key.Service, key.Method)
		}
	}

	os.Exit(1)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod")
		}
		dir = parent
	}
}

var (
	wasmRe = regexp.MustCompile(`^//adapter:wasm\s+(\w+)(?:\s+category=(\S+))?$`)
	httpRe = regexp.MustCompile(`^//adapter:http\s+(GET|POST|PUT|DELETE|PATCH)\s+(/\S+)$`)
	funcRe = regexp.MustCompile(`^func \(\w+ \*?(\w+)\) (\w+)\(`)
)

func parseAdapters(servicesDir string) (map[methodKey]adapterFlags, error) {
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	methods := make(map[methodKey]adapterFlags)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		filePath := filepath.Join(servicesDir, name)
		if err := parseAdapterFile(filePath, methods); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", name, err)
		}
	}

	return methods, nil
}

func parseAdapterFile(path string, methods map[methodKey]adapterFlags) error {
	file, err := os.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var pendingWasm bool
	var pendingHTTP bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if wasmRe.MatchString(line) {
			pendingWasm = true
			continue
		}
		if httpRe.MatchString(line) {
			pendingHTTP = true
			continue
		}

		matches := funcRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		if !pendingWasm && !pendingHTTP {
			continue
		}

		key := methodKey{Service: matches[1], Method: matches[2]}
		flags := methods[key]
		if pendingWasm {
			flags.hasWasm = true
		}
		if pendingHTTP {
			flags.hasHTTP = true
		}
		methods[key] = flags

		pendingWasm = false
		pendingHTTP = false
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning file: %w", err)
	}
	return nil
}
