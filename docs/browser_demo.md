# Browser Demo

The browser demo runs the real `go-dav-os` ISO in QEMU and exposes its VGA and
PS/2 console through noVNC. It is a shared, single-session demo: only one
browser can control the guest at a time. The reset button restarts QEMU and
clears guest memory. The page automatically retries while the console is busy
or the guest is restarting. Valid reset requests share a ten-second cooldown.

The guest has no network device, disk image, or QEMU monitor. The VNC and
control ports bind to loopback inside the container; nginx exposes the page and
WebSocket endpoint on the web service port. Keep the service at one instance so
the single-session limit and reset behavior apply globally.

## Deploy with Render

The **Browser demo** workflow builds and smoke-tests the Linux AMD64 image. On
pushes to `main`, it publishes that same image as
`ghcr.io/dmarro89/go-dav-os-demo` with the commit SHA and `latest` tags.

1. Wait for the first publish on `main`. GitHub Packages starts the package as
   private. To let Render pull it without credentials, make only the demo
   package public in its GitHub package settings.
2. Sync the Render Blueprint against `main`, using the root `render.yaml`.
   Manually sync again if Render tried to pull the image before it was available.
3. Wait for `/api/healthz` to pass, then open the service's `onrender.com` URL.

The Blueprint uses one free web instance in Singapore. Render pulls the image
from GHCR, and a free instance may spin down when idle. See the
[Render image deploy guide](https://render.com/docs/deploying-an-image).

To deploy after each successful publish, copy the image-backed Render service's
deploy hook URL into the GitHub Actions repository secret `RENDER_DEPLOY_HOOK_URL`.
The workflow then requests a deploy using the tested image's commit tag. Without
the secret, use **Manual Deploy → Deploy latest reference** in Render.

## Run locally

Build the ISO and browser-demo image with Docker:

```sh
docker build --platform linux/amd64 \
  --build-arg FINAL_STAGE=demo-runtime \
  -t go-dav-os-demo .
```

Run it with a single-session memory limit:

```sh
docker run --rm --platform linux/amd64 \
  --memory=512m --cpus=1 \
  -e PORT=8080 -p 8080:8080 \
  go-dav-os-demo
```

Open <http://localhost:8080>. The health endpoint is
<http://localhost:8080/api/healthz>.
