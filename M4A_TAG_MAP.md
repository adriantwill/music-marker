# M4A Tag Map (main.go -> MP4 atoms)

Source fields from `Song`/`ScrapedData` in `main.go`.

| Field | AtomicParsley flag | Atom key | Full path | `mp4.DataType` | Notes |
|---|---|---|---|---|---|
| `Title` | `--title` | `©nam` | `moov.udta.meta.ilst.©nam.data` | `mp4.DataTypeStringUTF8` | UTF-8 text |
| `Artist` | `--artist` | `©ART` | `moov.udta.meta.ilst.©ART.data` | `mp4.DataTypeStringUTF8` | Track artist |
| `Album` | `--album` | `©alb` | `moov.udta.meta.ilst.©alb.data` | `mp4.DataTypeStringUTF8` | Album title |
| `Genre` | `--genre` | `©gen` or `gnre` | `moov.udta.meta.ilst.(©gen\|gnre).data` | `mp4.DataTypeStringUTF8` (for `©gen`) | `gnre` can be numeric in some files |
| `Date` | `--year` | `©day` | `moov.udta.meta.ilst.©day.data` | `mp4.DataTypeStringUTF8` | Usually year/date string |
| `TrackNumber` + `TrackLength` | `--tracknum` (`n/t`) | `trkn` | `moov.udta.meta.ilst.trkn.data` | `mp4.DataTypeBinary` | Structured binary (current/total), not UTF-8 text |
| `DiskNumber` + `DiskLength` | `--disk` (`n/t`) | `disk` | `moov.udta.meta.ilst.disk.data` | `mp4.DataTypeBinary` | Structured binary (current/total), not UTF-8 text |
| `Explicit` | `--advisory` | `rtng` | `moov.udta.meta.ilst.rtng.data` | `mp4.DataTypeBinary` | Advisory/rating numeric payload |
| `Artwork` | `--artwork` | `covr` | `moov.udta.meta.ilst.covr.data` | `mp4.DataTypeStringJPEG` or `mp4.DataTypeBinary` | JPEG/PNG bytes (container/player-dependent) |
| `Lyrics` | `--lyrics` | `©lyr` | `moov.udta.meta.ilst.©lyr.data` | `mp4.DataTypeStringUTF8` | UTF-8 lyrics text |
| `FilePath` | N/A | N/A | N/A | N/A | File location, not MP4 metadata atom |
| Album artist | `--albumArtist` | `aART` | `moov.udta.meta.ilst.aART.data` | `mp4.DataTypeStringUTF8` |


## DataType quick refs

- `mp4.DataTypeStringUTF8`: regular text tags.
- `mp4.DataTypeBinary`: numeric/structured payloads (`trkn`, `disk`, `rtng`).
- `mp4.DataTypeStringJPEG`: JPEG artwork payload (some files still use binary-like handling).
