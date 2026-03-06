package main

import (
	"fmt"
	"os"

	mp4 "github.com/abema/go-mp4"
)

func testing() {
	file, err := os.Open("/Users/adrianwill/Music/Music/Media.localized/Music/21 Savage/american dream/02 all of me.m4a")
	if err != nil {
		fmt.Println(err)
		return
	}
	titleKey := mp4.BoxType{0xa9, 'n', 'a', 'm'}
	_, err = mp4.ReadBoxStructure(file, func(h *mp4.ReadHandle) (any, error) {

		if len(h.Path) >= 2 && h.Path[len(h.Path)-2] == titleKey &&
			h.BoxInfo.Type == mp4.BoxTypeData() {
			box, _, _ := h.ReadPayload()
			d := box.(*mp4.Data)
			fmt.Println("title:", string(d.Data))
		}
		fmt.Println(len(h.Path), h.BoxInfo.Type.String())
		_, e := h.Expand()
		return nil, e

	})
	if err != nil {
		fmt.Println(err)
	}
}
