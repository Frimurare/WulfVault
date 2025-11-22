# Email Domain Verification with Loopia DNS

This guide explains how to configure DNS records in Loopia for email domain verification with WulfVault's supported email providers (Resend, SendGrid, Mailgun, SMTP).

**Primary focus:** Resend.com (recommended email provider for WulfVault)

**See also:** [resend_loopia_setup.md](/resend_loopia_setup.md) for Resend-specific quick start guide.

---

## Why Domain Verification?

Modern email providers require domain verification before you can send emails from your custom domain. This prevents spam and ensures good email deliverability.

### Before Domain Verification (Test Mode):
- ❌ You can only send emails to your own email address
- ❌ Emails may be marked as spam
- ❌ Must use generic addresses like `onboarding@resend.dev`
- ❌ Cannot send to customers/users

### After Domain Verification (Production Mode):
- ✅ Send to anyone
- ✅ Better inbox placement and deliverability
- ✅ Professional "From" addresses (e.g., `noreply@wulfvault.se`)
- ✅ Full control over email branding

---

## Step-by-Step: Verifying a Domain in Loopia

### 1. Add Domain to Email Provider

First, add your domain in your email provider's control panel:

**Resend:**
- Go to https://resend.com/domains
- Click "Add Domain"
- Enter your domain (e.g., `wulfvault.se`)
- Copy the DNS records shown

**SendGrid / Mailgun:**
- Similar process in their respective control panels
- Each provider will give you DNS records to add

### 2. Understanding Loopia DNS Editor

**Important Loopia Behavior:**
- When you enter `resend._domainkey` in a field, Loopia automatically appends `.wulfvault.se`
- So you only type the subdomain part, NOT the full domain
- **Example:** Type `send`, Loopia creates `send.wulfvault.se`

**Accessing DNS Editor:**
1. Log in to Loopia
2. Go to "Domäner" (Domains)
3. Click on your domain (e.g., `wulfvault.se`)
4. Click "DNS-editor"

**DO NOT touch "Namnservrar" (Name servers)!** Leave them as `ns1.loopia.se` and `ns2.loopia.se`.

### 3. Common DNS Records for Email Verification

Most email providers require these 4 types of records:

#### DKIM (Domain Keys Identified Mail)
- **Type:** TXT
- **Name:** `resend._domainkey` (provider-specific prefix)
- **Value:** Long string starting with `p=MIGfMA0GCS...`
- **Purpose:** Cryptographic signature to prevent email spoofing

#### SPF MX (Sender Policy Framework - MX record)
- **Type:** MX
- **Name:** `send` (or provider-specific subdomain)
- **Priority:** 10
- **Value:** Provider's SMTP server (e.g., `feedback-smtp.eu-west-1.amazonses.com`)
- **Purpose:** Authorizes provider to send email on your behalf

#### SPF TXT (Sender Policy Framework - TXT record)
- **Type:** TXT
- **Name:** `send` (same as MX record)
- **Value:** `v=spf1 include:amazonses.com ~all` (provider-specific)
- **Purpose:** Lists authorized email senders for your domain

#### DMARC (Domain-based Message Authentication)
- **Type:** TXT
- **Name:** `_dmarc`
- **Value:** `v=DMARC1; p=none;`
- **Purpose:** Tells receiving servers what to do with unauthenticated emails

---

## Example: Verifying wulfvault.se with Resend (EU Region)

### DNS Records from Resend

When you add `wulfvault.se` to Resend (EU region), you get:

```
DKIM:
Type: TXT
Name: resend._domainkey
Value: p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDDt5GxNurNlVaJeS6NbIKRXUKA50xO2RXF8smEGoFChHtV+urdwXNK0aI6HbJFsugI1efSBn5oNt4Ze4LhPJO7ejZ7wTnRW7pnB1T9SUX2T/37X43NKQxHDvEPJpBkKSFoNo6LudYQDiU76XoY93x37rBPYxisyoSpni1pw7C/WwIDAQAB

SPF (MX):
Type: MX
Name: send
Priority: 10
Value: feedback-smtp.eu-west-1.amazonses.com

SPF (TXT):
Type: TXT
Name: send
Value: v=spf1 include:amazonses.com ~all

DMARC:
Type: TXT
Name: _dmarc
Value: v=DMARC1; p=none;
```

