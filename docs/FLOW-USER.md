# USER FLOW — Freelance OS
## Panduan Sederhana untuk Pengguna

> Dokumen ini menjelaskan alur kerja aplikasi dari sudut pandang pengguna biasa. Tidak ada istilah teknis — hanya langkah-langkah mudah.

---

## 🎯 Gambaran Besar

```
User daftar → Dapat Workspace + Free Plan → Buat Project → Buat Task → Kelola Client & Invoice
```

---

## 🔄 Alur #1: Pendaftaran

### Yang dilakukan USER:
1. Buka aplikasi → Klik **Daftar**
2. Isi formulir (nama, email, password, phone, address)
3. Klik **Buat Akun**

### Yang dilakukan SISTEM:
1. Buat akun user
2. Buat workspace pribadi otomatis (contoh: *"Budi's Workspace"*)
3. Set role: Owner di workspace sendiri
4. Set plan: Free
5. Langsung login otomatis (token JWT diberikan)

### Hasil:
- User langsung masuk dashboard
- Tidak perlu setup apapun — langsung bisa kerja

---

## 🔄 Alur #2: Login

### Yang dilakukan USER:
1. Masukkan email + password
2. Klik **Masuk**

### Yang dilakukan SISTEM:
1. Cek kredensial
2. Jika benar → generate JWT token
3. Load workspace user

### Hasil:
- User masuk ke dashboard
- License warning muncul jika plan Free atau expired

---

## 🔄 Alur #3: Buat Project

### Yang dilakukan USER:
1. Di dashboard → Klik **+ Buat Project**
2. Isi nama project
3. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Buat record project baru
2. Auto-generate 5 label: Todo, On Going, Done, Delivered, Canceled
3. Auto-generate 5 status task: Todo, On Progress, Done, Pending, Cancel
4. Set project status: Active
5. Set task default status: Todo (label pertama)

### Hasil:
- Project muncul di dashboard
- Langsung punya 5 label dan 5 status siap pakai

---

## 🔄 Alur #4: Buat Task

### Yang dilakukan USER:
1. Di dalam project → Klik **+ Buat Task**
2. Isi judul task
3. Pilih assignee (penanggung jawab)
4. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Buat record task
2. Set status default: Todo (label pertama)
3. Set priority default: Medium
4. Hubungkan task dengan assignee

### Hasil:
- Task muncul di kolom "Todo"
- Assignee bisa melihat task-nya

---

## 🔄 Alur #5: Pindahkan Task

### Yang dilakukan USER:
1. Klik task → Ubah status/label (drag & drop atau edit)
2. Pilih label baru: Todo → On Going → Done → Delivered → Canceled

### Yang dilakukan SISTEM:
1. Update task status
2. Reorder index task di kolom
3. Jika perlu, shift index task lain

### Hasil:
- Task berpindah kolom
- Semua member workspace bisa lihat perubahan

---

## 🔄 Alur #6: Invite Member

### Yang dilakukan USER:
1. Buka **Organization** → **Undang Anggota**
2. Masukkan email teman
3. Pilih role (Admin/Member)
4. Klik **Kirim Undangan**

### Yang dilakukan SISTEM:
1. Buat invitation record dengan token
2. Kirim email undangan
3. Token expired dalam 7 hari

### Teman USER:
1. Buka email → Klik link
2. Daftar/Login
3. Secara otomatis masuk workspace

### Hasil:
- Teman bisa akses project & task di workspace yang diundang

---

## 🔄 Alur #7: Buat Client & Invoice

### Yang dilakukan USER:
1. Menu **Clients** → **+ Tambah Client**
2. Isi data klien (nama, email, WhatsApp)
3. Klik **Simpan**
4. Buka client → **+ Buat Invoice**
5. Isi nominal & deskripsi
6. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Buat record client
2. Generate nomor invoice otomatis (contoh: INV-2026-001)
3. Set status invoice: Sent

### Saat klien bayar:
1. User klik **Mark Paid** di invoice
2. Isi jumlah yang dibayar + tanggal
3. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Update status invoice → Paid
2. Update total revenue client

### Hasil:
- Nomor invoice auto-generate
- Revenue client otomatis naik saat invoice dilunasi

---

## 🔄 Alur #8: Upgrade License

### Yang dilakukan USER:
1. Lihat warning banner di dashboard (jika Free/Expired)
2. Hubungi admin untuk dapat license key
3. Masuk menu **Activate License**
4. Masukkan license key
5. Klik **Aktifkan**

### Yang dilakukan SISTEM:
1. Validasi license key
2. Set plan sesuai license (Free/Pro/Team)
3. Set expiry date
4. Hapus warning banner

### Hasil:
- Banner license hilang
- Fitur premium aktif (jika ada)

---

## 🔄 Alur #9: Pindah Workspace

### Yang dilakukan USER:
1. Klik avatar/nama → **Switch Organization**
2. Pilih workspace yang dituju
3. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Validasi user adalah member workspace tersebut
2. Update context organization

### Hasil:
- User berpindah ke workspace lain
- Dashboard berubah menyesuaikan workspace baru

---

## 📊 Ringkasan Status

### Project Status:
- Active (hijau) — sedang dikerjakan
- On Hold (kuning) — dihentikan sementara
- Completed (biru) — selesai
- Archived (abu-abu) — diarsipkan

### Task Labels:
- Todo (abu-abu) — belum dikerjakan
- On Going (biru) — sedang dikerjakan
- Done (hijau) — selesai
- Delivered (ungu) — sudah dikirim/selesai
- Canceled (merah) — dibatalkan

### Task Status (internal):
- Todo — belum dikerjakan
- On Progress — sedang dikerjakan
- Done — selesai
- Pending — menunggu
- Cancel — dibatalkan

---

## ❓ Kalau Error

| Masalah | Solusi |
|---|---|
| Gagal daftar (email sudah ada) | Gunakan "Lupa Password" |
| Gagal invite member | Pastikan kamu Admin/Owner |
| Invoice tidak bisa di-paid | Pastikan invoice tidak canceled |
| Token expired | Minta admin resend invitation |

---

*Dokumen ini tidak bersifat teknis. Untuk dokumentasi developer, lihat file terpisah.*
