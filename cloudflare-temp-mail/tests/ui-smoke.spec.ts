import { test, expect } from '@playwright/test';
import { uiHtml } from '../src/ui/index-html';
import { uiScript } from '../src/ui/app-script';
import { uiStyles } from '../src/ui/styles-css';

test('generates address and refreshes empty inbox', async ({ page }) => {
  await page.route('https://mail.test/', (route) => route.fulfill({ contentType: 'text/html', body: uiHtml }));
  await page.route('**/assets/styles.css', (route) => route.fulfill({ contentType: 'text/css', body: uiStyles }));
  await page.route('**/assets/app.js', (route) => route.fulfill({ contentType: 'text/javascript', body: uiScript }));
  await page.route('**/api/v1/domains', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { domains: ['example.com'] }, error: null }) }));
  await page.route('**/api/v1/email/generate', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { email: 'tmp-ui@example.com', user: 'tmp-ui', domain: 'example.com' }, error: null }) }));
  await page.route('**/api/v1/email/example.com/tmp-ui/messages', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { messages: [] }, error: null }) }));
  await page.route('**/api/v1/email/example.com/tmp-ui/otp', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { otp: null, status: 'pending', receivedAt: null }, error: null }) }));

  await page.goto('https://mail.test/', { waitUntil: 'networkidle' });
  await expect(page.locator('#domain-select')).toHaveValue('example.com');

  await page.locator('#generate-email').click();
  await expect(page.locator('#email-output')).toHaveValue('tmp-ui@example.com');
  await expect(page.locator('#status-badge')).toHaveText('READY');

  await page.locator('#refresh-inbox').click();
  await expect(page.locator('#message-count')).toHaveText('0 messages');
  await expect(page.locator('#copy-otp')).toHaveText('No OTP yet');
});
