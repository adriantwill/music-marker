#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: ./test-artist.sh <song-id>"
    echo "Example: ./test-artist.sh 1789237364"
    exit 1
fi

SONG_ID=$1

echo "Fetching metadata for song ID: $SONG_ID"
echo ""

curl -s "https://itunes.apple.com/lookup?id=$SONG_ID" | jq -r '.results[0] | "Artist: \(.artistName)\nAlbum: \(.collectionName)\nTrack: \(.trackName)\nGenre: \(.primaryGenreName)\nExplicit: \(.trackExplicitness)"'

echo ""
echo "Raw artist name (with visible special chars):"
curl -s "https://itunes.apple.com/lookup?id=$SONG_ID" | jq -r '.results[0].artistName' | od -c
