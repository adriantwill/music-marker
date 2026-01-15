package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
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
}

func main() {
	fmt.Println("Enter Song Name: ", os.Args[1])
	fmt.Println("Enter Artist Name: ", os.Args[2])
	// file, err := os.Open(os.Args[4])
	// if err != nil {
	// 	fmt.Println("Error", err)
	// }
	// defer file.Close()
	// dir := fmt.Sprintf("/Users/adrianwill/Music/Music/Media.localized/Music/%s/%s", os.Args[3], os.Args[2])
	// os.MkdirAll(dir, 0755)
	switch os.Args[3] {
	case "1":
		metadataExtraction()
	case "2":
		metadataUpdate()
	case "3":
		getLyrics(os.Args[1], os.Args[2])
	default:
		fmt.Println("Invalid option")
	}
}

func scrapeAppleMusicMetadata(songTitle, artist string) (ScrapedData, error) {
	data := ScrapedData{}

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(fmt.Sprintf("https://music.apple.com/us/search?term=%s %s", songTitle, artist))
	page.MustWaitLoad()
	time.Sleep(2 * time.Second) // wait for JS

	// Click first result
	firstResult := page.MustElement(".click-action.svelte-c0t0j2")
	firstResult.MustClick()
	page.MustWaitLoad()
	time.Sleep(2 * time.Second)

	// Extract album
	if elem, err := page.Element("h1.headings__title.svelte-1uuona0 span"); err == nil {
		data.Album, _ = elem.Text()
	}

	// Extract artist
	if elem, err := page.Element(".headings__subtitles.svelte-1uuona0 a"); err == nil {
		data.Artist, _ = elem.Text()
	}

	// Extract genre
	if elem, err := page.Element(".headings__metadata-bottom.svelte-1uuona0"); err == nil {
		text, _ := elem.Text()
		data.Genre = strings.Split(text, " · ")[0]
	}

	// Extract date
	if elem, err := page.Element("p.description.svelte-1tm9k9g"); err == nil {
		text, _ := elem.Text()
		dateStr := strings.Split(text, "\n")[0]
		data.Date, _ = dateToUTC(dateStr)
	}

	// Count disks and find song
	diskElems := page.MustElements("div.songs-list.svelte-1nv3ko5.songs-list--album")
	data.DiskLength = len(diskElems)

	for diskIdx, diskElem := range diskElems {
		totalCount := 0
		songRows := diskElem.MustElements("div.songs-list-row__song-container.svelte-t6plbb")

		for _, row := range songRows {
			totalCount++

			nameElem := row.MustElement("div.songs-list-row__song-name.svelte-t6plbb")
			songName, _ := nameElem.Text()

			if songName == songTitle {
				data.DiskNumber = diskIdx + 1

				// Get track number
				if trackElem, err := row.Element("div.songs-list-row__column-data.svelte-t6plbb[data-testid='track-number']"); err == nil {
					trackStr, _ := trackElem.Text()
					data.TrackNumber, _ = strconv.Atoi(trackStr)
				}

				// Check explicit
				if _, err := row.Element("span.explicit-wrapper.svelte-j8a2wc"); err == nil {
					data.Explicit = Explicit
				} else {
					data.Explicit = Clean
				}
			}
		}
	}

	data.TrackLength = totalCount

	return data, nil
}

func metadataExtraction() {
	songTitle := os.Args[1]
	artist := os.Args[2]

	// Scrape Apple Music metadata
	scraped, err := scrapeAppleMusicMetadata(songTitle, artist)
	if err != nil {
		fmt.Println("Error scraping metadata:", err)
		return
	}

	// Build Song struct
	song := Song{
		Title:       songTitle,
		Album:       scraped.Album,
		Artist:      scraped.Artist,
		Date:        scraped.Date,
		Genre:       scraped.Genre,
		TrackNumber: scraped.TrackNumber,
		TrackLength: scraped.TrackLength,
		DiskNumber:  scraped.DiskNumber,
		DiskLength:  scraped.DiskLength,
		Explicit:    scraped.Explicit,
	}

	song.Lyrics = getLyrics(song.Title, song.Artist)

	fmt.Print(song)
}

func metadataUpdate() {
	// file, err := os.Open("/Users/adrianwill/Music/Music/Media.localized/Music/The Weeknd/Starboy/01 Starboy.m4a")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()
}

func dateToUTC(date string) (string, error) {
	months := map[string]string{
		"January":   "01",
		"February":  "02",
		"March":     "03",
		"April":     "04",
		"May":       "05",
		"June":      "06",
		"July":      "07",
		"August":    "08",
		"September": "09",
		"October":   "10",
		"November":  "11",
		"December":  "12",
	}
	parts := strings.Split(date, " ")
	parts[1] = strings.ReplaceAll(parts[1], ",", "")
	return fmt.Sprintf("%s-%s-%sT04:00:00Z", parts[2], months[parts[0]], parts[1]), nil
}

func getLyrics(song string, artist string) string {
	lyricsURL := fmt.Sprintf("https://api.lyrics.ovh/v1/%s/%s", artist, song)

	fmt.Printf("Fetching lyrics from: %s\n", lyricsURL)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", lyricsURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error fetching lyrics: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
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
			return strings.Join(nonEmpty, "\n")
		}
	}
	return ""
}
