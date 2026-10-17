# Screen-Captured Effect Demos

Record syscgo in a real terminal. Do not use VHS, terminal emulation, or reconstructed frames.

## Run an effect

Build the CLI from the repository root, then run the effect:

~~~bash
go build -o syscgo ./cmd/syscgo/
./syscgo -effect great-wave -theme nord -duration 9
~~~

The CLI and TUI preview update at 30 fps. Time-aware effects receive 33 ms steps. Frame-based effects advance 1.5 times faster than they did at 20 fps. Full cycles take 8 seconds for Great Wave, 16 for Wasteland, 12 for AK-47, and 14 for AK-47 Text. Add one second to the CLI duration so the recorder can trim its startup frames and keep the full cycle.

For text effects, pass a sample art file:

~~~bash
./syscgo -effect ak-47-text -file assets/SYSC.txt -theme nord -duration 15
~~~

Skip Decrypt until its animation is finished.

## Record and encode

Record the full-screen Ghostty window with GPU Screen Recorder or OBS at 30 fps. Use the testbench output's native resolution (DP-3 is 2560×1440 on the current testbench). Start the recorder in one shell and run the effect in the captured terminal after the recorder logs `update fps`. Keep the AK-47 credits visible, and keep the Wasteland credit in frame.

Start GPU Screen Recorder with:

~~~bash
gpu-screen-recorder -w DP-3 -f 30 -fm cfr -k h264 -encoder cpu -q ultra -ffmpeg-video-opts 'crf=0;preset=ultrafast' -cursor no -c mkv -o /tmp/capture.mkv
~~~

After the effect exits, stop only the recorder with SIGINT. Leave Ghostty open. Convert the recording without scaling:

~~~bash
ffmpeg -ss 1.2 -i /tmp/capture.mkv -t <duration> -an -vf fps=30,format=bgra -c:v libwebp_anim -lossless 1 -quality 100 -compression_level 6 -loop 0 assets/<effect-id>.webp
~~~

Do not pass a WebP preset. The local FFmpeg text preset produced lossy VP8 even with lossless mode enabled. The BGRA filter makes the lossless WebP preserve every decoded capture frame without another color conversion.

GitHub rejects files above 100 MiB. Keep 30 fps and native resolution. If a lossless WebP exceeds 100 MiB, shorten the recording while preserving a complete effect cycle. Check dimensions, frame count, frame rate, loop playback, credits, and file size with ffprobe, then review each clip in a browser.

## Showcase order

Keep Matrix as the hero. Put AK-47 Text second, then order the remaining effects newest to oldest. Put the TUI preview after the effect gallery. Leave Decrypt out of the showcase.
