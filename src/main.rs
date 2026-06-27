use std::path::Path;

use clap::Parser;
use directories::UserDirs;
use mp4ameta::Tag;
// #[derive(Parser, Debug)]
// struct Args {
//     #[arg(default_value_t=directories::UserDirs::download_dir(&self))]
//     dir: Path,
// }
fn main() {
    // let args = Args::parse();
    let user = UserDirs::new().expect("no user found");
    if let Some(downloads) = UserDirs::download_dir(&user) {
        let mut tag =
            Tag::read_from_path(downloads.join("WiseMan-FrankOcean-Revised.m4a")).unwrap();
        tag.set_artist("Frank Ocean");
        tag.write_to_path(downloads.join("WiseMan-FrankOcean-Revised.m4a"))
            .unwrap();
        println!("{}", tag.artist().unwrap());
    }
}
