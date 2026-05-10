#!/usr/bin/env node
// Node.js ESM Codex browser-login flow.
// Uses camofox REST API (redf0x1/camofox-browser) at localhost:9377 — zero npm deps.
import { randomBytes, createHash, createHmac } from 'node:crypto';
import { writeFileSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';
import { parseArgs } from 'node:util';
import { execSync, spawnSync } from 'node:child_process';

// ── OAuth constants ───────────────────────────────────────────────────────────

const AUTH_BASE    = 'https://auth.openai.com';
const CLIENT_ID    = 'app_EMoamEEZ73f0CkXaXp7hrann';
const REDIRECT_URI = 'http://localhost:1455/auth/callback';
const SCOPE        = 'openid email profile offline_access';

// ── crypto helpers ────────────────────────────────────────────────────────────

function generatePKCE() {
  const verifier  = randomBytes(43).toString('base64url');
  const challenge = createHash('sha256').update(verifier).digest('base64url');
  return { verifier, challenge };
}

// RFC 6238 TOTP (SHA-1, 30s window) — ported from test-camofox-oauth-flow.mjs
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
  const key     = Buffer.from(bytes);
  const counter = BigInt(Math.floor(Date.now() / 1000 / 30));
  const buf     = Buffer.alloc(8);
  buf.writeBigUInt64BE(counter);
  const hmac = createHmac('sha1', key).update(buf).digest();
  const off  = hmac[hmac.length - 1] & 0x0F;
  const code = (
    ((hmac[off]     & 0x7F) << 24) | ((hmac[off + 1] & 0xFF) << 16) |
    ((hmac[off + 2] & 0xFF) <<  8) |  (hmac[off + 3] & 0xFF)
  ) % 1_000_000;
  return code.toString().padStart(6, '0');
}

// ── camofox lifecycle ─────────────────────────────────────────────────────────

async function checkHealth(camoUrl) {
  try { const r = await fetch(`${camoUrl}/health`); return r.ok; }
  catch { return false; }
}

async function ensureCamofox(camoUrl) {
  if (await checkHealth(camoUrl)) return;

  const which = spawnSync('which', ['camofox'], { encoding: 'utf8' });
  if (which.status !== 0) {
    console.log('[setup] camofox not found, installing camofox-browser...');
    execSync('npm install -g camofox-browser', { stdio: 'inherit' });
  }

  console.log('[setup] starting camofox server...');
  try { execSync('camofox server start', { stdio: 'pipe' }); }
  catch { /* may already be starting */ }

  const deadline = Date.now() + 15_000;
  while (Date.now() < deadline) {
    if (await checkHealth(camoUrl)) { console.log('[setup] camofox ready'); return; }
    await new Promise(r => setTimeout(r, 1000));
  }
  throw new Error(`camofox server did not start within 15s at ${camoUrl}\nTry manually: camofox server start`);
}

// ── camofox REST helpers ──────────────────────────────────────────────────────

