/* WulfVault - Secure File Transfer System
 * Copyright (c) 2025 Ulf Holmström (Frimurare)
 * Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
 * You must retain this notice in any copy or derivative work.
 *
 * Mobile Navigation Handler
 */

(function() {
    'use strict';

    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initMobileNav);
    } else {
        initMobileNav();
    }

    function initMobileNav() {
        // Create hamburger button if it doesn't exist
        const header = document.querySelector('.header');
        if (!header) return;

        const nav = header.querySelector('nav');
        if (!nav) return;

        // Check if hamburger already exists
        let hamburger = header.querySelector('.hamburger');
        if (!hamburger) {
            // Create hamburger button
            hamburger = document.createElement('button');
            hamburger.className = 'hamburger';
            hamburger.setAttribute('aria-label', 'Toggle navigation');
            hamburger.setAttribute('aria-expanded', 'false');
            hamburger.innerHTML = '<span></span><span></span><span></span>';

            // Insert after logo
            const logo = header.querySelector('.logo');
            if (logo && logo.nextSibling) {
                header.insertBefore(hamburger, logo.nextSibling);
            } else {
                header.appendChild(hamburger);
            }
        }

        // Create overlay if it doesn't exist
        let overlay = document.querySelector('.mobile-nav-overlay');
        if (!overlay) {
            overlay = document.createElement('div');
            overlay.className = 'mobile-nav-overlay';
            document.body.appendChild(overlay);
        }

        // Toggle navigation
        function toggleNav() {
            const isActive = nav.classList.contains('active');

            if (isActive) {
                // Close nav
                nav.classList.remove('active');
                hamburger.classList.remove('active');
                overlay.classList.remove('active');
                hamburger.setAttribute('aria-expanded', 'false');
                document.body.style.overflow = '';
            } else {
                // Open nav
                nav.classList.add('active');
                hamburger.classList.add('active');
                overlay.classList.add('active');
                hamburger.setAttribute('aria-expanded', 'true');
                document.body.style.overflow = 'hidden'; // Prevent scrolling
            }
        }

        // Event listeners
        hamburger.addEventListener('click', toggleNav);
        overlay.addEventListener('click', toggleNav);

        // Close nav when clicking on a link
        const navLinks = nav.querySelectorAll('a');
        navLinks.forEach(link => {
            link.addEventListener('click', () => {
                if (window.innerWidth <= 768) {
                    toggleNav();
                }
            });
        });

        // Close nav on window resize if switching to desktop
        let resizeTimer;
        window.addEventListener('resize', () => {
            clearTimeout(resizeTimer);
            resizeTimer = setTimeout(() => {
                if (window.innerWidth > 768 && nav.classList.contains('active')) {
                    toggleNav();
                }
            }, 250);
        });

        // Handle escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && nav.classList.contains('active')) {
                toggleNav();
            }
        });

        // Add data-label attributes to table cells for mobile view
        addTableLabels();
    }

    function addTableLabels() {
        const tables = document.querySelectorAll('table');

        tables.forEach(table => {
            const headers = table.querySelectorAll('th');
            const headerTexts = Array.from(headers).map(th => th.textContent.trim());

            const rows = table.querySelectorAll('tbody tr');
            rows.forEach(row => {
                const cells = row.querySelectorAll('td');
                cells.forEach((cell, index) => {
                    if (headerTexts[index]) {
                        cell.setAttribute('data-label', headerTexts[index]);
                    }
                });
            });
        });
    }
})();
