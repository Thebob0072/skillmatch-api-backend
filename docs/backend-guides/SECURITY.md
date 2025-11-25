# 🔐 SkillMatch API - Security Documentation

## 🛡️ ระบบความปลอดภัยหลายชั้น (Multi-Layer Security)

### 1. 🔑 Authentication & Authorization

#### JWT Token Security
- **Algorithm**: HMAC SHA-256 (HS256)
- **Token Expiration**: 7 วัน (configurable)
- **Secret Key**: เก็บใน environment variables
- **Token Storage**: Client ต้องเก็บใน httpOnly cookie หรือ secure storage
- **Refresh Token**: ควรมีระบบ refresh token แยก (TODO)

```go
// Current Implementation
- JWT signed with HS256
- userID และ exp (expiration) claims
- Middleware validates token on every protected route
```

#### Password Security
- **Hashing**: bcrypt (cost factor 10)
- **Salt**: bcrypt generates unique salt per password
- **Plain text passwords**: ไม่เก็บในระบบ
- **Password Policy**: ควรบังคับ:
  - ความยาวขั้นต่ำ 8 ตัวอักษร
  - มีตัวพิมพ์ใหญ่ พิมพ์เล็ก ตัวเลข และอักขระพิเศษ

#### Role-Based Access Control (RBAC)
```
- User Roles: Client, Provider, Admin
- Tier-Based Access: Basic, Premium, VIP, Professional, GOD
- Verification Status: unverified, pending, verified, rejected
```

---

### 2. 🚫 Input Validation & Sanitization

#### SQL Injection Prevention
- ✅ **Parameterized Queries**: ใช้ `$1, $2, $3` placeholders
- ✅ **No String Concatenation**: ไม่ต่อ string โดยตรง
- ✅ **pgx Library**: ใช้ prepared statements อัตโนมัติ

```go
// ✅ SAFE
dbPool.QueryRow(ctx, "SELECT * FROM users WHERE user_id = $1", userID)

// ❌ UNSAFE (ห้ามใช้)
query := fmt.Sprintf("SELECT * FROM users WHERE user_id = %d", userID)
```

#### XSS Prevention
- Frontend ต้อง escape HTML output
- ใช้ React/Vue auto-escaping
- Content-Type headers ถูกต้อง
- CSP Headers (Content Security Policy)

#### CORS Configuration
```go
// ควร whitelist specific origins
AllowOrigins: []string{
    "http://localhost:5173",
    "https://yourdomain.com"
}
AllowCredentials: true
AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
```

---

### 3. 🔒 Data Protection

#### Sensitive Data Handling
- **Password Hash**: ไม่ส่งกลับใน API response (`json:"-"`)
- **Google ID**: ซ่อนจาก response (`json:"-"`)
- **KYC Documents**: เข้าถึงได้เฉพาะ Admin
- **Personal Info**: เก็บแยกใน `user_profiles`

#### Database Security
- **Connection String**: เก็บใน environment variables
- **SSL/TLS**: ควรเปิดใช้ connection encryption
- **Least Privilege**: แต่ละ service ใช้ DB user แยก
- **Backup**: ควร backup database สม่ำเสมอ

#### File Upload Security
- **GCS Signed URLs**: จำกัด expiry time (15 นาที)
- **File Type Validation**: ตรวจสอบ MIME type
- **File Size Limit**: จำกัดขนาดไฟล์
- **Virus Scanning**: ควรใช้ antivirus scan (TODO)

```go
// Current: Signed URLs with 15 min expiry
expires := time.Now().Add(15 * time.Minute)
```

---

### 4. 🌐 Network Security

#### HTTPS/TLS
- **Production**: ต้องใช้ HTTPS เท่านั้น
- **Certificate**: Let's Encrypt หรือ trusted CA
- **TLS Version**: ≥ 1.2
- **HSTS Header**: บังคับใช้ HTTPS

#### Rate Limiting
```go
// ควรเพิ่ม rate limiter middleware
// ป้องกัน brute force และ DDoS
- Login endpoint: 5 requests/minute
- API endpoints: 100 requests/minute
- File upload: 10 requests/hour
```

#### IP Whitelisting
- Admin endpoints ควร whitelist IP
- Production database ควร whitelist application servers

---

### 5. 🔍 Monitoring & Logging

#### Security Logging
```go
// ควร log:
- Failed login attempts
- Permission denied errors
- Suspicious activities
- KYC approval/rejection
- Admin actions
- Database errors
```

#### Audit Trail
```sql
-- ควรมีตาราง audit_logs
CREATE TABLE audit_logs (
    log_id SERIAL PRIMARY KEY,
    user_id INT,
    action VARCHAR(50),
    resource VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

### 6. 🛠️ API Security Best Practices

#### Request Validation
```go
// ✅ ใช้ binding validation
type CreateBookingRequest struct {
    ProviderID  int    `json:"provider_id" binding:"required"`
    PackageID   int    `json:"package_id" binding:"required"`
    BookingDate string `json:"booking_date" binding:"required"`
    StartTime   string `json:"start_time" binding:"required"`
}
```

#### Error Messages
```go
// ❌ ห้าม expose internal details
c.JSON(500, gin.H{"error": err.Error()}) // ❌