async function api(camo, method, path, body) {
  const r = await fetch(`${camo}${path}`, {
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

async function getTabUrl(camo, tabId, userId) {
  const r   = await api(camo, 'GET', `/tabs?userId=${encodeURIComponent(userId)}`);
  const tab = (r.tabs || []).find(t => t.id === tabId || t.tabId === tabId);
  if (!tab) throw new Error(`tab ${tabId} is missing (browser closed or crashed)`);
  return tab?.url || '';
}

async function waitUrl(camo, tabId, userId, predicate, timeout = 60_000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    const url = await getTabUrl(camo, tabId, userId);
    if (await predicate(url)) return url;
    await new Promise(r => setTimeout(r, 1500));
  }
  throw new Error(`waitUrl timeout after ${timeout}ms`);
}

async function waitNet(camo, tabId, userId, timeout = 6000) {
  return api(camo, 'POST', `/tabs/${tabId}/wait`, { userId, timeout, waitForNetwork: true });
}

function isTotpChallengeUrl(url) {
  let route = url.split(/[?#]/, 1)[0];
  try {
    route = new URL(url).pathname;
  } catch {
    // Fall back to the query-free string for non-absolute tab URLs.
  }
  const normalizedRoute = route.toLowerCase();
  return normalizedRoute.includes('mfa') || normalizedRoute.includes('totp') || normalizedRoute.includes('2fa');
}

async function typeInto(camo, tabId, userId, selector, text) {
  return api(camo, 'POST', `/tabs/${tabId}/type`, { userId, selector, text });
}

async function clickEl(camo, tabId, userId, selector) {
  return api(camo, 'POST', `/tabs/${tabId}/click`, { userId, selector });
}

async function snap(camo, tabId, userId) {
  return api(camo, 'GET', `/tabs/${tabId}/snapshot?userId=${encodeURIComponent(userId)}`);
}

function parseProxy(proxyUrl) {
  if (!proxyUrl) return null;
  let parsed;
  try {
    parsed = new URL(proxyUrl);
  } catch {
    throw new Error('--proxy must be a URL, e.g. http://user:pass@host:port');
  }
  const port = Number(parsed.port);
  if (!parsed.hostname || !Number.isInteger(port) || port <= 0) {
    throw new Error('--proxy must include host and port');
  }
  return {
    host: parsed.hostname,
    port,
    ...(parsed.username ? { username: decodeURIComponent(parsed.username) } : {}),
    ...(parsed.password ? { password: decodeURIComponent(parsed.password) } : {}),
  };
}

// ── mail OTP helpers ──────────────────────────────────────────────────────────

async function fetchOTP(mailUrl, email, freshSince, mailToken) {
  const [user, domain] = email.split('@');
  const url     = `${mailUrl}/api/v1/email/${domain}/${user}/otp`;
  const headers = { 'User-Agent': 'Mozilla/5.0 (compatible; codex-login/2.0)' };
  if (mailToken) headers['Authorization'] = `Bearer ${mailToken}`;
  const r = await fetch(url, { headers }).catch(() => null);
  if (!r || !r.ok) return null;
  const j = await r.json().catch(() => null);
  if (!j?.data?.otp) return null;
  // Guard against stale OTPs delivered before we submitted the form
  if (freshSince && new Date(j.data.receivedAt) < freshSince) return null;
  return j.data.otp;
}

async function waitForOTP(mailUrl, email, mailToken, timeout = 300, freshSince) {
  console.log(`[otp] waiting for OTP (up to ${timeout}s)...`);
  const deadline = Date.now() + timeout * 1000;
  let attempt = 0;
  while (Date.now() < deadline) {
    const otp = await fetchOTP(mailUrl, email, freshSince, mailToken);
    if (otp) { console.log('[otp] received'); return otp; }
    if (++attempt % 10 === 0) {
      const left = Math.ceil((deadline - Date.now()) / 1000);
      console.log(`[otp] still waiting... ${left}s left`);
    }
    await new Promise(r => setTimeout(r, 4000));
  }
  return null;
}

// ── ViOTP helpers ─────────────────────────────────────────────────────────────

async function rentPhone(token, serviceId) {
  const r = await fetch(
    `https://api.viotp.com/request/getv2?token=${token}&serviceId=${serviceId}`
  );
  const j = await r.json();
  if (!j?.data?.request_id) throw new Error(`ViOTP rent failed: ${JSON.stringify(j)}`);
  return j.data; // { phone_number, re_phone_number, request_id, countryCode, ... }
}

async function waitForSmsOTP(token, requestId, timeout = 120) {
  console.log('[phone] waiting for SMS OTP...');
  const deadline = Date.now() + timeout * 1000;
  while (Date.now() < deadline) {
    const r = await fetch(`https://api.viotp.com/session/getv2?token=${token}&requestId=${requestId}`);
    const j = await r.json();
    if (j?.data?.Status === 1) return j.data.Code;  // received
    if (j?.data?.Status === 2) throw new Error('ViOTP session expired');
    await new Promise(r => setTimeout(r, 5000));
  }
  throw new Error('ViOTP SMS OTP timeout');
}

// ── phone verification flow ───────────────────────────────────────────────────

async function handlePhoneVerification(camo, tabId, userId, viotpToken, viotpServiceId, maxRetries = 5) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    const phone = await rentPhone(viotpToken, viotpServiceId);
    console.log(`[phone] rented number ending ${String(phone.phone_number).slice(-4)} (attempt ${attempt + 1})`);

    // Select Vietnam +84
    console.log('[phone] selecting country (Vietnam)...');
    try {
      await clickEl(camo, tabId, userId, 'button[aria-label*="country"], button[aria-haspopup="listbox"]');
      const selResult = await api(camo, 'POST', `/tabs/${tabId}/evaluate`, {
        userId, expression: `
          (async function() {
            for (let i = 0; i < 30; i++) {
              const opts = Array.from(document.querySelectorAll('[role="option"], li'));
              const vn = opts.find(o => o.textContent.includes('Vietnam'));
              if (vn) {
                vn.click();
                return "Vietnam selected";
              }
              const lb = document.querySelector('[role="listbox"], .listbox');
              if (lb) lb.scrollTop += 800;
              await new Promise(r => setTimeout(r, 200));
            }
            return "Vietnam not found";
          })()
        `
      });
      console.log(`[phone] country selection: ${JSON.stringify(selResult)}`);
    } catch (selErr) {
      console.log(`[phone] country selection error: ${selErr.message}`);
    }

    await new Promise(r => setTimeout(r, 1500));
    let cleanPhone = String(phone.phone_number);
    if (cleanPhone.startsWith('0')) cleanPhone = cleanPhone.slice(1);
    if (cleanPhone.startsWith('84')) cleanPhone = cleanPhone.slice(2);
    
    console.log(`[phone] typing clean number: ${cleanPhone}`);
    await typeInto(camo, tabId, userId, 'input[type="tel"]:not([type="hidden"]), input[name*="phone"]:not([type="hidden"])', cleanPhone);
    await clickEl(camo, tabId, userId, 'button[type="submit"], button:has-text("Continue")');
    await new Promise(r => setTimeout(r, 3000));

    // Check if rejected (snapshot contains rejection keywords)
    const s = await snap(camo, tabId, userId);
    const snapText = JSON.stringify(s).toLowerCase();
    if (snapText.includes('not valid') || snapText.includes('invalid') || snapText.includes('not supported') || snapText.includes('something went wrong')) {
      console.log(`[phone] number rejected (${phone.phone_number}), retrying...`);
      continue;
    }

    // Wait for OTP input to appear; if OpenAI rejects late, catch and retry.
    try {
      await waitUrl(
        camo, tabId, userId,
        async () => JSON.stringify(await snap(camo, tabId, userId)).includes('one-time-code'),
        35_000,
      );
    } catch (waitErr) {
      console.log(`[phone] number likely rejected (late): ${waitErr.message}`);
      if (attempt + 1 < maxRetries) continue;
      throw new Error(`Phone verification failed after ${maxRetries} attempts (last: ${waitErr.message})`);
    }

    let smsCode;
    try {
      smsCode = await waitForSmsOTP(viotpToken, phone.request_id);
    } catch (smsErr) {
      throw new Error(`Phone verification failed after number was accepted: ${smsErr.message}`);
    }
    console.log('[phone] SMS OTP received');
    await typeInto(camo, tabId, userId, 'input[autocomplete="one-time-code"]', smsCode);
    await clickEl(camo, tabId, userId, 'button[type="submit"]');
    return;
  }
  throw new Error(`Phone verification failed after ${maxRetries} attempts`);
}

// ── main browser flow ─────────────────────────────────────────────────────────

async function doBrowserLogin(opts) {
  const { camo, userId, email, authorizeUrl, password, totpSecret, proxy,
          mailUrl, mailToken, viotpToken, viotpServiceId } = opts;

  console.log('[browser] navigating to authorize URL...');
  const opened = await api(camo, 'POST', '/tabs', {
    userId,
    sessionKey: userId,
    url: authorizeUrl,
    ...(proxy ? { proxy: parseProxy(proxy), geoMode: 'proxy-locked' } : {}),
  });
  const tabId  = opened.tabId || opened.id;
  if (!tabId) throw new Error(`no tabId in response: ${JSON.stringify(opened)}`);
  console.log(`[browser] tabId=${tabId}`);

  try {
    const deadline = Date.now() + 420_000;
    const freshSince = new Date(Date.now() - 60_000);
    console.log('[browser] starting interactive flow loop...');

    while (Date.now() < deadline) {
      let url = await getTabUrl(camo, tabId, userId);
      console.log(`[browser] loop: url=${url}`);
      if (url.includes('localhost:1455') && (url.includes('code=') || url.includes('error='))) {
        console.log(`[browser] callback reached: ${url}`);
        const code = new URL(url).searchParams.get('code');
        if (!code) throw new Error(`OAuth error in callback: ${url}`);
        return code;
      }

      const s = await snap(camo, tabId, userId);
      const t = JSON.stringify(s).toLowerCase();

      // 1. Email input
      if ((t.includes('email address') || t.includes('name="email"')) && !t.includes('password')) {
        console.log('[browser] typing email...');
        await typeInto(camo, tabId, userId, 'input[name="email"], input[type="email"], input[type="text"]', email);
        await clickEl(camo, tabId, userId, 'button[type="submit"] >> nth=0');
        await new Promise(r => setTimeout(r, 4000));
        continue;
      }

      // 2. Signup redirection
      if ((url.includes('/log-in') || url.includes('/login')) && (t.includes('incorrect email address') || t.includes('don\'t have an account') || t.includes('sign up'))) {
        console.log('[browser] account not found or login rejected, attempting signup redirection...');
        await clickEl(camo, tabId, userId, 'a[href*="signup"], a[href*="create-account"], button:has-text("Sign up") >> nth=0');
        await new Promise(r => setTimeout(r, 5000));
        continue;
      }

      // 3. Password input (Login or Signup)
      if (t.includes('password')) {
        console.log('[browser] typing password...');
        await typeInto(camo, tabId, userId, 'input[name="password"], input[type="password"]', password);
        await clickEl(camo, tabId, userId, 'button[type="submit"] >> nth=0');
        await new Promise(r => setTimeout(r, 6000));
        continue;
      }

      // 4. Email OTP / Verification
      if (t.includes('verify your email') || t.includes('check your inbox') || url.includes('email-verification')) {
        console.log('[browser] email verification detected');
        const otpBox = 'input[name="code"], input[autocomplete="one-time-code"], input[type="text"]';
        // check if box exists in snap
        if (t.includes('autocomplete="one-time-code"') || t.includes('name="code"')) {
           const otp = await waitForOTP(mailUrl, email, mailToken, 300, freshSince);
           await typeInto(camo, tabId, userId, otpBox, otp);
           await clickEl(camo, tabId, userId, 'button[name="intent"][value="validate"] >> nth=0');
           await new Promise(r => setTimeout(r, 6000));
           continue;
        } else {
           console.log('[browser] verification link? waiting for background activation...');
           await new Promise(r => setTimeout(r, 8000));
           continue;
        }
      }

      // 5. Onboarding (Name/Birthday)
      if (url.includes('onboarding') || t.includes('name="firstName"') || t.includes('birthday')) {
        console.log('[browser] onboarding step');
        await typeInto(camo, tabId, userId, 'input[name="firstName"]', 'Ethan');
        await typeInto(camo, tabId, userId, 'input[name="lastName"]', 'Martinez');
        await typeInto(camo, tabId, userId, 'input[name="birthday"]', '12/12/1995');
        await clickEl(camo, tabId, userId, 'button[type="submit"] >> nth=0');
        await new Promise(r => setTimeout(r, 6000));
        continue;
      }

      // 6. Phone verification
      if (url.includes('phone') || t.includes('verify your phone') || t.includes('phone-number')) {
        console.log('[browser] phone verification step');
        await handlePhoneVerification(camo, tabId, userId, viotpToken, viotpServiceId);
        await new Promise(r => setTimeout(r, 5000));
        continue;
      }

      // 7. TOTP
      if (isTotpChallengeUrl(url) && totpSecret) {
        console.log('[browser] TOTP step');
        const totp = generateTOTP(totpSecret);
        await typeInto(camo, tabId, userId, 'input[autocomplete="one-time-code"], input[name="code"]', totp);
        await clickEl(camo, tabId, userId, 'button[name="intent"][value="validate"], button[type="submit"]');
        await new Promise(r => setTimeout(r, 5000));
        continue;
      }

      // 8. General Continue/Authorize
      if (t.includes('button:has-text("continue")') || t.includes('button:has-text("agree")') || url.includes('consent')) {
        console.log('[browser] clicking continue/authorize...');
        await clickEl(camo, tabId, userId, 'button:has-text("Agree"), button:has-text("Authorize"), button:has-text("Continue"), button[type="submit"] >> nth=0');
        await new Promise(r => setTimeout(r, 5000));
        continue;
      }

      // If no match, wait
      process.stdout.write('.');
      await new Promise(r => setTimeout(r, 5000));
    }
    throw new Error('Timeout waiting for OAuth callback');

  } finally {
    await api(camo, 'DELETE', `/tabs/${tabId}`, { userId }).catch(() => {});
  }
}

// ── token exchange + panel JSON ───────────────────────────────────────────────

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
    throw new Error(`status=${r.status} error=${body.error || 'unknown'}`);
  }
  return body;
}

