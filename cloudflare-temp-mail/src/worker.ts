import { handleScheduled } from './handlers/cleanup-scheduler';
import { handleEmail, type InboundEmailMessage } from './handlers/email-handler';
import { isApiRequest, isAuthorized } from './lib/auth-token';
import { jsonError, jsonOk, textResponse } from './lib/http-response';
import { checkRateLimit } from './lib/rate-limit';
import { Router } from './lib/router';
import { listDomains, listRandomDomains } from './routes/domains';
import { generateEmail } from './routes/email-generate';
import { deleteMailbox, listMailbox, readMailboxMessage } from './routes/mailbox';
import { deleteMessage } from './routes/messages';
import { getLatestOtp } from './routes/otp';
import { uiHtml } from './ui/index-html';
import { uiScript } from './ui/app-script';
import { uiStyles } from './ui/styles-css';
import { domainHtml } from './ui/domain-html';
import { apiHtml } from './ui/api-html';
import type { Env } from './types';

const router = new Router()
  .add('GET', '/domains', listDomains)
  .add('GET', '/random-domains', listRandomDomains)
  .add('POST', '/email/generate', generateEmail)
  .add('GET', '/email/:domain/:user/messages', listMailbox)
  .add('GET', '/email/:domain/:user/messages/:id', readMailboxMessage)
  .add('DELETE', '/email/:domain/:user/messages/:id', deleteMessage)
  .add('DELETE', '/email/:domain/:user', deleteMailbox)
  .add('GET', '/email/:domain/:user/otp', getLatestOtp);

const serveUi = (request: Request) => {
  const path = new URL(request.url).pathname;
  if (path === '/') return textResponse(uiHtml, 'text/html; charset=utf-8');
  if (path === '/domain') return textResponse(domainHtml, 'text/html; charset=utf-8');
  if (path === '/api') return textResponse(apiHtml, 'text/html; charset=utf-8');
  if (path === '/assets/app.js') return textResponse(uiScript, 'text/javascript; charset=utf-8');
  if (path === '/assets/styles.css') return textResponse(uiStyles, 'text/css; charset=utf-8');
  return null;
};

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    if (!isAuthorized(request, env)) return jsonError(401, 'unauthorized', 'Missing or invalid bearer token');
    const uiResponse = serveUi(request);
    if (uiResponse) return uiResponse;
    if (new URL(request.url).pathname === '/health') return jsonOk({ ok: true });
    if (isApiRequest(request)) {
      const rateLimit = checkRateLimit(request, env);
      if (!rateLimit.allowed) return jsonError(429, 'rate_limited', 'Too many requests', { headers: { 'retry-after': String(rateLimit.retryAfterSeconds) } });
    }
    return router.handle(request, env, ctx);
  },

  async email(message: InboundEmailMessage, env: Env): Promise<void> {
    await handleEmail(message, env);
  },

  async scheduled(controller: ScheduledController, env: Env, ctx: ExecutionContext): Promise<void> {
    ctx.waitUntil(handleScheduled(controller, env));
  },
};
