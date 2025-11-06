# Nginx Configuration สำหรับ mostdata.site

## 📋 ภาพรวม

Configuration นี้ตั้งค่าให้:
- **Frontend**: Serve ที่ `https://mostdata.site` (root domain)
- **Backend API**: Proxy ที่ `https://mostdata.site/api` → `http://localhost:3000`

## 🚀 การติดตั้ง

### 1. Copy Configuration File

```bash
sudo cp nginx/mostdata.site.conf /etc/nginx/sites-available/mostdata.site
sudo ln -s /etc/nginx/sites-available/mostdata.site /etc/nginx/sites-enabled/
```

### 2. ตรวจสอบ Configuration

```bash
sudo nginx -t
```

### 3. Reload Nginx

```bash
sudo systemctl reload nginx
```

## 🔒 SSL Certificate Setup

### สำหรับ Production (Let's Encrypt)

```bash
# ติดตั้ง Certbot
sudo apt update
sudo apt install certbot python3-certbot-nginx

# ขอ SSL certificate
sudo certbot --nginx -d mostdata.site -d www.mostdata.site

# Certificate จะถูก renew อัตโนมัติ
```

### สำหรับ Development/Testing (Self-signed)

```bash
# สร้าง self-signed certificate
sudo mkdir -p /etc/letsencrypt/live/mostdata.site/
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/letsencrypt/live/mostdata.site/privkey.pem \
  -out /etc/letsencrypt/live/mostdata.site/fullchain.pem \
  -subj "/CN=mostdata.site"

# แก้ไข nginx config ให้ใช้ self-signed cert
```

## 📝 Configuration Details

### Ports
- **Frontend**: Served by Nginx (port 443 HTTPS, 80 HTTP redirect)
- **Backend**: Running on `localhost:3000` (internal only)

### Paths
- **Frontend files**: `/var/www/registration-system/frontend/dist`
- **Backend binary**: `/var/www/registration-system/backend/registration-api`

### Features
- ✅ HTTP to HTTPS redirect
- ✅ Gzip compression
- ✅ Static file caching
- ✅ Rate limiting for API
- ✅ Security headers
- ✅ WebSocket support (for future use)
- ✅ Health check endpoint

## 🔧 Environment Variables

ต้องตั้งค่า environment variables สำหรับ backend:

```bash
# ใน systemd service file หรือ .env
PORT=3000
CORS_ORIGINS=https://mostdata.site,https://www.mostdata.site
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
DB_SSL_MODE=disable
```

## 📊 Monitoring

### ดู Logs

```bash
# Access logs
sudo tail -f /var/log/nginx/mostdata.site.access.log

# Error logs
sudo tail -f /var/log/nginx/mostdata.site.error.log
```

### ตรวจสอบ Status

```bash
# Nginx status
sudo systemctl status nginx

# Backend status
sudo systemctl status registration-api
```

## 🔍 Troubleshooting

### Nginx ไม่ทำงาน
```bash
# ตรวจสอบ syntax
sudo nginx -t

# ดู error logs
sudo tail -n 50 /var/log/nginx/error.log

# Restart nginx
sudo systemctl restart nginx
```

### API ไม่ทำงาน
```bash
# ตรวจสอบว่า backend ทำงานอยู่
curl http://localhost:3000/api/health

# ตรวจสอบ backend logs
sudo journalctl -u registration-api -f
```

### SSL Certificate Issues
```bash
# Renew certificate manually
sudo certbot renew

# ตรวจสอบ certificate expiry
sudo certbot certificates
```

## 🔄 การอัพเดท Configuration

หลังจากแก้ไข configuration:

```bash
# ตรวจสอบ syntax
sudo nginx -t

# Reload (ไม่หยุด service)
sudo systemctl reload nginx

# หรือ Restart (หยุดและเริ่มใหม่)
sudo systemctl restart nginx
```

## 📚 Additional Resources

- [Nginx Documentation](https://nginx.org/en/docs/)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Certbot Documentation](https://certbot.eff.org/)

