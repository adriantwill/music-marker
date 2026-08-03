use clap::Parser;
use mp4ameta::Tag;
use serde::Deserialize;
use std::env;
use std::error::Error;
use std::ffi::OsStr;
use std::path::Path;
use walkdir::WalkDir;
#[derive(Parser, Debug)]
struct Args {
    id: i32,
}
#[derive(Deserialize)]
struct SongResult {
    results: Vec<ItunesTrack>,
}
#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ItunesTrack {
    artist_name: String,
    artist_id: u64,
    track_id: u64,
    collection_artist_name: Option<String>,
    collection_name: String,
    track_name: String,
    primary_genre_name: String,
    release_date: String,
    track_explicitness: String,
    track_number: u32,
    track_count: u32,
    disc_number: u32,
    disc_count: u32,
    artwork_url_100: String,
}

fn main() -> Result<(), Box<dyn Error>> {
    walk_dir()
}

fn walk_dir() -> Result<(), Box<dyn Error>> {
    let current_dir = env::current_dir().expect("no directory");
    let args = Args::parse();
    for entry in WalkDir::new(current_dir) {
        let entry = entry?;
        if entry.path().extension() == Some(OsStr::new("m4a"))
            && let path = entry.path().to_path_buf()
        {
            get_song(args.id)?;
            set_atributes(&path);
        }
    }
    Ok(())
}

fn set_atributes(path: &Path) {
    let mut tag = Tag::read_from_path(path).unwrap();
    tag.set_artist("artist");
    tag.set_album_artist("artist,test");
    tag.write_to_path("WiseMan-FrankOcean-Revised.m4a").unwrap();
    println!("{}", tag.artist().unwrap());
}
fn get_song(id: i32) -> Result<(), reqwest::Error> {
    let res: SongResult = reqwest::blocking::get(format!(
        "https://itunes.apple.com/lookup?id={id}&entity=song"
    ))?
    .json()?;
    print!("{}", res.results[0].artist_name);
    Ok(())
}

// fn get_song_json() -> Result<JSON>
