// Real-browser layout smoke test. Build first, then set CHROME_PATH if Chrome is not on PATH.
// Usage: CHROME_PATH=/path/to/chrome node scripts/footer-browser-check.mjs
import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { createServer } from 'node:http';
import { mkdtemp, readFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve, extname } from 'node:path';

const root = resolve('build');
const server = createServer(async (request, response) => {
  try {
    const name = request.url === '/' ? '/index.html' : new URL(request.url, 'http://localhost').pathname;
    const file = resolve(root, '.' + name);
    if (!file.startsWith(root + '/')) throw new Error('Invalid path');
    const mime = { '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css' };
    const content = await readFile(file);
    response.writeHead(200, { 'Content-Type': mime[extname(file)] || 'application/octet-stream' });
    response.end(content);
  } catch {
    response.writeHead(404).end();
  }
});
await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
const port = server.address().port;
const profile = await mkdtemp(join(tmpdir(), 'wdict-browser-check-'));
const chrome = spawn(process.env.CHROME_PATH || 'google-chrome', [
  '--headless=new', '--no-sandbox', '--disable-gpu', '--remote-debugging-port=0',
  '--user-data-dir=' + profile, 'about:blank'
], { stdio: 'ignore' });
let ws;
try {
  let debugPort;
  for (let i = 0; i < 100; i++) {
    try { debugPort = (await readFile(join(profile, 'DevToolsActivePort'), 'utf8')).split('\n')[0]; break; }
    catch { await new Promise(resolve => setTimeout(resolve, 100)); }
  }
  assert.ok(debugPort, 'Chrome did not start');
  const target = await fetch('http://127.0.0.1:' + debugPort + '/json/new?' + encodeURIComponent('http://127.0.0.1:' + port), { method: 'PUT' }).then(r => r.json());
  ws = new WebSocket(target.webSocketDebuggerUrl);
  await new Promise((resolve, reject) => { ws.onopen = resolve; ws.onerror = reject; });
  let id = 0;
  const pending = new Map();
  ws.onmessage = event => {
    const data = JSON.parse(event.data);
    if (!data.id) return;
    const item = pending.get(data.id);
    pending.delete(data.id);
    if (data.error) item.reject(new Error(data.error.message)); else item.resolve(data.result);
  };
  function command(method, params = {}) {
    return new Promise((resolve, reject) => {
      const n = ++id;
      pending.set(n, { resolve, reject });
      ws.send(JSON.stringify({ id: n, method, params }));
    });
  }
  async function evaluate(expression) {
    const result = await command('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true });
    if (result.exceptionDetails) throw new Error(result.exceptionDetails.text);
    return result.result.value;
  }
  await command('Page.enable');
  for (let i = 0; i < 100; i++) {
    if (await evaluate('Boolean(document.querySelector("footer select"))')) break;
    await new Promise(resolve => setTimeout(resolve, 100));
  }
  assert.equal(await evaluate('Boolean(document.querySelector("footer select"))'), true, 'App failed to render');
  for (const [width, height] of [[1024, 768], [360, 620]]) {
    await command('Emulation.setDeviceMetricsOverride', { width, height, deviceScaleFactor: 1, mobile: false });
    for (const screen of ['lookup', 'dictionaries']) {
      if (screen === 'dictionaries') await evaluate('document.querySelector("header button").click()');
      const result = await evaluate(`(() => {
        const main = document.querySelector('main'), footer = document.querySelector('footer');
        const marker = document.createElement('button');
        marker.textContent = 'End of long content';
        marker.style.marginTop = '1200px';
        main.append(marker);
        main.scrollTop = main.scrollHeight;
        const f = footer.getBoundingClientRect(), m = marker.getBoundingClientRect();
        const out = {scrollable: main.scrollHeight > main.clientHeight, scrollTop: main.scrollTop,
          footerBottom: f.bottom, viewportBottom: innerHeight, footerTop: f.top,
          markerBottom: m.bottom, markerTop: m.top, mainTop: main.getBoundingClientRect().top,
          horizontalOverflow: document.documentElement.scrollWidth > innerWidth,
          footerVisible: f.top >= 0 && f.bottom <= innerHeight};
        marker.focus();
        const focused = document.activeElement === marker && m.top >= out.mainTop && m.bottom <= out.footerTop;
        marker.remove();
        return {...out, focused};
      })()`);
      assert.ok(result.scrollable && result.scrollTop > 0, JSON.stringify({ width, screen, result }));
      assert.ok(Math.abs(result.footerBottom - result.viewportBottom) < 1 && result.footerVisible, JSON.stringify({ width, screen, result }));
      assert.ok(result.markerBottom <= result.footerTop && result.markerTop >= result.mainTop && result.focused, JSON.stringify({ width, screen, result }));
      assert.equal(result.horizontalOverflow, false, JSON.stringify({ width, screen, result }));
      const select = await evaluate(`(() => { const s=document.querySelector('footer select'); s.focus(); return document.activeElement === s; })()`);
      assert.equal(select, true, 'Language selector is focusable');
      await command('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Tab', code: 'Tab', windowsVirtualKeyCode: 9 });
      await command('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Tab', code: 'Tab', windowsVirtualKeyCode: 9 });
      assert.equal(await evaluate('document.activeElement.getAttribute("role")'), 'radio', 'Tab reaches appearance control');
      if (screen === 'dictionaries') await evaluate('document.querySelector("header button").click()');
      console.log(`PASS ${screen} ${width}x${height}: footer, scrolling and focus`);
    }
  }
} finally {
  ws?.close();
  chrome.kill();
  server.close();
  // Temporary Chrome profile is left under the system temp directory for cleanup by the OS.
}
