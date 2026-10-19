#!/bin/sh
# Regenerates testdata/dcpomatic-2.18.51-smpte-2k-ov.tar.gz with DCP-o-matic 2.18.
# Requires dcpomatic2_create, dcpomatic2_cli and ffmpeg on PATH.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 5-second 1080p test clip with a 440 Hz tone.
ffmpeg -loglevel error -y -f lavfi -i testsrc=size=1920x1080:rate=24 \
  -f lavfi -i sine=frequency=440:sample_rate=48000 -t 5 \
  -c:v libx264 -pix_fmt yuv420p -c:a aac -shortest "$work/clip.mp4"

dcpomatic2_create -n PKLTEST -c FTR --twok --standard SMPTE --no-encrypt \
  -o "$work/film" "$work/clip.mp4" >/dev/null
dcpomatic2_cli -r -n make-dcp "$work/film" >/dev/null

cd "$work/film"
tar -czf "$here/dcpomatic-2.18.51-smpte-2k-ov.tar.gz" Pkltest_*_OV
echo "wrote $here/dcpomatic-2.18.51-smpte-2k-ov.tar.gz"
