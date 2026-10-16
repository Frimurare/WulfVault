// WulfVault - Copyright (c) 2025 Ulf Holmström. AGPL-3.0.
(() => {
  const menu = document.querySelector('.wv-navigation');
  if (!menu) return;
  const mobile = window.matchMedia('(max-width: 900px)');
  const syncMenu = () => { menu.open = !mobile.matches; };
  syncMenu();
  mobile.addEventListener('change', syncMenu);
  menu.addEventListener('keydown', (event) => {
    if (event.key === 'Escape' && mobile.matches && menu.open) {
      menu.open = false;
      menu.querySelector('summary').focus();
    }
  });
  const path = location.pathname.replace(/\/$/, '') || '/';
  const links = [...menu.querySelectorAll('a[href]')];
  const active = links
    .filter((link) => path === link.pathname || path.startsWith(link.pathname + '/'))
    .sort((a, b) => b.pathname.length - a.pathname.length)[0];
  if (active) {
    links.forEach((link) => link.removeAttribute('aria-current'));
    active.setAttribute('aria-current', 'page');
    const group = active.closest('.wv-nav-group');
    if (group) group.open = true;
    const heading = document.querySelector('.wv-current-page');
    if (heading) heading.textContent = active.textContent.trim();
  }
  document.querySelectorAll('.wv-languages a').forEach((link) => {
    const url = new URL(link.href);
    url.searchParams.set('return', location.pathname + location.search);
    link.href = url.pathname + url.search;
  });
})();
