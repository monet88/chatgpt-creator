import { describe, expect, test } from 'vitest';
import { handleEmail, type InboundEmailMessage } from '../src/handlers/email-handler';
import { createTestEnv } from './test-env';

const createMessage = (to: string, raw: string, rawSize = raw.length): InboundEmailMessage & { rejected: string } => ({
  from: 'sender@test.dev',
  to,
  raw,
  rawSize,
  rejected: '',
  setReject(reason: string) {
    this.rejected = reason;
  },
});

describe('email handler', () => {
  test('stores inbound message metadata and R2 content', async () => {
    const { env, state } = createTestEnv();
    const message = createMessage('tmp@example.com', 'Subject: Login 654321\n\nYour code is 654321');
    await handleEmail(message, env);

    expect(message.rejected).toBe('');
    expect(state.messages).toHaveLength(1);
    expect(state.messages[0].otp).toBe('654321');
    expect(state.objects.size).toBeGreaterThan(0);
  });

  test('rejects unknown domain and oversized message', async () => {
    const { env } = createTestEnv();
    const unknown = createMessage('tmp@bad.dev', 'Subject: No');
    await handleEmail(unknown, env);
    expect(unknown.rejected).toBe('Unknown recipient domain');

    env.MAX_MESSAGE_BYTES = '2';
    const oversized = createMessage('tmp@example.com', 'Subject: Big', 100);
    await handleEmail(oversized, env);
    expect(oversized.rejected).toBe('Message too large');
  });

  test('rejects oversized decoded message when rawSize is missing', async () => {
    const { env } = createTestEnv();
    env.MAX_MESSAGE_BYTES = '2';
    const oversized = createMessage('tmp@example.com', 'Subject: Big');
    delete oversized.rawSize;
    await handleEmail(oversized, env);
    expect(oversized.rejected).toBe('Message too large');
  });
});
