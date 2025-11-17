# Data Processing Agreement (DPA)

**⚠️ ACTION REQUIRED: This is a template. Customize for your organization and have legal counsel review before use.**

**Agreement Date:** [DATE]
**DPA Version:** 1.0

---

## Purpose of This Document

This Data Processing Agreement ("DPA") is entered into between the **Data Controller** (your customer) and the **Data Processor** (you, operating WulfVault) to comply with **GDPR Article 28** requirements.

**When to Use This DPA:**
- You provide WulfVault as a service to business customers (B2B SaaS)
- Your customers use WulfVault to process their end-users' personal data
- You act as a "data processor" on behalf of your customers (the "data controllers")

**Not Needed If:**
- You only use WulfVault internally for your own organization
- You are the data controller for all data stored in WulfVault

---

## Parties to This Agreement

### Data Controller (Customer)

**Legal Name:** [CUSTOMER ORGANIZATION NAME]
**Address:** [CUSTOMER ADDRESS]
**Contact Person:** [CUSTOMER CONTACT NAME]
**Email:** [CUSTOMER CONTACT EMAIL]
**Registration Number:** [CUSTOMER REG NUMBER]

**Role:** Data Controller - Determines purposes and means of processing personal data

---

### Data Processor (Service Provider)

**Legal Name:** [YOUR ORGANIZATION NAME]
**Address:** [YOUR ADDRESS]
**Contact Person:** [YOUR CONTACT NAME]
**Email:** [YOUR CONTACT EMAIL]
**Registration Number:** [YOUR REG NUMBER]
**DPO (if applicable):** [DPO NAME AND EMAIL]

**Role:** Data Processor - Processes personal data on behalf of the Data Controller

---

## 1. Definitions

Terms used in this DPA have the meanings defined in the GDPR:

- **"GDPR"** means Regulation (EU) 2016/679 of the European Parliament and of the Council
- **"Personal Data"** means any information relating to an identified or identifiable natural person
- **"Processing"** means any operation performed on personal data (collection, storage, use, disclosure, deletion, etc.)
- **"Data Subject"** means the individual whose personal data is processed
- **"Supervisory Authority"** means an independent public authority established by an EU Member State
- **"Data Breach"** means a breach of security leading to accidental or unlawful destruction, loss, alteration, unauthorized disclosure of, or access to, personal data
- **"Sub-processor"** means any third-party processor engaged by the Processor

---

## 2. Scope and Application

### 2.1 Scope of Processing

This DPA applies to all Processing of Personal Data by the Processor on behalf of the Controller in connection with the provision of file storage and sharing services via WulfVault.

**Services Covered:**
- File storage and management
- File sharing via secure links
- User account management
- Audit logging and activity tracking
- [ADD OTHER SERVICES]

### 2.2 Duration

This DPA remains in effect for the duration of the Service Agreement between the parties, and survives termination until all Personal Data has been deleted or returned.

---

## 3. Nature and Purpose of Processing

### 3.1 Categories of Data Subjects

Personal Data processed under this DPA relates to the following categories of Data Subjects:

- [ ] **Employees** of the Controller
- [ ] **Customers** of the Controller
- [ ] **Business partners** of the Controller
- [ ] **End-users** of Controller's services
- [ ] **Other:** [SPECIFY]

### 3.2 Categories of Personal Data

The following types of Personal Data may be processed:

| Category | Data Types | Special Category? |
|----------|------------|-------------------|
| **Identification Data** | Name, email address, user ID | No |
| **Authentication Data** | Password hash, 2FA secrets | No |
| **Activity Data** | Login timestamps, file access logs, IP addresses (optional) | No |
| **File Metadata** | Filenames, file sizes, upload dates | No |
| **File Contents** | [SPECIFY - depends on what Controller uploads] | [YES/NO] |

**Special Categories of Personal Data (Art. 9 GDPR):**
- [ ] **NOT PERMITTED** - Controller must not upload special category data
- [ ] **PERMITTED** with explicit consent and additional safeguards (specify below)

