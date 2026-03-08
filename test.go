package main

import (
	"fmt"
	"os"

	mp4 "github.com/abema/go-mp4"
	"github.com/sunfish-shogi/bufseekio"
)

func testing() {
	inputPath := "/Users/adrianwill/Music/Music/Media.localized/Music/O A/test/test.m4a"
	inputFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer inputFile.Close()

	outputPath := "/Users/adrianwill/Music/Music/Media.localized/Music/O A/test/test.temp.m4a"
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return
	}
	defer outputFile.Close()
	r := bufseekio.NewReadSeeker(inputFile, 128*1024, 4)
	w := mp4.NewWriter(outputFile)
	titleKey := mp4.BoxType{0xa9, 'n', 'a', 'm'}
	artistKey := mp4.BoxType{0xa9, 'A', 'R', 'T'}
	albumArtistKey := mp4.BoxType{'a', 'A', 'R', 'T'}
	foundAlbumArtist := false
	_, err = mp4.ReadBoxStructure(r, func(h *mp4.ReadHandle) (any, error) {

		if !h.BoxInfo.IsSupportedType() || h.BoxInfo.Type == mp4.BoxTypeMdat() {
			// copy all data
			return nil, w.CopyBox(r, &h.BoxInfo)
		}

		// write header
		_, err := w.StartBox(&h.BoxInfo)
		if err != nil {
			return nil, err
		}
		// read payload
		box, _, err := h.ReadPayload()
		if err != nil {
			return nil, err
		}
		switch h.BoxInfo.Type {
		case mp4.BoxTypeData():
			if !(len(h.Path) >= 2) {
				break
			}
			if h.Path[len(h.Path)-2] == titleKey {
				box.(*mp4.Data).Data = []byte("test")
			}
			if h.Path[len(h.Path)-2] == artistKey {
				box.(*mp4.Data).Data = []byte("Frank Ocean, Tyler")
			}
			if h.Path[len(h.Path)-2] == albumArtistKey {
				foundAlbumArtist = true
				box.(*mp4.Data).Data = []byte("Frank Ocean")
			}
		}
		if _, err := mp4.Marshal(w, box, h.BoxInfo.Context); err != nil {
			return nil, err
		}
		if _, err := h.Expand(); err != nil {
			return nil, err
		}
		if foundAlbumArtist == false && h.BoxInfo.Type == mp4.BoxTypeIlst() {
			_, err := w.StartBox(&mp4.BoxInfo{Type: albumArtistKey})
			if err != nil {
				return nil, err
			}
			_, err = w.StartBox(&mp4.BoxInfo{Type: mp4.BoxTypeData()})
			if err != nil {
				return nil, err
			}
			ctx := h.BoxInfo.Context
			ctx.UnderIlst = true
			ctx.UnderIlstMeta = true
			fmt.Print(h.BoxInfo.Type.String())
			if _, err := mp4.Marshal(w, &mp4.Data{DataType: mp4.DataTypeStringUTF8, DataLang: 0, Data: []byte("Frank Ocean")}, ctx); err != nil {
				return nil, err
			}

			_, err = w.EndBox()
			if err != nil {
				return nil, err
			}
			_, err = w.EndBox()
			if err != nil {
				return nil, err
			}
			foundAlbumArtist = true
		}
		_, err = w.EndBox()
		return nil, err

	})
	fmt.Println(err)
}
