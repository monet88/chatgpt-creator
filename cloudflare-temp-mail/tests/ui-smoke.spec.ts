import { test, expect, type Page } from '@playwright/test';
import { uiHtml } from '../src/ui/index-html';
import { uiScript } from '../src/ui/app-script';
import { uiStyles } from '../src/ui/styles-css';

const routeShell = async (page: Page) => {
  await page.route('https://mail.test/', (route) => route.fulfill({ contentType: 'text/html', body: uiHtml }));
  await page.route('**/assets/styles.css', (route) => route.fulfill({ contentType: 'text/css', body: uiStyles }));
  await page.route('**/assets/app.js', (route) => route.fulfill({ contentType: 'text/javascript', body: uiScript }));
  await page.route('**/api/v1/domains', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { domains: ['example.com'] }, error: null }) }));
  await page.route('**/api/v1/email/example.com/tmp-ui/otp', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { otp: null, status: 'pending', receivedAt: null }, error: null }) }));
};

test('copies mailbox URL and shows message detail inline', async ({ context, page }) => {
  const mailboxUrl = 'https://mail.test/#tmp-ui%40example.com';

  await context.grantPermissions(['clipboard-read', 'clipboard-write'], { origin: 'https://mail.test' });
  await routeShell(page);
  await page.route('**/api/v1/email/example.com/tmp-ui/messages', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { messages: [{ id: 'msg-1', from: 'Thắng Minh <minhthang421992@gmail.com>', subject: 'tmp-ui@example.com', receivedAt: '2026-05-08T08:40:59.000Z' }] }, error: null }) }));
  await page.route('**/api/v1/email/example.com/tmp-ui/messages/msg-1', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { id: 'msg-1', from: 'Thắng Minh <minhthang421992@gmail.com>', subject: 'tmp-ui@example.com', receivedAt: '2026-05-08T08:40:59.000Z', text: 'tmp-ui@example.com' }, error: null }) }));

  await page.goto(mailboxUrl, { waitUntil: 'networkidle' });

  await expect(page.locator('#current-email')).toHaveText('tmp-ui@example.com');
  await expect(page.locator('#url-email')).toHaveText(mailboxUrl);
  await page.locator('#url-email').click();
  await expect.poll(() => page.evaluate<string>('navigator.clipboard.readText()')).toBe(mailboxUrl);

  await expect(page.locator('#pager-status')).toHaveText('Hiển thị 1-1 / 1 email');
  await page.locator('.btn-read').click();
  await expect(page.locator('#message-detail')).toBeVisible();
  await expect(page.locator('#detail-subject')).toHaveText('tmp-ui@example.com');
  await expect(page.locator('#detail-body')).toContainText('tmp-ui@example.com');
});

test('renders HTML email in a sandboxed inline iframe', async ({ page }) => {
  await routeShell(page);
  await page.route('**/api/v1/email/example.com/tmp-ui/messages', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { messages: [{ id: 'msg-html', from: 'Sender <sender@example.com>', subject: 'HTML mail', receivedAt: '2026-05-08T08:40:59.000Z' }] }, error: null }) }));
  await page.route('**/api/v1/email/example.com/tmp-ui/messages/msg-html', (route) => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, data: { id: 'msg-html', from: 'Sender <sender@example.com>', subject: 'HTML mail', receivedAt: '2026-05-08T08:40:59.000Z', html: '<main><p>HTML body</p></main>' }, error: null }) }));

  await page.goto('https://mail.test/#tmp-ui%40example.com', { waitUntil: 'networkidle' });
  await page.locator('.btn-read').click();

  const iframe = page.locator('#detail-body iframe');
  await expect(iframe).toBeVisible();
  await expect(iframe).toHaveAttribute('sandbox', '');
});
