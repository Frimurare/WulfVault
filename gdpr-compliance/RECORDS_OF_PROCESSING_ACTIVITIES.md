# Records of Processing Activities (ROPA)

**⚠️ ACTION REQUIRED: Customize this template and maintain it as a living document.**

**Organization:** [YOUR ORGANIZATION NAME]
**Last Updated:** [DATE]
**Reviewed By:** [DPO or PRIVACY CONTACT NAME]
**Next Review Date:** [DATE]

---

## Purpose

This document fulfills the requirement under **GDPR Article 30** to maintain records of processing activities. It documents all personal data processing operations performed by WulfVault.

**Who Needs This:**
- ✅ **Required:** Organizations with 250+ employees
- ✅ **Required:** Smaller organizations if processing:
  - Is not occasional
  - Includes special category data (Article 9)
  - Involves high risk to rights and freedoms

**Best Practice:** Even if not legally required, maintaining ROPA demonstrates GDPR compliance commitment.

---

## Organization Information

**Data Controller:**
- **Legal Name:** [YOUR ORGANIZATION LEGAL NAME]
- **Registration Number:** [COMPANY REG NUMBER]
- **Address:** [REGISTERED ADDRESS]
- **Country:** [COUNTRY]
- **Contact Email:** [GENERAL CONTACT EMAIL]
- **Phone:** [PHONE NUMBER]
- **Website:** [WEBSITE URL]

**Data Protection Officer (if appointed):**
- **Name:** [DPO NAME]
- **Email:** [DPO EMAIL]
- **Phone:** [DPO PHONE]
- **Address:** [DPO ADDRESS]

**EU Representative (if outside EU and required by Art. 27):**
- **Name:** [REP NAME]
- **Email:** [REP EMAIL]
- **Address:** [REP EU ADDRESS]
- [ ] Not Applicable (organization is within EU)

---

## Processing Activity #1: User Account Management

### 1.1 Purpose of Processing

**Primary Purpose:**
Managing user accounts for file storage and sharing service

**Specific Purposes:**
- User authentication and authorization
- Account creation and maintenance
- Password management and 2FA
- User role and permission management

**Legal Basis (GDPR Article 6):**
- [X] Contractual necessity (Art. 6(1)(b)) - Providing file storage service
- [ ] Consent (Art. 6(1)(a))
- [ ] Legal obligation (Art. 6(1)(c))
- [X] Legitimate interest (Art. 6(1)(f)) - Security and fraud prevention
- [ ] Vital interests (Art. 6(1)(d))
- [ ] Public task (Art. 6(1)(e))

**Legitimate Interest Assessment (if applicable):**
We have a legitimate interest in maintaining secure authentication to prevent unauthorized access and protect user data. This interest is not overridden by data subject rights as authentication is essential for service security.

### 1.2 Categories of Data Subjects

- [X] Service users (system users with accounts)
- [X] Download account holders (external recipients)
- [X] Administrators
- [ ] Employees
- [ ] Customers of customers
- [ ] Other: _______________

### 1.3 Categories of Personal Data

| Data Category | Data Elements | Special Category (Art. 9)? |
|---------------|---------------|---------------------------|
| **Identification Data** | Name, email address, user ID | No |
| **Authentication Data** | Password hash (bcrypt), 2FA secret (encrypted), backup codes (hashed) | No |
| **Account Data** | Creation date, role (Admin/Manager/User), account status (active/inactive) | No |
| **Storage Quota Data** | Allocated storage, used storage | No |

**Special Category Data:** [ ] Yes [X] No

### 1.4 Categories of Recipients

**Internal Recipients:**
- System administrators (for account management and support)
- Technical support staff (when troubleshooting user issues)

**External Recipients:**
- [ ] None currently
- [ ] Third-party email service for transactional emails: [SPECIFY IF APPLICABLE]
- [ ] Cloud hosting provider: [SPECIFY IF APPLICABLE]

**International Transfers:** [ ] Yes [X] No (if No, data stays within [COUNTRY/REGION])

If Yes, safeguards:
- [ ] Standard Contractual Clauses (SCCs)
- [ ] Adequacy Decision
- [ ] Binding Corporate Rules (BCRs)
- [ ] Other: _______________

### 1.5 Retention Period

