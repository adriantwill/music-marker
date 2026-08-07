use clap::Parser;
use mp4ameta::{Img, Tag};
use serde::Deserialize;
use std::collections::HashMap;
use std::error::Error;
use std::ffi::OsStr;
use std::path::{Path, PathBuf};
use std::process::Command;
#[derive(Parser, Debug)]
struct Args {
    id: i32,
    path: PathBuf,
}
#[derive(Deserialize)]
struct SongResult {
    results: Vec<Song>,
}
#[derive(Debug, Deserialize)]
struct Lyrics {
    lyrics: String,
}
#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct Song {
    wrapper_type: String,
    artist_name: String,
    collection_artist_name: Option<String>,
    artist_id: u64,
    collection_name: String,
    track_name: String,
    primary_genre_name: String,
    release_date: String,
    track_explicitness: String,
    track_number: u16,
    track_count: u16,
    disc_number: u16,
    disc_count: u16,
    artwork_url_100: String,
}
#[derive(Deserialize)]
struct Artist {
    artist_name: String,
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let song_id = args.id;
    if args.path.extension() == Some(OsStr::new("m4a")) {
        let res: SongResult =
            reqwest::blocking::get(format!("https://itunes.apple.com/lookup?id={song_id}"))?
                .json()?;
        let artist_id = res.results[0].artist_id;
        let artist: Artist =
            reqwest::blocking::get(format!("https://itunes.apple.com/lookup?id={artist_id}"))?
                .json()?;
        if res.results[0].wrapper_type != "track" {
            return Err("Not 1 Song".into());
        }
        set_atributes(&args.path, &res.results[0], artist.artist_name)?;
    } else if args.path.is_dir() {
        let res: SongResult = reqwest::blocking::get(format!(
            "https://itunes.apple.com/lookup?id={song_id}&entity=song"
        ))?
        .json()?;
        let artist_id = res.results[0].artist_id;
        let artist: Artist =
            reqwest::blocking::get(format!("https://itunes.apple.com/lookup?id={artist_id}"))?
                .json()?;
        let mut song_lookup: HashMap<String, Song> = HashMap::new();
        for song in res.results {
            if song.wrapper_type == "track" {
                song_lookup.insert(song.track_name.clone(), song);
            }
        }
        let path = args.path.display();
        for entry in glob::glob(&format!("{path}/*.m4a"))? {
            let path = entry?;
            let tag = Tag::read_from_path(&path).unwrap();
            let Some(song_name) = tag.title().or_else(|| path.file_stem()?.to_str()) else {
                print!("no title found in file {}", path.display());
                continue;
            };
            let Some(song) = song_lookup.get(song_name) else {
                print!("no song found called {}", song_name);
                continue;
            };
            set_atributes(&path, song, artist.artist_name.clone())?;
        }
    } else {
        return Err("Not dir or m4a file".into());
    }
    Ok(())
}

fn set_atributes(path: &Path, song: &Song, album_artist: String) -> Result<(), Box<dyn Error>> {
    let mut tag = Tag::read_from_path(path)?;
    tag.set_title(&song.track_name);
    tag.set_artist(&song.artist_name);
    tag.set_album(&song.collection_name);
    tag.set_album_artist(album_artist);
    tag.set_genre(&song.primary_genre_name);
    tag.set_year(&song.release_date);
    tag.set_track_number(song.track_number);
    tag.set_total_tracks(song.track_count);
    tag.set_disc_number(song.disc_number);
    tag.set_total_discs(song.disc_count);
    let artwork = reqwest::blocking::get(
        song.artwork_url_100
            .replace("100x100bb.jpg", "3000x3000bb.jpg"),
    )?
    .bytes()?;
    tag.set_artwork(Img::jpeg(artwork.to_vec()));
    if &song.track_explicitness == "explicit" {
        tag.set_advisory_rating(mp4ameta::AdvisoryRating::Explicit);
    }
    let artist_name = &song.artist_name;
    let song_name = &song.track_name;
    let lyrics: Lyrics = reqwest::blocking::get(format!(
        "https://api.lyrics.ovh/v1/{artist_name}/{song_name}/"
    ))?
    .json()?;
    tag.set_lyrics(lyrics.lyrics);
    tag.write_to_path(path).unwrap();
    Command::new("open").arg(path).status()?;
    Ok(())
}
