# Cookie Policy

**⚠️ ACTION REQUIRED: This is a template. Replace all [PLACEHOLDERS] with your information before publishing.**

**Last Updated:** [DATE]
**Effective Date:** [DATE]

---

## What Are Cookies?

Cookies are small text files that are stored on your device (computer, tablet, or mobile) when you visit a website. They help websites remember your preferences and provide essential functionality.

This Cookie Policy explains what cookies are, how we use them, and your choices regarding their use.

---

## Cookies We Use

### Essential Cookies (No Consent Required)

These cookies are **strictly necessary** for the Service to function and cannot be disabled in our systems. They are usually only set in response to actions you make that amount to a request for services, such as logging in or filling in forms.

| Cookie Name | Purpose | Duration | Type |
|-------------|---------|----------|------|
| **session** | Authentication and session management | 24 hours | Essential (HTTP-Only, Secure, SameSite=Lax) |

**What This Cookie Does:**
- Keeps you logged in while you use the Service
- Maintains your authentication state
- Prevents unauthorized access to your account
- Essential for security and functionality

**Security Features:**
- **HttpOnly**: Cannot be accessed via JavaScript (prevents XSS attacks)
- **Secure**: Only transmitted over HTTPS
- **SameSite=Lax**: Protects against CSRF attacks
- **Expires**: After 24 hours or when you log out

**Legal Basis:** Strictly necessary for service provision (no consent required under ePrivacy Directive)

---

## Cookies We Do NOT Use

**WulfVault is privacy-focused and does NOT use:**

- ❌ **Analytics Cookies** (Google Analytics, Matomo, etc.)
- ❌ **Advertising Cookies** (Google Ads, Facebook Pixel, etc.)
- ❌ **Tracking Cookies** (third-party trackers)
- ❌ **Social Media Cookies** (Facebook, Twitter, etc.)
- ❌ **Profiling Cookies** (behavioral tracking)
- ❌ **Marketing Cookies** (remarketing, conversion tracking)

**We prioritize your privacy.** Our Service operates with minimal data collection and no unnecessary tracking.

---

## How to Control Cookies

### Browser Settings

You can control and manage cookies through your browser settings:

#### Google Chrome
1. Settings → Privacy and security → Cookies and other site data
2. Choose your preferred cookie setting
3. Block third-party cookies or all cookies

#### Mozilla Firefox
1. Settings → Privacy & Security → Cookies and Site Data
2. Choose "Block cookies" or "Delete cookies and site data when Firefox is closed"

#### Safari
1. Preferences → Privacy
2. Choose "Block all cookies" or "Prevent cross-site tracking"

#### Microsoft Edge
1. Settings → Privacy, search, and services
2. Choose "Block third-party cookies" or "Block all cookies"

### Impact of Blocking Essential Cookies

**⚠️ Important:** If you block the `session` cookie, you will NOT be able to:
- Log in to your account
- Access your files
- Use the Service

The session cookie is **essential** for authentication and cannot be bypassed.

---

## Cookie Consent Banner

When you first visit [YOUR WULFVAULT INSTANCE URL], you will see a cookie consent banner informing you about our use of essential cookies.

**What You Need to Know:**
- We only use **essential cookies** for authentication
- These cookies do NOT require your explicit consent under ePrivacy Directive Article 5(3) (they are strictly necessary)
- The banner is informational and helps you understand our cookie usage
- You can dismiss the banner by clicking "Accept" or closing it

**No Tracking:** We do not set any cookies until you log in. Browsing the public pages does not set cookies.

---

## Cookie Retention

| Cookie | Retention Period | Deletion |
|--------|------------------|----------|
| `session` | 24 hours from last activity | Automatically deleted after 24 hours or when you log out |

**How to Manually Delete Cookies:**
- **Log Out**: Click "Logout" to immediately invalidate your session
- **Clear Browser Data**: Use your browser's "Clear browsing data" feature
- **Session Expiry**: Wait 24 hours for automatic expiration

