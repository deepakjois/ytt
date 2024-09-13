package main

import (
	"flag"
	"fmt"
	"html"
	"os"
	"strings"

	"github.com/fatih/color"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("%s <youtube_url>\n", os.Args[0])
		flag.PrintDefaults()
	}

	// Define flags
	noTimestamps := flag.Bool("no-timestamps", false, "Don't print timestamps")
	filepath := flag.String("o", "", "Output filename (defaults to stdout)")
	lang := flag.String("lang", "en", "Language code for the desired transcript")

	// Parse flags
	flag.Parse()

	// Check for YouTube URL argument
	if flag.NArg() < 1 {
		printFancyError("YouTube URL is required")
		os.Exit(1)
	}

	videoID, err := ExtractVideoID(flag.Arg(0))
	if err != nil {
		printFancyError(fmt.Sprintf("failed to extract video ID: %v", err))
		os.Exit(1)
	}

	transcriptList, err := ListTranscripts(videoID)
	if err != nil {
		printFancyError(fmt.Sprintf("failed to list transcripts: %v", err))
		os.Exit(1)
	}

	// Choose the transcript with the specified language code
	var transcript *Transcript
	if *lang != "" {
		transcript, err = transcriptList.FindTranscript(*lang)
		if err != nil {
			printFancyError(fmt.Sprintf("No transcript found for language code '%s'", *lang))
			os.Exit(1)
		}
	} else {
		// If no language specified, choose the first available transcript
		for _, t := range transcriptList.ManuallyCreatedTranscripts {
			transcript = t
			break
		}
		if transcript == nil {
			for _, t := range transcriptList.GeneratedTranscripts {
				transcript = t
				break
			}
		}
	}
	if transcript == nil {
		printFancyError("No transcript available")
		os.Exit(1)
	}

	entries, err := transcript.Fetch()
	if err != nil {
		printFancyError(fmt.Sprintf("Failed to fetch transcript: %v", err))
		os.Exit(1)
	}

	var sb strings.Builder
	for _, entry := range entries {
		if !*noTimestamps {
			sb.WriteString(fmt.Sprintf("%.2f:%.2f\t", entry.Start, entry.Start+entry.Duration))
		}
		sb.WriteString(html.UnescapeString(entry.Text))
		sb.WriteString("\n")
	}

	output := sb.String()

	if *filepath != "" {
		err = os.WriteFile(*filepath, []byte(output), 0644)
		if err != nil {
			printFancyError(fmt.Sprintf("Failed to write to file: %v", err))
			os.Exit(1)
		}
		fmt.Printf("Transcript written to %s\n", *filepath)
	} else {
		fmt.Print(output)
	}

}

func printFancyError(message string) {
	errorEmoji := "❌"
	errorColor := color.New(color.FgRed, color.Bold)
	_, _ = errorColor.Fprintf(os.Stderr, "%s Error: %s\n", errorEmoji, message)
}
