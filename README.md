# pklVerifier

A DCP validation tool written in Go. It checks each asset listed in a
SMPTE PackingList (PKL) against the files on disk, comparing SHA-1 hashes and
sizes, and cross-checks the ASSETMAP.

> Still under development. Review the results before relying on them.

## Build

```
$ go build -o pklVerifier ./cmd
```

Version info can be embedded at build time:

```
$ go build -ldflags "-X main.Version=v1.1.0 -X main.Commit=$(git rev-parse --short HEAD)" -o pklVerifier ./cmd
```

Prebuilt binaries for Linux, macOS, and Windows are attached to each
[release](https://github.com/mijho/pklVerifier/releases).

## Usage

```
$ pklVerifier validate -i /path/to/DCP [-o results.json]
```

| Flag | Alias | Description |
|------|-------|-------------|
| `--input` | `-i`, `-d` | Path to the DCP directory (required) |
| `--output` | `-o` | Write JSON results to a file instead of stdout |

Results are JSON:

- `results`: one entry per PKL asset, plus one per ASSETMAP entry that no PKL lists.
- `unlisted`: files on disk that the ASSETMAP does not mention.
- `errors`: the total number of failures across the above.

Exit codes: `0` when everything passes, `2` when any check fails, `1` for
fatal problems such as no PKL being found or unparseable XML.

```json
{
  "results": [
    {
      "name": "VIDEO/example_aud.mxf",
      "reported_size": "27031694",
      "actual_size": 27031694,
      "reported_hash": "XXXXXXXXXXXXXXXXXXXXXXXXX+s=",
      "actual_hash": "XXXXXXXXXXXXXXXXXXXXXXXXX+s=",
      "hash_valid": true,
      "size_valid": true
    }
  ],
  "unlisted": [],
  "errors": 0
}
```

## Notes

- File locations come from the ASSETMAP when present, otherwise from the
  PKL's `OriginalFileName`.
- Assets with several ASSETMAP chunks are hashed as the concatenation of
  their chunks, in order, and their size is the sum of the chunk sizes.
- PKL and ASSETMAP filenames are matched case-insensitively. `VOLINDEX.xml`
  is ignored by the unlisted-file check.
- Assets whose paths resolve outside the DCP directory are rejected.

## Licence

MIT. See [LICENCE](LICENCE).

## Testing

```
$ go test -race ./cmd
```

The real-DCP tests extract `testdata/dcpomatic-2.18.51-smpte-2k-ov.tar.gz`, a
2K SMPTE DCP made by DCP-o-matic 2.18.51, and check that it validates. They
also check a copy in a film folder and a copy with one picture byte flipped.

### Try it on the sample DCP

```
$ mkdir -p /tmp/sample && tar -xzf testdata/dcpomatic-2.18.51-smpte-2k-ov.tar.gz -C /tmp/sample
$ go run ./cmd validate -i /tmp/sample/Pkltest_FTR-1_F-178_XX-XX_10_2K_20261019_SMPTE_OV
```

Expect `"errors": 0` and three results: the CPL, the picture MXF and the sound MXF.

To check the film-folder layout, pass the parent directory instead:

```
$ go run ./cmd validate -i /tmp/sample
```

Results are then prefixed with the DCP directory name.

To check that a change is caught, flip a byte in the picture track and rerun:

```
$ printf '\xff' | dd of=/tmp/sample/Pkltest_*_OV/j2c_*.mxf bs=1 seek=1000000 conv=notrunc
$ go run ./cmd validate -i /tmp/sample/Pkltest_FTR-1_F-178_XX-XX_10_2K_20261019_SMPTE_OV
```

Expect `"errors": 1` and `"hash_valid": false` on the picture track.

### Regenerating the fixture

`testdata/generate.sh` rebuilds the sample from a 5-second test clip. It needs
`dcpomatic2_create`, `dcpomatic2_cli` and `ffmpeg` on `PATH`. The output
contains no real content, so it can be committed.
