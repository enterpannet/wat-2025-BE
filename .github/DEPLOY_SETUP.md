# เอกสารการตั้งค่า Deploy ด้วย GitHub Actions

## 📋 ภาพรวม

ระบบ deploy นี้ใช้ GitHub Actions เพื่อ deploy ทั้ง backend (Go) และ frontend (React/TypeScript) ไปยัง server ผ่าน SSH

## 🔧 การตั้งค่า GitHub Secrets

### สำหรับ Backend

ไปที่: **Settings → Secrets and variables → Actions → New repository secret**

เพิ่ม secrets ต่อไปนี้:

| Secret Name | คำอธิบาย | ตัวอย่าง |
|------------|---------|---------|
| `BACKEND_HOST` | IP address หรือ domain ของ server | `192.168.1.100` หรือ `api.example.com` |
| `BACKEND_USERNAME` | Username สำหรับ SSH | `ubuntu` หรือ `root` |
| `BACKEND_SSH_KEY` | Private SSH key สำหรับ authentication | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `BACKEND_PORT` | Port สำหรับ SSH (ถ้าไม่ใช่ 22) | `22` หรือ `2222` |

### สำหรับ Frontend

| Secret Name | คำอธิบาย | ตัวอย่าง |
|------------|---------|---------|
| `FRONTEND_HOST` | IP address หรือ domain ของ server | `192.168.1.100` หรือ `www.example.com` |
| `FRONTEND_USERNAME` | Username สำหรับ SSH | `ubuntu` หรือ `root` |
| `FRONTEND_SSH_KEY` | Private SSH key สำหรับ authentication | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `FRONTEND_PORT` | Port สำหรับ SSH (ถ้าไม่ใช่ 22) | `22` หรือ `2222` |

> **หมายเหตุ:** ถ้า backend และ frontend อยู่บน server เดียวกัน สามารถใช้ secrets เดียวกันได้ (ใช้ `BACKEND_HOST`, `BACKEND_USERNAME` เป็นต้น)

## 🔑 การสร้าง SSH Key

### 1. สร้าง SSH Key บน Local Machine

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/github_actions_deploy
```

### 2. Copy Public Key ไปยัง Server

```bash
ssh-copy-id -i ~/.ssh/github_actions_deploy.pub username@your-server-ip
```

### 3. Copy Private Key ไปยัง GitHub Secrets

```bash
cat ~/.ssh/github_actions_deploy
```

Copy ทั้งหมด (รวม `-----BEGIN` และ `-----END`) แล้วใส่ใน GitHub Secrets

## 📁 การตั้งค่า Server

### Backend Service (systemd)

สร้างไฟล์ `/etc/systemd/system/registration-api.service`:

```ini
[Unit]
Description=Registration System API
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/registration-system/backend
ExecStart=/var/www/registration-system/backend/registration-api
Restart=always
RestartSec=5

# Environment variables
Environment="PORT=3000"
Environment="CORS_ORIGIN=https://your-frontend-domain.com"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=your_db_user"
Environment="DB_PASSWORD=your_db_password"
Environment="DB_NAME=your_db_name"
Environment="DB_SSL_MODE=disable"
# Environment="DB_CHANNEL_BINDING=prefer" # ถ้าต้องการ

[Install]
WantedBy=multi-user.target
```

### เริ่มใช้งาน Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable registration-api
sudo systemctl start registration-api
sudo systemctl status registration-api
```

### Frontend (Nginx)

ตัวอย่าง configuration สำหรับ `/etc/nginx/sites-available/registration-system`:

```nginx
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;

    root /var/www/registration-system/frontend;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json;
}
```

เปิดใช้งาน:

```bash
sudo ln -s /etc/nginx/sites-available/registration-system /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## 🚀 Workflows ที่มี

### 1. `deploy-backend.yml`
- Deploy เฉพาะ backend
- Trigger เมื่อมีการ push ไฟล์ใน `backend/` หรือแก้ไข workflow file
- สามารถรัน manual ได้

### 2. `deploy-frontend.yml`
- Deploy เฉพาะ frontend
- Trigger เมื่อมีการ push ไฟล์ใน `frontend/` หรือแก้ไข workflow file
- สามารถรัน manual ได้

### 3. `deploy.yml`
- Deploy ทั้ง backend และ frontend พร้อมกัน (parallel)
- Trigger เมื่อ push ไปยัง `main` หรือ `master`
- สามารถรัน manual ได้

### 4. `ci.yml`
- Run tests และ build validation
- Trigger เมื่อ push หรือ pull request

## 📝 การใช้งาน

### Automatic Deploy
- Push code ไปยัง `main` หรือ `master` branch
- GitHub Actions จะทำงานอัตโนมัติ

### Manual Deploy
1. ไปที่ **Actions** tab ใน GitHub repository
2. เลือก workflow ที่ต้องการ (Deploy Backend, Deploy Frontend, หรือ Deploy All)
3. คลิก **Run workflow**
4. เลือก branch และคลิก **Run workflow**

## 🔍 Troubleshooting

### Backend ไม่ทำงาน
```bash
# ตรวจสอบ logs
sudo journalctl -u registration-api -f

# ตรวจสอบ status
sudo systemctl status registration-api

# Restart service
sudo systemctl restart registration-api
```

### Frontend ไม่แสดงผล
```bash
# ตรวจสอบ Nginx logs
sudo tail -f /var/log/nginx/error.log

# ตรวจสอบ configuration
sudo nginx -t

# ตรวจสอบ permissions
ls -la /var/www/registration-system/frontend
```

### SSH Connection Failed
- ตรวจสอบว่า SSH key ถูกต้อง
- ตรวจสอบว่า server allow SSH connection
- ตรวจสอบ firewall rules
- ตรวจสอบ port ถูกต้อง

## 📚 ข้อมูลเพิ่มเติม

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [SSH Action Documentation](https://github.com/appleboy/ssh-action)
- [SCP Action Documentation](https://github.com/appleboy/scp-action)

