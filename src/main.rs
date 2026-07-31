use std::path::{self, Path};
use walkdir::{Error, WalkDir};
//
// use clap::Parser;
use directories::UserDirs;
use mp4ameta::Tag;
// #[derive(Parser, Debug)]
// struct Args {
//     #[arg(default_value_t=directories::UserDirs::download_dir(&self))]
//     dir: Path,
// }
//
// struct Song {
//     title: String,
// }
fn main() -> Result<(), Error> {
    // let args = Args::parse();
    walk_dir()

    // let title = "test";
    // let artist = "test";
    // let path = "test";
    // let user = UserDirs::new().expect("no user found");
    // print!(get_song(1657472093));
    // if let Some(downloads) = UserDirs::download_dir(&user) {
    //     let mut tag = Tag::read_from_path(downloads.join(path)).unwrap();
    //     tag.set_artist(artist);
    //     tag.set_album_artist(artist);
    //     tag.write_to_path(downloads.join("WiseMan-FrankOcean-Revised.m4a"))
    //         .unwrap();
    //     println!("{}", tag.artist().unwrap());
    // }
}

fn walk_dir() -> Result<(), Error> {
    let user = UserDirs::new().expect("no user found");
    if let Some(downloads) = UserDirs::download_dir(&user) {
        for entry in WalkDir::new(downloads) {
            println!("{:?}", entry?.file_type());
        }
    }
    Ok(())
}

// fn set_atributes(path: Path) -> Result<(), Error> {
//     let user = UserDirs::new().expect("no user found");
//     if let Some(downloads) = UserDirs::download_dir(&user) {
//         let mut tag = Tag::read_from_path(downloads.join(temp)).unwrap();
//     }
//     Ok(())
// }
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
