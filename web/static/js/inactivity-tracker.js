// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

// Inactivity Timer
// Logs out users after 10 minutes of inactivity (no mouse/keyboard activity and no active transfers)

(function() {
    const INACTIVITY_TIMEOUT = 10 * 60 * 1000; // 10 minutes in milliseconds
    const WARNING_TIME = 60 * 1000; // Show warning 1 minute before timeout
    const CHECK_INTERVAL = 5 * 1000; // Check every 5 seconds

    let lastActivityTime = Date.now();
    let warningShown = false;
    let checkInterval = null;
    let activeTransfer = false; // Flag to track if upload/download is in progress

    // Update last activity time
    function updateActivity() {
        lastActivityTime = Date.now();
        warningShown = false;
        hideWarning();
    }

    // Check for inactivity
    function checkInactivity() {
        // Don't check if there's an active transfer
        if (activeTransfer) {
            return;
        }

        const now = Date.now();
        const timeSinceActivity = now - lastActivityTime;
        const timeUntilLogout = INACTIVITY_TIMEOUT - timeSinceActivity;

        // Show warning 1 minute before logout
        if (timeUntilLogout <= WARNING_TIME && !warningShown) {
            showWarning(Math.ceil(timeUntilLogout / 1000));
            warningShown = true;
        }

        // Logout if timeout reached
        if (timeSinceActivity >= INACTIVITY_TIMEOUT) {
            logout();
        }
    }

    // Show warning banner
    function showWarning(secondsRemaining) {
        // Remove existing warning if any
        hideWarning();

        const warning = document.createElement('div');
        warning.id = 'inactivity-warning';
        warning.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
            color: white;
            padding: 16px;
            text-align: center;
            z-index: 10000;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            animation: slideDown 0.3s ease-out;
        `;

        warning.innerHTML = `
            <div style="max-width: 800px; margin: 0 auto; display: flex; align-items: center; justify-content: center; gap: 16px;">
                <svg style="width: 24px; height: 24px; flex-shrink: 0;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                </svg>
                <div style="flex: 1;">
                    <strong style="font-size: 16px; display: block; margin-bottom: 4px;">Inaktivitetsvarning</strong>
                    <span style="font-size: 14px; opacity: 0.95;">Du kommer att loggas ut om <span id="seconds-remaining">${secondsRemaining}</span> sekunder på grund av inaktivitet.</span>
                </div>
                <button onclick="window.inactivityTracker.stayLoggedIn()" style="
                    background: white;
                    color: #ff6b6b;
                    border: none;
                    padding: 10px 20px;
                    border-radius: 6px;
                    font-weight: 600;
                    cursor: pointer;
                    font-size: 14px;
                    transition: all 0.2s;
                    flex-shrink: 0;
                " onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">
                    Stanna inloggad
                </button>
            </div>
        `;

        // Add slide-down animation
        const style = document.createElement('style');
        style.textContent = `
            @keyframes slideDown {
                from {
                    transform: translateY(-100%);
                    opacity: 0;
                }
                to {
                    transform: translateY(0);
                    opacity: 1;
                }
            }
        `;
        document.head.appendChild(style);

        document.body.insertBefore(warning, document.body.firstChild);

        // Update countdown every second
        const countdownInterval = setInterval(() => {
            const element = document.getElementById('seconds-remaining');
            if (!element) {
                clearInterval(countdownInterval);
                return;
            }

            const now = Date.now();
            const timeSinceActivity = now - lastActivityTime;
            const timeUntilLogout = INACTIVITY_TIMEOUT - timeSinceActivity;
            const secondsLeft = Math.ceil(timeUntilLogout / 1000);

            if (secondsLeft > 0) {
                element.textContent = secondsLeft;
            } else {
                clearInterval(countdownInterval);
            }
        }, 1000);
    }

    // Hide warning banner
    function hideWarning() {
        const warning = document.getElementById('inactivity-warning');
        if (warning) {
            warning.remove();
        }
    }

    // Stay logged in (reset timer)
    function stayLoggedIn() {
        updateActivity();
    }

    // Logout
    function logout() {
        window.location.href = '/login?timeout=1';
    }

    // Mark transfer as active (called when upload/download starts)
    function markTransferActive() {
        activeTransfer = true;
        console.log('Transfer started - inactivity timer paused');
    }

    // Mark transfer as inactive (called when upload/download ends)
    function markTransferInactive() {
        activeTransfer = false;
        updateActivity(); // Reset activity time when transfer completes
        console.log('Transfer completed - inactivity timer resumed');
    }

    // Initialize
    function init() {
        // Listen for user activity
        const activityEvents = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click'];
        activityEvents.forEach(event => {
            document.addEventListener(event, updateActivity, { passive: true });
        });

        // Start checking for inactivity
        checkInterval = setInterval(checkInactivity, CHECK_INTERVAL);

        // Initial activity time
        updateActivity();

        console.log('Inactivity tracker initialized (10 minute timeout)');
    }

    // Expose public API
    window.inactivityTracker = {
        init: init,
        stayLoggedIn: stayLoggedIn,
        markTransferActive: markTransferActive,
        markTransferInactive: markTransferInactive,
        updateActivity: updateActivity
    };

    // Auto-initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
