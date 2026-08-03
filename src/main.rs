use clap::Parser;
use directories::UserDirs;
use mp4ameta::Tag;
use std::env;
use std::path::Path;
use std::{ffi::OsStr, path::PathBuf};
use walkdir::{Error, WalkDir};
#[derive(Parser, Debug)]
struct Args {
    dir: String,
}
//
// struct Song {
//     title: String,
// }
fn main() -> Result<(), Error> {
    walk_dir()
}

fn walk_dir() -> Result<(), Error> {
    let current_dir = env::current_dir().expect("no directory");
    for entry in WalkDir::new(current_dir) {
        let entry = entry?;
        if entry.path().extension() == Some(OsStr::new("m4a"))
            && let path = entry.path().to_path_buf()
        {
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
//
// async fn get_song(id: i32) -> String {
//     let response = reqwest::get(format!(
//         "https://itunes.apple.com/lookup?id={id}&entity=song"
//     ))
//     .await
//     .text()
//     .await;
//     return response;
// }

// fn get_song_json() -> Result<JSON>