If special category data is permitted, specify:
- Types: [e.g., health data, biometric data]
- Legal basis: [e.g., explicit consent, substantial public interest]
- Additional safeguards: [e.g., encryption, access controls]

### 3.3 Purpose of Processing

The Processor shall process Personal Data solely for the following purposes:

1. **Primary Purpose:** Providing file storage and sharing services as specified in the Service Agreement
2. **Security:** Detecting and preventing unauthorized access, fraud, and abuse
3. **Compliance:** Maintaining audit logs for legal and regulatory compliance
4. **Support:** Providing technical support and troubleshooting services

**Processor shall NOT:**
- Process Personal Data for any other purpose without prior written authorization from the Controller
- Use Personal Data for own business purposes (marketing, analytics, product improvement)
- Disclose Personal Data to third parties except as permitted in this DPA

---

## 4. Processor's Obligations (GDPR Article 28(3))

### 4.1 Processing Instructions

The Processor shall:
- Process Personal Data **only** on documented instructions from the Controller
- Inform the Controller immediately if instructions violate GDPR or other data protection laws
- Not transfer Personal Data outside the EU/EEA without Controller's prior written authorization

**Documented Instructions:**
Instructions are provided through:
- This DPA and its appendices
- The Service Agreement
- Written instructions via [SPECIFY METHOD, e.g., support tickets, email to [EMAIL]]

### 4.2 Confidentiality

The Processor shall ensure that:
- All personnel authorized to process Personal Data are bound by confidentiality obligations
- Access to Personal Data is limited to personnel who need it to perform their duties
- Confidentiality obligations survive termination of employment or contracts

### 4.3 Security Measures (Article 32)

The Processor implements the following technical and organizational measures:

#### Technical Measures

| Measure | Implementation |
|---------|----------------|
| **Encryption in Transit** | TLS 1.2+ for all connections |
| **Encryption at Rest** | [OPTIONAL: SQLCipher for database encryption] |
| **Password Security** | bcrypt hashing (cost factor 12) |
| **Two-Factor Authentication** | TOTP-based 2FA (optional for users) |
| **Access Control** | Role-based access control (Admin, Manager, User) |
| **Session Security** | Secure, HttpOnly, SameSite cookies with timeout |
| **Audit Logging** | Comprehensive activity tracking |

#### Organizational Measures

| Measure | Description |
|---------|-------------|
| **Access Control Policy** | Written policy restricting access to authorized personnel |
| **Background Checks** | [YES/NO] - Employee screening procedures |
| **Security Training** | Annual GDPR and security awareness training for staff |
| **Incident Response Plan** | Documented procedure for handling data breaches |
| **Business Continuity** | [DESCRIBE BACKUP AND DISASTER RECOVERY] |
| **Regular Audits** | [FREQUENCY, e.g., Annual security assessments] |

### 4.4 Sub-processors (Article 28(2) & 28(4))

**General Authorization:**
The Controller provides general authorization for the Processor to engage Sub-processors, subject to the following conditions:

1. **Prior Notice:** Processor must notify Controller of any new Sub-processor at least [30 DAYS] in advance
2. **Objection Right:** Controller may object to new Sub-processor on reasonable grounds within [14 DAYS]
3. **Same Obligations:** Sub-processors must be bound by same data protection obligations as in this DPA
4. **Processor Liability:** Processor remains fully liable for Sub-processor's compliance

**Current Sub-processors:**

| Sub-processor | Service | Location | Data Processed |
|---------------|---------|----------|----------------|
| [EXAMPLE: AWS] | [Cloud hosting] | [EU-Frankfurt] | [All data] |
| [EXAMPLE: SendGrid] | [Email delivery] | [US] | [Email addresses] |
| [ADD OTHERS] | | | |

**Sub-processor List:** Maintained at [URL] and updated when changes occur.

### 4.5 Data Subject Rights (Articles 15-22)

The Processor shall:

