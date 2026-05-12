package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	mp4 "github.com/abema/go-mp4"
	"github.com/sunfish-shogi/bufseekio"
)

// TODO the image function might not be working
func testing(dir string, title string, artist string, genre string,
	year string, album string, trackNum int, trackLength int, diskNum int,
	diskLength int, artwork string, lyrics string, explicit int, collectionArtist string) error {
	// homeDir, _ := os.UserHomeDir()
	// directory := filepath.Join(homeDir, "Downloads")

	inputFile, err := os.Open(dir)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer inputFile.Close()

	outputPath := filepath.Join(os.TempDir(), "TEMP.m4a")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	r := bufseekio.NewReadSeeker(inputFile, 128*1024, 4)
	w := mp4.NewWriter(outputFile)
	titleKey := mp4.BoxType{0xa9, 'n', 'a', 'm'}
	artistKey := mp4.BoxType{0xa9, 'A', 'R', 'T'}
	albumKey := mp4.BoxType{0xa9, 'a', 'l', 'b'}
	genreKey := mp4.BoxType{0xa9, 'g', 'e', 'n'}
	dateKey := mp4.BoxType{0xa9, 'd', 'a', 'y'}
	trackKey := mp4.BoxType{'t', 'r', 'k', 'n'}
	diskKey := mp4.BoxType{'d', 'i', 's', 'k'}
	ratingKey := mp4.BoxType{'r', 't', 'n', 'g'}
	artworkKey := mp4.BoxType{'c', 'o', 'v', 'r'}
	lyricsKey := mp4.BoxType{0xa9, 'l', 'y', 'r'}
	albumArtistKey := mp4.BoxType{'a', 'A', 'R', 'T'}
	foundTitle := false
	foundArtist := false
	foundAlbum := false
	foundGenre := false
	foundDate := false
	foundTrack := false
	foundDisk := false
	foundRating := false
	foundArtwork := false
	foundLyrics := false
	foundAlbumArtist := false
	_, writeErr := mp4.ReadBoxStructure(r, func(h *mp4.ReadHandle) (any, error) {

		if !h.BoxInfo.IsSupportedType() || h.BoxInfo.Type == mp4.BoxTypeMdat() {
			return nil, w.CopyBox(r, &h.BoxInfo)
		}

		_, err := w.StartBox(&h.BoxInfo)
		if err != nil {
			return nil, err
		}
		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		if h.BoxInfo.Type == mp4.BoxTypeData() {
			lenPath := len(h.Path)
			if lenPath >= 2 {
				d := box.(*mp4.Data)
				key := h.Path[lenPath-2]
				switch key {
				case titleKey:
					setData(d, mp4.DataTypeStringUTF8, []byte(title), &foundTitle)
				case artistKey:
					setData(d, mp4.DataTypeStringUTF8, []byte(artist), &foundArtist)
				case albumKey:
					setData(d, mp4.DataTypeStringUTF8, []byte(album), &foundAlbum)
				case genreKey:
					setData(d, mp4.DataTypeStringUTF8, []byte(genre), &foundGenre)
				case dateKey:
					setData(d, mp4.DataTypeStringUTF8, []byte(year), &foundDate)
				case trackKey:
					setData(d, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, byte(trackNum), 0x00, byte(trackLength), 0x00, 0x00}, &foundTrack)
				case diskKey:
					setData(d, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, byte(diskNum), 0x00, byte(diskLength)}, &foundDisk)
				case ratingKey:
					if explicit == 1 {
						setData(d, mp4.DataTypeBinary, []byte{byte(explicit)}, &foundRating)
					}
					foundRating = true
				case artworkKey:
					if artwork != "" {
						resp, err := http.Get(artwork)
						if err != nil {
							return nil, err
						}
						defer resp.Body.Close()
						if resp.StatusCode != http.StatusOK {
							return nil, fmt.Errorf("artwork download failed: %s", resp.Status)
						}
						imgBytes, err := io.ReadAll(resp.Body)
						if err != nil {
							return nil, err
						}
						setData(d, mp4.DataTypeBinary, imgBytes, &foundArtwork)
					}
				case lyricsKey:
					if lyrics != "" {
						setData(d, mp4.DataTypeStringUTF8, []byte(lyrics), &foundLyrics)
					}
				case albumArtistKey:
					if collectionArtist != "" {
						setData(d, mp4.DataTypeStringUTF8, []byte(collectionArtist), &foundAlbumArtist)
					} else {
						setData(d, mp4.DataTypeStringUTF8, []byte(artist), &foundAlbumArtist)
					}
				}
			}
		}
		if _, err := mp4.Marshal(w, box, h.BoxInfo.Context); err != nil {
			return nil, err
		}
		if _, err := h.Expand(); err != nil {
			return nil, err
		}
		if h.BoxInfo.Type == mp4.BoxTypeIlst() {
			if !foundTitle {
				if err := createBox(h, titleKey, w, mp4.DataTypeStringUTF8, []byte(title), &foundTitle); err != nil {
					return nil, err
				}
			}
			if !foundArtist {
				if err := createBox(h, artistKey, w, mp4.DataTypeStringUTF8, []byte(artist), &foundArtist); err != nil {
					return nil, err
				}
			}
			if !foundAlbum {
				if err := createBox(h, albumKey, w, mp4.DataTypeStringUTF8, []byte(album), &foundAlbum); err != nil {
					return nil, err
				}
			}
			if !foundGenre {
				if err := createBox(h, genreKey, w, mp4.DataTypeStringUTF8, []byte(genre), &foundGenre); err != nil {
					return nil, err
				}
			}
			if !foundDate {
				if err := createBox(h, dateKey, w, mp4.DataTypeStringUTF8, []byte(year), &foundDate); err != nil {
					return nil, err
				}
			}
			if !foundTrack {
				if err := createBox(h, trackKey, w, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, byte(trackNum), 0x00, byte(trackLength), 0x00, 0x00}, &foundTrack); err != nil {
					return nil, err
				}
			}
			if !foundDisk {
				if err := createBox(h, diskKey, w, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, byte(diskNum), 0x00, byte(diskLength)}, &foundDisk); err != nil {
					return nil, err
				}
			}
			if !foundRating && explicit == 1 {
				if err := createBox(h, ratingKey, w, mp4.DataTypeBinary, []byte{byte(explicit)}, &foundRating); err != nil {
					return nil, err
				}
			}
			if !foundArtwork && artwork != "" {
				resp, err := http.Get(artwork)
				if err != nil {
					return nil, err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return nil, fmt.Errorf("artwork download failed: %s", resp.Status)
				}
				imgBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, err
				}
				if err := createBox(h, artworkKey, w, mp4.DataTypeBinary, imgBytes, &foundArtwork); err != nil {
					return nil, err
				}
			}
			if !foundLyrics && lyrics != "" {
				if err := createBox(h, lyricsKey, w, mp4.DataTypeStringUTF8, []byte(lyrics), &foundLyrics); err != nil {
					return nil, err
				}
			}
			if !foundAlbumArtist && collectionArtist != "" {
				if err := createBox(h, albumArtistKey, w, mp4.DataTypeStringUTF8, []byte(collectionArtist), &foundAlbumArtist); err != nil {
					return nil, err
				}
			}
		}
		_, err = w.EndBox()
		return nil, err
	})
	fmt.Println(err)
	fmt.Println("foundTitle:", foundTitle)
	fmt.Println("foundArtist:", &foundArtist)
	fmt.Println("foundAlbum:", foundAlbum)
	fmt.Println("foundGenre:", foundGenre)
	fmt.Println("foundDate:", foundDate)
	fmt.Println("foundTrack:", foundTrack)
	fmt.Println("foundDisk:", foundDisk)
	fmt.Println("foundRating:", foundRating)
	fmt.Println("foundArtwork:", foundArtwork)
	fmt.Println("foundLyrics:", foundLyrics)
	fmt.Println("foundAlbumArtist:", foundAlbumArtist)
	closeErr := outputFile.Close()

	if writeErr != nil {
		return fmt.Errorf("write mp4 metadata: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close output file: %w", closeErr)
	}
	openCmd := exec.Command("open", outputPath)
	if err := openCmd.Run(); err != nil {
		fmt.Println("failed to open file: %w", err)
		return err
	}
	if err := os.Remove(dir); err != nil {
		fmt.Println("failed to remove file: %w", err)
		return err
	}
	return nil
}

func setData(d *mp4.Data, dt uint32, b []byte, foundBool *bool) {
	d.DataType = dt
	d.Data = b
	*foundBool = true
}

func createBox(h *mp4.ReadHandle, key mp4.BoxType, w *mp4.Writer, dt uint32, b []byte, foundBool *bool) error {
	_, err := w.StartBox(&mp4.BoxInfo{Type: key})
	if err != nil {
		return err
	}
	_, err = w.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeData()})
	if err != nil {
		return err
	}
	ctx := h.BoxInfo.Context
	ctx.UnderIlst = true
	ctx.UnderIlstMeta = true
	fmt.Print(h.BoxInfo.Type.String())
	if _, err := mp4.Marshal(w, &mp4.Data{DataType: dt, DataLang: 0, Data: b}, ctx); err != nil {
		return err
	}

	_, err = w.EndBox()
	if err != nil {
		return err
	}
	_, err = w.EndBox()
	if err != nil {
		return err
	}
	*foundBool = true
	return err
}
