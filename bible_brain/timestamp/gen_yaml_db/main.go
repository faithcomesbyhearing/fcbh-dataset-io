package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Testament  string
	TextType   string
	StreamType string
	OutputDir  string
	Template   string
	BibleId    string
	Verbose    bool
}

func main() {
	var config Config

	flag.StringVar(&config.Testament, "testament", "", "Testament scope: n1|n2|o1|o2")
	flag.StringVar(&config.TextType, "text", "", "Text format: usx|plain")
	flag.StringVar(&config.StreamType, "stream", "hls", "Stream format: hls|dash (default: hls)")
	flag.StringVar(&config.OutputDir, "output", "", "Output directory")
	flag.StringVar(&config.Template, "template", "", "Custom template file (optional)")
	flag.StringVar(&config.BibleId, "bible", "", "Generate for specific Bible ID (optional)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")

	flag.Parse()

	// Validate required arguments
	if config.Testament == "" || config.TextType == "" || config.OutputDir == "" {
		fmt.Println("Usage: yaml_generator -testament {n1|n2|o1|o2} -text {usx|plain} -output <dir> [-stream {hls|dash}] [-template <file>] [-bible <id>] [-verbose]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  ./yaml_generator -testament n1 -text usx -output ./n1_usx/")
		fmt.Println("  ./yaml_generator -testament n2 -text plain -stream dash -output ./n2_plain/ -verbose")
		fmt.Println("  ./yaml_generator -testament n1 -text usx -template ./my_template.yaml -output ./custom/")
		fmt.Println("  ./yaml_generator -testament n1 -text usx -bible ABPWBT -output ./single/")
		os.Exit(1)
	}

	// Validate testament and text type
	if !isValidTestament(config.Testament) {
		log.Fatalf("Invalid testament: %s. Must be one of: n1, n2, o1, o2", config.Testament)
	}

	if !isValidTextType(config.TextType) {
		log.Fatalf("Invalid text type: %s. Must be one of: usx, plain", config.TextType)
	}

	if !isValidStreamType(config.StreamType) {
		log.Fatalf("Invalid stream type: %s. Must be one of: hls, dash", config.StreamType)
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory %s: %v", config.OutputDir, err)
	}

	// Initialize generator
	generator, err := NewYAMLGenerator(config)
	if err != nil {
		log.Fatalf("Failed to initialize generator: %v", err)
	}

	// Generate YAMLs
	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate YAMLs: %v", err)
	}

	fmt.Printf("YAML generation completed successfully!\n")
}

func isValidTestament(testament string) bool {
	switch testament {
	case "n1", "n2", "o1", "o2":
		return true
	default:
		return false
	}
}

func isValidTextType(textType string) bool {
	switch textType {
	case "usx", "plain":
		return true
	default:
		return false
	}
}

func isValidStreamType(streamType string) bool {
	switch streamType {
	case "hls", "dash":
		return true
	default:
		return false
	}
}
