#!/usr/bin/env python3

import http.server
import logging
import os
from pathlib import Path
import signal
import subprocess
import threading
import time
from urllib.parse import urlsplit


ISO_PATH = "/opt/go-dav-os/dav-go-os.iso"
DEBUG_LOG = Path("/tmp/go-dav-os-debug.log")
RESET_COOLDOWN_SECONDS = 10
QEMU_COMMAND = [
    "qemu-system-x86_64",
    "-m",
    "128M",
    "-smp",
    "1",
    "-cdrom",
    ISO_PATH,
    "-snapshot",
    "-nic",
    "none",
    "-display",
    "none",
    "-vnc",
    "127.0.0.1:0,websocket=5700",
    "-monitor",
    "none",
    "-serial",
    "none",
    "-no-reboot",
    "-no-shutdown",
    "-debugcon",
    f"file:{DEBUG_LOG}",
]


class ResetRateLimited(Exception):
    pass


class VMController:
    def __init__(self):
        self._lock = threading.Lock()
        self._process = None
        self._last_reset = None

    def start(self):
        DEBUG_LOG.unlink(missing_ok=True)
        self._process = subprocess.Popen(
            QEMU_COMMAND,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.STDOUT,
            close_fds=True,
        )

    def is_ready(self):
        process = self._process
        if process is None or process.poll() is not None:
            return False
        try:
            return "Welcome to DavOS" in DEBUG_LOG.read_text(errors="ignore")
        except FileNotFoundError:
            return False

    def reset(self):
        with self._lock:
            now = time.monotonic()
            if (
                self._last_reset is not None
                and now - self._last_reset < RESET_COOLDOWN_SECONDS
            ):
                raise ResetRateLimited
            self._last_reset = now
            self._stop_locked()
            self.start()

    def close(self):
        with self._lock:
            self._stop_locked()

    def _stop_locked(self):
        process = self._process
        self._process = None
        if process is None or process.poll() is not None:
            return
        process.terminate()
        try:
            process.wait(timeout=3)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()


class ControlHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path != "/healthz":
            self.send_error(404)
            return
        if not self.server.vm.is_ready():
            self.send_error(503, "Guest is still booting")
            return
        self.send_response(200)
        self.send_header("Content-Length", "0")
        self.end_headers()

    def do_POST(self):
        if self.path != "/reset":
            self.send_error(404)
            return
        origin = urlsplit(self.headers.get("Origin", ""))
        host = self.headers.get("Host", "").lower()
        if not origin.netloc or origin.netloc.lower() != host:
            self.send_error(403, "Reset requests must come from this site")
            return
        try:
            self.server.vm.reset()
        except ResetRateLimited:
            self.send_error(429, "Please wait before resetting again")
            return
        except OSError:
            logging.exception("Could not restart the demo VM")
            self.send_error(503, "Could not restart the demo VM")
            return
        self.send_response(202)
        self.send_header("Content-Length", "0")
        self.end_headers()

    def log_message(self, format, *args):
        logging.info("%s - %s", self.address_string(), format % args)


def write_nginx_config(port):
    template = Path("/opt/go-dav-os/demo/nginx.conf.template")
    config = Path("/tmp/go-dav-os-nginx.conf")
    Path("/tmp/nginx-client-body").mkdir(exist_ok=True)
    Path("/tmp/nginx-fastcgi").mkdir(exist_ok=True)
    Path("/tmp/nginx-proxy").mkdir(exist_ok=True)
    Path("/tmp/nginx-scgi").mkdir(exist_ok=True)
    Path("/tmp/nginx-uwsgi").mkdir(exist_ok=True)
    config.write_text(template.read_text().replace("__PORT__", str(port)))
    return config


def main():
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    port = int(os.environ.get("PORT", "8080"))
    if port < 1 or port > 65535:
        raise ValueError("PORT must be between 1 and 65535")

    vm = VMController()
    vm.start()
    control = http.server.ThreadingHTTPServer(("127.0.0.1", 8081), ControlHandler)
    control.vm = vm
    threading.Thread(target=control.serve_forever, daemon=True).start()

    nginx = subprocess.Popen(
        ["nginx", "-p", "/tmp", "-c", str(write_nginx_config(port)), "-g", "daemon off;"],
        stdin=subprocess.DEVNULL,
    )

    def stop(signum, frame):
        raise SystemExit(0)

    signal.signal(signal.SIGTERM, stop)
    signal.signal(signal.SIGINT, stop)
    try:
        nginx.wait()
        if nginx.returncode != 0:
            raise SystemExit(nginx.returncode)
    finally:
        control.shutdown()
        control.server_close()
        vm.close()
        if nginx.poll() is None:
            nginx.terminate()
            nginx.wait(timeout=3)


if __name__ == "__main__":
    main()
