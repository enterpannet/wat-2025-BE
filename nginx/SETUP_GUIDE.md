# 🚀 คู่มือการ Setup Server สำหรับ mostdata.site

## 📋 Prerequisites

- Ubuntu 20.04+ หรือ Debian 11+
- Root หรือ sudo access
- Domain `mostdata.site` ชี้มาที่ server IP
- PostgreSQL database setup

## 🔧 ขั้นตอนการติดตั้ง

### 1. อัพเดทระบบ

```bash
sudo apt update && sudo apt upgrade -y
```

### 2. ติดตั้ง Dependencies

```bash
# Nginx
sudo apt install nginx -y

# PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Certbot (สำหรับ SSL)
sudo apt install certbot python3-certbot-nginx -y

# System tools
sudo apt install curl wget git -y
```

### 3. สร้าง PostgreSQL Database และ User

```bash
sudo -u postgres psql

-- ใน PostgreSQL prompt:
CREATE DATABASE registration_db;
CREATE USER registration_user WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE registration_db TO registration_user;
\q
```

### 4. สร้าง Directory Structure

```bash
sudo mkdir -p /var/www/registration-system/{backend,frontend/dist}
sudo chown -R www-data:www-data /var/www/registration-system
```

### 5. ติดตั้ง Nginx Configuration

```bash
# Copy configuration
sudo cp nginx/mostdata.site.conf /etc/nginx/sites-available/mostdata.site

# Enable site
sudo ln -s /etc/nginx/sites-available/mostdata.site /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

### 6. ขอ SSL Certificate (Let's Encrypt)

```bash
# ขอ certificate
sudo certbot --nginx -d mostdata.site -d www.mostdata.site

# Certbot จะอัพเดท nginx config อัตโนมัติ
# ตรวจสอบว่า SSL ทำงาน
sudo nginx -t && sudo systemctl reload nginx
```

### 7. Setup Backend Service

```bash
# Copy service file
sudo cp nginx/registration-api.service /etc/systemd/system/

# แก้ไข environment variables ตามข้อมูลจริง
sudo nano /etc/systemd/system/registration-api.service

# Reload systemd
sudo systemctl daemon-reload

# Enable service (จะ start อัตโนมัติตอน reboot)
sudo systemctl enable registration-api

# Start service
sudo systemctl start registration-api

# ตรวจสอบ status
sudo systemctl status registration-api
```

### 8. Firewall Configuration

```bash
# เปิด firewall สำหรับ HTTP, HTTPS
sudo ufw allow 'Nginx Full'
sudo ufw allow ssh
sudo ufw enable

# ตรวจสอบ status
sudo ufw status
```

## ✅ การตรวจสอบ

### ตรวจสอบ Nginx

```bash
# ตรวจสอบ status
sudo systemctl status nginx

# ตรวจสอบ configuration
sudo nginx -t

# ดู logs
sudo tail -f /var/log/nginx/mostdata.site.access.log
sudo tail -f /var/log/nginx/mostdata.site.error.log
```

### ตรวจสอบ Backend API

```bash
# ตรวจสอบ service status
sudo systemctl status registration-api

# ดู logs
sudo journalctl -u registration-api -f

# ทดสอบ API
curl http://localhost:3000/api/public/provinces
```

### ตรวจสอบ Frontend

```bash
# ตรวจสอบว่าไฟล์ frontend อยู่ถูกที่
ls -la /var/www/registration-system/frontend/dist/

# ทดสอบจาก browser
# เปิด https://mostdata.site
```

## 🔄 Workflow สำหรับ Deploy

หลังจาก setup แล้ว การ deploy จะทำงานผ่าน GitHub Actions:

1. **Backend Deploy**: 
   - Build Go binary
   - Upload ไปที่ `/var/www/registration-system/backend/`
   - Restart `registration-api` service

2. **Frontend Deploy**:
   - Build React app ด้วย bun
   - Upload ไปที่ `/var/www/registration-system/frontend/dist/`
   - Reload nginx

## 🔐 Security Checklist

- [ ] เปลี่ยน database password ให้ปลอดภัย
- [ ] ตั้งค่า firewall (UFW)
- [ ] ใช้ SSL certificate (Let's Encrypt)
- [ ] อัพเดทระบบให้เป็นเวอร์ชันล่าสุด
- [ ] ตั้งค่า backup database
- [ ] ตั้งค่า log rotation
- [ ] ตรวจสอบ file permissions

## 📝 Environment Variables สำหรับ Backend

แก้ไขใน `/etc/systemd/system/registration-api.service`:

```ini
Environment="PORT=3000"
Environment="CORS_ORIGINS=https://mostdata.site,https://www.mostdata.site"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=registration_user"
Environment="DB_PASSWORD=your_secure_password"
Environment="DB_NAME=registration_db"
Environment="DB_SSL_MODE=disable"
```

## 🔧 Troubleshooting

### Backend ไม่ start

```bash
# ดู error logs
sudo journalctl -u registration-api -n 50

# ตรวจสอบ database connection
sudo -u postgres psql -d registration_db -U registration_user

# Test binary manually
sudo -u www-data /var/www/registration-system/backend/registration-api
```

### Nginx 404 Error

```bash
# ตรวจสอบว่าไฟล์ frontend อยู่ถูกที่
ls -la /var/www/registration-system/frontend/dist/

# ตรวจสอบ nginx configuration
sudo nginx -t
sudo cat /etc/nginx/sites-available/mostdata.site | grep root
```

### API ไม่ทำงาน

```bash
# ตรวจสอบว่า backend ฟังที่ port 3000
sudo netstat -tlnp | grep 3000

# ทดสอบ API จาก server
curl http://localhost:3000/api/health

# ตรวจสอบ nginx proxy configuration
sudo nginx -t
```

## 📚 Additional Resources

- [Nginx Documentation](https://nginx.org/en/docs/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [systemd Documentation](https://www.freedesktop.org/software/systemd/man/systemd.service.html)

