package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/faithcomesbyhearing/fcbh-dataset-io/db"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/revise_audio"
	"github.com/faithcomesbyhearing/fcbh-dataset-io/utility/ffmpeg"
)

func main() {
	ctx := context.Background()

	// Ensure GOPROJ is set
	if os.Getenv("GOPROJ") == "" {
		wd, _ := os.Getwd()
		if filepath.Base(wd) == "test_vits_jude" {
			repoRoot := filepath.Join(wd, "..", "..", "..")
			os.Setenv("GOPROJ", repoRoot)
		} else if filepath.Base(filepath.Dir(wd)) == "cmd" {
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

	basePath := os.Getenv("FCBH_DATASET_FILES")
	if basePath == "" {
		fmt.Println("Error: FCBH_DATASET_FILES not set")
		os.Exit(1)
	}

	dbPath := "/Users/jrstear/fcbh/arti/old/engniv2011/00005/database/engniv2011.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	fmt.Println("=== So-VITS-SVC Test: Jude 1:1 ===\n")

	// Step 1: Extract training data (if model doesn't exist)
	modelDir := filepath.Join(soVitsRoot, "logs", "44k")
	modelPath := filepath.Join(modelDir, "G_*.pth")
	
	// Check if model exists
	matches, _ := filepath.Glob(modelPath)
	if len(matches) == 0 {
		fmt.Println("No trained model found. Need to train first.")
		fmt.Println("This will extract training data and set up training...")
		
		// Extract training data
		trainingDataDir := filepath.Join(soVitsRoot, "dataset_raw")
		os.MkdirAll(trainingDataDir, 0755)
		
		goproj := os.Getenv("GOPROJ")
		prepareScript := filepath.Join(goproj, "revise_audio", "vits", "python", "prepare_training_data.py")
		
		pythonPath := os.Getenv("FCBH_VITS_PYTHON")
		if pythonPath == "" {
			// Try conda
			if condaPrefix := os.Getenv("CONDA_PREFIX"); condaPrefix != "" {
				pythonPath = filepath.Join(condaPrefix, "bin", "python")
			} else {
				pythonPath = "python3"
			}
		}
		
		fmt.Printf("Extracting training data from Jude...\n")
		cmd := exec.Command(pythonPath, prepareScript, dbPath, basePath, "JUD", "1", trainingDataDir, "jude_narrator")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error extracting training data: %v\n", err)
			fmt.Println("\nNote: Training data extraction failed. You may need to:")
			fmt.Println("1. Manually extract audio segments")
			fmt.Println("2. Place them in dataset_raw/jude_narrator/")
			fmt.Println("3. Run So-VITS-SVC preprocessing and training")
			os.Exit(1)
		}
		
		fmt.Println("\nTraining data extracted. Next steps:")
		fmt.Println("1. cd $SO_VITS_SVC_ROOT")
		fmt.Println("2. python resample.py")
		fmt.Println("3. python preprocess_flist_config.py")
		fmt.Println("4. python preprocess_hubert_f0.py")
		fmt.Println("5. python train.py -c configs/config.json")
		fmt.Println("\nOr use the training script in revise_audio/vits/python/")
		os.Exit(0)
	}

	// Step 2: Test inference on Jude 1:1
	fmt.Println("Model found. Testing inference on Jude 1:1...\n")

	// Open database
	conn := db.NewDBAdapter(ctx, dbPath)

	// Get original chapter audio
	query := `SELECT audio_file FROM scripts WHERE book_id = 'JUD' AND chapter_num = 1 AND audio_file != '' LIMIT 1`
	var audioFileName string
	err := conn.DB.QueryRow(query).Scan(&audioFileName)
	if err != nil {
		fmt.Printf("Error getting chapter audio: %v\n", err)
		os.Exit(1)
	}

	originalAudioFile := filepath.Join(basePath, "engniv2011/n1da", audioFileName)
	if _, err := os.Stat(originalAudioFile); err != nil {
		filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == audioFileName {
				originalAudioFile = path
				return filepath.SkipAll
			}
			return nil
		})
	}

	// Extract verse 1 segment
	fmt.Println("1. Extracting Jude 1:1 segment...")
	verseQuery := `SELECT script_begin_ts, script_end_ts FROM scripts 
		WHERE book_id = 'JUD' AND chapter_num = 1 AND verse_str = '1' LIMIT 1`
	var beginTS, endTS float64
	err = conn.DB.QueryRow(verseQuery).Scan(&beginTS, &endTS)
	if err != nil {
		fmt.Printf("Error getting verse timestamps: %v\n", err)
		os.Exit(1)
	}

	tempDir := os.TempDir()
	
	// ChopOneSegment signature: (ctx, tempDir, inputFile, beginTS, endTS) -> outputFile
	verseSegmentPath, status := ffmpeg.ChopOneSegment(ctx, tempDir, originalAudioFile, beginTS, endTS)
	if status != nil {
		fmt.Printf("Error extracting segment: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Extracted: %s (%.2fs - %.2fs)\n", verseSegmentPath, beginTS, endTS)

	// Step 3: Use original audio segment for voice conversion test
	// (In production, this would be TTS-generated audio from revised text)
	fmt.Println("\n2. Using original audio segment for voice conversion test...")
	fmt.Println("   (Note: In production, this would be TTS-generated audio from revised text)")
	sourceAudioPath := verseSegmentPath
	fmt.Printf("   Source audio: %s\n", sourceAudioPath)

	// Step 4: Voice convert using So-VITS-SVC
	fmt.Println("\n3. Voice converting with So-VITS-SVC...")
	
	// Use specific checkpoint if provided as argument, otherwise find latest
	var latestModel string
	if len(os.Args) > 2 {
		// Use specific checkpoint provided as second argument (e.g., "1600")
		specificCheckpoint := filepath.Join(modelDir, fmt.Sprintf("G_%s.pth", os.Args[2]))
		if _, err := os.Stat(specificCheckpoint); err == nil {
			latestModel = specificCheckpoint
			fmt.Printf("   Using specified checkpoint: %s\n", filepath.Base(latestModel))
		} else {
			fmt.Printf("Warning: Specified checkpoint not found: %s, using latest instead\n", specificCheckpoint)
		}
	}
	
	// If no specific checkpoint or it wasn't found, find the latest
	if latestModel == "" {
		matches, _ := filepath.Glob(filepath.Join(modelDir, "G_*.pth"))
		if len(matches) == 0 {
			fmt.Println("Error: No model checkpoint found")
			os.Exit(1)
		}
		
		// Get the latest checkpoint
		var latestEpoch int
		for _, match := range matches {
			// Extract epoch number from filename like G_100.pth
			var epoch int
			fmt.Sscanf(filepath.Base(match), "G_%d.pth", &epoch)
			if epoch > latestEpoch {
				latestEpoch = epoch
				latestModel = match
			}
		}
		fmt.Printf("   Using latest checkpoint: %s\n", filepath.Base(latestModel))
	}
	
	configPath := filepath.Join(soVitsRoot, "configs", "config.json")
	if _, err := os.Stat(configPath); err != nil {
		// Try logs/44k/config.json
		configPath = filepath.Join(modelDir, "config.json")
	}
	
	if _, err := os.Stat(configPath); err != nil {
		fmt.Printf("Error: Config file not found: %s\n", configPath)
		os.Exit(1)
	}

	vcConfig := revise_audio.VoiceConversionConfig{
		F0Method: "rmvpe",
		Device:   "auto",
	}
	
	vitsAdapter := revise_audio.NewVITSAdapter(ctx, vcConfig)
	defer vitsAdapter.Close()

	// Convert voice (speaker name from config - typically "jude_narrator" or "speaker0")
	speakerName := "jude_narrator" // This should match the speaker name in the trained model
	convertedPath, status := vitsAdapter.ConvertVoice(sourceAudioPath, latestModel, configPath, speakerName)
	if status != nil {
		fmt.Printf("Error converting voice: %v\n", status)
		os.Exit(1)
	}
	fmt.Printf("   Converted: %s\n", convertedPath)

	// Step 5: Save final result
	outputDir := os.Getenv("HOME")
	if outputDir == "" {
		outputDir = "/tmp"
	}
	outputPath := filepath.Join(outputDir, "tmp", "arti_revised_audio", "JUD_01_v1_vits.wav")
	os.MkdirAll(filepath.Dir(outputPath), 0755)
	
	// Copy converted file to output
	cmd := exec.Command("cp", convertedPath, outputPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error copying output: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("\n✅ Complete! Output saved to: %s\n", outputPath)
	fmt.Println("\nNote: This is a test. For production, you would also:")
	fmt.Println("- Match prosody to surrounding context")
	fmt.Println("- Stitch into the full chapter audio")
}

