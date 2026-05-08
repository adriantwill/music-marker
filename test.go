package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	mp4 "github.com/abema/go-mp4"
	"github.com/sunfish-shogi/bufseekio"
)

// TODO the image function might not be working
func testing() {
	homeDir, _ := os.UserHomeDir()
	directory := filepath.Join(homeDir, "Downloads")

	inputPath := "/Users/adrianwill/Music/Music/Media.localized/Music/O A/test/test.temp.m4a"
	inputFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer inputFile.Close()

	outputPath := "/Users/adrianwill/Music/Music/Media.localized/Music/O A/test/test.m4a"
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return
	}
	defer outputFile.Close()
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
	_, err = mp4.ReadBoxStructure(r, func(h *mp4.ReadHandle) (any, error) {

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
					setData(d, mp4.DataTypeStringUTF8, []byte("Golden Girls"), &foundTitle)
				case artistKey:
					setData(d, mp4.DataTypeStringUTF8, []byte("Frank Ocean, Tyler"), &foundArtist)
				case albumKey:
					setData(d, mp4.DataTypeStringUTF8, []byte("chanel orange"), &foundAlbum)
				case genreKey:
					setData(d, mp4.DataTypeStringUTF8, []byte("alternative"), &foundGenre)
				case dateKey:
					setData(d, mp4.DataTypeStringUTF8, []byte("2015-02-13T12:00:00Z"), &foundDate)
				case trackKey:
					setData(d, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, 0x03, 0x00, 0x0C, 0x00, 0x00}, &foundTrack)
				case diskKey:
					setData(d, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x01}, &foundDisk)
				case ratingKey:
					setData(d, mp4.DataTypeBinary, []byte{0x01}, &foundRating)
				case artworkKey:
					resp, err := http.Get("https://is1-ssl.mzstatic.com/image/thumb/Music125/v4/27/9a/8c/279a8c66-9914-add2-9c7f-912f2946fb8a/15UMGIM08570.rgb.jpg/3000x3000bb.jpg")
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
				case lyricsKey:
					setData(d, mp4.DataTypeStringUTF8, []byte("2015-02-13T12:00:00Z"), &foundLyrics)
				case albumArtistKey:
					setData(d, mp4.DataTypeStringUTF8, []byte("Frank Ocean"), &foundAlbumArtist)
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
				if err := createBox(h, titleKey, w, mp4.DataTypeStringUTF8, []byte("Golden Girls"), &foundTitle); err != nil {
					return nil, err
				}
			}
			if !foundArtist {
				if err := createBox(h, artistKey, w, mp4.DataTypeStringUTF8, []byte("Frank Ocean, Tyler"), &foundArtist); err != nil {
					return nil, err
				}
			}
			if !foundAlbum {
				if err := createBox(h, albumKey, w, mp4.DataTypeStringUTF8, []byte("chanel orange"), &foundAlbum); err != nil {
					return nil, err
				}
			}
			if !foundGenre {
				if err := createBox(h, genreKey, w, mp4.DataTypeStringUTF8, []byte("alternative"), &foundGenre); err != nil {
					return nil, err
				}
			}
			if !foundDate {
				if err := createBox(h, dateKey, w, mp4.DataTypeStringUTF8, []byte("2015-02-13T12:00:00Z"), &foundDate); err != nil {
					return nil, err
				}
			}
			if !foundTrack {
				if err := createBox(h, trackKey, w, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, 0x03, 0x00, 0x0C, 0x00, 0x00}, &foundTrack); err != nil {
					return nil, err
				}
			}
			if !foundDisk {
				if err := createBox(h, diskKey, w, mp4.DataTypeBinary, []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x01}, &foundDisk); err != nil {
					return nil, err
				}
			}
			if !foundRating {
				if err := createBox(h, ratingKey, w, mp4.DataTypeBinary, []byte{0x01}, &foundRating); err != nil {
					return nil, err
				}
			}
			if !foundArtwork {
				resp, err := http.Get("https://is1-ssl.mzstatic.com/image/thumb/Music125/v4/27/9a/8c/279a8c66-9914-add2-9c7f-912f2946fb8a/15UMGIM08570.rgb.jpg/3000x3000bb.jpg")
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
			if !foundLyrics {
				if err := createBox(h, lyricsKey, w, mp4.DataTypeStringUTF8, []byte("2015-02-13T12:00:00Z"), &foundLyrics); err != nil {
					return nil, err
				}
			}
			if !foundAlbumArtist {
				if err := createBox(h, albumArtistKey, w, mp4.DataTypeStringUTF8, []byte("Frank Ocean"), &foundAlbumArtist); err != nil {
					return nil, err
				}
			}
		}
		_, err = w.EndBox()
		return nil, err
	})
	fmt.Println(err)
	fmt.Println("foundTitle:", foundTitle)
	fmt.Println("foundArtist:", foundArtist)
	fmt.Println("foundAlbum:", foundAlbum)
	fmt.Println("foundGenre:", foundGenre)
	fmt.Println("foundDate:", foundDate)
	fmt.Println("foundTrack:", foundTrack)
	fmt.Println("foundDisk:", foundDisk)
	fmt.Println("foundRating:", foundRating)
	fmt.Println("foundArtwork:", foundArtwork)
	fmt.Println("foundLyrics:", foundLyrics)
	fmt.Println("foundAlbumArtist:", foundAlbumArtist)
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
