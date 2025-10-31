# 🔧 Troubleshooting 403 Forbidden Error

## ปัญหา: 403 Forbidden เมื่อเข้าถึง https://www.mostdata.site/

### 🔍 การตรวจสอบ

#### 1. ตรวจสอบว่า Frontend Directory มีอยู่หรือไม่

```bash
# ตรวจสอบ directory
ls -la /var/www/registration-system/frontend/dist/

# ถ้าไม่มี directory ให้สร้าง
sudo mkdir -p /var/www/registration-system/frontend/dist
sudo chown -R www-data:www-data /var/www/registration-system/frontend
sudo chmod -R 755 /var/www/registration-system/frontend
```

#### 2. ตรวจสอบว่า Frontend Files ถูก Deploy หรือยัง

```bash
# ดูว่ามี index.html หรือไม่
ls -la /var/www/registration-system/frontend/dist/index.html

# ถ้าไม่มี แสดงว่ายังไม่ได้ deploy frontend
```

#### 3. ตรวจสอบ Permissions

```bash
# ตรวจสอบ ownership
ls -la /var/www/registration-system/frontend/

# ควรเป็น www-data:www-data
# ถ้าไม่ใช่ แก้ไขด้วย:
sudo chown -R www-data:www-data /var/www/registration-system/frontend
sudo chmod -R 755 /var/www/registration-system/frontend
```

#### 4. ตรวจสอบ Nginx Error Logs

```bash
# ดู error logs เพื่อหาสาเหตุ
sudo tail -f /var/log/nginx/mostdata.site.error.log

# ดู access logs
sudo tail -f /var/log/nginx/mostdata.site.access.log
```

#### 5. ตรวจสอบ SELinux (ถ้าใช้)

```bash
# ตรวจสอบว่า SELinux block หรือไม่
getenforce

# ถ้าเป็น Enforcing อาจต้อง set context
sudo setsebool -P httpd_read_user_content 1
sudo chcon -R -t httpd_sys_content_t /var/www/registration-system/frontend
```

### 🔧 การแก้ไข

#### กรณีที่ 1: Frontend ยังไม่ได้ Deploy

**Deploy Frontend ผ่าน GitHub Actions:**

1. ไปที่ GitHub Repository → Actions
2. เลือก workflow "Deploy Frontend to Ubuntu"
3. คลิก "Run workflow"
4. รอให้ deploy เสร็จ

**หรือ Deploy Manual:**

```bash
# บน local machine หรือ CI/CD
cd frontend
bun install
bun run build

# Upload ไปที่ server
scp -r dist/* user@mostdata.site:/tmp/frontend-deploy/
ssh user@mostdata.site

# บน server
sudo mkdir -p /var/www/registration-system/frontend/dist
sudo cp -r /tmp/frontend-deploy/* /var/www/registration-system/frontend/dist/
sudo chown -R www-data:www-data /var/www/registration-system/frontend/dist
sudo chmod -R 755 /var/www/registration-system/frontend/dist
sudo systemctl reload nginx
```

#### กรณีที่ 2: Permissions ผิด

```bash
# แก้ไข ownership
sudo chown -R www-data:www-data /var/www/registration-system/frontend

# แก้ไข permissions
sudo chmod -R 755 /var/www/registration-system/frontend
sudo find /var/www/registration-system/frontend/dist -type f -exec chmod 644 {} \;
sudo find /var/www/registration-system/frontend/dist -type d -exec chmod 755 {} \;
```

#### กรณีที่ 3: Nginx Configuration Error

```bash
# ตรวจสอบ syntax
sudo nginx -t

# ถ้ามี error แก้ไข config แล้ว reload
sudo systemctl reload nginx

# หรือ restart
sudo systemctl restart nginx
```

#### กรณีที่ 4: Directory Structure ผิด

```bash
# ตรวจสอบ structure
tree -L 3 /var/www/registration-system/frontend/

# ควรเป็น:
# /var/www/registration-system/frontend/
# └── dist/
#     ├── index.html
#     ├── assets/
#     └── ...
```

### 📋 Checklist

- [ ] Directory `/var/www/registration-system/frontend/dist` มีอยู่
- [ ] มีไฟล์ `index.html` ใน dist directory
- [ ] Ownership เป็น `www-data:www-data`
- [ ] Permissions ถูกต้อง (755 สำหรับ directory, 644 สำหรับ files)
- [ ] Nginx config syntax ถูกต้อง (`nginx -t`)
- [ ] Nginx service ทำงานอยู่ (`systemctl status nginx`)
- [ ] Frontend files ถูก deploy แล้ว

### 🐛 Debug Endpoint

ใช้ debug endpoint เพื่อตรวจสอบ:

```
https://mostdata.site/debug-frontend
```

จะแสดงข้อมูลว่า directory มีอยู่หรือไม่

### 📚 เพิ่มเติม

- [Nginx 403 Forbidden Solutions](https://www.nginx.com/blog/nginx-403-forbidden/)
- [File Permissions Guide](https://www.linux.com/training-tutorials/understanding-linux-file-permissions/)

