import { defineConfig } from '@playwright/test';

export default defineConfig({
  testMatch: /ui-smoke\.spec\.ts/,
  use: {
    browserName: 'chromium',
  },
});