**Assist the Controller** in responding to Data Subject requests:
- Right of access (Article 15) - Via `/api/v1/user/export-data` endpoint
- Right to rectification (Article 16) - Via user settings interface
- Right to erasure (Article 17) - Via `/api/v1/gdpr/delete-account` endpoint
- Right to data portability (Article 20) - JSON/CSV export available
- Right to restrict processing (Article 18) - Manual intervention by Processor
- Right to object (Article 21) - Manual intervention by Processor

**Response Time:** Within [10 BUSINESS DAYS] of receiving Controller's request

**Charges:**
- First request per Data Subject per year: No charge
- Additional requests: [SPECIFY FEE SCHEDULE, if applicable]

### 4.6 Data Breach Notification (Article 33)

In the event of a Personal Data breach, the Processor shall:

1. **Notify Controller** without undue delay and **within 24 hours** of becoming aware
2. **Provide information** including:
   - Nature of the breach (categories and approximate number of Data Subjects affected)
   - Likely consequences of the breach
   - Measures taken or proposed to address the breach
   - Contact point for more information
3. **Cooperate** with Controller's breach investigation
4. **Document** the breach in accordance with Article 33(5)

**Notification Method:** [EMAIL TO SPECIFIED ADDRESS / SECURITY HOTLINE / SUPPORT TICKET]

**Controller's Responsibility:**
The Controller is responsible for notifying Supervisory Authority and affected Data Subjects as required by Articles 33 and 34.

### 4.7 Audit Rights (Article 28(3)(h))

The Controller has the right to:

**Audit the Processor's compliance** with this DPA, including:
- Requesting evidence of security measures
- Requesting copies of certifications (ISO 27001, SOC 2, etc.)
- Conducting on-site inspections (with reasonable notice)

**Audit Frequency:** [SPECIFY, e.g., "Maximum once per year unless breach occurs"]
**Notice Period:** [30 DAYS] written notice required
**Processor's Costs:** [FREE FOR FIRST AUDIT PER YEAR / CONTROLLER BEARS COSTS / OTHER]

**Self-Certification:**
The Processor shall provide annual self-certification of compliance with this DPA.

---

## 5. Data Transfers Outside EU/EEA

### 5.1 Transfer Mechanism

If Personal Data is transferred outside the EU/EEA, the Processor shall ensure adequate protection through:

