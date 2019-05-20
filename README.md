# url-extract
Extract URLs from websites using a headless browser

## Getting started

This project uses a headless chromium instance for navigating websites.

```bash
docker run -it --name headless-chromium --rm -p 127.0.0.1:9222:9222 --entrypoint "chromium-browser" zenika/alpine-chrome --headless --disable-gpu --no-sandbox --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222 --enable-logging --autoplay-policy=no-user-gesture-required --disable-software-rasterizer --disable-dev-shm-usage --disable-sync --disable-background-networking --no-first-run --no-pings --metrics-recording-only --safebrowsing-disable-auto-update --mute-audio

```

Some of the switches might be deprecated or entirely useless. Sources are
- https://github.com/GoogleChrome/puppeteer/issues/940#issuecomment-336423912
- https://peter.sh/experiments/chromium-command-line-switches/
- https://github.com/obsproject/obs-browser/issues/105


## Running

Once the docker container is running, and navigating to `127.0.0.1:9222` works in your local browser, we can try navigating to a website using the headless instance. By default we are trying to extract media files from our target (see `main.go`).

```bash
go build && ./url-extract -quiet -heuristics -url https://castr.io/hlsplayer
```

The result should look something like `https://cstr-x.castr.io/castr/live_x/index.m3u8`.

### Detection

The request headers when accessing a website (at the time of writing) are

```
Connection: keep-alive
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/73.0.3683.103 Safari/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8
Accept-Encoding: gzip, deflate, br
```

## Limitations

Generally, only URLs that are automatically loaded via the network are found. URLs that are only loaded after user interaction, e.g. on-click will not be found. The autoplay-policy `no-user-gesture-required` allows websites to play some media files without user-interaction.

### Heuristics

We try to click elements with ids and classes that "look like" media players to start playing media that will not auto play. Check `heuristics.go` for more information. This is always open to improvement or adjustment. Note that using such heuristics can lead to false-negatives by interacting with the website in a way that stops loading a resource that should normally be found.

### Todo

Heuristics are not able to access sites that embed e.g. media players inside iframes. Properly accessing iframes seems to be an open issue in the `chromedp` project.