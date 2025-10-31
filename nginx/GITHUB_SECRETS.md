# 🔐 การตั้งค่า GitHub Secrets สำหรับ mostdata.site

## 📋 Secrets ที่ต้องตั้งค่า

เนื่องจาก deploy ไปที่ **server เดียวกัน** คุณสามารถใช้ **secrets เดียวกัน** สำหรับทั้ง backend และ frontend

### สำหรับ Server เดียวกัน

ไปที่: **GitHub Repository → Settings → Secrets and variables → Actions → New repository secret**

| Secret Name | คำอธิบาย | ตัวอย่าง |
|------------|---------|---------|
| `HOST` | IP address หรือ domain ของ server | `mostdata.site` หรือ `192.168.1.100` |
| `USERNAME` | Username สำหรับ SSH | `ubuntu` หรือ `root` หรือ `www-data` |
| `SSH_KEY` | Private SSH key สำหรับ authentication | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `PORT` | Port สำหรับ SSH (ถ้าไม่ใช่ 22) | `22` หรือ `2222` |

### สำหรับ Backend Workflow

Workflow จะใช้ secrets เหล่านี้ (สามารถใช้ชื่อเดียวกับด้านบนได้):

- `BACKEND_HOST` หรือใช้ `HOST` (แนะนำ)
- `BACKEND_USERNAME` หรือใช้ `USERNAME` (แนะนำ)
- `BACKEND_SSH_KEY` หรือใช้ `SSH_KEY` (แนะนำ)
- `BACKEND_PORT` หรือใช้ `PORT` (แนะนำ)

### สำหรับ Frontend Workflow

Workflow ใช้ secrets เหล่านี้ (สามารถใช้ชื่อเดียวกับด้านบนได้):

- `FRONTEND_HOST` หรือใช้ `HOST` (แนะนำ)
- `FRONTEND_USERNAME` หรือใช้ `USERNAME` (แนะนำ)
- `FRONTEND_SSH_KEY` หรือใช้ `SSH_KEY` (แนะนำ)
- `FRONTEND_PORT` หรือใช้ `PORT` (แนะนำ)

## 💡 แนะนำ

**ถ้า deploy ไปที่ server เดียวกัน** แนะนำให้ใช้ชื่อ secrets เดียวกัน:
- `HOST`
- `USERNAME`
- `SSH_KEY`
- `PORT`

จากนั้นแก้ไข workflows ให้ใช้ชื่อเดียวกัน:

### แก้ไข Backend Workflow

```yaml
# แทนที่ BACKEND_HOST ด้วย HOST
host: ${{ secrets.HOST }}

# แทนที่ BACKEND_USERNAME ด้วย USERNAME
username: ${{ secrets.USERNAME }}

# แทนที่ BACKEND_SSH_KEY ด้วย SSH_KEY
key: ${{ secrets.SSH_KEY }}

# แทนที่ BACKEND_PORT ด้วย PORT
port: ${{ secrets.PORT || 22 }}
```

### Frontend Workflow

Frontend workflow ใช้ชื่อ `HOST`, `USERNAME`, `SSH_KEY`, `PORT` อยู่แล้ว ไม่ต้องแก้ไข

## 🔑 การสร้าง SSH Key

### 1. สร้าง SSH Key

```bash
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/github_actions_deploy
```

### 2. Copy Public Key ไปยัง Server

```bash
ssh-copy-id -i ~/.ssh/github_actions_deploy.pub username@mostdata.site
```

หรือ copy manually:

```bash
cat ~/.ssh/github_actions_deploy.pub
# แล้ว SSH เข้า server และ paste ลงใน ~/.ssh/authorized_keys
```

### 3. Copy Private Key ไปยัง GitHub Secrets

```bash
cat ~/.ssh/github_actions_deploy
```

Copy ทั้งหมด (รวม `-----BEGIN OPENSSH PRIVATE KEY-----` และ `-----END OPENSSH PRIVATE KEY-----`) แล้วใส่ใน GitHub Secret `SSH_KEY`

## ✅ การทดสอบ SSH Connection

```bash
# ทดสอบ SSH connection
ssh -i ~/.ssh/github_actions_deploy username@mostdata.site

# ถ้าทำงานได้ แสดงว่าพร้อมสำหรับ GitHub Actions
```

## 🚨 Security Notes

- **อย่า commit private key ลง repository**
- ใช้ SSH key แยกสำหรับ GitHub Actions
- ตั้งค่า key permissions ให้ถูกต้อง (`chmod 600 ~/.ssh/github_actions_deploy`)
- หมั่น rotate keys เป็นระยะ