- [ ] **Standard Contractual Clauses (SCCs)** - Appendix A
- [ ] **Adequacy Decision** - [SPECIFY COUNTRY]
- [ ] **Binding Corporate Rules (BCRs)**
- [ ] **Explicit Consent** from Data Subjects (Controller's responsibility)

**Current Data Location:**
[SPECIFY: e.g., "All data stored within EU - Frankfurt, Germany"]

### 5.2 Supplementary Measures

Where required, the Processor implements supplementary measures to ensure data protection equivalent to EU standards:
- Encryption of data in transit and at rest
- Strict access controls
- Legal guarantees regarding government access requests
- Transparency about legal obligations

---

## 6. Data Retention and Deletion

### 6.1 Retention Periods

| Data Category | Retention Period | Legal Basis |
|---------------|------------------|-------------|
| **User Account Data** | Until account deletion | Service provision |
| **Uploaded Files** | Until deleted by user or account deletion | Service provision |
| **Audit Logs** | [90 DAYS / CONTROLLER-SPECIFIED] | Compliance, security |
| **Backup Data** | [30 DAYS] after source deletion | Disaster recovery |

### 6.2 Data Deletion Upon Termination

Upon termination of the Service Agreement, the Processor shall:

**Option 1 - Controller Chooses:**
- [ ] **Delete** all Personal Data within [30 DAYS]
- [ ] **Return** all Personal Data to Controller in [JSON/CSV] format

**Option 2 - Standard Procedure:**
At Controller's choice, either delete or return Personal Data within 30 days.

**Exceptions:**
Processor may retain Personal Data to the extent required by applicable law, provided it remains confidential and is not further processed.

**Certification of Deletion:**
Upon request, Processor shall provide written certification that data has been deleted.

---

## 7. Controller's Obligations

The Controller shall:

1. **Provide clear instructions** for Processing Personal Data
2. **Ensure legal basis** exists for all Processing activities
3. **Inform Data Subjects** about the Processing (transparency obligation)
4. **Respond to Data Subject requests** (Processor assists but Controller is responsible)
5. **Notify Processor** of any restrictions or special requirements for Processing
6. **Conduct Data Protection Impact Assessment (DPIA)** if required by Article 35

---

## 8. Liability and Indemnification

### 8.1 Liability

**GDPR Article 82 Liability:**
- Each party shall be liable for damages caused by Processing that violates GDPR
- Processor is liable only if it has not complied with obligations specifically directed to processors or has acted outside or contrary to lawful instructions
- Controller and Processor may be held jointly and severally liable for damages

### 8.2 Indemnification

Each party agrees to indemnify the other against:
- Fines or penalties imposed by Supervisory Authorities for breach of this DPA
- Claims by Data Subjects for GDPR violations caused by the indemnifying party
- Reasonable legal costs incurred in defending such claims

**Limitation:**
Subject to limitations in the main Service Agreement (if applicable).

---

## 9. Term and Termination

### 9.1 Term

This DPA commences on [START DATE] and continues until:
- Termination of the Service Agreement, OR
- All Personal Data has been deleted or returned

### 9.2 Termination for Breach

Either party may terminate this DPA if:
- The other party materially breaches this DPA and fails to cure within [30 DAYS]
- Continued Processing would violate applicable data protection laws

### 9.3 Effects of Termination

Upon termination:
- Processor ceases all Processing except as necessary for deletion/return
- Processor deletes or returns Personal Data per Section 6.2
- This DPA's confidentiality and liability provisions survive

---

## 10. Governing Law and Disputes

**Governing Law:**
This DPA is governed by the laws of [YOUR JURISDICTION] and GDPR.

**Dispute Resolution:**
Disputes shall be resolved through:
1. Good faith negotiations between the parties
2. [MEDIATION / ARBITRATION] if negotiations fail
3. Courts of [JURISDICTION] as final resort

**Supervisory Authority Jurisdiction:**
Notwithstanding the above, Data Subjects retain the right to lodge complaints with their Supervisory Authority.

---

## 11. Amendment and Modification

This DPA may only be amended:
- By written agreement signed by both parties
- As necessary to comply with changes in data protection laws (Processor shall notify Controller)

**Precedence:**
In case of conflict between this DPA and the Service Agreement, this DPA prevails on matters related to data protection.

---

## 12. Contact Information

### Controller Contact

**Data Protection Inquiries:**
Email: [CONTROLLER EMAIL]
Address: [CONTROLLER ADDRESS]

### Processor Contact

**Data Protection Officer (if applicable):**
Name: [DPO NAME]
Email: [DPO EMAIL]
Address: [DPO ADDRESS]

**Security Incident Reporting:**
Email: [SECURITY EMAIL]
Phone: [24/7 HOTLINE, if applicable]

---

## Signatures

### Data Controller (Customer)

**Signed by:** _______________________________
**Name:** [PRINT NAME]
**Title:** [TITLE]
**Date:** [DATE]

### Data Processor (Service Provider)

**Signed by:** _______________________________
**Name:** [PRINT NAME]
**Title:** [TITLE]
**Date:** [DATE]

---

## Appendix A: Standard Contractual Clauses (SCCs)

**Note:** If data transfers outside EU/EEA occur, attach the European Commission's Standard Contractual Clauses:

**For Controller-to-Processor Transfers:**
Use the SCCs adopted by Commission Implementing Decision (EU) 2021/914 of 4 June 2021.

**Download from:**
https://ec.europa.eu/info/law/law-topic/data-protection/international-dimension-data-protection/standard-contractual-clauses-scc_en

**Modules to Use:**
- **Module Two:** Controller-to-Processor transfers

**Completion Instructions:**
- Annex I: Complete with information from Sections 3 and 5 of this DPA
- Annex II: Complete with security measures from Section 4.3
- Annex III: Complete with Sub-processor list from Section 4.4

---

## Appendix B: Technical and Organizational Measures

### Detailed Security Documentation

**1. Access Control (Prevent Unauthorized Access)**

| Measure | Description |
|---------|-------------|
| Physical Access | [DESCRIBE: e.g., "Servers hosted in ISO 27001-certified data centers with 24/7 surveillance"] |
| Logical Access | Role-based access control (RBAC) with least privilege principle |
| Authentication | Password policy (min. 12 characters) + optional 2FA |
| Authorization | Three user roles: User, Manager, Admin with granular permissions |

**2. Transmission Control (Secure Data Transfers)**

| Measure | Description |
|---------|-------------|
| Encryption in Transit | TLS 1.2+ (AES-256-GCM cipher suite) |
| API Security | Session-based authentication with CSRF protection |
| File Transfers | All uploads/downloads encrypted via HTTPS |

**3. Input Control (Audit Trail)**

| Measure | Description |
|---------|-------------|
| Audit Logging | All actions logged (login, file access, user management, settings) |
| Log Retention | [90 days / CONFIGURABLE] |
| Log Protection | Logs stored in tamper-evident format, write-only access |
| Log Export | CSV export available for compliance analysis |

**4. Availability Control (Business Continuity)**

| Measure | Description |
|---------|-------------|
| Backups | [DESCRIBE: e.g., "Daily automated backups, 30-day retention"] |
| Disaster Recovery | [DESCRIBE: e.g., "RPO: 24 hours, RTO: 4 hours"] |
| Redundancy | [DESCRIBE: e.g., "Multi-zone deployment with automatic failover"] |

**5. Separation Control (Multi-tenancy, if applicable)**

| Measure | Description |
|---------|-------------|
| Data Isolation | [DESCRIBE: e.g., "Each organization uses separate database instance"] |
| Logical Separation | User data isolated via database-level access controls |

**6. Pseudonymization and Encryption**

| Measure | Description |
|---------|-------------|
| Password Storage | bcrypt hashing (cost factor 12), irreversible |
| 2FA Secrets | AES-256-GCM encryption with key rotation |
| Deleted Accounts | Email pseudonymization: `deleted-user-[ID]@deleted.local` |
| Optional Encryption | [IF ENABLED: SQLCipher for database encryption at rest] |

---

## Appendix C: Sub-processor List

**Last Updated:** [DATE]

| Sub-processor | Service | Data Processed | Location | Safeguards |
|---------------|---------|----------------|----------|------------|
| [EXAMPLE: Amazon Web Services (AWS)] | Cloud hosting | All data | EU (Frankfurt) | DPA in place, ISO 27001, SOC 2 |
| [EXAMPLE: SendGrid / Twilio] | Email delivery | Email addresses, transactional emails | US | Standard Contractual Clauses |
| [ADD OTHERS] | | | | |

**Change Notification:**
Controller will be notified of Sub-processor changes via email to [CONTROLLER EMAIL] at least 30 days in advance.

---

## 📝 Customization Checklist

Before using this DPA:

- [ ] Replace all [PLACEHOLDER] text with actual information
- [ ] Complete Appendix A if international transfers occur
- [ ] Complete Appendix B with detailed security measures
- [ ] Complete Appendix C with current Sub-processor list
- [ ] Have legal counsel review for your jurisdiction
- [ ] Ensure consistency with your Service Agreement
- [ ] Both parties sign and date the agreement
- [ ] Provide copy to Controller and retain your copy
- [ ] Review annually for updates to GDPR or other laws

**Need Help?** Consult with a qualified data protection lawyer experienced in GDPR Article 28 compliance.

---

**END OF DATA PROCESSING AGREEMENT TEMPLATE**
