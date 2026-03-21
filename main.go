package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type ExplicitType string

const (
	Clean    ExplicitType = "clean"
	Explicit ExplicitType = "explicit"
)

type Song struct {
	Title       string
	Album       string
	Artist      string
	FilePath    string
	Date        string
	Genre       string
	TrackNumber int
	TrackLength int
	DiskNumber  int
	DiskLength  int
	Explicit    ExplicitType
	Artwork     string
	Lyrics      string
}

type ScrapedData struct {
	Title       string
	Album       string
	Artist      string
	Genre       string
	Date        string
	TrackNumber int
	TrackLength int
	DiskNumber  int
	DiskLength  int
	Explicit    ExplicitType
	Artwork     string
}

type iTunesResponse struct {
	ResultCount int           `json:"resultCount"`
	Results     []iTunesTrack `json:"results"`
}

type iTunesTrack struct {
	ArtistName           string `json:"artistName"`
	ArtistId             int    `json:"artistId"`
	CollectionArtistName string `json:"collectionArtistName"`
	CollectionName       string `json:"collectionName"`
	TrackName            string `json:"trackName"`
	PrimaryGenreName     string `json:"primaryGenreName"`
	ReleaseDate          string `json:"releaseDate"`
	TrackExplicitness    string `json:"trackExplicitness"`
	TrackNumber          int    `json:"trackNumber"`
	TrackCount           int    `json:"trackCount"`
	DiscNumber           int    `json:"discNumber"`
	DiscCount            int    `json:"discCount"`
	ArtworkUrl100        string `json:"artworkUrl100"`
}

func main() {
	homeDir, _ := os.UserHomeDir()
	directory := filepath.Join(homeDir, "Downloads")
	if len(os.Args) > 1 {
		directory = strings.TrimSpace(os.Args[1])
	}
	info, err := os.Stat(directory)
	matches, _ := filepath.Glob(filepath.Join(directory, "*.m4a"))
	if err == nil || (info.IsDir() && len(matches) < 1) {
		fmt.Print("Enter a valid directory with m4a files")
		return
	}
	if !info.IsDir() && strings.ToLower(filepath.Ext(directory)) == ".m4a" {
		fmt.Print("Enter a valid .m4a file path")
		return
	}
	fmt.Print("Welcome to Music Metadata Marker")

	n, err := strconv.Atoi(input)
	for {
		n, err = strconv.Atoi(input)
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}
		input, _ := reader.ReadString('\n')
		//trims input
		input = strings.TrimSpace(strings.ToLower(input))
		break
	}
	metadataUpdate(directory, matches)
}

func listDirectoriesAtDepth(root string, targetDepth int) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == "." {
			return nil
		}
		depth := strings.Count(filepath.ToSlash(rel), "/") + 1
		if depth == targetDepth {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}

