import { beforeEach, describe, expect, test } from 'vitest';
import { handleEmail, type InboundEmailMessage } from '../src/handlers/email-handler';
import { resetRateLimitForTests } from '../src/lib/rate-limit';
import worker from '../src/worker';
import { createCtx, createTestEnv } from './test-env';

const readJson = async (response: Response) => response.json() as Promise<{ success: boolean; data: any; error: any }>;

const createMessage = (to: string): InboundEmailMessage & { rejected: string } => ({
  from: 'sender@test.dev',
  to,
  raw: 'Subject: Login 654321\n\nYour code is 654321',
  rawSize: 41,
  rejected: '',
  setReject(reason: string) {
    this.rejected = reason;
  },
});

describe('api routes', () => {
  beforeEach(() => resetRateLimitForTests());

  test('lists domains and generates address under /api/v1', async () => {
    const { env } = createTestEnv();
    const domains = await readJson(await worker.fetch(new Request('https://mail.test/api/v1/domains'), env, createCtx() as ExecutionContext));
    expect(domains.data.domains).toEqual(['example.com']);

    const generated = await readJson(
      await worker.fetch(
        new Request('https://mail.test/api/v1/email/generate', { method: 'POST', body: JSON.stringify({ domain: 'example.com' }) }),
        env,
        createCtx() as ExecutionContext,
      ),
    );
    expect(generated.success).toBe(true);
    expect(generated.data.email).toMatch(/@example\.com$/);
  });

  test('serves health and random domains routes', async () => {
    const { env } = createTestEnv({ domains: ['example.com', 'mail.test'] });
    const health = await readJson(await worker.fetch(new Request('https://mail.test/health'), env, createCtx() as ExecutionContext));
    expect(health.data).toEqual({ ok: true });

    const random = await readJson(await worker.fetch(new Request('https://mail.test/api/v1/random-domains?limit=1'), env, createCtx() as ExecutionContext));
    expect(random.data.domains).toHaveLength(1);
    expect(['example.com', 'mail.test']).toContain(random.data.domains[0]);
  });

  test('does not emit CORS headers for cross-origin browser calls', async () => {
    const { env } = createTestEnv();
    const response = await worker.fetch(new Request('https://mail.test/api/v1/domains', { headers: { origin: 'https://app.example' } }), env, createCtx() as ExecutionContext);
    expect(response.status).toBe(200);
    expect(response.headers.get('access-control-allow-origin')).toBeNull();
  });

  test('does not rate limit health checks', async () => {
    const { env } = createTestEnv();
    env.RATE_LIMIT_MAX_REQUESTS = '1';
    await worker.fetch(new Request('https://mail.test/api/v1/domains', { headers: { 'cf-connecting-ip': '203.0.113.2' } }), env, createCtx() as ExecutionContext);
    await worker.fetch(new Request('https://mail.test/api/v1/domains', { headers: { 'cf-connecting-ip': '203.0.113.2' } }), env, createCtx() as ExecutionContext);

    const health = await worker.fetch(new Request('https://mail.test/health', { headers: { 'cf-connecting-ip': '203.0.113.2' } }), env, createCtx() as ExecutionContext);
    expect(health.status).toBe(200);
  });

  test('rate limits requests per client', async () => {
    const { env } = createTestEnv();
    env.RATE_LIMIT_MAX_REQUESTS = '1';
    env.RATE_LIMIT_WINDOW_SECONDS = '60';
    const first = await worker.fetch(new Request('https://mail.test/api/v1/domains', { headers: { 'cf-connecting-ip': '203.0.113.1' } }), env, createCtx() as ExecutionContext);
    expect(first.status).toBe(200);

    const second = await worker.fetch(new Request('https://mail.test/api/v1/domains', { headers: { 'cf-connecting-ip': '203.0.113.1' } }), env, createCtx() as ExecutionContext);
    const body = await readJson(second);
    expect(second.status).toBe(429);
    expect(second.headers.get('retry-after')).toBe('60');
    expect(body.error.code).toBe('rate_limited');
  });

  test('reads mailbox, message detail, otp, and delete-all', async () => {
    const { env, state } = createTestEnv({
      messages: [
        {
          id: 'm1', domain: 'example.com', user: 'tmp', sender: 'a@test.dev', recipient: 'tmp@example.com', subject: 'Code', received_at: '2026-05-07T00:00:00.000Z', raw_key: 'raw', text_key: 'text', html_key: null, otp: '123456', size: 12, deleted_at: null, purged_at: null,
        },
      ],
      objects: new Map([['text', 'Your code is 123456']]),
    });

    const list = await readJson(await worker.fetch(new Request('https://mail.test/api/v1/email/example.com/tmp/messages'), env, createCtx() as ExecutionContext));
    expect(list.data.messages).toHaveLength(1);

    const detail = await readJson(await worker.fetch(new Request('https://mail.test/api/v1/email/example.com/tmp/messages/m1'), env, createCtx() as ExecutionContext));
    expect(detail.data.body).toBe('Your code is 123456');

    const otp = await readJson(await worker.fetch(new Request('https://mail.test/api/v1/email/example.com/tmp/otp'), env, createCtx() as ExecutionContext));
    expect(otp.data).toMatchObject({ otp: '123456', status: 'received' });

    await worker.fetch(new Request('https://mail.test/api/v1/email/example.com/tmp', { method: 'DELETE' }), env, createCtx() as ExecutionContext);
    expect(state.messages[0].deleted_at).toBeTruthy();
    expect(state.objects.has('text')).toBe(true);

    const afterDelete = createMessage('tmp@example.com');
    await handleEmail(afterDelete, env);
    expect(afterDelete.rejected).toBe('Unknown mailbox');
  });

  test('scopes message delete to mailbox route ownership', async () => {
    const { env, state } = createTestEnv({
      messages: [
        { id: 'm1', domain: 'example.com', user: 'tmp', sender: 'a@test.dev', recipient: 'tmp@example.com', subject: 'Code', received_at: '2026-05-07T00:00:00.000Z', raw_key: 'raw', text_key: null, html_key: null, otp: null, size: 12, deleted_at: null, purged_at: null },
      ],
    });

    await worker.fetch(new Request('https://mail.test/api/v1/email/example.com/other/messages/m1', { method: 'DELETE' }), env, createCtx() as ExecutionContext);
    expect(state.messages[0].deleted_at).toBeNull();

    await worker.fetch(new Request('https://mail.test/api/v1/email/example.com/tmp/messages/m1', { method: 'DELETE' }), env, createCtx() as ExecutionContext);
    expect(state.messages[0].deleted_at).toBeTruthy();
  });

  test('rejects API requests when token is missing or invalid', async () => {
    const { env } = createTestEnv();
    delete env.AUTH_DISABLED;
    const missing = await worker.fetch(new Request('https://mail.test/api/v1/domains'), env, createCtx() as ExecutionContext);
    expect(missing.status).toBe(401);

    env.API_TOKEN = 'secret';
    const invalid = await worker.fetch(new Request('https://mail.test/api/v1/domains'), env, createCtx() as ExecutionContext);
    expect(invalid.status).toBe(401);
  });

  test('allows explicit auth-disabled development mode and bearer token auth', async () => {
    const { env } = createTestEnv();
    const dev = await worker.fetch(new Request('https://mail.test/api/v1/domains'), env, createCtx() as ExecutionContext);
    expect(dev.status).toBe(200);

    delete env.AUTH_DISABLED;
    env.API_TOKEN = 'secret';
    const authed = await worker.fetch(new Request('https://mail.test/api/v1/domains', { headers: { authorization: 'Bearer secret' } }), env, createCtx() as ExecutionContext);
    expect(authed.status).toBe(200);
  });

  test('returns 400 for malformed URL-encoded params', async () => {
    const { env } = createTestEnv();
    const response = await worker.fetch(new Request('https://mail.test/api/v1/email/example.com/%/messages'), env, createCtx() as ExecutionContext);
    const body = await readJson(response);
    expect(response.status).toBe(400);
    expect(body.error.code).toBe('invalid_path_param');
  });
});
