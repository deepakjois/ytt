package ytt

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var videoRegexpList = []*regexp.Regexp{
	regexp.MustCompile(`(?:v|embed|shorts|watch\?v)(?:=|/)([^"&?/=%]{11})`),
	regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
	regexp.MustCompile(`([^"&?/=%]{11})`),
}

var (
	ErrInvalidCharactersInVideoID = errors.New("invalid characters in video id")
	ErrVideoIDMinLength           = errors.New("the video id must be at least 10 characters long")
	ErrTranscriptsDisabled        = errors.New("transcripts disabled")
	ErrTranscriptsUnavailable     = errors.New("transcripts disabled or video unavailable")
	ErrNoTranscriptFound          = errors.New("no transcript found for the given language codes")
	ErrInvalidFormat              = errors.New("invalid captions tracks format")
)

// ExtractVideoID extracts the videoID from the given string for a YouTube URL.
func ExtractVideoID(videoID string) (string, error) {
	if strings.Contains(videoID, "youtu") || strings.ContainsAny(videoID, "\"?&/<%=") {
		for _, re := range videoRegexpList {
			if matches := re.FindStringSubmatch(videoID); matches != nil {
				videoID = matches[1]
				break
			}
		}
	}

	if strings.ContainsAny(videoID, "?&/<%=") {
		return "", ErrInvalidCharactersInVideoID
	}

	if len(videoID) < 10 {
		return "", ErrVideoIDMinLength
	}

	return videoID, nil
}

const (
	watchURL = "https://www.youtube.com/watch?v=%s"
)

// TranscriptList represents a list of transcripts for a YouTube video.
type TranscriptList struct {
	VideoID                    string
	ManuallyCreatedTranscripts map[string]*Transcript
	GeneratedTranscripts       map[string]*Transcript
}

// Transcript represents a transcript for a YouTube video.
type Transcript struct {
	VideoID      string
	URL          string
	Language     string
	LanguageCode string
	IsGenerated  bool
}

// TranscriptEntry represents a transcript entry for a YouTube video.
type TranscriptEntry struct {
	Text     string  `xml:",chardata"`
	Start    float64 `xml:"start,attr"`
	Duration float64 `xml:"dur,attr"`
}

// ListTranscripts lists the transcripts for the given videoID.
func ListTranscripts(videoID string) (*TranscriptList, error) {
	html, err := fetchVideoHTML(videoID)
	if err != nil {
		return nil, err
	}

	captionsJSON, err := extractCaptionsJSON(html, videoID)
	if err != nil {
		return nil, err
	}

	return buildTranscriptList(videoID, captionsJSON)
}

func fetchVideoHTML(videoID string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(watchURL, videoID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func extractCaptionsJSON(html, videoID string) (map[string]interface{}, error) {
	parts := strings.Split(html, `"captions":`)
	if len(parts) <= 1 {
		return nil, ErrTranscriptsUnavailable
	}

	jsonPart := strings.Split(parts[1], `,"videoDetails"`)[0]
	jsonPart = strings.ReplaceAll(jsonPart, "\n", "")

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonPart), &result)
	if err != nil {
		return nil, err
	}

	captionsJSON, ok := result["playerCaptionsTracklistRenderer"].(map[string]interface{})
	if !ok {
		return nil, ErrTranscriptsDisabled
	}

	return captionsJSON, nil
}

func buildTranscriptList(videoID string, captionsJSON map[string]interface{}) (*TranscriptList, error) {
	manualTranscripts := make(map[string]*Transcript)
	generatedTranscripts := make(map[string]*Transcript)

	captionTracks, ok := captionsJSON["captionTracks"].([]interface{})
	if !ok {
		return nil, ErrInvalidFormat
	}

	for _, captionTrack := range captionTracks {
		track, _ := captionTrack.(map[string]interface{})
		languageCode, _ := track["languageCode"].(string)
		baseURL, _ := track["baseUrl"].(string)
		name, _ := track["name"].(map[string]interface{})
		simpleText, _ := name["simpleText"].(string)
		kind, _ := track["kind"].(string)

		transcript := &Transcript{
			VideoID:      videoID,
			URL:          baseURL,
			Language:     simpleText,
			LanguageCode: languageCode,
			IsGenerated:  kind == "asr",
		}

		if kind == "asr" {
			generatedTranscripts[languageCode] = transcript
		} else {
			manualTranscripts[languageCode] = transcript
		}
	}

	return &TranscriptList{
		VideoID:                    videoID,
		ManuallyCreatedTranscripts: manualTranscripts,
		GeneratedTranscripts:       generatedTranscripts,
	}, nil
}

// Fetch fetches the transcript from the transcript URL.
func (t *Transcript) Fetch() ([]TranscriptEntry, error) {
	resp, err := http.Get(t.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseTranscript(string(body))
}

func parseTranscript(xmlData string) ([]TranscriptEntry, error) {
	var transcript struct {
		Entries []TranscriptEntry `xml:"text"`
	}

	err := xml.Unmarshal([]byte(xmlData), &transcript)
	if err != nil {
		return nil, err
	}

	for i := range transcript.Entries {
		transcript.Entries[i].Text = removeHTMLTags(transcript.Entries[i].Text)
	}

	return transcript.Entries, nil
}

func removeHTMLTags(text string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(text, "")
}

// FindTranscript finds the first transcript that matches the language codes.
func (tl *TranscriptList) FindTranscript(languageCodes ...string) (*Transcript, error) {
	for _, code := range languageCodes {
		if t, ok := tl.ManuallyCreatedTranscripts[code]; ok {
			return t, nil
		}
		if t, ok := tl.GeneratedTranscripts[code]; ok {
			return t, nil
		}
	}
	return nil, ErrNoTranscriptFound
}