---

## Third-Party Cookies

**We do NOT use third-party cookies.**

All cookies set by [YOUR WULFVAULT INSTANCE URL] are **first-party cookies** (set directly by our domain).

**No Third-Party Services:**
- No analytics providers (Google Analytics, etc.)
- No advertising networks
- No social media integrations
- No CDNs that set cookies
- No tracking scripts

**Exception:** If you deploy WulfVault with optional integrations (e.g., custom analytics, email delivery services), you must:
1. Update this Cookie Policy with details of third-party cookies
2. Obtain user consent if cookies are not strictly necessary
3. Provide opt-out mechanisms

---

## Cookie Details (Technical Specifications)

### Session Cookie

```
Name: session
Value: [encrypted session token]
Domain: [YOUR DOMAIN]
Path: /
Expires: 24 hours from last activity
HttpOnly: Yes
Secure: Yes (HTTPS only)
SameSite: Lax
```

**What Is Stored:**
The session cookie contains an encrypted token that links to your authenticated session in our database. It does NOT contain:
- Your password
- Personal data
- File contents
- Browsing history

**Encryption:**
Session tokens are randomly generated and have no meaning outside our system. They are stored securely in our database and matched to your account when you make requests.

---

## Updates to This Cookie Policy

We may update this Cookie Policy to reflect:
- Changes in cookie usage
- New legal requirements
- Service improvements

**Notification of Changes:**
- Updated "Last Updated" date at the top of this policy
- [OPTIONAL: Email notification for material changes]

**Your Continued Use:**
Continued use of the Service after changes constitutes acceptance of the updated Cookie Policy.

---

## Your Rights Under GDPR

Even though we only use essential cookies, you have rights regarding your data:

