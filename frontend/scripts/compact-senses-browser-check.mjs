// Real-browser layout smoke test. Build first, then set CHROME_PATH if Chrome is not on PATH.
// Usage: CHROME_PATH=/path/to/chrome node scripts/compact-senses-browser-check.mjs
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { createServer } from "node:http";
import { mkdtemp, readFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve, extname } from "node:path";

const root = resolve("build");
const server = createServer(async (request, response) => {
  try {
    if (request.url === "/wails/runtime.js") {
      response.writeHead(200, { "Content-Type": "text/javascript" });
      response.end(`const handlers = {};
export const Events = { On(name, fn) { handlers[name] = fn; return () => {}; } };
export const Call = { async ByName(name) {
  if (name.endsWith('.Lookup')) {
    setTimeout(() => handlers['lookup:result']?.({data:{SearchID:'1',Language:'es',Result:{Status:'found',Origin:'offline',Entries:[{
      Headword:'touch', Language:'es', Pronunciations:[{Text:'[ˈtowtʃ]'}],
      Source:{Name:'Example',URL:'https://example.com',LicenseURL:'https://example.com/license'},
      Meanings:Array.from({length:6},(_,i)=>({PartOfSpeech:'Sustantivo masculino con una denominación extraordinariamente larga',SenseGroup:'g1',Definitions:[i === 0 ? 'Una definición larga con supercalifragilisticexpialidocioussupercalifragilisticexpialidocious que debe envolver correctamente en una pantalla estrecha.' : 'Sentido ' + (i+1)]}))
    }]}}}),0); return '1';
  }
  if (name.endsWith('.Dictionaries') || name.endsWith('.ListVoices')) return [];
  return {};
} };`);
      return;
    }
    const name =
      request.url === "/"
        ? "/index.html"
        : new URL(request.url, "http://localhost").pathname;
    const file = resolve(root, "." + name);
    if (!file.startsWith(root + "/")) throw new Error("Invalid path");
    const mime = {
      ".html": "text/html",
      ".js": "text/javascript",
      ".css": "text/css",
    };
    const content = await readFile(file);
    response.writeHead(200, {
      "Content-Type": mime[extname(file)] || "application/octet-stream",
    });
    response.end(content);
  } catch {
    response.writeHead(404).end();
  }
});
await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
const port = server.address().port;
const profile = await mkdtemp(join(tmpdir(), "wdict-browser-check-"));
const chrome = spawn(
  process.env.CHROME_PATH || "google-chrome",
  [
    "--headless=new",
    "--no-sandbox",
    "--disable-gpu",
    "--remote-debugging-port=0",
    "--user-data-dir=" + profile,
    "about:blank",
  ],
  { stdio: "ignore" },
);
let ws;
try {
  let debugPort;
  for (let i = 0; i < 100; i++) {
    try {
      debugPort = (
        await readFile(join(profile, "DevToolsActivePort"), "utf8")
      ).split("\n")[0];
      break;
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }
  assert.ok(debugPort, "Chrome did not start");
  const target = await fetch(
    "http://127.0.0.1:" +
      debugPort +
      "/json/new?" +
      encodeURIComponent("http://127.0.0.1:" + port),
    { method: "PUT" },
  ).then((r) => r.json());
  ws = new WebSocket(target.webSocketDebuggerUrl);
  await new Promise((resolve, reject) => {
    ws.onopen = resolve;
    ws.onerror = reject;
  });
  let id = 0;
  const pending = new Map();
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (!data.id) return;
    const item = pending.get(data.id);
    pending.delete(data.id);
    if (data.error) item.reject(new Error(data.error.message));
    else item.resolve(data.result);
  };
  function command(method, params = {}) {
    return new Promise((resolve, reject) => {
      const n = ++id;
      pending.set(n, { resolve, reject });
      ws.send(JSON.stringify({ id: n, method, params }));
    });
  }
  async function evaluate(expression) {
    const result = await command("Runtime.evaluate", {
      expression,
      returnByValue: true,
      awaitPromise: true,
    });
    if (result.exceptionDetails) throw new Error(result.exceptionDetails.text);
    return result.result.value;
  }
  await command("Page.enable");
  await command("Emulation.setDeviceMetricsOverride", {
    width: 320,
    height: 720,
    deviceScaleFactor: 1,
    mobile: false,
  });
  for (let i = 0; i < 100; i++) {
    if (await evaluate('Boolean(document.querySelector("#query"))')) break;
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  await evaluate(
    `(() => { const input = document.querySelector('#query'); input.value = 'touch'; input.dispatchEvent(new Event('input', {bubbles:true})); document.querySelector('form').requestSubmit(); })()`,
  );
  for (let i = 0; i < 100; i++) {
    if (await evaluate('Boolean(document.querySelector("article section li"))'))
      break;
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  for (const theme of ["light", "dark"]) {
    await evaluate(`document.querySelector('[data-mode="${theme}"]').click()`);
    const result = await evaluate(`(() => {
      const article = document.querySelector('article');
      if (!article) return {missing:true};
      const main = document.querySelector('main'), section = article.querySelector('section');
      const nodes = [...section.querySelectorAll('h3,li')];
      const bounds = section.getBoundingClientRect();
      return {theme:document.documentElement.dataset.theme, lists:section.querySelectorAll('ol').length,
        items:section.querySelectorAll('li').length,
        overflow:document.documentElement.scrollWidth > innerWidth || main.scrollWidth > main.clientWidth || article.scrollWidth > article.clientWidth,
        clipped:nodes.some(n => { const r=n.getBoundingClientRect(); return r.left < bounds.left - 1 || r.right > bounds.right + 1 || n.scrollWidth > n.clientWidth + 1; }),
        overlap:nodes.some((n,i) => i && n.getBoundingClientRect().top < nodes[i-1].getBoundingClientRect().bottom - 1)};
    })()`);
    assert.deepEqual(
      result,
      {
        theme,
        lists: 1,
        items: 6,
        overflow: false,
        clipped: false,
        overlap: false,
      },
      JSON.stringify(result),
    );
    await evaluate(`document.querySelector('#query').focus()`);
    let reached = false;
    for (let i = 0; i < 12; i++) {
      await command("Input.dispatchKeyEvent", {
        type: "keyDown",
        key: "Tab",
        code: "Tab",
        windowsVirtualKeyCode: 9,
      });
      await command("Input.dispatchKeyEvent", {
        type: "keyUp",
        key: "Tab",
        code: "Tab",
        windowsVirtualKeyCode: 9,
      });
      if (
        await evaluate(
          'document.activeElement.classList.contains("play-button")',
        )
      ) {
        reached = true;
        break;
      }
    }
    assert.ok(reached, "Playback button must be reachable by Tab");
    console.log(`PASS compact senses at 320px ${theme}: layout and keyboard`);
  }
} finally {
  ws?.close();
  chrome.kill();
  server.close();
}
