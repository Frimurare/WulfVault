# Privacy Policy

**⚠️ ACTION REQUIRED: This is a template. Replace all [PLACEHOLDERS] with your information before publishing.**

**Last Updated:** [DATE]
**Effective Date:** [DATE]

---

## 1. Introduction

[YOUR ORGANIZATION NAME] ("we," "us," or "our") operates [YOUR WULFVAULT INSTANCE URL] (the "Service"). This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our file storage and sharing service powered by WulfVault.

**By using the Service, you agree to the collection and use of information in accordance with this Privacy Policy.**

If you do not agree with this Privacy Policy, please do not access or use the Service.

---

## 2. Information We Collect

### 2.1 Personal Information You Provide

We collect information that you voluntarily provide when using the Service:

| Data Type | Purpose | Legal Basis |
|-----------|---------|-------------|
| **Name** | User identification and account management | Contractual necessity (GDPR Art. 6(1)(b)) |
| **Email Address** | Authentication, notifications, password recovery | Contractual necessity (GDPR Art. 6(1)(b)) |
| **Password** | Account security (stored as bcrypt hash, never plaintext) | Contractual necessity (GDPR Art. 6(1)(b)) |
| **Two-Factor Authentication Data** | Enhanced account security (optional, encrypted) | Contractual necessity (GDPR Art. 6(1)(b)) |

### 2.2 Files and Content

We store:
- **Files you upload** to the Service
- **Metadata** associated with files (filename, size, upload date, MIME type)
- **Share links** you create and their settings
- **File access history** (who accessed what and when)

**Legal Basis:** Contractual necessity (GDPR Art. 6(1)(b)) - storage is the core purpose of the Service.

### 2.3 Automatically Collected Information

We may automatically collect:

| Data Type | Collected | Purpose | Legal Basis |
|-----------|-----------|---------|-------------|
| **IP Address** | [YES/NO - Optional, configurable] | Security, abuse prevention | Legitimate interest (GDPR Art. 6(1)(f)) |
| **Browser Type** | No | N/A | N/A |
| **Device Information** | No | N/A | N/A |
| **Usage Analytics** | No (no tracking by default) | N/A | N/A |
| **Cookies** | Yes (session only) | Authentication | Strictly necessary |

**Note:** WulfVault is privacy-focused and does NOT use tracking cookies, analytics, or advertising technologies.

### 2.4 Activity Logs (Audit Trail)

We maintain an **audit log** of user actions for security and compliance:

- Login/logout events
- File uploads, downloads, deletions
- Share link creation and access
- User management actions (admin only)
- Settings changes
- Account deletion requests

**Retention Period:** [90 DAYS / YOUR CONFIGURED PERIOD]
**Legal Basis:** Legitimate interest (GDPR Art. 6(1)(f)) - security, fraud prevention, compliance
**Your Rights:** You can export your audit log at any time via Settings → Export My Data

---

## 3. How We Use Your Information

We use your information for the following purposes:

| Purpose | Legal Basis |
|---------|-------------|
| **Account Management** - Creating and maintaining your account | Contractual necessity (Art. 6(1)(b)) |
| **Service Provision** - Storing, managing, and sharing your files | Contractual necessity (Art. 6(1)(b)) |
| **Security** - Detecting and preventing unauthorized access | Legitimate interest (Art. 6(1)(f)) |
| **Compliance** - Maintaining audit logs for legal requirements | Legal obligation (Art. 6(1)(c)) |
| **Communication** - Sending service notifications and security alerts | Contractual necessity (Art. 6(1)(b)) |
| **Support** - Responding to your inquiries and technical issues | Contractual necessity (Art. 6(1)(b)) |

**We do NOT:**
- Sell your personal data to third parties
- Use your data for advertising or marketing (unless you opt-in)
- Analyze your file contents for profiling or behavioral tracking
- Share your data with third parties except as described in Section 4

---

## 4. Information Sharing and Disclosure

### 4.1 When We Share Your Information

We may share your information only in these limited circumstances:

**With Your Consent:**
- When you explicitly authorize us to share your data

**Service Providers:**
- [LIST ANY THIRD-PARTY SERVICES USED, e.g., cloud hosting, email delivery]
- These providers are bound by Data Processing Agreements (DPAs) and process data only on our instructions

**Legal Requirements:**
- To comply with legal obligations, court orders, or government requests
- To protect our rights, property, or safety, or that of our users

**Business Transfers:**
- In the event of a merger, acquisition, or sale, your data may be transferred to the acquiring entity (you will be notified in advance)

### 4.2 When You Share Files

When you create a **share link**:
- Anyone with the link can access the file (if link is not password-protected)
- Access is logged in audit trail (if enabled)
- You are responsible for controlling who receives the link

**Important:** Do not share links containing sensitive data via insecure channels.

---

