package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type ExplicitType string

const (
	Clean    ExplicitType = "clean"
	Explicit ExplicitType = "explicit"
)

type Song struct {
	Title       string       //done
	Album       string       //done
	Artist      string       //done
	FilePath    string       //done
	Date        string       //done in UTC
	Genre       string       //done
	TrackNumber int          //done
	TrackLength int          //done
	DiskNumber  int          //done
	DiskLength  int          //done
	Explicit    ExplicitType //done
	Artwork     string       //done
	Lyrics      string
}

type ScrapedData struct {
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
	if len(os.Args) < 2 {
		fmt.Println("Processes all .m4a files in Downloads folder")
		return
	}

	switch os.Args[1] {
	case "1":
		albumUpdate()
	case "2":
		metadataUpdate("", "")
	default:
		fmt.Println("Invalid option. Use '2' to update metadata.")
	}
}

func albumUpdate() {
	directory := os.Args[2]
	parts := strings.Split(filepath.Clean(directory), string(filepath.Separator))
	albumName := parts[len(parts)-1]
	artistName := parts[len(parts)-2]
	searchTerm := artistName + " " + albumName
	metadataUpdate(directory, searchTerm)
}

func metadataUpdate(dir string, searchTerm string) {
	homeDir, _ := os.UserHomeDir()
	targetDir := filepath.Join(homeDir, "Downloads")
	if dir != "" {
		targetDir = dir
	}
	matches, err := filepath.Glob(filepath.Join(targetDir, "*.m4a"))
	if err != nil || len(matches) == 0 {
		println(dir)
		fmt.Println("No m4a files found in target directory")
		return
	}

	fmt.Printf("Found %d files to process\n\n", len(matches))

	// Process each file
	for i, m4aFile := range matches {
		fmt.Printf("=== Processing file %d/%d ===\n", i+1, len(matches))
		fmt.Printf("File: %s\n", filepath.Base(m4aFile))

		// Extract title from filename
		baseName := filepath.Base(m4aFile)
		songTitle := strings.TrimSuffix(baseName, ".m4a")
		fmt.Printf("Song title: %s\n", songTitle)
		artist := searchTerm
		// Prompt for artist
		if dir == "" {
			fmt.Print("Enter artist name: ")
			reader := bufio.NewReader(os.Stdin)
			artist, _ = reader.ReadString('\n')
			artist = strings.TrimSpace(artist)
			if artist == "" {
				fmt.Println("Skipping (no artist provided)")
				continue
			}
		}

		// Try to get song ID from search
		songID, err := extractSongIDFromSearch(songTitle, artist)
		if err != nil {
			fmt.Printf("Search failed: %v\n", err)
			return
		}
		fmt.Printf("Found song ID: %s\n", songID)

		// Fetch metadata from iTunes
		scraped, err := getMetadataFromiTunes(songID)
		if err != nil {
			fmt.Printf("Error fetching metadata: %v\n", err)
			fmt.Println("Skipping")
			continue
		}
		if dir != "" {
			lyrics := getLyrics(songTitle, scraped.Artist)
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

			if err := processFile(m4aFile, songTitle, scraped); err != nil {
				fmt.Printf("Error processing file: %v\n\n", err)
				continue
			}
		}
		// Process file (same as current implementation)

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
	lyricsURL := fmt.Sprintf("https://api.lyrics.ovh/v1/%s/%s/", artist, song)

	fmt.Printf("Fetching lyrics from: %s\n", lyricsURL)

	client := &http.Client{Timeout: 15 * time.Second}

	// Retry up to 3 times
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			fmt.Printf("Retry attempt %d/%d...\n", attempt, maxRetries)
			time.Sleep(2 * time.Second) // Wait between retries
		}

		req, _ := http.NewRequest("GET", lyricsURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)

		if err != nil {
			if attempt == maxRetries {
				fmt.Printf("Lyrics unavailable after %d attempts, continuing without lyrics\n", maxRetries)
				return ""
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			if attempt == maxRetries {
				fmt.Printf("Lyrics unavailable (status %d), continuing without lyrics\n", resp.StatusCode)
			}
			return ""
		}

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
			fmt.Println("✓ Lyrics fetched successfully")
			return strings.Join(nonEmpty, "\n")
		}

		fmt.Println("No lyrics found")
		return ""
	}

	return ""
}

func extractSongIDFromSearch(songTitle, artist string) (string, error) {
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

	// Strip non-alphanumeric/space chars from search term
	searchTerm := songTitle + " " + artist
	re := regexp.MustCompile(`[^a-zA-Z0-9 ]`)
	cleaned := re.ReplaceAllString(searchTerm, "")
	searchURL := fmt.Sprintf("https://music.apple.com/us/search?term=%s", cleaned)
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

	// Prefer collectionArtistName, fallback to artistName
	artist := track.CollectionArtistName
	if artist == "" {
		artist = track.ArtistName
	}
	if idx := strings.Index(artist, " &"); idx != -1 {
		artist = artist[:idx]
	}
	if idx := strings.Index(artist, ", "); idx != -1 {
		artist = artist[:idx]
	}

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

	data.Artwork = strings.ReplaceAll(track.ArtworkUrl100, "100x100bb.jpg", "100000x100000-999.jpg")

	return data, nil
}

func processFile(m4aFile, songTitle string, scraped ScrapedData) error {
	// Get lyrics
	lyrics := getLyrics(songTitle, scraped.Artist)

	// Download artwork with fallbacks
	artworkPath := ""
	if scraped.Artwork != "" {
		artworkPath = filepath.Join(os.TempDir(), "artwork.jpg")

		// Try high-res first, then fallback to lower resolutions
		artworkURLs := []string{
			scraped.Artwork, // 100000x100000-999.jpg
			strings.ReplaceAll(scraped.Artwork, "100000x100000-999.jpg", "3000x3000bb.jpg"),
			strings.ReplaceAll(scraped.Artwork, "100000x100000-999.jpg", "600x600bb.jpg"),
		}

		downloaded := false
		for _, artURL := range artworkURLs {
			if err := downloadFile(artURL, artworkPath); err == nil {
				// Verify file exists and has content
				if info, err := os.Stat(artworkPath); err == nil && info.Size() > 0 {
					fmt.Printf("Downloaded artwork: %s (%.1f KB)\n", artURL, float64(info.Size())/1024)
					downloaded = true
					break
				}
			}
		}

		if !downloaded {
			fmt.Println("Warning: Could not download artwork, continuing without it")
			artworkPath = "" // Reset so we don't try to use it
		}
	}

	// Build AtomicParsley command
	args := []string{m4aFile, "--overWrite",
		"--artwork", "REMOVE_ALL", // Remove all existing artwork first
		"--title", songTitle,
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

	// Run AtomicParsley
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

	// Create destination directory

	destDir := filepath.Join("/Users/adrianwill/Music/Music/Media.localized/Music", scraped.Artist, scraped.Album)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Move file
	newFileName := fmt.Sprintf("%s.m4a", songTitle)
	destPath := filepath.Join(destDir, newFileName)
	if err := os.Rename(m4aFile, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	fmt.Printf("Moved to: %s\n", destPath)

	// Clean up artwork file
	if artworkPath != "" {
		os.Remove(artworkPath)
	}

	// Open file
	openCmd := exec.Command("open", destPath)
	if err := openCmd.Run(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	return nil
}
