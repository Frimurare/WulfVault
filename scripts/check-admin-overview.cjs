// WulfVault - Copyright (c) 2025 Ulf Holmström. AGPL-3.0.
// Run by the dashboard validation workflow against generated, disposable pages.
const { chromium } = require('playwright');
const AxeBuilder = require('@axe-core/playwright').default;
const fs = require('node:fs/promises');
const assert = require('node:assert/strict');

(async () => {
  const browser = await chromium.launch();
  await fs.mkdir('ui-screenshots', { recursive: true });
  try {
    for (const lang of ['en', 'sv']) {
      for (const width of [1440, 1024, 768, 390, 320]) {
        const context = await browser.newContext({ viewport: { width, height: 1000 } });
        const page = await context.newPage();
        await page.goto('http://127.0.0.1:8765/' + lang + '.html');
        await page.waitForFunction(() => document.querySelector('.wv-navigation').open === (innerWidth > 900));
        assert.equal(await page.locator('html').getAttribute('lang'), lang);
        assert.ok(await page.locator('body').evaluate((body) => body.scrollWidth <= innerWidth), 'horizontal overflow at ' + width + '/' + lang);
        assert.equal(await page.locator('.wv-nav-link[aria-current="page"]').count(), 1);
        if (width <= 900) {
          const menu = page.locator('.wv-navigation');
          await menu.locator(':scope > summary').click();
          assert.equal(await menu.getAttribute('open'), '');
          await page.locator('.wv-nav-group').first().locator('summary').click();
          assert.ok(await page.locator('.wv-navigation a[href="/admin/duplicates"]').isVisible());
          await page.locator('.wv-navigation a[href="/admin/duplicates"]').focus();
          await page.keyboard.press('Escape');
          assert.equal(await menu.getAttribute('open'), null);
          assert.ok(await menu.locator(':scope > summary').evaluate((el) => el === document.activeElement));
        } else {
          await page.locator('.wv-nav-group').first().locator('summary').focus();
          await page.keyboard.press('Enter');
          assert.ok(await page.locator('.wv-navigation a[href="/admin/duplicates"]').isVisible());
        }
        await page.locator('.wv-more').first().locator('summary').click();
        assert.ok(await page.locator('.wv-detail-grid').isVisible());
        assert.ok(await page.locator('body').evaluate((body) => body.scrollWidth <= innerWidth), 'expanded content overflow at ' + width + '/' + lang);
        const result = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21aa']).analyze();
        assert.deepEqual(result.violations.map((v) => ({ id: v.id, nodes: v.nodes.map((n) => n.target) })), [], 'accessibility violations');
        await page.locator('.wv-more').first().locator('summary').click();
        await page.screenshot({ path: 'ui-screenshots/' + lang + '-' + width + '.png', fullPage: true });
        if (lang === 'sv' && [1440, 390].includes(width)) {
          const png = await fs.readFile('ui-screenshots/' + lang + '-' + width + '.png');
          const encoded = png.toString('base64');
          for (let i = 0; i < encoded.length; i += 4000) console.log('WV_PREVIEW_' + width + ':' + encoded.slice(i, i + 4000));
        }
        await context.close();
      }
    }
    const noJS = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 844 } });
    const page = await noJS.newPage();
    await page.goto('http://127.0.0.1:8765/sv.html');
    assert.ok(await page.locator('a[href="/admin/users"]').first().isVisible(), 'navigation must work without JS');
    await noJS.close();
    console.log('PASS: EN/SV at 1440, 1024, 768, 390 and 320 px; keyboard, no-JS navigation, overflow and WCAG checks.');
  } finally {
    await browser.close();
  }
})().catch((error) => { console.error(error); process.exitCode = 1; });