// ✅ ใช้ generic message
c.JSON(500, gin.H{"error": "Internal server error"}) // ✅
c.JSON(404, gin.H{"error": "Resource not found"}) // ✅
```

#### HTTP Headers
```go
// Security Headers ที่ควรมี:
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Strict-Transport-Security: max-age=31536000
- Content-Security-Policy: default-src 'self'
```

---

### 7. 🎭 KYC Verification Security

#### Document Verification
- ✅ **3-Document Check**: บัตร + สุขภาพ + Selfie
- ✅ **Manual Review**: Admin ตรวจทุกคำขอ
- ✅ **Face Matching**: เปรียบเทียบใบหน้า
- ✅ **Age Verification**: ≥ 20 ปี
- ✅ **Expiry Check**: บัตรไม่หมดอายุ

#### Privacy Protection
- KYC files เข้าถึงผ่าน signed URLs (10 นาที)
- เฉพาะ Admin ดูเอกสารได้
- Files ไม่ public accessible
- ควรเข้ารหัสไฟล์ at rest (GCS encryption)

---

### 8. 💳 Payment Security

#### Stripe Integration
- ✅ **No Card Storage**: Stripe handles card data
- ✅ **PCI Compliance**: Stripe is PCI-DSS certified
- ✅ **Webhook Verification**: Verify stripe signature
- ✅ **Idempotency**: ป้องกัน duplicate charges

```go
// Webhook signature verification
stripe.VerifySignature(payload, signature, webhookSecret)
```

---

### 9. 🚨 Security Vulnerabilities to Prevent

#### Common Attacks

**1. SQL Injection** ✅ PROTECTED
- ใช้ parameterized queries ทั้งหมด

**2. XSS (Cross-Site Scripting)** ⚠️ FRONTEND RESPONSIBILITY
- Frontend ต้อง escape output
- ใช้ React/Vue default escaping

**3. CSRF (Cross-Site Request Forgery)** ⚠️ TODO
- ควรใช้ CSRF tokens
- SameSite cookie attribute

**4. Brute Force** ⚠️ TODO
- ควรเพิ่ม rate limiting
- Account lockout after N failed attempts

**5. DDoS** ⚠️ INFRASTRUCTURE
- ใช้ CloudFlare, AWS Shield
- Rate limiting at API Gateway

**6. Session Hijacking** ✅ PARTIAL
- JWT expiration
- ควรใช้ refresh tokens
- Logout invalidation (TODO)

**7. Man-in-the-Middle** ✅ HTTPS REQUIRED
- ใช้ HTTPS/TLS เท่านั้น
- Certificate pinning (mobile apps)

**8. File Upload Attacks** ⚠️ PARTIAL
- ตรวจสอบ file type
- ควรใช้ virus scanner
- File size limits

**9. API Abuse** ⚠️ TODO
- Rate limiting
- API key management
- Request throttling

**10. Privilege Escalation** ✅ PROTECTED
- Role-based checks
- Tier-based access control

---

### 10. 🔐 Environment Variables Security

```bash
# ❌ ห้าม hardcode ในโค้ด
# ✅ ใช้ .env file (ต้องอยู่ใน .gitignore)

# Required Environment Variables:
DATABASE_URL=postgresql://user:pass@host:5432/dbname
JWT_SECRET=your-super-secret-key-at-least-32-chars
REDIS_URL=redis://localhost:6379
STRIPE_SECRET_KEY=sk_test_xxxxx
GCS_BUCKET_NAME=your-bucket
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

#### .env.example (สำหรับ reference)
```bash
DATABASE_URL=postgresql://localhost:5432/skillmatch
JWT_SECRET=change-me-in-production
REDIS_URL=redis://localhost:6379
STRIPE_SECRET_KEY=sk_test_changeme
GCS_BUCKET_NAME=your-bucket-name
```

---

### 11. 🔄 Secure Development Lifecycle

#### Code Review Checklist
- [ ] ใช้ parameterized queries
- [ ] Validate all user inputs
- [ ] Check authorization on protected routes
- [ ] No sensitive data in logs
- [ ] No hardcoded secrets
- [ ] Error messages don't expose internals
- [ ] Rate limiting implemented
- [ ] HTTPS enforced

#### Dependency Management
```bash
# ตรวจสอบ vulnerabilities
go list -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

#### Security Testing
- **Penetration Testing**: ควรทำ pen test ก่อน production
- **Vulnerability Scanning**: ใช้เครื่องมือ automated scan
- **Code Analysis**: ใช้ static code analyzer (gosec)

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...
```

---

### 12. 📊 Security Metrics to Monitor

```
- Failed login attempts (per IP)
- Unusual API usage patterns
- Multiple account creation from same IP
- Rapid file uploads
- Admin action frequency
- Database query performance
- Error rate spikes
- Unauthorized access attempts
```

---

