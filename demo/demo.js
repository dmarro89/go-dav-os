import RFB from "/novnc/core/rfb.js";

const screen = document.getElementById("screen");
const status = document.getElementById("status");
const resetButton = document.getElementById("reset");
let rfb;

function connect() {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  rfb = new RFB(screen, `${protocol}//${window.location.host}/websockify`);
  rfb.scaleViewport = true;
  rfb.resizeSession = false;
  rfb.addEventListener("connect", () => {
    status.textContent = "Connected — click the console to type";
    resetButton.disabled = false;
  });
  rfb.addEventListener("disconnect", () => {
    status.textContent = "Disconnected — another visitor may be using the console";
    resetButton.disabled = false;
  });
  rfb.addEventListener("credentialsrequired", () => {
    status.textContent = "The demo needs to be restarted";
  });
  rfb.addEventListener("securityfailure", () => {
    status.textContent = "Could not connect to the demo";
  });
}

resetButton.addEventListener("click", async () => {
  resetButton.disabled = true;
  status.textContent = "Resetting the shared machine…";
  try {
    const response = await fetch("/api/reset", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    });
    if (!response.ok) {
      throw new Error(response.status === 429 ? "Wait a few seconds before resetting again." : "Reset failed.");
    }
    if (rfb) {
      rfb.disconnect();
    }
    screen.replaceChildren();
    window.setTimeout(connect, 500);
    status.textContent = "Machine restarted — reconnecting…";
  } catch (error) {
    status.textContent = error.message;
    resetButton.disabled = false;
  }
});

connect();