### Adding Records in Loopia

**Step 1: Add DKIM Record**
1. In DNS-editor, click "Lägg till record"
2. Fill in:
   - **Namn/Host:** `resend._domainkey`
   - **Typ:** TXT
   - **Data:** `p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDDt5GxNurNlVaJeS6NbIKRXUKA50xO2RXF8smEGoFChHtV+urdwXNK0aI6HbJFsugI1efSBn5oNt4Ze4LhPJO7ejZ7wTnRW7pnB1T9SUX2T/37X43NKQxHDvEPJpBkKSFoNo6LudYQDiU76XoY93x37rBPYxisyoSpni1pw7C/WwIDAQAB`
3. Click "Lägg till" (Add)

**Step 2: Add SPF MX Record**
1. Click "Lägg till record" again
2. Fill in:
   - **Namn/Host:** `send`
   - **Typ:** MX
   - **Prio:** 10
   - **Data:** `feedback-smtp.eu-west-1.amazonses.com`
3. Click "Lägg till"

**Step 3: Add SPF TXT Record**
1. Click "Lägg till record" again
2. Fill in:
   - **Namn/Host:** `send`
   - **Typ:** TXT
   - **Data:** `v=spf1 include:amazonses.com ~all`
3. Click "Lägg till"

**Step 4: Add DMARC Record**
1. Click "Lägg till record" again
2. Fill in:
   - **Namn/Host:** `_dmarc`
   - **Typ:** TXT
   - **Data:** `v=DMARC1; p=none;`
3. Click "Lägg till"

### What You'll See After Adding

Your DNS editor should show NEW sections like this:

```
resend._domainkey
  TXT  36000  p=MIGfMA0GCS... [long DKIM key]

send
  MX   36000  10  feedback-smtp.eu-west-1.amazonses.com
  TXT  36000  v=spf1 include:amazonses.com ~all

_dmarc
  TXT  36000  v=DMARC1; p=none;
```

**DO NOT add these records under `*`, `@`, or `www` sections!** They should be their own separate sections.

---

## Common Mistakes

### ❌ Mistake 1: Adding to Wrong Section
**Wrong:**
```
www
  TXT  p=MIGfMA0GCS... [DKIM key added here]
```

**Right:**
```
resend._domainkey
  TXT  p=MIGfMA0GCS... [DKIM key in its own section]
```

### ❌ Mistake 2: Including Full Domain Name
**Wrong:** Type `resend._domainkey.wulfvault.se`

**Right:** Type `resend._domainkey` (Loopia adds .wulfvault.se automatically)

### ❌ Mistake 3: Changing Name Servers
**Wrong:** Changing DNS servers to Resend/SendGrid nameservers

**Right:** Keep Loopia nameservers (`ns1.loopia.se`, `ns2.loopia.se`), only add DNS records

### ❌ Mistake 4: Not Waiting for Propagation
**Wrong:** Click "Verify" immediately after adding records

**Right:** Wait 10-30 minutes for DNS propagation before verifying

---

## Verification Process

### After Adding DNS Records

1. **Wait 10-30 minutes** for DNS propagation
2. **Check DNS propagation** (optional):
   ```bash
   dig TXT resend._domainkey.wulfvault.se
   dig MX send.wulfvault.se
   dig TXT send.wulfvault.se
   dig TXT _dmarc.wulfvault.se
   ```
3. **Verify in provider:**
   - Resend: Click "Verify" button next to domain
   - SendGrid/Mailgun: Similar verify button in control panel
4. **Status changes:**
   - Before: "Looking for DNS records..." (⏳ Pending)
   - After: "Verified" (✅ Active)

### If Verification Fails

**Check these:**
1. Did you add all 4 records?
2. Are they in their OWN sections (not under `*`, `@`, `www`)?
3. Did you wait at least 10-30 minutes?
4. Are the values EXACTLY as provided (no extra spaces)?
5. Is the TTL value reasonable (36000 is fine, default)?