## 5. Data Storage and Security

### 5.1 Where We Store Your Data

Your data is stored on servers located in:

**[SPECIFY YOUR SERVER LOCATION, e.g., "EU - Frankfurt, Germany" or "US - Oregon, USA"]**

**Data Transfers:**
- [IF DATA LEAVES EU]: We use Standard Contractual Clauses (SCCs) approved by the European Commission for data transfers outside the EU/EEA.
- [IF DATA STAYS IN EU]: Your data does not leave the EU/EEA.

### 5.2 How We Protect Your Data

We implement industry-standard security measures:

| Security Measure | Implementation |
|------------------|----------------|
| **Encryption in Transit** | TLS 1.2+ (HTTPS) for all connections |
| **Encryption at Rest** | [ENABLED/OPTIONAL - SQLCipher for database] |
| **Password Storage** | bcrypt hashing (cost factor 12, never plaintext) |
| **Two-Factor Authentication** | TOTP-based 2FA with encrypted secret storage |
| **Access Control** | Role-based permissions (Admin, Manager, User) |
| **Session Security** | Secure, HttpOnly, SameSite cookies with 24-hour expiry |
| **Audit Logging** | Comprehensive activity tracking with tamper-evident logs |

**Security Incident Response:**
See our [Breach Notification Procedure](BREACH_NOTIFICATION_PROCEDURE.md) for how we handle data breaches.

---

## 6. Data Retention

We retain your personal data only as long as necessary for the purposes outlined in this Privacy Policy:

| Data Type | Retention Period | Reason |
|-----------|------------------|--------|
| **Account Information** | Until account deletion + 30 days | Contractual necessity, fraud prevention |
| **Uploaded Files** | Until you delete them or account is deleted | Service provision |
| **Audit Logs** | [90 days / YOUR CONFIGURED PERIOD] | Security, compliance, legal defense |
| **Deleted Accounts** | Email anonymized, audit trail preserved | GDPR compliance (soft deletion) |
| **Backup Data** | [SPECIFY BACKUP RETENTION, e.g., 30 days] | Disaster recovery |

**Automated Deletion:**
Audit logs older than the retention period are automatically purged daily at [TIME, e.g., 3:00 AM UTC].

**Account Deletion:**
When you delete your account:
1. Your email is anonymized (replaced with `deleted-user-[ID]@deleted.local`)
2. Your files are permanently deleted
3. Your audit trail is preserved (anonymized) for compliance
4. Your account cannot be recovered after 30 days

---

## 7. Your Rights Under GDPR

As a data subject in the EU/EEA, you have the following rights:

### 7.1 Right of Access (Art. 15)
**What:** Obtain a copy of your personal data
**How:** Settings → Export My Data → Download JSON file
**Response Time:** Immediate (automated export)

### 7.2 Right to Rectification (Art. 16)
**What:** Correct inaccurate personal data
**How:** Settings → Profile → Update your name, email, or password
**Response Time:** Immediate

### 7.3 Right to Erasure / "Right to be Forgotten" (Art. 17)
**What:** Request deletion of your personal data
**How:** Settings → Delete My Account → Confirm deletion
**Response Time:** Immediate (soft deletion with email anonymization)
**Note:** Audit logs are preserved (anonymized) for legal compliance

### 7.4 Right to Data Portability (Art. 20)
**What:** Receive your data in a machine-readable format
**How:** Settings → Export My Data → Download JSON/CSV
**Format:** JSON (structured data) and CSV (audit logs)
**Response Time:** Immediate

### 7.5 Right to Restrict Processing (Art. 18)
**What:** Limit how we process your data
**How:** Contact us at [YOUR PRIVACY CONTACT EMAIL]
**Response Time:** 30 days

### 7.6 Right to Object (Art. 21)
**What:** Object to processing based on legitimate interests
**How:** Contact us at [YOUR PRIVACY CONTACT EMAIL]
**Response Time:** 30 days

### 7.7 Right to Withdraw Consent (Art. 7(3))
**What:** Withdraw consent for optional data processing
**How:** Contact us or disable optional features
**Note:** Does not affect lawfulness of prior processing

### 7.8 Right to Lodge a Complaint
**What:** File a complaint with a supervisory authority
**Where:** [YOUR LOCAL DATA PROTECTION AUTHORITY]
**EU Authorities:** https://edpb.europa.eu/about-edpb/about-edpb/members_en

---

## 8. Cookies and Tracking

### 8.1 Cookies We Use

| Cookie Name | Purpose | Type | Duration | Consent Required? |
|-------------|---------|------|----------|-------------------|
| `session` | User authentication and session management | Essential | 24 hours | No (strictly necessary) |

### 8.2 Cookie Settings

You can control cookies through your browser settings:
- **Block all cookies:** The Service will not function (authentication requires session cookie)
- **Delete cookies:** You will be logged out