func metadataUpdate(dir string, matches []string) {
	// choice = strings.TrimSpace(choice)
	repeat := true
	fmt.Printf("Found %d files to process\n\n", len(matches))

	for i, m4aFile := range matches {
		reader := bufio.NewReader(os.Stdin)
		baseName := filepath.Base(m4aFile)
		searchTerm := strings.TrimSuffix(baseName, ".m4a")
		var result iTunesResponse
		if repeat {
			fmt.Print("Enter an Album or Song ID to find songs under, or a string to search for (leave blank to serach just using the file's name)")
			choice, _ := reader.ReadString('\n')
			if choice != "\n" {
				searchTerm = strings.TrimSpace(choice)
			}
			if isASCIIDigitsOnly(searchTerm) {
				resp, err := http.Get(fmt.Sprintf("https://itunes.apple.com/lookup?id=%s&entity=song", searchTerm))
				if err != nil {
					fmt.Printf("Error fetching metadata: %v\n", err)
					continue
				}
				defer resp.Body.Close()
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					fmt.Printf("Error decoding response: %v\n", err)
					continue
				}
			}
			if result.ResultCount != 1 {
				fmt.Print("Re prompt for every file (Search via file name) (Y/n): ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(strings.ToLower(input))
				if input == "n" || input == "no" {
					repeat = false
				}
			}
		}
		switch result.ResultCount {
		case 0:

			songID, err = extractSongIDFromSearch(searchTerm)
			if err != nil {
				fmt.Printf("Search failed: %v\n", err)
				return
			}
			fmt.Printf("Found song ID: %s\n", songID)
		case 1:
		default:
		}
		fmt.Printf("=== Processing file %d/%d ===\n", i+1, len(matches))
		fmt.Printf("File: %s\n", filepath.Base(m4aFile))
		// if len(songTitle) >= 2 {
		// 	songTitle = strings.TrimSpace(songTitle[2:])
		// }

		fmt.Printf("Song title: %s\n", searchTerm)
		songID := ""
		// if choice == "1" {
		// 	fmt.Print("Enter song ID: ")
		// 	reader := bufio.NewReader(os.Stdin)
		// 	songID, _ = reader.ReadString('\n')
		// 	songID = strings.TrimSpace(songID)
		// 	if songID == "" {
		// 		fmt.Println("Skipping (no artist provided)")
		// 		continue
		// 	}
		// } else {
		// }
		scraped, err := getMetadataFromiTunes(songID)
		if err != nil {
			fmt.Printf("Error fetching metadata: %v\n", err)
			fmt.Println("Skipping")
			continue
		}
		if dir != "" {
			lyrics := getLyrics(scraped.Title, scraped.Artist)
			args := []string{m4aFile, "--overWrite",
				"--genre", scraped.Genre,
				"--year", scraped.Date,
			}
			if lyrics != "" {
				args = append(args, "--lyrics", lyrics)
			}
			if scraped.Explicit == Explicit {
				args = append(args, "--advisory", "explicit")
			}
			fmt.Println("Running AtomicParsley...")
			cmd := exec.Command("AtomicParsley", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("AtomicParsley failed: %v\nOutput: %s", err, output)
				return
			}
		} else {
			if err := processFile(m4aFile, scraped); err != nil {
				fmt.Printf("Error processing file: %v\n\n", err)
				continue
			}
		}
		fmt.Printf("✓ Successfully processed\n\n")
	}

	fmt.Println("All files processed!")
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func getLyrics(song string, artist string) string {
	lyricsURL := fmt.Sprintf(
		"https://api.lyrics.ovh/v1/%s/%s/",
		url.PathEscape(artist),
		url.PathEscape(song),
	)
	fmt.Printf("Fetching lyrics from: %s\n", lyricsURL)
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", lyricsURL, nil)
	if err != nil {
		fmt.Println("Error creating lyrics request:", err)
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching lyrics:", err)
		return ""
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error fetching lyrics:", resp.Status)
		resp.Body.Close()
		return ""
	}
	defer resp.Body.Close()
	var result struct {
		Lyrics string `json:"lyrics"`
	}
	if json.NewDecoder(resp.Body).Decode(&result) == nil && result.Lyrics != "" {
		lines := strings.Split(result.Lyrics, "\n")
		var nonEmpty []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmpty = append(nonEmpty, line)
			}
		}
		fmt.Println("Lyrics fetched successfully")
		return strings.Join(nonEmpty, "\n")
	}
	fmt.Println("No lyrics found")
	return ""
}

func extractSongIDFromSearch(songTitle string) (string, error) {
	var songID string
	var foundResult bool

	c := colly.NewCollector()

	c.OnHTML("a.click-action", func(e *colly.HTMLElement) {
		if !foundResult {
			href := e.Attr("href")
			fmt.Println(href)
			if _, after, ok := strings.Cut(href, "?i="); ok {
				songID = after
				foundResult = true
			}
		}
	})
	searchTerm := normalizeAppleSearchTerm(songTitle)
	searchURL := fmt.Sprintf("https://music.apple.com/us/search?term=%s", encodeAppleSearchTerm(searchTerm))
	fmt.Println(searchURL)

	err := c.Visit(searchURL)
	if err != nil {
		return "", fmt.Errorf("failed to visit search page: %w", err)
	}

	if !foundResult {
		return "", fmt.Errorf("no results found for '%s' by '%s'", songTitle, artist)
	}

	return songID, nil
}

func normalizeAppleSearchTerm(songTitle string) string {
	songTitle = strings.ReplaceAll(songTitle, "+", " ")
	combined := strings.TrimSpace(songTitle)
	combined = strings.NewReplacer("(", " ", ")", " ").Replace(combined)
	return strings.Join(strings.Fields(combined), " ")
}

func encodeAppleSearchTerm(searchTerm string) string {
	encoded := url.QueryEscape(searchTerm)
	return strings.ReplaceAll(encoded, "+", "%20")
}

func isASCIIDigitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func getMetadataFromiTunes(songID string) (ScrapedData, error) {
	data := ScrapedData{}
	apiURL := fmt.Sprintf("https://itunes.apple.com/lookup?id=%s", songID)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return data, fmt.Errorf("iTunes API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result iTunesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return data, fmt.Errorf("failed to parse iTunes response: %w", err)
	}

	if result.ResultCount == 0 || len(result.Results) == 0 {
		return data, fmt.Errorf("no results from iTunes API for ID %s", songID)
	}

	track := result.Results[0]
	artistId := track.ArtistId
	artist := getArtistName(artistId)

	data.Title = track.TrackName
	data.Album = track.CollectionName
	data.Artist = artist
	data.Genre = track.PrimaryGenreName
	data.Date = track.ReleaseDate
	data.TrackNumber = track.TrackNumber
	data.TrackLength = track.TrackCount
	data.DiskNumber = track.DiscNumber
	data.DiskLength = track.DiscCount

	if track.TrackExplicitness == "explicit" {
		data.Explicit = Explicit
	}

	data.Artwork = strings.ReplaceAll(track.ArtworkUrl100, "100x100bb.jpg", "3000x3000bb.jpg")

	return data, nil
}

func getArtistName(artistId int) string {

	apiURL := fmt.Sprintf("https://itunes.apple.com/lookup?id=%d", artistId)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result iTunesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	if result.ResultCount == 0 || len(result.Results) == 0 {
		return ""
	}

	track := result.Results[0]
	artist := track.ArtistName
	return artist
}

func processFile(m4aFile string, scraped ScrapedData) error {
	// Get lyrics
	lyrics := getLyrics(scraped.Title, scraped.Artist)

	// Download artwork with fallbacks
	artworkPath := ""
	if scraped.Artwork != "" {
		artworkPath = filepath.Join(os.TempDir(), "artwork.jpg")
		downloaded := false
		if err := downloadFile(scraped.Artwork, artworkPath); err == nil {
			if info, err := os.Stat(artworkPath); err == nil && info.Size() > 0 {
				fmt.Printf("Downloaded artwork: %s (%.1f KB)\n", scraped.Artwork, float64(info.Size())/1024)
				downloaded = true
			}
		}

		if !downloaded {
			fmt.Println("Warning: Could not download artwork, continuing without it")
			artworkPath = ""
		}
	}
	args := []string{m4aFile, "--overWrite",
		"--artwork", "REMOVE_ALL",
		"--title", scraped.Title,
		"--artist", scraped.Artist,
		"--album", scraped.Album,
		"--genre", scraped.Genre,
		"--year", scraped.Date,
		"--tracknum", fmt.Sprintf("%d/%d", scraped.TrackNumber, scraped.TrackLength),
		"--disk", fmt.Sprintf("%d/%d", scraped.DiskNumber, scraped.DiskLength),
	}
	if artworkPath != "" {
		fmt.Printf("Removing old artwork and adding new: %s\n", artworkPath)
		args = append(args, "--artwork", artworkPath)
	} else {
		fmt.Println("Removing old artwork (no new artwork to add)")
	}
	if lyrics != "" {
		args = append(args, "--lyrics", lyrics)
	}
	if scraped.Explicit == Explicit {
		args = append(args, "--advisory", "explicit")
	}

	fmt.Println("Running AtomicParsley...")
	cmd := exec.Command("AtomicParsley", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if artworkPath != "" {
			os.Remove(artworkPath)
		}
		return fmt.Errorf("AtomicParsley failed: %v\nOutput: %s", err, output)
	}

	fmt.Println("✓ Metadata successfully applied")

	destDir := filepath.Join("/Users/adrianwill/Music/Music/Media.localized/Music", scraped.Artist, scraped.Album)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	newFileName := fmt.Sprintf("%s.m4a", scraped.Title)
	destPath := filepath.Join(destDir, newFileName)
	if err := os.Rename(m4aFile, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	fmt.Printf("Moved to: %s\n", destPath)
	if artworkPath != "" {
		os.Remove(artworkPath)
	}
	openCmd := exec.Command("open", destPath)
	if err := openCmd.Run(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	return nil
}

func normalizeTitle(title string) string {
	title, _, _ = transform.String(
		transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC),
		title,
	)
	title = strings.ToLower(title)
	title = regexp.MustCompile(`[^a-z0-9 ]+`).ReplaceAllString(title, " ")
	title = strings.Join(strings.Fields(title), " ")
	return title
}
