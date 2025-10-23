# TLS Certificates Directory

This directory contains TLS/SSL certificates and keys for HTTPS.

## ⚠️ Security Notice

- **Private keys are NOT committed to git** - they are in `.gitignore`
- Each deployment generates its own certificates
- For development: Self-signed certs are auto-generated on first run
- For production: Use CA-signed certificates from your provider

## Development Setup

Certificates are automatically generated on first run:

```bash
go run .
# Creates: server.crt, server.key
```

## Production Setup

Replace auto-generated certs with proper certificates:

1. Obtain certificate from Certificate Authority (Let's Encrypt, etc.)
2. Place `server.crt` in this directory
3. Place `server.key` in this directory (keep secure, .gitignore protected)
4. Set proper file permissions: `chmod 600 server.key`
5. Restart the application

## File Permissions

Ensure private key has restricted permissions:

```bash
chmod 600 server.key
chmod 644 server.crt
```

## Certificate Details

- **Format**: PEM (Privacy Enhanced Mail)
- **Key Type**: RSA 2048-bit (dev), RSA 2048-bit+ (prod)
- **Validity**: 365 days (dev), 1+ years (prod)
- **Algorithm**: TLS 1.2+

## Testing Self-Signed Certificates

Accept self-signed certs in curl:

```bash
curl -k https://localhost:8443/health
# or
curl --insecure https://localhost:8443/health
```

In PowerShell:

```powershell
$PSDefaultParameterValues['Invoke-WebRequest:SkipCertificateCheck'] = $true
Invoke-WebRequest https://localhost:8443/health
```

---

**Last Updated**: October 23, 2025
