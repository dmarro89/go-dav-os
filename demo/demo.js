import RFB from "/novnc/core/rfb.js";

const screen = document.getElementById("screen");
const status = document.getElementById("status");
const resetButton = document.getElementById("reset");
const initialRetryMs = 500;
const maxRetryMs = 8000;
let rfb = null;
let retryTimer = null;
let retryDelay = initialRetryMs;
let resetting = false;

function cancelRetry() {
  window.clearTimeout(retryTimer);
  retryTimer = null;
}

function scheduleRetry() {
  if (resetting || rfb || retryTimer !== null) {
    return;
  }
  retryTimer = window.setTimeout(connect, retryDelay);
  retryDelay = Math.min(retryDelay * 2, maxRetryMs);
}

function connect() {
  cancelRetry();
  if (rfb || resetting) {
    return;
  }

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const session = new RFB(screen, `${protocol}//${window.location.host}/websockify`);
  rfb = session;
  session.scaleViewport = true;
  session.resizeSession = false;
  session.addEventListener("connect", () => {
    if (rfb !== session) {
      return;
    }
    retryDelay = initialRetryMs;
    if (!resetting) {
      status.textContent = "Connected — click the console to type";
      resetButton.disabled = false;
    }
  });
  session.addEventListener("disconnect", () => {
    if (rfb !== session) {
      return;
    }
    rfb = null;
    if (!resetting) {
      status.textContent = "Console unavailable — retrying…";
      resetButton.disabled = false;
      scheduleRetry();
    }
  });
  session.addEventListener("credentialsrequired", () => {
    if (rfb === session && !resetting) {
      status.textContent = "The demo needs to be restarted";
      resetButton.disabled = false;
    }
  });
  session.addEventListener("securityfailure", () => {
    if (rfb === session && !resetting) {
      status.textContent = "Could not connect to the demo";
    }
  });
}

resetButton.addEventListener("click", async () => {
  resetting = true;
  resetButton.disabled = true;
  cancelRetry();
  status.textContent = "Resetting the shared machine…";
  try {
    const response = await fetch("/api/reset", { method: "POST" });
    if (!response.ok) {
      throw new Error(response.status === 429 ? "Wait a few seconds before resetting again." : "Reset failed.");
    }
    const previous = rfb;
    rfb = null;
    previous?.disconnect();
    screen.replaceChildren();
    retryDelay = initialRetryMs;
    status.textContent = "Machine restarted — reconnecting…";
  } catch (error) {
    status.textContent = error.message;
  } finally {
    resetting = false;
    resetButton.disabled = false;
    scheduleRetry();
  }
});

connect();
