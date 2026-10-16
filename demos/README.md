# Screen-Captured Effect Demos

Showcase captures must record `syscgo` running in a real terminal. Do not use VHS for new captures.

## Run an effect

From the repository root, build the CLI and start an effect:

```bash
go build -o syscgo ./cmd/syscgo/
./syscgo -effect great-wave -theme nord -duration 0
```

For effects that use ASCII input, pass an art file:

```bash
./syscgo -effect decrypt -file assets/SYSC.txt -theme nord -duration 0
```

Press `Ctrl+C` to stop the effect.

## Record and encode

Record the live terminal window with a screen recorder such as GPU Screen Recorder or OBS at 1920×1080 or the monitor's native resolution when it is higher. Keep the terminal readable and capture a full effect cycle. Encode the recording as an animated WebP:

```bash
ffmpeg -i capture.mkv \
  -vf "fps=15" -an \
  -c:v libwebp_anim -lossless 0 -preset text -quality 85 -loop 0 \
  assets/<effect-id>.webp
```

Keep each WebP at or below 5 MiB when the animation remains readable. Shorten the clip or reduce the frame rate before lowering the capture resolution. Review the result in a browser to confirm that it loops and that the terminal text stays legible.

## Convert existing showcase GIFs

Re-encode existing showcase GIFs so the README can use the smaller WebP files:

```bash
ffmpeg -i assets/<source>.gif \
  -vf "fps=15" -an \
  -c:v libwebp_anim -lossless 0 -preset text -quality 85 -loop 0 \
  assets/<effect-id>.webp
```

Review each output in a browser before changing README links. Remove a GIF only after all references to it have moved to the verified WebP.
