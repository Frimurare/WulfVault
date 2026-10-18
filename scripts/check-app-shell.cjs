// WulfVault - Copyright (c) 2025 Ulf Holmström. AGPL-3.0.
const { chromium } = require('playwright');
const AxeBuilder = require('@axe-core/playwright').default;
const fs = require('node:fs/promises');
const path = require('node:path');
const assert = require('node:assert/strict');

const pages = [
  ['users', '/admin/users'], ['user-edit', '/admin/users/edit'],
  ['recipient-edit', '/admin/download-accounts/edit'], ['files', '/admin/files'],
  ['duplicates', '/admin/duplicates'], ['trash', '/admin/trash'],
  ['branding', '/admin/branding'], ['server-settings', '/admin/settings'],
  ['identity', '/admin/identity-providers'], ['email', '/admin/email-settings'],
  ['audit', '/admin/audit-logs'], ['logs', '/admin/server-logs'],
  ['monitor', '/admin/sysmonitor-logs'], ['admin-teams', '/admin/teams'],
  ['teams', '/teams'], ['team-files', '/teams/files'],
  ['my-files', '/dashboard'], ['settings', '/settings'],
  ['account', '/settings/account'], ['downloads', '/download/dashboard'],
  ['download-password', '/download/change-password'], ['download-account', '/download/account-settings'],
  ['login', '/login', true], ['forgot-password', '/forgot-password', true],
  ['download-splash', '/s/demo-report', true]
];

(async () => {
  const browser = await chromium.launch();
  const failures = [];
  await fs.mkdir('ui-screenshots', { recursive: true });
  try {
    for (const lang of ['en', 'sv']) {
      for (const width of [1440, 768, 390, 320]) {
        for (const [name, routePath, isPublic] of pages) {
          const context = await browser.newContext({ viewport: { width, height: 1000 } });
          const page = await context.newPage();
          const errors = [];
          page.on('pageerror', (error) => errors.push(error.message));
          // Serve real Go-rendered fixtures at their real route paths so active
          // navigation and language return URLs behave as they do in the app.
          await page.route('**/*', async (route) => {
            const request = route.request();
            const url = new URL(request.url());
            if (url.pathname.startsWith('/static/')) return route.continue();
            if (request.isNavigationRequest()) {
              const html = await fs.readFile(path.join(process.env.WV_UI_PREVIEW_DIR, name + '-' + lang + '.html'));
              return route.fulfill({ contentType: 'text/html; charset=utf-8', body: html });
            }
            if (request.resourceType() === 'fetch' || request.resourceType() === 'xhr') {
              return route.fulfill({ contentType: 'application/json', body: JSON.stringify({ success: true, logs: [], teams: [], files: [], emails: [], total: 0, total_count: 0, count: 0 }) });
            }
            return route.abort();
          });
          try {
            await page.goto('http://127.0.0.1:8765' + routePath + '?fixture_lang=' + lang);
            await page.waitForLoadState('networkidle');
            if (!isPublic) {
              assert.equal(await page.locator('.wv-sidebar').count(), 1);
              assert.equal(await page.locator('html').getAttribute('lang'), lang);
              assert.equal(await page.locator('.wv-navigation').getAttribute('open') !== null, width > 900);
              assert.equal(await page.locator('#main').count(), 1);
              const selected = page.locator('.wv-nav-link[aria-current="page"]');
              assert.ok(await selected.count() >= 1, 'no active route');
              const shellAxe = await new AxeBuilder({ page }).include('.wv-sidebar').include('.wv-topbar').withTags(['wcag2a', 'wcag2aa', 'wcag21aa']).analyze();
              assert.deepEqual(shellAxe.violations.map((v) => ({ id: v.id, nodes: v.nodes.map((n) => n.target) })), [], 'shared navigation accessibility');
              const languageLink = page.locator('.wv-languages a').first();
              assert.ok((await languageLink.getAttribute('href')).includes('return='), 'language switch must preserve current route');
              if (width <= 900) {
                await page.locator('.wv-navigation > summary').click();
                assert.ok(await page.locator('.wv-navigation > nav').isVisible());
                await page.locator('.wv-navigation > summary').focus();
                await page.keyboard.press('Escape');
                assert.equal(await page.locator('.wv-navigation').getAttribute('open'), null);
              }
            }
            if (name === 'admin-teams') {
              const action = page.locator('.team-actions .btn-action').first();
              assert.equal(await action.evaluate((el) => getComputedStyle(el).backgroundColor), 'rgb(255, 255, 255)', 'team actions must use the shared neutral style');
              assert.ok(!/\p{Extended_Pictographic}/u.test(await page.locator('.team-actions').innerText()), 'team actions must not contain legacy emoji');
            }
            const overflow = await page.evaluate(() => ({
              page: document.documentElement.scrollWidth,
              width: innerWidth,
              offenders: [...document.querySelectorAll('body *')].filter((el) => el.getBoundingClientRect().right > innerWidth + 1).slice(0, 8).map((el) => el.tagName + '.' + el.className)
            }));
            assert.ok(overflow.page <= width + 1, 'horizontal overflow: ' + JSON.stringify(overflow));
            assert.deepEqual(errors, [], 'page script errors');
            await page.evaluate(() => { document.activeElement?.blur(); window.scrollTo(0, 0); });
            if (lang === 'sv' && [1440, 390].includes(width)) {
              await page.screenshot({ path: 'ui-screenshots/' + name + '-' + width + '.png', fullPage: true });
              if (['admin-teams', 'my-files', 'files'].includes(name)) {
                const png = await fs.readFile('ui-screenshots/' + name + '-' + width + '.png');
                const encoded = png.toString('base64');
                for (let i = 0; i < encoded.length; i += 4000) console.log('WV_APP_PREVIEW_' + name + '_' + width + ':' + encoded.slice(i, i + 4000));
              }
            }
          } catch (error) {
            failures.push(name + '/' + lang + '/' + width + ': ' + error.message);
            await page.screenshot({ path: 'ui-screenshots/FAILED-' + name + '-' + lang + '-' + width + '.png', fullPage: true });
          } finally {
            await context.close();
          }
        }
      }
    }
  } finally {
    await browser.close();
  }
  assert.deepEqual(failures, [], 'application page checks failed');
  console.log('PASS: 25 application pages in EN/SV at 1440, 768, 390 and 320 px; shared shell accessibility, responsive navigation, active routes and page scripts.');
})().catch((error) => { console.error(error); process.exitCode = 1; });
