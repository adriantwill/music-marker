use clap::Parser;
use mp4ameta::{Img, Tag};
use serde::Deserialize;
use std::collections::HashMap;
use std::env::current_dir;
use std::error::Error;
use std::ffi::OsStr;
use std::io;
use std::path::{Path, PathBuf};
use std::process::Command;
#[derive(Parser, Debug)]
struct Args {
    id: Option<i32>,
    path: Option<PathBuf>,
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
        if res.len() != 1 || res[0].wrapper_type != "track" {
            return Err("Provided id not a song".into());
        }
        set_atributes(&path, &res[0], artist)?;
    } else if path.is_dir() {
        let lookup = if let Some(song_id) = args.id {
            let (res, artist) = get_song_metadata(song_id)?;
            let mut song_lookup = HashMap::new();

            for song in res {
                if song.wrapper_type == "track" {
                    song_lookup.insert(song.track_name.clone(), song);
                }
            }

            Some((song_lookup, artist))
        } else {
            None
        };
        let path = path.display();
        for entry in glob::glob(&format!("{path}/*.m4a"))? {
            let path = entry?;
            match &lookup {
                Some(lookup) => {
                    let tag = Tag::read_from_path(&path)?;
                    let Some(song_name) = tag.title().or_else(|| path.file_stem()?.to_str()) else {
                        print!("no title found in file {}", path.display());
                        continue;
                    };
                    let Some(song) = lookup.0.get(song_name) else {
                        print!("no song found called {}", song_name);
                        continue;
                    };
                    set_atributes(&path, &song, lookup.1.clone())?;
                }
                None => {
                    let song_id = get_user_input()?;
                    let (res, artist) = get_song_metadata(song_id)?;
                    if res.len() != 1 || res[0].wrapper_type != "track" {
                        return Err("Provided id not a song".into());
                    }
                    set_atributes(&path, &res[0], artist.clone())?;
                }
            };
        }
    } else {
        return Err("Not dir or m4a file".into());
    }
    Ok(())
}

fn get_user_input() -> Result<i32, std::num::ParseIntError> {
    let mut input = String::new();
    io::stdin()
        .read_line(&mut input)
        .expect("couldnt read input");
    Ok(input.trim().parse()?)
}

fn get_song_metadata(id: i32) -> Result<(Vec<Song>, String), reqwest::Error> {
    let res: SongResult = reqwest::blocking::get(format!(
        "https://itunes.apple.com/lookup?id={id}&entity=song"
    ))?
    .json()?;
    let artist_id = res.results[0].artist_id;
    let artist: Artist =
        reqwest::blocking::get(format!("https://itunes.apple.com/lookup?id={artist_id}"))?
            .json()?;
    Ok((res.results, artist.artist_name))
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
    tag.write_to_path(path)?;
    Command::new("open").arg(path).status()?;
    Ok(())
}