function parseIdToken(idToken) {
  if (!idToken) return {};
  const parts = idToken.split('.');
  if (parts.length !== 3) return {};
  try {
    return JSON.parse(Buffer.from(parts[1], 'base64url').toString());
  } catch { return {}; }
}

function buildPanelEntry(email, tokens) {
  const now       = new Date();
  const expiresIn = tokens.expires_in || 3600;
  const claims    = parseIdToken(tokens.id_token || '');
  const auth      = claims['https://api.openai.com/auth'] || {};
  const accountId = auth.chatgpt_account_id || '';
  const userId    = auth.chatgpt_user_id || '';
  const planType  = auth.chatgpt_plan_type || 'free';
  const orgs      = auth.organizations || [];
  const orgId     = orgs[0]?.id || '';

  return {
    _token_version:    Date.now(),
    access_token:      tokens.access_token || '',
    account_id:        accountId,
    chatgpt_account_id: accountId,
    chatgpt_user_id:   userId,
    disabled:          false,
    email,
    expired:           new Date(now.getTime() + 7 * 86400_000).toISOString(),
    expires_at:        new Date(now.getTime() + expiresIn * 1000).toISOString(),
    id_token:          tokens.id_token || '',
    last_refresh:      now.toISOString(),
    model_mapping:     {},
    organization_id:   orgId,
    plan_type:         planType,
    refresh_token:     tokens.refresh_token || '',
    type:              'codex',
  };
}

