package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/kkdai/youtube/v2"
)

func main() {
	// Define flags
	noTimestamps := flag.Bool("no-timestamps", false, "Don't print timestamps")
	filepath := flag.String("o", "", "Output filename (defaults to stdout)")

	// Parse flags
	flag.Parse()

	// Check for YouTube URL argument
	if flag.NArg() < 1 {
		printFancyError("YouTube URL is required")
		os.Exit(1)
	}

	// Extract Transcript
	youtubeClient := youtube.Client{}

	video, err := youtubeClient.GetVideo(flag.Arg(0))
	if err != nil {
		printFancyError(fmt.Sprintf("failed to get video info: %v", err))
		os.Exit(1)
	}

	transcript, err := youtubeClient.GetTranscript(video, "en")
	if err != nil {
		printFancyError(fmt.Sprintf("failed to get transcript info: %v", err))
	}

	var transcriptStr strings.Builder
	if !*noTimestamps {
		transcriptStr.WriteString(transcript.String())
	} else {
		for _, tr := range transcript {
			transcriptStr.WriteString(strings.TrimSpace(tr.Text) + "\n")
		}
	}

	if *filepath == "" {
		fmt.Println(transcriptStr.String())
	} else {
		if err = os.WriteFile(*filepath, []byte(transcriptStr.String()), 0644); err != nil {
			printFancyError(fmt.Sprintf("failed to write cleaned transcript: %v", err))
		}
	}
}

func printFancyError(message string) {
	errorEmoji := "❌"
	errorColor := color.New(color.FgRed, color.Bold)
	_, _ = errorColor.Fprintf(os.Stderr, "%s Error: %s\n", errorEmoji, message)
}
