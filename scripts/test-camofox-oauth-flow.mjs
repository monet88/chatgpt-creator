#!/usr/bin/env node
/**
 * Test: can camofox REST API handle the Codex OAuth flow?
 * Steps: authorize URL → email submit → OTP/TOTP/password → consent → callback
 * Starts local HTTP server on :1455 to capture OAuth code.
 */
import { createServer } from 'node:http';
import { createHmac, randomBytes, createHash } from 'node:crypto';
import { writeFileSync } from 'node:fs';

const AUTH_BASE = 'https://auth.openai.com';
const CLIENT_ID = 'app_EMoamEEZ73f0CkXaXp7hrann';
const REDIRECT_URI = 'http://localhost:1455/auth/callback';
const SCOPE = 'openid email profile offline_access';
const CAMO = 'http://localhost:9377';
const USER = 'codex-test';

// ── helpers ──────────────────────────────────────────────────────────────────

function generatePKCE() {
  const verifier = randomBytes(43).toString('base64url');
  const challenge = createHash('sha256').update(verifier).digest('base64url');
  return { verifier, challenge };
}

function generateTOTP(secret) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  const clean = secret.toUpperCase().replace(/[=\s]/g, '');
  const bytes = [];
  let bits = 0, value = 0;
  for (const c of clean) {
    const idx = chars.indexOf(c);
    if (idx === -1) continue;
    value = (value << 5) | idx;
    bits += 5;
    if (bits >= 8) { bits -= 8; bytes.push((value >> bits) & 0xFF); }
  }
  const key = Buffer.from(bytes);
  const counter = BigInt(Math.floor(Date.now() / 1000 / 30));
  const buf = Buffer.alloc(8);
  buf.writeBigUInt64BE(counter);
  const hmac = createHmac('sha1', key).update(buf).digest();
  const off = hmac[hmac.length - 1] & 0x0F;
  const code = (
    ((hmac[off] & 0x7F) << 24) | ((hmac[off + 1] & 0xFF) << 16) |
    ((hmac[off + 2] & 0xFF) << 8) | (hmac[off + 3] & 0xFF)
  ) % 1_000_000;
  return code.toString().padStart(6, '0');
}

async function api(method, path, body) {
  const r = await fetch(`${CAMO}${path}`, {
    method,
    headers: { 'Content-Type': 'application/json' },
    ...(body ? { body: JSON.stringify(body) } : {}),
  });
  const text = await r.text();
  let payload;
  try { payload = JSON.parse(text); } catch { payload = { raw: text }; }
  if (!r.ok) {
    const message = typeof payload?.error === 'string' ? payload.error : text.slice(0, 300);
    throw new Error(`${method} ${path} failed (${r.status}): ${message}`);
  }
  return payload;
}

async function shot(tabId, label) {
  const r = await api('GET', `/tabs/${tabId}/screenshot?userId=${encodeURIComponent(USER)}`);
  if (r.data) {
    const p = `/tmp/camo-test-${label}.png`;
    writeFileSync(p, Buffer.from(r.data, 'base64'));
    console.log(`  [shot] ${p}`);
  }
}

async function snap(tabId) {
  return api('GET', `/tabs/${tabId}/snapshot?userId=${encodeURIComponent(USER)}`);
}

async function waitNet(tabId, timeout = 5000) {
  return api('POST', `/tabs/${tabId}/wait`, { userId: USER, timeout, waitForNetwork: true });
}

// Poll tab URL via tabs list (no direct /tabs/:id/url endpoint)
async function getTabUrl(tabId) {
  const r = await api('GET', `/tabs?userId=${encodeURIComponent(USER)}`);
  const tabs = r.tabs || [];
  const tab = tabs.find(t => t.id === tabId || t.tabId === tabId);
  return tab?.url || '';
}

async function waitUrl(tabId, predicate, timeout = 60000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    const url = await getTabUrl(tabId);
    if (predicate(url)) return url;
    await new Promise(r => setTimeout(r, 1500));
  }
  throw new Error('waitUrl timeout');
}

// ── OAuth callback server ────────────────────────────────────────────────────

function startCallbackServer(timeout = 120_000) {
  return new Promise((resolve, reject) => {
    const server = createServer((req, res) => {
      try {
        const u = new URL(req.url, 'http://localhost:1455');
        if (u.pathname === '/auth/callback') {
          const code = u.searchParams.get('code');
          res.writeHead(200, { 'Content-Type': 'text/html' });
          res.end('<h1>OK — code received</h1>');
          if (code) finish(code);
          else fail(new Error('Callback missing code param'));
        } else {
          res.writeHead(404); res.end();
        }
      } catch { res.writeHead(500); res.end(); }
    });
    let settled = false;
    let timeoutId;
    const close = done => {
      if (server.listening) server.close(done);
      else done();
    };
    const finish = code => {
      if (settled) return;
      settled = true;
      clearTimeout(timeoutId);
      close(() => resolve(code));
    };
    const fail = err => {
      if (settled) return;
      settled = true;
      clearTimeout(timeoutId);
      close(() => reject(err));
    };
    server.on('error', fail);
    server.on('listening', () => console.log('[callback] :1455 ready'));
    server.listen(1455, '127.0.0.1');
    timeoutId = setTimeout(() => fail(new Error('OAuth callback timeout')), timeout);
  });
}

