# url-extract
Extract URLs from websites using a headless browser

## Get started

This project uses a headless chromium instance for navigating websites.

```bash
docker container run -it --name headless-chromium --rm -p 127.0.0.1:9222:9222 --entrypoint "chromium-browser" zenika/alpine-chrome --headless --disable-gpu --no-sandbox --disable-software-rasterizer --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222 --enable-logging --disable-dev-shm-usage --disable-sync --disable-background-networking --no-first-run --no-pings --metrics-recording-only --safebrowsing-disable-auto-update --mute-audio

```

Some of the switches might be deprecated or entirely useless. Sources are
- https://github.com/GoogleChrome/puppeteer/issues/940#issuecomment-336423912
- https://peter.sh/experiments/chromium-command-line-switches/

## Running

Once the docker container is running, and navigating to `127.0.0.1:9222` works in your local browser, we can try navigating to a website using the headless instance. By default we are trying to extract `.m3u8` files from our target (see `main.go`).

```bash
go build && ./url-extract -url https://castr.io/hlsplayer
```

The result should look something like `https://cstr-x.castr.io/castr/live_x/index.m3u8`.

## Limitations

This will only find URLs (in particular `m3u8` ones) that will be loaded without any user interaction. Media that is only played e.g. on-click will not be found.