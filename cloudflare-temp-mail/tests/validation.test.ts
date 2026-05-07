import { describe, expect, test } from 'vitest';
import { parseEmailAddress } from '../src/lib/email-address';
import { extractOtp } from '../src/services/otp-extractor';

describe('validation helpers', () => {
  test('parses valid email and rejects invalid addresses', () => {
    expect(parseEmailAddress('User.Name@example.com')).toEqual({ user: 'user.name', domain: 'example.com', email: 'user.name@example.com' });
    expect(parseEmailAddress('bad-address')).toBeNull();
    expect(parseEmailAddress('x@localhost')).toBeNull();
  });

  test('extracts numeric OTP from subject or body', () => {
    expect(extractOtp('Login code 123456')).toBe('123456');
    expect(extractOtp('hello', 'Your code is 998877')).toBe('998877');
    expect(extractOtp('no code')).toBeNull();
  });
});