**Retention:** Until account deletion + 30 days for backup purge

**Justification:** Contractual necessity (service provision) and fraud prevention

**Deletion Method:** Soft deletion (email anonymized, account marked deleted) via `SoftDeleteUser` function. Permanent deletion after 30 days.

### 1.6 Technical and Organizational Measures

**Security Measures:**
- bcrypt password hashing (cost factor 12)
- AES-256-GCM encryption for 2FA secrets
- Session-based authentication (24-hour timeout)
- HTTPS/TLS for all communications (TLS 1.2+)
- Role-based access control (RBAC)
- Audit logging of all account actions

**Access Controls:**
- Only authorized administrators can access user account data
- Principle of least privilege applied
- 2FA required for administrator accounts

---

## Processing Activity #2: File Storage and Management

### 2.1 Purpose of Processing

**Primary Purpose:**
Storing, managing, and enabling sharing of user files

**Specific Purposes:**
- File upload and storage
- File download and access
- Share link creation and management
- File metadata management
- Trash/deleted files management

**Legal Basis (GDPR Article 6):**
- [X] Contractual necessity (Art. 6(1)(b)) - Core service provision
- [ ] Consent (Art. 6(1)(a))
- [ ] Legal obligation (Art. 6(1)(c))
- [X] Legitimate interest (Art. 6(1)(f)) - Security and abuse prevention
- [ ] Vital interests (Art. 6(1)(d))
- [ ] Public task (Art. 6(1)(e))

### 2.2 Categories of Data Subjects

- [X] Service users who upload files
- [X] Share link recipients who download files
- [ ] Other: _______________

### 2.3 Categories of Personal Data

| Data Category | Data Elements | Special Category (Art. 9)? |
|---------------|---------------|---------------------------|
| **File Metadata** | Filename, file size, MIME type, upload date, owner ID, file ID | No |
| **File Contents** | User-uploaded files (varies by user) | **⚠️ Users must not upload special category data unless necessary** |
| **Share Data** | Share link URL, expiry date, password (if set), access count | No |

**Special Category Data:** [ ] No [X] Potentially (user-controlled - users are responsible for not uploading prohibited data)

**Note:** Users are informed via Terms of Service not to upload special category data (health, biometric, etc.) unless they have a legal basis to do so.

### 2.4 Categories of Recipients

**Internal Recipients:**
- File owner (uploader)
- System administrators (for technical support only)

**External Recipients:**
- Share link recipients (when users explicitly create share links)
- [ ] Cloud storage provider: [SPECIFY IF USING EXTERNAL STORAGE]

**International Transfers:** [ ] Yes [X] No

If Yes, safeguards: _______________

### 2.5 Retention Period

**Retention:**
- Active files: Until deleted by user
- Deleted files: 5 days in trash, then permanently deleted
- Account deletion: All files deleted immediately

**Justification:** Service provision and user control

**Deletion Method:**
- Soft deletion: Moved to trash (5-day recovery period)
- Permanent deletion: File and metadata permanently removed from database and storage

### 2.6 Technical and Organizational Measures

**Security Measures:**
- Optional encryption at rest (SQLCipher for database)
- HTTPS/TLS for file transfers
- Access control (only file owner and share link holders can access)
- File upload restrictions (size limits, file type filtering)
- Share link password protection (optional)
- Share link expiry dates (configurable)
- Audit logging of file access

**Access Controls:**
- Users can only access their own files
- Share links grant temporary access
- Administrators can access files only for technical support with user permission

---

## Processing Activity #3: Audit Logging and Activity Tracking

### 3.1 Purpose of Processing

**Primary Purpose:**
Security monitoring, compliance, and audit trail maintenance

**Specific Purposes:**
- Detecting unauthorized access
- Investigating security incidents
- Compliance with legal obligations
- Providing data subject access to activity logs
- Fraud prevention

**Legal Basis (GDPR Article 6):**
- [ ] Contractual necessity (Art. 6(1)(b))
- [ ] Consent (Art. 6(1)(a))
- [X] Legal obligation (Art. 6(1)(c)) - Compliance with security requirements
- [X] Legitimate interest (Art. 6(1)(f)) - Security, fraud prevention, legal defense
- [ ] Vital interests (Art. 6(1)(d))
- [ ] Public task (Art. 6(1)(e))

