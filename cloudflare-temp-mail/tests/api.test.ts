import { describe, expect, test } from 'vitest';
import worker from '../src/worker';
import { createCtx, createTestEnv } from './test-env';

const readJson = async (response: Response) => response.json() as Promise<{ success: boolean; data: any; error: any }>;

describe('api routes', () => {
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

  test('rejects API requests when token is enabled', async () => {
    const { env } = createTestEnv();
    env.API_TOKEN = 'secret';
    const response = await worker.fetch(new Request('https://mail.test/api/v1/domains'), env, createCtx() as ExecutionContext);
    expect(response.status).toBe(401);
  });
});