### 13. 🚀 Production Security Checklist

#### Pre-Deployment
- [ ] HTTPS/TLS enabled
- [ ] Environment variables set
- [ ] Database backup configured
- [ ] Rate limiting enabled
- [ ] Monitoring/alerting setup
- [ ] Error tracking (Sentry, etc.)
- [ ] Security headers configured
- [ ] CORS properly configured
- [ ] Admin accounts secured (2FA)
- [ ] Audit logging enabled

#### Post-Deployment
- [ ] Security scan completed
- [ ] Penetration test passed
- [ ] Backup restore tested
- [ ] Incident response plan ready
- [ ] Security team contacts defined
- [ ] Compliance requirements met

---

### 14. 🆘 Incident Response Plan

#### If Security Breach Detected:

1. **Immediate Actions**
   - Isolate affected systems
   - Revoke compromised credentials
   - Enable maintenance mode

2. **Investigation**
   - Review audit logs
   - Identify attack vector
   - Assess damage scope

3. **Remediation**
   - Patch vulnerability
   - Reset affected passwords
   - Notify affected users (if required)

4. **Post-Incident**
   - Document lessons learned
   - Update security measures
   - Conduct security review

---

### 15. 📞 Security Contacts

```
Security Team: security@yourdomain.com
Bug Bounty: (if applicable)
Emergency Hotline: (for critical issues)
```

---

### 16. 🔄 Regular Security Tasks

#### Daily
- Monitor error logs
- Check failed login attempts
- Review unusual activities

#### Weekly
- Review new user registrations
- Check KYC pending queue
- Scan for vulnerabilities

#### Monthly
- Update dependencies
- Review access permissions
- Security training for team
- Backup integrity check

#### Quarterly
- Full security audit
- Penetration testing
- Policy review and update
- Disaster recovery drill

---

## 🎓 Security Training Resources

### For Developers
- OWASP Top 10: https://owasp.org/www-project-top-ten/
- Go Security Cheat Sheet: https://cheatsheetseries.owasp.org/
- JWT Best Practices: https://tools.ietf.org/html/rfc8725

### For Admins
- KYC Verification Guidelines
- Phishing Detection
- Social Engineering Awareness

---

## ⚠️ Known Limitations & TODO

### Current Security Gaps (Priority Order):

1. **HIGH PRIORITY**
   - [ ] Add rate limiting middleware
   - [ ] Implement refresh token system
   - [ ] Add CSRF protection
   - [ ] Enable security headers middleware
   - [ ] Add audit logging table

2. **MEDIUM PRIORITY**
   - [ ] Implement account lockout (brute force)
   - [ ] Add virus scanner for uploads
   - [ ] Implement IP-based restrictions for admin
   - [ ] Add 2FA for admin accounts
   - [ ] Password complexity requirements

3. **LOW PRIORITY**
   - [ ] Automated security scanning in CI/CD
   - [ ] Bug bounty program
   - [ ] Compliance certifications
   - [ ] Advanced threat detection

---

## 📝 Compliance Notes

### GDPR (EU)
- User consent for data processing
- Right to be forgotten (delete account)
- Data portability
- Privacy policy required

### PDPA (Thailand)
- User consent required
- Data retention policy
- Security measures documented
- Privacy policy in Thai

### Data Retention
- KYC documents: 5 years (legal requirement)
- User data: Until account deletion
- Logs: 90 days
- Backups: 30 days

---

## 🔒 Encryption at Rest & Transit

### In Transit
- ✅ HTTPS/TLS for all API calls
- ✅ Encrypted connection to database (TLS)
- ✅ Encrypted connection to Redis (TLS)
- ✅ GCS uses HTTPS

### At Rest
- ⚠️ Database encryption (should enable)
- ✅ GCS default encryption enabled
- ⚠️ Redis data encryption (optional)
- ⚠️ Backup encryption (should enable)

---

## 📖 Additional Resources

- **Go Security**: https://go.dev/doc/security
- **Stripe Security**: https://stripe.com/docs/security
- **Google Cloud Security**: https://cloud.google.com/security
- **PostgreSQL Security**: https://www.postgresql.org/docs/current/security.html

---

## 🎯 Summary

SkillMatch API ใช้แนวทางความปลอดภัยหลายชั้น:

1. ✅ **Authentication**: JWT + bcrypt passwords
2. ✅ **Authorization**: Role-based + Tier-based
3. ✅ **Input Validation**: Parameterized queries + binding
4. ✅ **KYC Verification**: 3-document + manual review
5. ✅ **Data Protection**: Sensitive data hidden
6. ⚠️ **Rate Limiting**: TODO (critical)
7. ⚠️ **Security Headers**: TODO (important)
8. ✅ **HTTPS**: Required in production
9. ⚠️ **Audit Logging**: TODO (important)
10. ✅ **Payment Security**: Stripe handles sensitive data

**Overall Security Status**: 🟡 **Good, with room for improvement**

**Next Steps**: ลำดับความสำคัญคือ Rate Limiting → Security Headers → Audit Logging