**Legitimate Interest Assessment:**
We have a legitimate interest in maintaining audit logs for:
1. Detecting and preventing security breaches
2. Investigating suspicious activity
3. Legal defense in case of disputes
4. Compliance with industry best practices

This interest is not overridden by data subject rights as audit logging is essential for system security and integrity.

### 3.2 Categories of Data Subjects

- [X] Service users
- [X] Download account holders
- [X] Administrators

### 3.3 Categories of Personal Data

| Data Category | Data Elements | Special Category (Art. 9)? |
|---------------|---------------|---------------------------|
| **Identity Data** | User ID, user name, email | No |
| **Activity Data** | Action type (login, file upload, etc.), timestamp, IP address (optional) | No |
| **Session Data** | Session ID, user agent (browser) | No |
| **Target Data** | Resource accessed (file ID, settings changed) | No |

**Special Category Data:** [ ] Yes [X] No

**IP Address Logging:** [X] Optional [ ] Always Enabled [ ] Disabled
- If enabled: Legitimate interest (security and fraud detection)
- Users can request IP logging be disabled for their account

### 3.4 Categories of Recipients

**Internal Recipients:**
- System administrators
- Security team (for incident investigation)
- Data subjects themselves (can export their own audit log)

**External Recipients:**
- [ ] None
- [ ] Supervisory authority (only if required for breach investigation)

**International Transfers:** [ ] Yes [X] No

### 3.5 Retention Period

**Retention:** [90 DAYS / YOUR CONFIGURED PERIOD]

**Justification:**
- Security monitoring (detect anomalies over time)
- Compliance with audit requirements
- Legal defense (statute of limitations)

**Deletion Method:**
- Automated daily cleanup job (`internal/cleanup/cleanup.go`)
- Logs older than retention period permanently deleted

### 3.6 Technical and Organizational Measures

**Security Measures:**
- Audit logs stored in tamper-evident format
- Write-only access for log generation (users cannot modify logs)
- Database-level access controls
- Encryption in transit and at rest
- CSV export available for compliance analysis

**Access Controls:**
- Only administrators can view all audit logs
- Users can view their own audit logs
- Export functionality requires authentication

---

## Processing Activity #4: Download Tracking (for Share Links)

### 4.1 Purpose of Processing

**Primary Purpose:**
Tracking file downloads via share links for sender notification and abuse prevention

**Specific Purposes:**
- Notifying file owners of downloads
- Detecting abuse (excessive downloads, bots)
- Providing download statistics to file owners

**Legal Basis (GDPR Article 6):**
- [X] Contractual necessity (Art. 6(1)(b)) - Service feature
- [ ] Consent (Art. 6(1)(a))
- [ ] Legal obligation (Art. 6(1)(c))
- [X] Legitimate interest (Art. 6(1)(f)) - Abuse prevention
- [ ] Vital interests (Art. 6(1)(d))
- [ ] Public task (Art. 6(1)(e))

### 4.2 Categories of Data Subjects

- [X] Share link recipients (file downloaders)
- [X] File owners (who receive download notifications)

### 4.3 Categories of Personal Data

| Data Category | Data Elements | Special Category (Art. 9)? |
|---------------|---------------|---------------------------|
| **Identity Data** | Email address (optional - recipient may be unauthenticated) | No |
| **Download Data** | Download timestamp, file ID, share link used | No |
| **Network Data** | IP address (optional, configurable) | No |

**Special Category Data:** [ ] Yes [X] No

