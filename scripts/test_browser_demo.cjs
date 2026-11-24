const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const { chromium } = require("playwright-core");

async function checkDemo(url) {
  const browser = await chromium.launch({
    channel: "chrome",
    headless: true,
  });
  try {
    const firstContext = await browser.newContext();
    const secondContext = await browser.newContext();
    const first = await firstContext.newPage();
    const second = await secondContext.newPage();
    const connected = (page) => page.getByText(
      "Connected — click the console to type", { exact: true }
    ).waitFor({ timeout: 30000 });

    await first.goto(url);
    await connected(first);
    await first.locator("#screen canvas").click();
    await first.keyboard.type("version");
    await first.keyboard.press("Enter");

    await second.goto(url);
    await second.getByText("Console unavailable — retrying…", {
      exact: true,
    }).waitFor({ timeout: 20000 });
    await firstContext.close();
    await connected(second);

    const guestLog = execFileSync("docker", [
      "exec", "davos-browser-demo", "cat", "/tmp/go-dav-os-debug.log",
    ], { encoding: "utf8" });
    assert.match(guestLog, /DavOS [0-9.]+ \(64bit\)/);
    console.log("Guest keyboard and automatic session handoff verified");

    const reset = second.waitForResponse((response) => response.url().endsWith("/api/reset"));
    await second.locator("#reset").click();
    assert.equal((await reset).status(), 202);
    await connected(second);
    console.log("Browser reconnected after the reset");
  } finally {
    await browser.close();
  }
}

checkDemo(process.argv[2]).catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
