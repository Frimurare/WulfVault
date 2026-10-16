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
})();
