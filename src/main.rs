use clap::Parser;
use mp4ameta::{Img, Tag};
use serde::Deserialize;
use serde_json::json;
use std::collections::HashMap;
use std::env::current_dir;
use std::error::Error;
use std::ffi::OsStr;
use std::io::{self, Write};
use std::path::{Path, PathBuf};
use std::process::Command;
#[derive(Parser, Debug)]
struct Args {
    id: Option<i32>,
    path: Option<PathBuf>,
}
#[derive(Deserialize)]
struct TrackResult {
    results: Vec<ApiItem>,
}
#[derive(Deserialize)]
struct ArtistResult {
    results: Vec<Artist>,
}
#[derive(Debug, Deserialize)]
struct Lyrics {
    lyrics: String,
}
#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct Track {
    artist_name: String,
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
#[serde(rename_all = "camelCase")]
struct Artist {
    artist_name: String,
}
#[derive(Deserialize)]
#[serde(
    tag = "wrapperType",
    rename_all = "camelCase",
    rename_all_fields = "camelCase"
)]
enum ApiItem {
    Track(Track),
    Collection { artist_id: u64 },
}
impl ApiItem {
    fn artist_id(&self) -> u64 {
        match self {
            Self::Track(track) => track.artist_id,
            Self::Collection { artist_id } => *artist_id,
        }
    }
}

fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let path = match args.path {
        Some(path) => path,
        None => current_dir()?,
    };
    if path.extension() == Some(OsStr::new("m4a")) {
        let song_id = match args.id {
            Some(song_id) => song_id,
            None => get_user_input()?,
        };
        let (res, artist) = get_song_metadata(song_id)?;
        validate_song(&path, &res[0], artist)?
    } else if path.is_dir() {
        let lookup = match args.id {
            Some(song_id) => {
                let (res, artist) = get_song_metadata(song_id)?;
                let track_lookup: HashMap<String, Track> = res
                    .into_iter()
                    .filter_map(|item| match item {
                        ApiItem::Track(track) => Some((track.track_name.clone(), track)),
                        ApiItem::Collection { .. } => None,
                    })
                    .collect();
                Some((track_lookup, artist))
            }
            None => None,
        };
        let path = path.display();
        for entry in glob::glob(&format!("{path}/*.m4a"))? {
            let path = entry?;
            match &lookup {
                Some(lookup) => {
                    let tag = Tag::read_from_path(&path)?;
                    let Some(song_name) = tag.title().or_else(|| path.file_stem()?.to_str()) else {
                        println!("no title found in file {}", path.display());
                        continue;
                    };
                    let Some(song) = lookup.0.get(song_name) else {
                        println!("no song found called {}", song_name);
                        continue;
                    };
                    set_atributes(&path, &song, lookup.1.clone())?;
                }
                None => {
                    let song_id = get_user_input()?;
                    let (res, artist) = get_song_metadata(song_id)?;
                    validate_song(&path, &res[0], artist)?
                }
            };
        }
    } else {
        return Err("Not dir or m4a file".into());
    }
    Ok(())
}

fn get_user_input() -> Result<i32, Box<dyn Error>> {
    print!("Enter song id: ");
    io::stdout().flush()?;
    let mut input = String::new();
    io::stdin().read_line(&mut input)?;
    Ok(input.trim().parse()?)
}

fn validate_song(path: &Path, res: &ApiItem, album_artist: String) -> Result<(), Box<dyn Error>> {
    match res {
        ApiItem::Track(track) => set_atributes(&path, &track, album_artist),
        _ => {
            return Err("Provided id not a song".into());
        }
    }
}

fn get_song_metadata(id: i32) -> Result<(Vec<ApiItem>, String), Box<dyn Error>> {
    let res: TrackResult = reqwest::blocking::get(format!(
        "https://itunes.apple.com/lookup?id={id}&entity=song"
    ))?
    .json()?;
    if res.results.len() < 1 {
        return Err("No matching song/album".into());
    }
    let artist_id = res.results[0].artist_id();
    let artist: ArtistResult =
        reqwest::blocking::get(format!("https://itunes.apple.com/lookup?id={artist_id}"))?
            .json()?;
    let artist_name = artist.results[0].artist_name.clone();
    Ok((res.results, artist_name))
}

fn set_atributes(path: &Path, song: &Track, album_artist: String) -> Result<(), Box<dyn Error>> {
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
    let Ok(lyrics): Result<Lyrics, reqwest::Error> = reqwest::blocking::get(format!(
        "https://api.lyrics.ovh/v1/{artist_name}/{song_name}/"
    ))?
    .json() else {
        println!("couldnt get lyrics");
        return Ok(());
    };
    tag.set_lyrics(lyrics.lyrics);
    tag.write_to_path(path)?;
    Command::new("open").arg(path).status()?;
    Ok(())
}
