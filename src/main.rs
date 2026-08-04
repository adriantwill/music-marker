use clap::Parser;
use mp4ameta::{Img, Tag};
use serde::Deserialize;
use std::env;
use std::error::Error;
use std::ffi::OsStr;
use std::fs;
use std::path::Path;
use std::process::Command;
use walkdir::WalkDir;
#[derive(Parser, Debug)]
struct Args {
    id: i32,
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
    artist_name: String,
    collection_artist_name: Option<String>,
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

fn main() -> Result<(), Box<dyn Error>> {
    walk_dir()
}

fn walk_dir() -> Result<(), Box<dyn Error>> {
    let current_dir = env::current_dir().expect("no directory");
    let args = Args::parse();
    let song_id = args.id;
    for entry in WalkDir::new(current_dir) {
        let entry = entry?;
        if entry.path().extension() == Some(OsStr::new("m4a"))
            && let path = entry.path().to_path_buf()
        {
            print!("{:?}", path);
            let res: SongResult = reqwest::blocking::get(format!(
                "https://itunes.apple.com/lookup?id={song_id}&entity=song"
            ))?
            .json()?;
            if res.results.len() != 1 {
                return Err("Not 1 Song".into());
            }
            set_atributes(&path, &res.results[0])?;
            Command::new("open").arg(&path).status()?;
        }
    }
    Ok(())
}

fn set_atributes(path: &Path, song: &Song) -> Result<(), reqwest::Error> {
    let mut tag = Tag::read_from_path(path).unwrap();
    tag.set_title(&song.track_name);
    tag.set_artist(&song.artist_name);
    tag.set_album(&song.collection_name);
    if let Some(album_artist) = &song.collection_artist_name {
        tag.set_album_artist(album_artist);
    }
    tag.set_genre(&song.primary_genre_name);
    tag.set_year(&song.release_date);
    tag.set_track_number(song.track_number);
    tag.set_total_tracks(song.track_count);
    tag.set_disc_number(song.disc_number);
    tag.set_total_discs(song.disc_count);
    let artwork = reqwest::blocking::get(
        song.artwork_url_100
            .replace("100x100bb.jpg", "3000x3000bb.jpg"),
    )
    .unwrap()
    .bytes()
    .unwrap();
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
    println!("{}", tag.artist().unwrap());
    Ok(())
}
