# 🚀 SkillMatch API - Production Deployment Guide

## 📋 Pre-Deployment Checklist

### ⚠️ สิ่งที่ต้องทำก่อน Deploy

#### 1. Environment Variables
```bash
# แก้ไขค่าใน .env หรือสร้างจาก template
cp .env.production .env

# ค่าที่ต้องตั้ง:
# - JWT_SECRET_KEY (ใช้: openssl rand -base64 64)
# - DB_PASSWORD
# - GOOGLE_CLIENT_ID & GOOGLE_CLIENT_SECRET
# - STRIPE_SECRET_KEY & STRIPE_WEBHOOK_SECRET
```

#### 2. SSL Certificates
```bash
# ติดตั้ง certbot
sudo apt install certbot

# สร้าง SSL certificate
sudo certbot certonly --standalone -d your-domain.com

# Copy certificates
sudo mkdir -p ./ssl
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem ./ssl/
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem ./ssl/
```

#### 3. Update nginx.conf
```bash
# แก้ไข domain ใน nginx.conf
sed -i 's/your-domain.com/actual-domain.com/g' nginx.conf
```

#### 4. Database Backup
```bash
# ตั้งค่า cron job สำหรับ backup
crontab -e

# เพิ่ม:
0 2 * * * /path/to/skillmatch-api/backup.sh >> /var/log/backup.log 2>&1
```

---

## 🐳 Docker Deployment

### Option 1: Manual Deployment

```bash
# Build และเริ่ม services
docker-compose -f docker-compose.prod.yml up -d

# ดู logs
docker-compose -f docker-compose.prod.yml logs -f

# ตรวจสอบสถานะ
curl http://localhost:8080/ping
```

### Option 2: Automated Deployment (GitHub Actions)

1. **Setup GitHub Secrets:**
   - `DOCKER_USERNAME` / `DOCKER_PASSWORD`
   - `DEPLOY_HOST` / `DEPLOY_USER` / `DEPLOY_SSH_KEY`

2. **Push to main:**
   ```bash
   git push origin main
   ```

3. **GitHub Actions จะ:**
   - Run tests
   - Build Docker image
   - Deploy to server
   - Health check

---

## 🔧 Production Configuration

### Database Connection Pool
```go
// config ใน database.go
MaxConns = 100
MinConns = 10
MaxConnLifetime = 1 hour
MaxConnIdleTime = 30 minutes
```

### Nginx Rate Limiting
```nginx
API: 60 requests/minute
Auth: 10 requests/minute
```

### Security Headers
- Strict-Transport-Security
- X-Frame-Options
- X-Content-Type-Options
- X-XSS-Protection
- Content-Security-Policy

---

## 📊 Monitoring & Maintenance

### Health Check
```bash
# API
curl http://localhost:8080/ping

# Database
docker exec -it postgres-db psql -U admin -d skillmatch_db -c "SELECT NOW();"

# Redis
docker exec -it redis-cache redis-cli ping
```

### View Logs
```bash
# ดู logs ทั้งหมด
docker-compose -f docker-compose.prod.yml logs -f

# ดูแค่ API
docker-compose -f docker-compose.prod.yml logs -f api
```

### Database Management
```bash
# Backup
./backup.sh

# Restore
gunzip < backups/backup_file.sql.gz | \
  docker exec -i postgres-db psql -U admin -d skillmatch_db

# เข้า database
docker exec -it postgres-db psql -U admin -d skillmatch_db
```

### Performance
```bash
# Container stats
docker stats

# Database connections
docker exec postgres-db psql -U admin -d skillmatch_db \
  -c "SELECT count(*) FROM pg_stat_activity;"

# Redis memory
docker exec redis-cache redis-cli info memory
```

---

## 🔄 Update & Rollback

### Update
```bash
git pull origin main
docker-compose -f docker-compose.prod.yml build
docker-compose -f docker-compose.prod.yml up -d
```

### Rollback
```bash
docker images skillmatch-api
docker tag skillmatch-api:previous skillmatch-api:latest
docker-compose -f docker-compose.prod.yml up -d
```

---

## 🚨 Troubleshooting

### Migration Errors
```bash
# ตรวจสอบไฟล์ migration
ls -la docs/sql-migrations/

# ดู logs
docker-compose logs api | grep Migration
```

### Port Already in Use
```bash
# หา process ที่ใช้ port
lsof -ti:8080 | xargs kill -9
```

### SSL Certificate Renewal
```bash
sudo certbot renew
sudo cp /etc/letsencrypt/live/your-domain.com/*.pem ./ssl/
docker-compose restart nginx
```

### Database Connection
```bash
# ตรวจสอบ database
docker ps | grep postgres

# ทดสอบการเชื่อมต่อ
docker exec api ping postgres-db
```

---

## 📈 Scaling (Optional)

### Multiple API Instances
```yaml
services:
  api:
    deploy:
      replicas: 3
```

### Database Read Replicas
```yaml
postgres-read-replica:
  image: postgres:15-alpine
```

### Redis Cluster
```yaml
redis-sentinel:
  image: redis:7-alpine
```

---

## 🔐 Security

### ระบบความปลอดภัย
- JWT authentication
- Password hashing (bcrypt)
- SQL injection prevention
- HTTPS/TLS
- Rate limiting
- Security headers
- Non-root Docker user
- Health checks

### เพิ่มเติม (แนะนำ)
- Fail2ban
- VPN for database
- 2FA for admin
- Security monitoring

---

## 📚 เอกสารอ้างอิง

- **API Documentation:** `docs/`
- **Database Schema:** `DATABASE_STRUCTURE.md`
- **Features Guide:** `PROVIDER_SYSTEM_GUIDE.md`, `FINANCIAL_SYSTEM_GUIDE.md`

---

## ✅ สถานะระบบ

### ระบบหลัก (พร้อมใช้งาน)
- Docker containerization ✅
- nginx reverse proxy ✅
- Database (PostgreSQL) ✅
- Cache (Redis) ✅
- Authentication & Authorization ✅
- Payment (Stripe) ✅
- Messaging (WebSocket) ✅
- File upload (GCS) ✅
- Email verification ✅
- Rate limiting ✅
- Security headers ✅
- Health checks ✅
- Backup script ✅
- CI/CD pipeline ✅

### ฟีเจอร์ครบถ้วน
- User management ✅
- Provider system ✅
- Booking system ✅
- Review system ✅
- Messaging system ✅
- Financial system ✅
- Admin panel ✅
- Analytics ✅
- Notifications ✅

---

**Updated:** November 24, 2025  
**Tech Stack:** Go + Gin + PostgreSQL + Redis + nginx
