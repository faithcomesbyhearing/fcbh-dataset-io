package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/revise_audio"
)

func main() {
	ctx := context.Background()

	// Ensure GOPROJ is set
	if os.Getenv("GOPROJ") == "" {
		wd, _ := os.Getwd()
		if filepath.Base(filepath.Dir(wd)) == "cmd" {
			repoRoot := filepath.Join(wd, "..", "..")
			os.Setenv("GOPROJ", repoRoot)
		} else {
			os.Setenv("GOPROJ", wd)
		}
	}

	// Check environment variables
	soVitsRoot := os.Getenv("SO_VITS_SVC_ROOT")
	if soVitsRoot == "" {
		fmt.Println("Error: SO_VITS_SVC_ROOT not set")
		os.Exit(1)
	}

	// Input: raw TTS segment
	inputFile := filepath.Join(os.Getenv("HOME"), "tmp", "arti_revised_audio", "JUD_01_v1_raw_tts_segment.wav")
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
	}

	if _, err := os.Stat(inputFile); err != nil {
		fmt.Printf("Error: Input file not found: %s\n", inputFile)
		os.Exit(1)
	}

	fmt.Println("=== Voice Conversion: TTS Segment ===\n")
	fmt.Printf("Input: %s\n", inputFile)

	// Find latest checkpoint
	modelDir := filepath.Join(soVitsRoot, "logs", "44k")
	checkpointArg := "13600" // Latest checkpoint
	if len(os.Args) > 2 {
		checkpointArg = os.Args[2]
	}
	modelPath := filepath.Join(modelDir, fmt.Sprintf("G_%s.pth", checkpointArg))
	if _, err := os.Stat(modelPath); err != nil {
		fmt.Printf("Error: Model checkpoint not found: %s\n", modelPath)
		os.Exit(1)
	}
	fmt.Printf("Model: %s\n", filepath.Base(modelPath))

	// Find config
	configPath := filepath.Join(soVitsRoot, "configs", "config.json")
	if _, err := os.Stat(configPath); err != nil {
		configPath = filepath.Join(modelDir, "config.json")
	}
	if _, err := os.Stat(configPath); err != nil {
		fmt.Printf("Error: Config file not found: %s\n", configPath)
		os.Exit(1)
	}
	fmt.Printf("Config: %s\n", configPath)

	// Setup VITS adapter
	vcConfig := revise_audio.VoiceConversionConfig{
		F0Method: "rmvpe",
		Device:   "auto",
	}

	vitsAdapter := revise_audio.NewVITSAdapter(ctx, vcConfig)
	defer vitsAdapter.Close()

	speakerName := "jude_narrator"
	fmt.Printf("Speaker: %s\n", speakerName)

	// Convert voice
	fmt.Println("\nConverting voice...")
	convertedPath, status := vitsAdapter.ConvertVoice(inputFile, modelPath, configPath, speakerName)
	if status != nil {
		fmt.Printf("Error converting voice: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("Converted: %s\n", convertedPath)

	// Save to output directory
	outputDir := filepath.Join(os.Getenv("HOME"), "tmp", "arti_revised_audio")
	os.MkdirAll(outputDir, 0755)
	outputWav := filepath.Join(outputDir, "JUD_01_v1_tts_voice_converted.wav")
	outputMp3 := filepath.Join(outputDir, "JUD_01_v1_tts_voice_converted.mp3")

	// Copy WAV
	data, err := os.ReadFile(convertedPath)
	if err != nil {
		fmt.Printf("Error reading converted file: %v\n", err)
		os.Exit(1)
	}
	err = os.WriteFile(outputWav, data, 0644)
	if err != nil {
		fmt.Printf("Error writing WAV: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("WAV saved: %s\n", outputWav)

	// Convert to MP3
	fmt.Println("\nConverting to MP3...")
	cmd := exec.Command("ffmpeg", "-i", outputWav, "-codec:a", "libmp3lame", "-b:a", "192k", outputMp3, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error converting to MP3: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Complete! Output saved to: %s\n", outputMp3)
}

