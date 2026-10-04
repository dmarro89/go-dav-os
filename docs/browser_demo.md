# Browser Demo

The browser demo runs the real `go-dav-os` ISO in QEMU and exposes its VGA and
PS/2 console through noVNC. It is a shared, single-session demo: only one
browser can control the guest at a time. The reset button restarts QEMU and
clears guest memory.

The guest has no network device, disk image, or QEMU monitor. The VNC and
control ports bind to loopback inside the container; nginx exposes the page and
WebSocket endpoint on the web service port. Keep the service at one instance so
the single-session limit and reset behavior apply globally.

## Deploy with Render

1. Merge this change into `main`.
2. In Render, create a Blueprint from the repository and sync its root
   `render.yaml`.
3. Wait for the `/api/healthz` check to pass, then use the public `onrender.com`
   URL assigned to the `go-dav-os-demo` service.
4. Add that URL to the README's browser-demo link after the service is live.

The Blueprint selects the `demo-runtime` image stage and uses one free web
instance in Singapore. Render's free service may spin down when idle; use an
always-on plan if the demo needs uninterrupted availability. See Render's
[Blueprint reference](https://render.com/docs/blueprint-spec) and [free
instance details](https://render.com/docs/free).

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