**Debug with dig:**
```bash
# Should return your DKIM key
dig TXT resend._domainkey.wulfvault.se +short

# Should return MX record
dig MX send.wulfvault.se +short

# Should return SPF policy
dig TXT send.wulfvault.se +short

# Should return DMARC policy
dig TXT _dmarc.wulfvault.se +short
```

---

## Region-Specific Differences

### Resend EU (eu-west-1)
- MX: `feedback-smtp.eu-west-1.amazonses.com`
- Used for: European domains, GDPR compliance

### Resend US (us-east-1)
- MX: `feedback-smtp.us-east-1.amazonses.com`
- Used for: US-based domains, faster US delivery

**Make sure to use the records YOUR provider gives you!** Don't copy-paste from this guide verbatim.

---

## After Successful Verification

### Update WulfVault Configuration

1. **Go to:** Admin → Email Settings
2. **Select:** Resend tab
3. **Update "From Email":** Change from `onboarding@resend.dev` to `noreply@wulfvault.se`
4. **Click:** Save Configuration
5. **Click:** Test Email
6. **Result:** Email should now work! ✅

### Professional Email Addresses

Common formats:
- `noreply@yourdomain.com` - For automated emails
- `files@yourdomain.com` - For file sharing system
- `support@yourdomain.com` - For support emails
- `notifications@yourdomain.com` - For notification emails

---

## Troubleshooting

### "Domain not verified" error after 30 minutes

**Possible causes:**
1. DNS records not added correctly
2. TTL too high (change to 3600 or lower)
3. Loopia caching issue (flush cache in Loopia panel)
4. Wrong region (US vs EU records)

**Solution:**
- Verify with `dig` commands
- Check Loopia DNS editor shows records correctly
- Wait another 30 minutes
- Contact Loopia support if records don't propagate

### "Can only send to your own email" error

**Cause:** Domain not verified yet OR using wrong From address

**Solution:**
1. Verify domain is verified in provider control panel
2. Update From Email to use verified domain (`noreply@wulfvault.se`, not `noreply@gmail.com`)
3. Test again

### Emails going to spam

**Possible causes:**
1. Missing SPF/DKIM/DMARC records
2. Wrong DMARC policy (`p=reject` too strict)
3. Low sender reputation (new domain)

**Solutions:**
- Verify all 4 DNS records are present
- Use `p=none` for DMARC initially
- Gradually increase sending volume to build reputation
- Add proper email content (not just links)

---

## Related Documentation

- [WulfVault Email Configuration](../README.md#email--notifications)
- [Resend Documentation](https://resend.com/docs)
- [SendGrid DNS Records](https://docs.sendgrid.com/ui/account-and-settings/how-to-set-up-domain-authentication)
- [Mailgun DNS Verification](https://documentation.mailgun.com/en/latest/user_manual.html#verifying-your-domain)

---

## Success Story: wulfvault.se Verification

**Date:** 2025-11-22

**Setup:**
- Domain: `wulfvault.se`
- Registrar: Loopia
- Email Provider: Resend (EU region)
- From Address: `noreply@wulfvault.se`

**Timeline:**
1. **12:35** - Added domain to Resend
2. **12:40** - Added 4 DNS records in Loopia (initial confusion with sections)
3. **12:42** - Deleted wrong records (under *, @, www)
4. **12:45** - Added correct records in proper sections
5. **12:55** - DNS propagated (10 minutes)
6. **12:56** - Clicked "Verify" in Resend → SUCCESS ✅
7. **12:58** - Updated WulfVault to use `noreply@wulfvault.se`
8. **12:59** - Sent test email → WORKED! 🎉

**Key Learnings:**
- Don't add records under existing sections (`*`, `@`, `www`)
- Each DNS record needs its OWN section
- 10-30 minute wait is necessary
- Use exact values from provider (EU vs US region matters!)

---

**Last Updated:** 2025-11-22
**Maintained by:** WulfVault Team
**License:** AGPL-3.0
