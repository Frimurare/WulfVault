# Email Sending Setup (Resend + Loopia DNS)

This guide explains how to connect your custom domain to **Resend** so your WulfVault instance can send emails (password resets, notifications, onboarding, etc.). The process only needs to be done once.

## 1. Create Your Resend Account
1. Go to **https://resend.com**
2. Create an account
3. Verify your personal email address
4. Go to the left menu → **Domains**
5. Click **Add Domain**
6. Enter your domain, e.g. `wulfvault.se`

You will now see a list of DNS records that need to be added to your DNS provider.

## 2. Locate the DNS Settings in Loopia
If you are using **Loopia**, follow these steps:

1. Log in to https://customerzone.loopia.se
2. Click **Domain names**
3. Select your domain
4. Click **DNS-editor**
5. Scroll down to see the list of DNS records
6. Use **Add record** to create each entry Resend requires

## 3. Add the Required DNS Records

### 3.1 DKIM (TXT Record)
```
Type: TXT
Host: resend._domainkey
Value: p=MIGfMA0GCSq... (long key)
TTL: 3600
```

### 3.2 SPF (TXT Record)
```
Type: TXT
Host: send
Value: v=spf1 include:amazonses.com ~all
TTL: 3600
```

### 3.3 MX (Mail Exchange)
```
Type: MX
Host: send
Value: feedback-smtp.eu-west-1.amazonses.com.
Priority: 10
TTL: 3600
```

### 3.4 (Optional) DMARC
```
Type: TXT
Host: _dmarc
Value: v=DMARC1; p=none;
TTL: 3600
```

## 4. Wait for DNS Propagation
DNS changes may take 15 minutes to several hours.  
Check propagation via https://www.whatsmydns.net

Search:
```
resend._domainkey.yourdomain.se (TXT)
```

Once most locations show the DKIM record, Resend will automatically mark the domain as Verified.

## 5. Sending Emails (Test Mode vs Production Mode)

### Before domain verification:
- You can only send emails to the email address used when creating the Resend account.
- You must use:
```
From: onboarding@resend.dev
```

### After verification:
- You can send to any recipient.
- You may use:
```
From: WulfVault <noreply@yourdomain.se>
```

## 6. Example PowerShell Email Test

```
$apiKey = "YOUR_API_KEY"

$body = @{
    from    = "WulfVault <noreply@yourdomain.se>"
    to      = "your-email@yourdomain.se"
    subject = "WulfVault Email Test"
    html    = "<p>This is a test email from WulfVault.</p>"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "https://api.resend.com/emails" `
  -Method Post `
  -Headers @{
      "Authorization" = "Bearer $apiKey";
      "Content-Type"  = "application/json"
  } `
  -Body $body
```

## 7. Using Resend in WulfVault
Set:
```
RESEND_API_KEY=your_sending_key
MAIL_FROM=noreply@yourdomain.se
```

## 8. Troubleshooting

### DKIM stuck on Pending
Loopia sometimes delays TXT propagation.  
If whatsmydns.net shows the correct key → Resend will verify soon.

### 403: Domain not verified
Use:
```
from: onboarding@resend.dev
```

### 403: Only allowed to send to your own email
You are still in test mode.

## Summary
1. Add domain in Resend  
2. Copy DNS records into Loopia  
3. Wait for propagation  
4. Use Resend API key in WulfVault  