**Note:** Recipients do not need to provide personal data to download files via share links (unless required by file owner's settings).

### 4.4 Categories of Recipients

**Internal Recipients:**
- File owner (sees download count and optionally downloader info)
- System administrators (for abuse investigation)

**External Recipients:**
- [ ] None

**International Transfers:** [ ] Yes [X] No

### 4.5 Retention Period

**Retention:**
- Download logs: [90 DAYS / YOUR CONFIGURED PERIOD] (aligned with audit log retention)
- Download accounts: Until account deletion (GDPR self-service deletion available)

**Justification:** Contractual necessity and abuse prevention

**Deletion Method:**
- Automated cleanup via daily job
- Download account deletion anonymizes email address in logs

### 4.6 Technical and Organizational Measures

**Security Measures:**
- IP logging optional (disabled by default for privacy)
- Download logs stored securely
- Access restricted to file owner and administrators
- HTTPS for all download activity

**Access Controls:**
- File owners can only see downloads for their own files
- Download account holders can request account deletion anytime

---

## Processing Activity #5: Email Communications

### 5.1 Purpose of Processing

**Primary Purpose:**
Sending transactional emails required for service operation

**Specific Purposes:**
- Account registration confirmation
- Password reset requests
- Download notifications (to file owners)
- Account deletion confirmations
- Security alerts (suspicious login, etc.)

**Legal Basis (GDPR Article 6):**
- [X] Contractual necessity (Art. 6(1)(b)) - Essential service communications
- [ ] Consent (Art. 6(1)(a)) - Not required for transactional emails
- [X] Legal obligation (Art. 6(1)(c)) - Required notifications (e.g., breach notification)
- [X] Legitimate interest (Art. 6(1)(f)) - Security alerts
- [ ] Vital interests (Art. 6(1)(d))
- [ ] Public task (Art. 6(1)(e))

### 5.2 Categories of Data Subjects

- [X] Service users
- [X] Download account holders

### 5.3 Categories of Personal Data

| Data Category | Data Elements | Special Category (Art. 9)? |
|---------------|---------------|---------------------------|
| **Contact Data** | Email address, name | No |
| **Communication Data** | Email subject, body (containing service information only) | No |

**Special Category Data:** [ ] Yes [X] No

**Note:** We do NOT use email for marketing purposes (no newsletters, promotions, etc.).

### 5.4 Categories of Recipients

**Internal Recipients:**
- Email system (internal SMTP or external email service)

**External Recipients:**
- [ ] Email delivery service provider: [SPECIFY IF USING SENDGRID, AWS SES, etc.]
  - If yes, DPA in place: [ ] Yes [ ] No

**International Transfers:** [ ] Yes [X] No (if using EU-based email provider)

If Yes, safeguards: _______________

### 5.5 Retention Period

**Retention:**
- Sent emails: Not retained (transactional only)
- Email addresses: As long as user account exists

**Justification:** Transactional necessity

**Deletion Method:**
- Email addresses deleted upon account deletion

### 5.6 Technical and Organizational Measures

**Security Measures:**
- TLS for email transmission
- No sensitive data in email bodies (use secure links instead)
- Rate limiting to prevent spam
- Email templates reviewed for GDPR compliance

**Access Controls:**
- Only automated system sends emails
- Administrators cannot access sent email contents

---

## Processing Activity #6: User Support and Troubleshooting

### 6.1 Purpose of Processing

**Primary Purpose:**
Providing technical support to users

**Specific Purposes:**
- Responding to support requests
- Troubleshooting technical issues
- Account recovery assistance

**Legal Basis (GDPR Article 6):**
- [X] Contractual necessity (Art. 6(1)(b)) - Service support
- [ ] Consent (Art. 6(1)(a))
- [ ] Legal obligation (Art. 6(1)(c))
- [X] Legitimate interest (Art. 6(1)(f)) - Customer service
- [ ] Vital interests (Art. 6(1)(d))
- [ ] Public task (Art. 6(1)(e))

### 6.2 Categories of Data Subjects

- [X] Service users requesting support

### 6.3 Categories of Personal Data

| Data Category | Data Elements | Special Category (Art. 9)? |
|---------------|---------------|---------------------------|
| **Identity Data** | Name, email address, user ID | No |
| **Support Data** | Support request details, issue description, correspondence | No |
| **Account Data** | Account status, usage details (to diagnose issues) | No |

**Special Category Data:** [ ] Yes [X] No

**Note:** Users are instructed not to include special category data in support requests.

### 6.4 Categories of Recipients

**Internal Recipients:**
- Technical support staff
- System administrators (if escalated)

**External Recipients:**
- [ ] None (support handled internally)
- [ ] Support ticket system provider: [SPECIFY IF APPLICABLE]

**International Transfers:** [ ] Yes [X] No

### 6.5 Retention Period

**Retention:**
- Support tickets: [2 YEARS / YOUR POLICY] after case closure

**Justification:** Legal defense, quality assurance, training

**Deletion Method:**
- Automated deletion after retention period
- Support data deleted upon account deletion (if user requests)

### 6.6 Technical and Organizational Measures

**Security Measures:**
- Support ticket system access controls
- HTTPS for all communications
- Confidentiality agreements for support staff
- Secure communication channels (no sensitive data in plain email)

**Access Controls:**
- Only assigned support staff can access tickets
- Administrators have override access (audited)

---

## Data Flows Summary

### Data Flow Diagram

```
[User]
  ↓ (HTTPS)
[WulfVault Web Server]
  ↓
[Application Server]
  ↓
[SQLite Database] ← [Audit Logs]
  ↓
[File Storage]

External Data Flows (if applicable):
[WulfVault] → [Email Service] (transactional emails only)
[WulfVault] → [Cloud Hosting Provider] (all data)
```

### Third-Party Processors (Sub-Processors)

| Processor | Service | Data Processed | Location | DPA Signed |
|-----------|---------|----------------|----------|------------|
| [EXAMPLE: AWS] | [Cloud hosting] | [All data] | [EU-Frankfurt] | [X] Yes [ ] No |
| [EXAMPLE: SendGrid] | [Email delivery] | [Email addresses, names] | [US] | [X] Yes [ ] No |
| [ADD OTHERS] | | | | |

**Note:** Update this table whenever new third-party services are added.

---

## Data Subject Rights Implementation

| Right (GDPR Article) | Implementation Method | Response Time |
|---------------------|----------------------|---------------|
| **Right of Access (Art. 15)** | `/api/v1/user/export-data` endpoint, audit log CSV export | Immediate (automated) |
| **Right to Rectification (Art. 16)** | User settings page (profile update, password change) | Immediate (self-service) |
| **Right to Erasure (Art. 17)** | `/settings/delete-account` page, soft deletion with anonymization | Immediate (automated) |
| **Right to Restrict Processing (Art. 18)** | Contact DPO, manual account suspension | 30 days |
| **Right to Data Portability (Art. 20)** | JSON export (same as access right) | Immediate (automated) |
| **Right to Object (Art. 21)** | Contact DPO, disable optional processing (e.g., IP logging) | 30 days |
| **Right Not to Be Subject to Automated Decision-Making (Art. 22)** | N/A (no automated decision-making) | N/A |

---

## Data Protection Impact Assessment (DPIA)

**DPIA Required?** [ ] Yes [X] No

**Assessment:**
WulfVault does NOT require a DPIA under GDPR Article 35 because it does NOT involve:
- Large-scale systematic monitoring of publicly accessible areas
- Large-scale processing of special category data (Art. 9) or criminal data (Art. 10)
- Automated decision-making with legal or similarly significant effects

**However,** if you:
- Deploy WulfVault to process special category data (health, biometric, etc.)
- Process data of children at scale
- Add automated decision-making or profiling features

Then you MUST conduct a DPIA.

**Last DPIA Review:** [DATE] - N/A

---

## Changes to Processing Activities

**Change Log:**

| Date | Processing Activity | Change Description | Updated By |
|------|--------------------|--------------------|------------|
| [DATE] | Initial ROPA | Created initial records | [NAME] |
| | | | |
| | | | |

**Review Frequency:** Annually or whenever processing activities change significantly

**Next Review Date:** [DATE]

---

## Compliance Attestation

I attest that the information in this Records of Processing Activities document is accurate and complete to the best of my knowledge as of [DATE].

**Name:** [DPO or RESPONSIBLE PERSON NAME]
**Title:** [TITLE]
**Signature:** _______________
**Date:** [DATE]

---

## 📝 Customization Checklist

Before finalizing this document:

- [ ] Replace all [PLACEHOLDER] text
- [ ] Add any additional processing activities specific to your deployment
- [ ] Document all third-party sub-processors
- [ ] Specify data retention periods per your policy
- [ ] Complete data flow diagram with your architecture
- [ ] Review with legal counsel or DPO
- [ ] Update whenever processing activities change
- [ ] Review annually at minimum
- [ ] Store securely for audit purposes

**Need Help?** Consult with a data protection lawyer or certified DPO.

---

**END OF RECORDS OF PROCESSING ACTIVITIES**

**Disclaimer:** This template does not constitute legal advice. Organizations should consult with qualified legal professionals for compliance verification.
