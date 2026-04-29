#!/usr/bin/env zsh

[[ -n "$1" ]] || {
  print -u2 "usage: $0 <url>"
  exit 1
}

f="$(
  yt-dlp \
    -f 'ba[ext=m4a]/ba' \
    -o "$HOME/Downloads/%(title)s.%(ext)s" \
    --print after_move:filepath \
    "$1"
)"

[[ "${f##*.}" = m4a ]] || afconvert "$f" -f m4af "${f%.*}.m4a"