function panelFilename(email, planType) {
  const digest    = createHash('sha256').update(`${email}\x00${planType}`).digest('hex').slice(0, 12);
  const safeEmail = email.replace(/[^a-zA-Z0-9@._\-]/g, '_').slice(0, 120);
  const safePlan  = planType.replace(/[^a-zA-Z0-9@._\-]/g, '_').slice(0, 120);
  return `codex-${safeEmail}-${safePlan}-${digest}.json`;
}

// ── CLI entrypoint ────────────────────────────────────────────────────────────

async function main() {
  const { values } = parseArgs({
    options: {
      email:             { type: 'string' },
      out:               { type: 'string' },
      'mail-url':        { type: 'string',  default: 'https://mail.monet.uno' },
      'mail-token':      { type: 'string',  default: '' },
      proxy:             { type: 'string',  default: '' },
      password:          { type: 'string',  default: '' },
      'totp-secret':     { type: 'string',  default: '' },
      'camofox-url':     { type: 'string',  default: 'http://localhost:9377' },
      'viotp-token':     { type: 'string',  default: process.env.VIOTP_TOKEN ?? '' },
      'viotp-service-id':{ type: 'string',  default: process.env.VIOTP_SERVICE_ID ?? '0' },
    },
  });

  if (!values.email || !values.out) {
    console.error('Usage: codex-browser-login.mjs --email <email> --out <dir> [options]');
    process.exit(1);
  }

  const camo = values['camofox-url'];
  await ensureCamofox(camo);

  const email  = values.email;
  const userId = email.replace(/[@.]/g, '-');
  console.log(`[codex-login] email=${email}`);

  const { verifier, challenge } = generatePKCE();
  const state = randomBytes(16).toString('hex');
  const authorizeUrl = `${AUTH_BASE}/oauth/authorize?` + new URLSearchParams({
    response_type: 'code', client_id: CLIENT_ID,
    redirect_uri: REDIRECT_URI, scope: SCOPE, state,
    code_challenge: challenge, code_challenge_method: 'S256',
    id_token_add_organizations: 'true',
    codex_cli_simplified_flow: 'true', prompt: 'login',
  });

  let code;
  try {
    code = await doBrowserLogin({
      camo, userId, email, authorizeUrl,
      password:       values.password,
      totpSecret:     values['totp-secret'],
      proxy:          values.proxy,
      mailUrl:        values['mail-url'],
      mailToken:      values['mail-token'],
      viotpToken:     values['viotp-token'],
      viotpServiceId: values['viotp-service-id'],
    });
  } catch (e) {
    console.error(`[codex-login] FAILED (browser): ${e.message}`);
    process.exit(1);
  }

  console.log('[codex-login] exchanging code for tokens...');
  let tokens;
  try {
    tokens = await exchangeCode(code, verifier);
  } catch (e) {
    console.error(`[codex-login] FAILED (token exchange): ${e.message}`);
    process.exit(1);
  }

  if (!tokens.access_token) {
    console.error(`[codex-login] FAILED (token exchange): missing access_token`);
    process.exit(1);
  }

  const entry   = buildPanelEntry(email, tokens);
  const fname   = panelFilename(email, entry.plan_type);
  const outDir  = values.out;
  mkdirSync(outDir, { recursive: true });
  const outPath = join(outDir, fname);
  writeFileSync(outPath, JSON.stringify(entry, null, 2));
  console.log(`[codex-login] OK → ${outPath}`);
}

await main().catch(e => { console.error(`[codex-login] ERROR: ${e.message}`); process.exit(1); });