// ── token exchange ────────────────────────────────────────────────────────────

async function exchangeCode(code, verifier) {
  const r = await fetch(`${AUTH_BASE}/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: CLIENT_ID,
      code,
      redirect_uri: REDIRECT_URI,
      code_verifier: verifier,
    }).toString(),
  });
  const body = await r.json().catch(() => ({}));
  if (!r.ok) {
    throw new Error(`token exchange failed with status ${r.status}`);
  }
  return body;
}

// ── main ──────────────────────────────────────────────────────────────────────

async function main() {
  const EMAIL = process.env.TEST_EMAIL;
  const PASSWORD = process.env.TEST_PASSWORD;
  const TOTP_SECRET = process.env.TEST_TOTP_SECRET;

  if (!EMAIL || !PASSWORD || !TOTP_SECRET) {
    console.error('Set TEST_EMAIL, TEST_PASSWORD, TEST_TOTP_SECRET env vars');
    process.exit(1);
  }

  // Verify server
  const health = await api('GET', '/health').catch(() => null);
  if (!health) { console.error('[error] camofox server not reachable at :9377'); process.exit(1); }
  console.log('[server] OK port=9377');

  // Start local callback server to capture OAuth code (primary method).
  const codePromise = startCallbackServer(120_000);

  // Build PKCE + auth URL
  const { verifier, challenge } = generatePKCE();
  const state = randomBytes(16).toString('hex');
  const authUrl = `${AUTH_BASE}/oauth/authorize?` + new URLSearchParams({
    response_type: 'code', client_id: CLIENT_ID,
    redirect_uri: REDIRECT_URI, scope: SCOPE, state,
    code_challenge: challenge, code_challenge_method: 'S256',
    id_token_add_organizations: 'true',
    codex_cli_simplified_flow: 'true', prompt: 'login',
  });

  console.log('[flow] opening auth URL...');
  const opened = await api('POST', '/tabs', { userId: USER, sessionKey: 'default', url: authUrl });
  const tabId = opened.tabId || opened.id;
  if (!tabId) { console.error('[error] no tabId:', JSON.stringify(opened)); process.exit(1); }
  console.log(`[flow] tabId=${tabId}`);

  await waitNet(tabId, 8000);
  await shot(tabId, '01-initial');
  const s1 = await snap(tabId);
  console.log('[page] url:', await getTabUrl(tabId));
  console.log('[page] snapshot (300c):', JSON.stringify(s1).slice(0, 300));

  // Fill email
  console.log('[flow] filling email...');
  await api('POST', `/tabs/${tabId}/type`, { userId: USER, selector: 'input[name="email"]', text: EMAIL });
  await shot(tabId, '02-email-filled');
  await api('POST', `/tabs/${tabId}/click`, { userId: USER, selector: 'button[type="submit"]' });
  console.log('[flow] email submitted, waiting for next step...');
  await new Promise(r => setTimeout(r, 4000));
  await waitNet(tabId, 6000);
  await shot(tabId, '03-after-email-submit');
  const url3 = await getTabUrl(tabId);
  const s3 = await snap(tabId);
  console.log('[page] url:', url3);
  console.log('[page] snapshot (500c):', JSON.stringify(s3).slice(0, 500));

  // Detect step via URL (most reliable) + snapshot fallback
  const snap3Text = JSON.stringify(s3).toLowerCase();
  const hasPasswordInput = url3.includes('/password') || snap3Text.includes('textbox "password"') || snap3Text.includes('"enter your password"');
  const hasOtpInput = url3.includes('email-verification') || url3.includes('/otp') || snap3Text.includes('one-time') || snap3Text.includes('verification code');
  const hasTotpInput = url3.includes('totp') || url3.includes('2fa') || snap3Text.includes('authenticator') || snap3Text.includes('two-factor');

  console.log(`[detect] url=${url3} otp=${hasOtpInput} password=${hasPasswordInput} totp=${hasTotpInput}`);

  if (hasPasswordInput) {
    console.log('[flow] password step detected, filling...');
    await api('POST', `/tabs/${tabId}/type`, { userId: USER, selector: 'input[type="password"]', text: PASSWORD });
    await shot(tabId, '04-password-filled');
    await api('POST', `/tabs/${tabId}/click`, { userId: USER, selector: 'button[type="submit"]' });
    await new Promise(r => setTimeout(r, 4000));
    await waitNet(tabId, 8000);
    await shot(tabId, '05-after-password');
    const url5 = await getTabUrl(tabId);
    const s5 = await snap(tabId);
    console.log('[page] url:', url5);
    console.log('[page] snapshot (500c):', JSON.stringify(s5).slice(0, 500));
  }

  // Re-check after password step
  const url5b = await getTabUrl(tabId);
  const s5b = await snap(tabId);
  const snap5Text = JSON.stringify(s5b).toLowerCase();
  const hasTotpNow = url5b.includes('totp') || url5b.includes('2fa') || url5b.includes('mfa') ||
    snap5Text.includes('authenticator') || snap5Text.includes('two-factor') || snap5Text.includes('6-digit');
  console.log(`[detect-post-pw] url=${url5b} totp=${hasTotpNow}`);
  if (hasTotpNow || hasTotpInput || snap3Text.includes('authenticator')) {
    const totp = generateTOTP(TOTP_SECRET);
    console.log('[flow] TOTP step detected');
    const totpSel = 'input[autocomplete="one-time-code"], input[name="code"], input[type="text"]';
    await api('POST', `/tabs/${tabId}/type`, { userId: USER, selector: totpSel, text: totp });
    await shot(tabId, '06-totp-filled');
    await api('POST', `/tabs/${tabId}/click`, { userId: USER, selector: 'button[type="submit"]' });
    await new Promise(r => setTimeout(r, 4000));
    await waitNet(tabId, 8000);
    await shot(tabId, '07-after-totp');
    console.log('[page] url:', await getTabUrl(tabId));
  }

  if (hasOtpInput) {
    console.log('[flow] email OTP step detected — cannot auto-fill (Gmail account)');
    console.log('[flow] check /tmp/camo-test-03-after-email-submit.png for page state');
    console.log('[info] this confirms camofox CAN navigate and fill forms up to OTP step');
    // Don't exit — let callback server time out to clean up
  }

  // Handle consent if present
  await new Promise(r => setTimeout(r, 2000));
  const urlC = await getTabUrl(tabId);
  if (urlC.includes('consent')) {
    console.log('[flow] consent page, clicking allow...');
    await api('POST', `/tabs/${tabId}/click`, {
      userId: USER,
      selector: 'button[type="submit"], button:has-text("Allow"), button:has-text("Continue")',
    });
    await new Promise(r => setTimeout(r, 4000));
    await shot(tabId, '08-after-consent');
  }

  // Wait for OAuth code — prefer callback server, fallback to URL polling/snapshot.
  console.log('[flow] waiting for OAuth code (callback server + URL polling)...');
  let code;

  // Race: callback server vs URL polling
  const urlPollPromise = (async () => {
    try {
      const callbackUrl = await waitUrl(
        tabId,
        url => url.includes('localhost:1455') && url.includes('code='),
        90_000,
      );
      return new URL(callbackUrl).searchParams.get('code');
    } catch {
      return null;
    }
  })();

  try {
    code = await Promise.any([
      codePromise.then(c => { console.log('[flow] code received via callback server'); return c; }),
      urlPollPromise.then(c => { if (!c) throw new Error('poll empty'); console.log('[flow] code captured via URL polling'); return c; }),
    ]);
  } catch {
    // Last resort: extract from snapshot text
    const s = await snap(tabId);
    await shot(tabId, '09-final-state');
    const snapStr = JSON.stringify(s);
    const m = snapStr.match(/code=([A-Za-z0-9_\-]+)/);
    if (m) {
      code = m[1];
      console.log('[flow] code extracted from snapshot text');
    } else {
      console.log('[flow] OAuth code not captured');
      console.log('[result] PARTIAL — camofox navigates/fills correctly, callback server did not receive code');
      process.exit(0);
    }
  }

  // Exchange code for tokens
  console.log('[flow] exchanging code for tokens...');
  const tokens = await exchangeCode(code, verifier);
  if (tokens.access_token) {
    console.log('[result] SUCCESS — full OAuth flow completed');
    console.log('[token] plan_type (from id_token claims) — checking...');
    // Decode id_token to get plan
    const parts = (tokens.id_token || '').split('.');
    if (parts.length === 3) {
      const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
      const auth = payload['https://api.openai.com/auth'] || {};
      console.log(`[token] email=${payload.email} plan=${auth.chatgpt_plan_type}`);
    }
    console.log('[result] camofox REST API CAN handle this flow');
  } else {
    console.log('[result] token exchange failed: missing access_token');
  }

  await api('DELETE', `/tabs/${tabId}`, { userId: USER }).catch(() => {});
}

main().catch(e => { console.error('[fatal]', e.message); process.exit(1); });
