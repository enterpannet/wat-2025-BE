# 🔐 การจัดการ Environment Variables สำหรับ Backend

## 📋 Environment Variables ที่ต้องใช้

Backend application ใช้ environment variables ต่อไปนี้:

| Variable | คำอธิบาย | Default | Required |
|----------|---------|---------|----------|
| `PORT` | Port ที่ backend จะ listen | `3000` | No |
| `CORS_ORIGIN` | Frontend origin สำหรับ CORS | `http://localhost:5173` | No |
| `DB_HOST` | PostgreSQL host | - | **Yes** |
| `DB_PORT` | PostgreSQL port | `5432` | No |
| `DB_USER` | PostgreSQL username | - | **Yes** |
| `DB_PASSWORD` | PostgreSQL password | - | **Yes** |
| `DB_NAME` | PostgreSQL database name | - | **Yes** |
| `DB_SSL_MODE` | SSL mode สำหรับ database | `disable` | No |
| `DB_CHANNEL_BINDING` | Channel binding สำหรับ SSL | - | No |

## 🔧 วิธีที่ 1: ใช้ Systemd Service File (แนะนำ ⭐)

### ข้อดี
- ✅ ปลอดภัยกว่า (ไม่ต้องส่งผ่าน GitHub Actions)
- ✅ จัดการง่าย (แก้ไขบน server โดยตรง)
- ✅ ไม่ต้อง redeploy เมื่อเปลี่ยน env
- ✅ ใช้ environment variables จาก system level

### การตั้งค่า

แก้ไขไฟล์ `/etc/systemd/system/registration-api.service`:

```ini
[Service]
# Environment variables
Environment="PORT=3000"
Environment="CORS_ORIGIN=https://mostdata.site"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=registration_user"
Environment="DB_PASSWORD=your_secure_password"
Environment="DB_NAME=registration_db"
Environment="DB_SSL_MODE=disable"
```

หลังจากแก้ไข:

```bash
sudo systemctl daemon-reload
sudo systemctl restart registration-api
```

## 🔧 วิธีที่ 2: ใช้ GitHub Secrets + .env File

### ข้อดี
- ✅ จัดการผ่าน GitHub UI
- ✅ Version controlled (แต่ไม่ควรเก็บ secrets ใน code)
- ✅ Deploy อัตโนมัติพร้อมกับ code

### ข้อเสีย
- ❌ Secrets ต้องส่งผ่าน GitHub Actions
- ❌ ต้อง redeploy เมื่อเปลี่ยน env

### การตั้งค่า GitHub Secrets

ไปที่: **GitHub Repository → Settings → Secrets and variables → Actions**

เพิ่ม secrets ต่อไปนี้:

| Secret Name | คำอธิบาย | ตัวอย่าง |
|------------|---------|---------|
| `PORT` | Port สำหรับ backend | `3000` |
| `CORS_ORIGIN` | Frontend URL | `https://mostdata.site` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database username | `registration_user` |
| `DB_PASSWORD` | Database password | `your_secure_password` |
| `DB_NAME` | Database name | `registration_db` |
| `DB_SSL_MODE` | SSL mode | `disable` หรือ `require` |
| `DB_CHANNEL_BINDING` | Channel binding (optional) | `prefer` |

### การทำงาน

เมื่อ deploy, workflow จะสร้าง `.env` file ใน `/var/www/registration-system/backend/.env` จาก GitHub Secrets

## 🔧 วิธีที่ 3: ใช้ .env File บน Server (Manual)

สร้างไฟล์ `/var/www/registration-system/backend/.env`:

```bash
sudo nano /var/www/registration-system/backend/.env
```

ใส่ค่า:

```env
PORT=3000
CORS_ORIGIN=https://mostdata.site
DB_HOST=localhost
DB_PORT=5432
DB_USER=registration_user
DB_PASSWORD=your_secure_password
DB_NAME=registration_db
DB_SSL_MODE=disable
```

ตั้งค่า permissions:

```bash
sudo chown www-data:www-data /var/www/registration-system/backend/.env
sudo chmod 600 /var/www/registration-system/backend/.env
```

## 📊 Priority Order

Backend จะอ่าน environment variables ตามลำดับนี้:

1. **System environment variables** (สูงสุด)
2. **.env file** (ถ้ามี)
3. **Default values** ใน code (ต่ำสุด)

> **หมายเหตุ:** ถ้าใช้ systemd service file, environment variables จะถูกตั้งค่าเป็น system level (priority สูงสุด)

## ✅ แนะนำ

**สำหรับ Production:**
- ใช้ **Systemd Service File** (วิธีที่ 1) เพราะปลอดภัยและจัดการง่าย

**สำหรับ Development:**
- ใช้ **.env file** (วิธีที่ 3) เพราะง่ายต่อการแก้ไข

**สำหรับ CI/CD:**
- ใช้ **GitHub Secrets** (วิธีที่ 2) ถ้าต้องการ deploy อัตโนมัติ

## 🔒 Security Best Practices

1. **อย่า commit `.env` file ลง repository**
2. **ใช้ strong passwords** สำหรับ database
3. **ตั้งค่า file permissions** ให้ถูกต้อง (600)
4. **ใช้ SSL/TLS** สำหรับ database connection ใน production
5. **หมั่น rotate passwords** เป็นระยะ

## 🔍 การตรวจสอบ

### ตรวจสอบ Environment Variables

```bash
# ดู environment variables ของ service
sudo systemctl show registration-api --property=Environment

# หรือดูใน .env file
sudo cat /var/www/registration-system/backend/.env
```

### ตรวจสอบว่ามีการโหลด env ถูกต้อง

ดู logs:

```bash
sudo journalctl -u registration-api -f
```

หา log message: `"Database connected successfully"` (แสดงว่า env ถูกต้อง)

## 🐛 Troubleshooting

### Backend ไม่เชื่อมต่อ Database

```bash
# ตรวจสอบ env variables
sudo systemctl show registration-api --property=Environment

# ตรวจสอบ .env file
sudo cat /var/www/registration-system/backend/.env

# ทดสอบ database connection
sudo -u www-data /var/www/registration-system/backend/registration-api
```

### Service ไม่ใช้ env จาก .env file

- ตรวจสอบว่า service ไม่มี `Environment=` ใน service file (จะ override .env file)
- หรือใช้ systemd service file แทน