| Right | How to Exercise |
|-------|-----------------|
| **Right to Information** | Read this Cookie Policy |
| **Right to Access** | Settings → Export My Data |
| **Right to Deletion** | Settings → Delete My Account |
| **Right to Object** | Disable cookies (but Service won't function) |
| **Right to Complain** | Contact your Supervisory Authority |

**Contact Us:** See Section 10 below for privacy inquiries.

---

## Legal Basis

### ePrivacy Directive (Cookie Law)

**Article 5(3) of Directive 2002/58/EC:**
Cookies that are **strictly necessary** for the provision of a service explicitly requested by the user do NOT require consent.

**Our Compliance:**
The `session` cookie is strictly necessary for:
- Authentication (you explicitly request login)
- Service provision (file storage requires authentication)
- Security (prevents unauthorized access)

**Result:** No consent required, but we provide transparency through this Cookie Policy and our consent banner.

### GDPR

**Article 6(1)(b) - Contractual Necessity:**
Processing of session data is necessary to provide the Service you requested (file storage and sharing).

---

## Contact Us

If you have questions about our Cookie Policy:

**[YOUR ORGANIZATION NAME]**
**Email:** [PRIVACY CONTACT EMAIL]
**Address:** [POSTAL ADDRESS]
**Data Protection Officer:** [DPO EMAIL] (if applicable)

**Response Time:** We aim to respond within 30 days.

---

## Supervisory Authority

You have the right to lodge a complaint with a data protection authority:

**[YOUR LOCAL SUPERVISORY AUTHORITY]**
**Website:** [AUTHORITY WEBSITE]

**EU Supervisory Authorities:**
https://edpb.europa.eu/about-edpb/about-edpb/members_en

---

## Appendix: Cookie Audit Log

**Current Cookies (As of [DATE]):**

| Cookie | Set By | Purpose | Type | Consent? | Duration |
|--------|--------|---------|------|----------|----------|
| session | WulfVault | Authentication | Essential | No | 24 hours |

**Change Log:**
- [DATE]: Initial Cookie Policy - Only session cookie used

**Future Changes:**
If we add new cookies, this table will be updated and users notified.

---

## Frequently Asked Questions (FAQ)

### Q: Why don't you use analytics cookies?

**A:** WulfVault is designed with privacy-by-design principles. We don't track user behavior, page views, or usage patterns. Our philosophy is to collect only the minimum data necessary for service provision.

### Q: Can I use WulfVault without cookies?

**A:** No. The session cookie is essential for authentication. Without it, you cannot log in or access your files. This is a fundamental security requirement.

### Q: Do you sell my data to third parties?

**A:** No. We do not sell, share, or monetize user data. We do not use advertising, analytics, or tracking technologies.

### Q: Will you add more cookies in the future?

**A:** Our commitment is to minimize data collection. If we ever add new cookies (e.g., for opt-in analytics or user-requested features), we will:
1. Update this Cookie Policy
2. Notify users
3. Obtain consent if required by law
4. Provide opt-out mechanisms

### Q: How can I verify what cookies are set?

**A:** Use your browser's developer tools:
- **Chrome/Edge:** F12 → Application → Cookies
- **Firefox:** F12 → Storage → Cookies
- **Safari:** Develop → Show Web Inspector → Storage → Cookies

You should only see the `session` cookie from our domain after logging in.

### Q: Are cookies shared across subdomains?

**A:** [SPECIFY YOUR CONFIGURATION, e.g., "No, cookies are scoped to the exact domain you access" OR "Yes, cookies work across *.yourdomain.com"]

### Q: What happens to my session cookie when I log out?

**A:** The session is immediately invalidated in our database, and the cookie is deleted from your browser. Even if someone obtains the cookie value after logout, it cannot be used to access your account.

---

## Technical Details for Developers

### Cookie Implementation

**Backend:** WulfVault uses Go's `net/http` package with secure cookie settings:

```go
http.SetCookie(w, &http.Cookie{
    Name:     "session",
    Value:    sessionToken,
    MaxAge:   86400, // 24 hours
    Path:     "/",
    HttpOnly: true,
    Secure:   true, // HTTPS only
    SameSite: http.SameSiteLaxMode,
})
```

**Session Storage:** Server-side database (not client-side storage)

**Token Generation:** Cryptographically secure random tokens

**Validation:** Every request validates session token against database

---

## Compliance Certifications

**Standards We Follow:**
- ✅ GDPR (General Data Protection Regulation)
- ✅ ePrivacy Directive (2002/58/EC, amended 2009/136/EC)
- ✅ UK GDPR
- ✅ [ADD OTHER APPLICABLE STANDARDS, e.g., CCPA, LGPD]

**Audits:**
[OPTIONAL: "This Cookie Policy has been reviewed by [LEGAL FIRM] on [DATE]"]

---

## 📝 Customization Checklist

Before publishing this Cookie Policy:

- [ ] Replace all [PLACEHOLDER] text with your organization details
- [ ] Verify that you only use the session cookie (or add others to the table)
- [ ] If you add third-party services, document their cookies
- [ ] Specify your server location and jurisdiction
- [ ] Add your supervisory authority contact information
- [ ] Update the effective date
- [ ] Have legal counsel review (if required)
- [ ] Link from your cookie consent banner
- [ ] Link from website footer
- [ ] Make accessible at /cookie-policy or similar URL

**Need Help?** Consult with a data protection lawyer in your jurisdiction.

---

**END OF COOKIE POLICY TEMPLATE**

---

## Integration with WulfVault

### Where to Link This Policy

1. **Cookie Consent Banner** - "Learn more about cookies" link
2. **Website Footer** - "Cookie Policy" link
3. **Privacy Policy** - Reference and link to this Cookie Policy
4. **User Settings** - "Privacy & Cookies" section

### Banner Implementation

The cookie consent banner is implemented in WulfVault at:
- Template: `web/templates/cookie_consent_banner.html` (to be created)
- JavaScript: `web/static/js/cookie-consent.js` (to be created)

See implementation in Task #9 of the GDPR compliance project.

---

**Disclaimer:** This template does not constitute legal advice. Consult with qualified legal professionals for compliance verification in your jurisdiction.
