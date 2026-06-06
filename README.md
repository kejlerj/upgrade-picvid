# upgrade-picvid

A CLI tool to modernize old video files so they can be opened and previewed by modern apps like **macOS Photos**.

## Why?

macOS Photos refuses to open or preview videos encoded with legacy codecs (MPEG-1, MPEG-2, DV, Cinepak, etc.) or stored in old containers (AVI, MTS, VOB, FLV…). These formats were common on digital cameras and camcorders from the 2000s and early 2010s.

`upgrade-picvid` scans a folder for those legacy files, re-encodes them to H.264/AAC inside an MP4 container — the format Photos and virtually every modern app understands — and moves the originals to `~/legacy-picvid` so nothing is lost.

## Requirements

- [ffmpeg](https://ffmpeg.org/) and `ffprobe` (bundled with ffmpeg)

If they are not installed, the tool will offer to install them via Homebrew.

## Installation

```bash
git clone https://github.com/kejlerj/upgrade-picvid.git
cd upgrade-picvid
go build -o upgrade-picvid
```

## Usage

```bash
./upgrade-picvid -folder /path/to/videos [options]
```

### Parameters

| Flag | Default | Description |
|------|---------|-------------|
| `-folder` | *(required)* | Path to the folder containing videos to convert |
| `-recursive` | `false` | Also scan subfolders |
| `-crf` | `18` | Quality level: `0` (lossless) → `51` (worst). `18` is visually near-lossless, `23` is ffmpeg's default |
| `-preset` | `medium` | Encoding speed vs. file size trade-off (see below) |

### Presets

From fastest to slowest encoding (slower = smaller file, same quality):

`ultrafast` → `superfast` → `veryfast` → `faster` → `fast` → **`medium`** → `slow` → `slower` → `veryslow`

A slower preset does **not** affect quality — only how much the encoder works to compress the output. `medium` is a good default for most use cases.

### Examples

Convert all legacy videos in a folder:
```bash
./upgrade-picvid -folder ~/Movies/old-camcorder
```

Recursively scan subfolders, with slightly smaller output files:
```bash
./upgrade-picvid -folder ~/Movies/old-camcorder -recursive -preset slow
```

Prioritize speed over file size (useful for large batches):
```bash
./upgrade-picvid -folder ~/Movies/old-camcorder -preset fast -crf 20
```

## What happens to the original files?

Original files are **not deleted**. They are moved to `~/legacy-picvid` after a successful conversion. You can delete them manually once you've verified the output looks good.

## Supported legacy formats

**Codecs:** MPEG-1, MPEG-2, MJPEG, Theora, Sorenson Video (SVQ1/SVQ3), Cinepak, FLV1, VP6, DV, H.263

**Containers:** AVI, ASF, MTS, M2TS, TS, VOB, FLV, 3GP, 3G2, MPG, MPEG

## License

MIT