**We do NOT use:**
- Analytics cookies (Google Analytics, etc.)
- Advertising cookies
- Social media tracking pixels
- Third-party tracking technologies

For more information, see our [Cookie Policy](COOKIE_POLICY_TEMPLATE.md).

---

## 9. Children's Privacy

The Service is **not intended for children under 16** (or the minimum age in your jurisdiction).

We do not knowingly collect personal data from children. If you believe we have collected information from a child, please contact us immediately at [YOUR PRIVACY CONTACT EMAIL], and we will delete it promptly.

**Parental Consent:** If you are under 18, please obtain parental consent before using the Service.

---

## 10. International Data Transfers

**For EU/EEA Users:**

If your data is transferred outside the EU/EEA, we ensure adequate protection through:
- **Standard Contractual Clauses (SCCs)** approved by the European Commission
- **Adequacy Decisions** for countries with equivalent data protection laws
- **Your explicit consent** where applicable

**Current Data Locations:**
[SPECIFY: e.g., "All data is stored within the EU" OR "Data is stored in US with SCC protections"]

---

## 11. Data Protection Officer

[CHOOSE ONE:]

**Option A - DPO Appointed:**
We have appointed a Data Protection Officer (DPO) to oversee GDPR compliance:

**Name:** [DPO NAME]
**Email:** [DPO EMAIL]
**Address:** [DPO POSTAL ADDRESS]

**Option B - No DPO Required:**
Our organization is not required to appoint a Data Protection Officer under GDPR Art. 37. For privacy inquiries, contact:

**Privacy Contact:** [PRIVACY CONTACT NAME]
**Email:** [PRIVACY CONTACT EMAIL]
**Address:** [POSTAL ADDRESS]

---

## 12. Changes to This Privacy Policy

We may update this Privacy Policy from time to time. Changes will be communicated by:
- Posting the updated policy on this page
- Updating the "Last Updated" date at the top
- [OPTIONAL: Sending email notifications for material changes]

**Your continued use of the Service after changes constitutes acceptance of the updated Privacy Policy.**

**Version History:**
- v1.0 - [DATE] - Initial policy

---

## 13. Contact Us

If you have questions or concerns about this Privacy Policy or our data practices:

**[YOUR ORGANIZATION NAME]**
**Email:** [PRIVACY CONTACT EMAIL]
**Address:** [POSTAL ADDRESS]
**Phone:** [PHONE NUMBER - Optional]

**Response Time:** We aim to respond to all privacy inquiries within 30 days (GDPR requirement).

---

## 14. Legal Information

**Data Controller:**
[YOUR ORGANIZATION LEGAL NAME]
[REGISTRATION NUMBER, if applicable]
[REGISTERED ADDRESS]

**Jurisdiction:**
This Privacy Policy is governed by [YOUR JURISDICTION] law and GDPR for EU/EEA users.

**Supervisory Authority:**
[YOUR LOCAL DATA PROTECTION AUTHORITY NAME AND WEBSITE]

---

## Appendix: Technical Details

### Data Export Format

When you export your data, you receive a JSON file containing:

```json
{
  "user": {
    "id": "...",
    "name": "...",
    "email": "...",
    "role": "...",
    "created_at": "...",
    "quota_bytes": "...",
    "used_bytes": "..."
  },
  "files": [
    {
      "filename": "...",
      "size": "...",
      "uploaded_at": "...",
      "share_links": [...]
    }
  ],
  "audit_logs": [
    {
      "action": "...",
      "timestamp": "...",
      "ip_address": "..."
    }
  ]
}
```

### Account Deletion Process

Technical implementation of GDPR-compliant account deletion:

1. **Soft Deletion:** Account marked as deleted (`deleted_at` timestamp set)
2. **Email Anonymization:** `deleted-user-[ID]@deleted.local`
3. **File Deletion:** All uploaded files permanently removed
4. **Share Link Invalidation:** All share links disabled
5. **Audit Trail Preservation:** Activity logs preserved (anonymized) for legal compliance
6. **Irreversible After:** 30 days (permanent purge from backups)

---

**END OF PRIVACY POLICY TEMPLATE**

---

## 📝 Customization Checklist

Before publishing, ensure you have:

- [ ] Replaced all [PLACEHOLDER] text with your organization details
- [ ] Specified data storage locations (servers, jurisdiction)
- [ ] Configured audit log retention period
- [ ] Listed any third-party service providers
- [ ] Appointed DPO or designated privacy contact
- [ ] Specified your supervisory authority
- [ ] Updated effective date
- [ ] Had legal counsel review the policy
- [ ] Published at accessible URL (e.g., /privacy-policy)
- [ ] Linked from footer, registration page, and cookie banner

**Need help?** Consult with a qualified data protection lawyer in your jurisdiction.
