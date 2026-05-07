import { describe, expect, test } from 'vitest';
import { runCleanup } from '../src/services/cleanup-service';
import { createTestEnv } from './test-env';

describe('cleanup service', () => {
  test('purges expired messages and leaves active messages visible', async () => {
    const oldDate = new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString();
    const activeDate = new Date().toISOString();
    const { env, state } = createTestEnv({
      messages: [
        { id: 'old', domain: 'example.com', user: 'tmp', sender: 'a', recipient: 'tmp@example.com', subject: '', received_at: oldDate, raw_key: 'old-raw', text_key: null, html_key: null, otp: null, size: 1, deleted_at: null, purged_at: null },
        { id: 'new', domain: 'example.com', user: 'tmp', sender: 'a', recipient: 'tmp@example.com', subject: '', received_at: activeDate, raw_key: 'new-raw', text_key: null, html_key: null, otp: null, size: 1, deleted_at: null, purged_at: null },
      ],
      objects: new Map([['old-raw', 'old'], ['new-raw', 'new']]),
    });

    await runCleanup(env);
    expect(state.messages.find((message) => message.id === 'old')?.purged_at).toBeTruthy();
    expect(state.messages.find((message) => message.id === 'new')?.purged_at).toBeNull();
    expect(state.objects.has('old-raw')).toBe(false);
    expect(state.objects.has('new-raw')).toBe(true);
  });

  test('purges messages across multiple cleanup batches', async () => {
    const oldDate = new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString();
    const { env, state } = createTestEnv({
      messages: [
        { id: 'old-1', domain: 'example.com', user: 'tmp', sender: 'a', recipient: 'tmp@example.com', subject: '', received_at: oldDate, raw_key: 'old-1-raw', text_key: null, html_key: null, otp: null, size: 1, deleted_at: null, purged_at: null },
        { id: 'old-2', domain: 'example.com', user: 'tmp', sender: 'a', recipient: 'tmp@example.com', subject: '', received_at: oldDate, raw_key: 'old-2-raw', text_key: null, html_key: null, otp: null, size: 1, deleted_at: null, purged_at: null },
      ],
      objects: new Map([['old-1-raw', 'old'], ['old-2-raw', 'old']]),
    });
    env.CLEANUP_BATCH_SIZE = '1';

    expect(await runCleanup(env)).toEqual({ purged: 1 });
    expect(state.messages.filter((message) => message.purged_at)).toHaveLength(1);

    expect(await runCleanup(env)).toEqual({ purged: 1 });
    expect(state.messages.filter((message) => message.purged_at)).toHaveLength(2);
    expect(state.objects.size).toBe(0);
  });

  test('does not mark message purged when R2 delete fails', async () => {
    const oldDate = new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString();
    const { env, state } = createTestEnv({
      messages: [
        { id: 'old', domain: 'example.com', user: 'tmp', sender: 'a', recipient: 'tmp@example.com', subject: '', received_at: oldDate, raw_key: 'old-raw', text_key: null, html_key: null, otp: null, size: 1, deleted_at: null, purged_at: null },
      ],
      objects: new Map([['old-raw', 'old']]),
    });
    env.MAIL_BUCKET.delete = async () => {
      throw new Error('R2 unavailable');
    };

    const result = await runCleanup(env);
    expect(result.purged).toBe(0);
    expect(state.messages[0].purged_at).toBeNull();
  });
});
