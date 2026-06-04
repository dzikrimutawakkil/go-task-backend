# USER FLOW — Freelance OS
## Panduan Sederhana untuk Pengguna

> Dokumen ini menjelaskan alur kerja aplikasi dari sudut pandang pengguna biasa. Tidak ada istilah teknis — hanya langkah-langkah mudah.

---

## 🎯 Gambaran Besar

```
User daftar → Dapat Workspace (Personal) + Plan Free → Buat Project → Buat Task → Kelola Client & Invoice
```

---

## 🔄 Alur #1: Pendaftaran

### Yang dilakukan USER:
1. Buka aplikasi → Klik **Daftar**
2. Isi formulir (nama, email, password, phone, address)
3. Klik **Buat Akun**

### Yang dilakukan SISTEM:
1. Buat akun user
2. Set plan: **Free** secara default
3. Buat workspace pribadi otomatis (contoh: *"Budi's Workspace"*)
4. Set role: **Owner** di workspace sendiri
5. Langsung login otomatis (token JWT diberikan)

### Hasil:
- User langsung masuk dashboard
- Tidak perlu setup apapun — langsung bisa kerja
- Bisa pakai 1 workspace + 3 project + 50 tasks gratis

---

## 🔄 Alur #2: Login

### Yang dilakukan USER:
1. Masukkan email + password
2. Klik **Masuk**

### Yang dilakukan SISTEM:
1. Cek kredensial
2. Jika benar → generate JWT token
3. Load workspace user
4. Tampilkan tier info jika Free atau expired

### Hasil:
- User masuk ke dashboard
- Tier info banner muncul jika plan Free atau expired

---

## 🔄 Alur #3: Buat Project

### Yang dilakukan USER:
1. Di dashboard → Klik **+ Buat Project**
2. Isi nama project
3. Opsional: pilih client yang sudah ada atau buat client baru langsung
4. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Buat record project baru
2. Auto-generate 3 urgency label: **Urgent**, **Normal**, **Low**
3. Auto-generate 5 task status: **Todo**, **On Progress**, **Done**, **Pending**, **Cancel**
4. Set project status: **Active**
5. Set task default status: **Todo** (status pertama)
6. Opsional: link ke client atau buat client baru inline

### Hasil:
- Project muncul di dashboard
- Langsung punya 3 urgency label dan 5 status siap pakai
- Jika Free user sudah punya 3 project → tidak bisa tambah (HTTP 403)

---

## 🔄 Alur #4: Buat Task

### Yang dilakukan USER:
1. Di dalam project → Klik **+ Buat Task**
2. Isi judul task
3. Pilih urgency label (default: Normal)
4. Pilih assignee (penanggung jawab dari member workspace)
5. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Buat record task
2. Set status default: **Todo** (status pertama)
3. Set priority default: **Medium**
4. Hubungkan task dengan assignee
5. Jika Free user sudah punya 50 tasks di project tsb → tidak bisa tambah (HTTP 403)

### Hasil:
- Task muncul di kolom "Todo"
- Assignee bisa melihat task-nya

---

## 🔄 Alur #5: Ubah Urgency Task

### Yang dilakukan USER:
1. Klik task → Ubah urgency (ubah label)
2. Pilih urgency baru: **Low** → **Normal** → **Urgent**

### Yang dilakukan SISTEM:
1. Update task label (urgency)
2. Tampilkan visual indicator urgency (warna)

### Hasil:
- Task urgency berubah
- Semua member workspace bisa lihat perubahan

---

## 🔄 Alur #6: Invite Member

### Yang dilakukan USER (Owner/Admin):
1. Buka **Workspace** → **Undang Anggota**
2. Masukkan email teman
3. Pilih role (**Admin** / **Member**)
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
- Free plan hanya bisa 1 member (Owner saja)
- Pro plan bisa sampai 5 member
- Ultimate plan bisa sampai 15 member

---

## 🔄 Alur #7: Buat Client & Invoice

### Yang dilakukan USER:
1. Menu **Clients** → **+ Tambah Client**
2. Isi data klien (nama, email, WhatsApp, company)
3. Klik **Simpan**
4. Buka client → **+ Buat Invoice**
5. Isi nominal & deskripsi
6. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Buat record client
2. Generate nomor invoice otomatis (contoh: **INV-2026-001**)
3. Set status invoice: **Sent**

### Saat klien bayar:
1. User klik **Mark Paid** di invoice
2. Isi jumlah yang dibayar + tanggal
3. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Update status invoice → **Paid**
2. Update `total_revenue` client (otomatis)
3. Set `paid_at` timestamp

### Hasil:
- Nomor invoice auto-generate
- Revenue client otomatis naik saat invoice dilunasi
- Jika Free user sudah punya 10 invoice bulan ini → tidak bisa tambah (HTTP 403)

---

## 🔄 Alur #8: Link Project ke Client

### Yang dilakukan USER:
1. Buat project baru → pilih **Client** dari dropdown
2. Atau buat client baru langsung saat bikin project (**Inline Create**)
3. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Link project ke client
2. Client bisa di-ubah atau di-unlink kapan saja

### Hasil:
- Project menampilkan info client terkait
- Filter project berdasarkan client

---

## 🔄 Alur #9: Upgrade Plan (Admin Aktifasi Manual)

### Yang dilakukan USER:
1. Lihat banner di dashboard (jika plan **Free** atau expired)
2. Hubungi admin untuk konfirmasi pembayaran
3. Admin mengaktifkan plan via sistem

### Yang dilakukan ADMIN:
1. Buka menu **Admin**
2. Pilih workspace yang akan di-upgrade
3. Pilih plan: **Pro** atau **Ultimate**
4. Pilih durasi: 1–24 bulan
5. Klik **Aktifkan**

### Yang dilakukan SISTEM:
1. Update workspace tier
2. Set expiry date
3. Semua member workspace mendapat fitur sesuai tier
4. Banner tier info hilang

### Hasil:
- Banner hilang
- Fitur premium aktif sesuai tier workspace
- Tier diikat ke **workspace** (bukan ke user)
- User bisa punya workspace berbeda dengan tier berbeda

---

## 🔄 Alur #10: Pindah Workspace

### Yang dilakukan USER:
1. Klik avatar/nama di pojok kanan atas
2. Pilih **Switch Workspace**
3. Pilih workspace yang dituju
4. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Validasi user adalah member workspace tersebut
2. Update context workspace (via `X-Workspace-ID` header)

### Hasil:
- User berpindah ke workspace lain
- Dashboard berubah menyesuaikan workspace baru
- Tier yang berlaku sesuai workspace tujuan

---

## 🔄 Alur #11: Buat Workspace Baru (Team Workspace)

### Yang dilakukan USER (sesuai tier):
1. Klik **+ Buat Workspace** di dashboard
2. Isi nama workspace
3. Klik **Simpan**

### Yang dilakukan SISTEM:
1. Buat record workspace baru
2. Set user sebagai **Owner**
3. Workspace mulai dengan tier **Free**
4. Quota check: Free=1 workspace, Pro=2, Ultimate=4

### Hasil:
- Workspace baru muncul di daftar
- Bisa invite member ke workspace baru
- Workspace baru berdiri sendiri dengan tier masing-masing

---

## 📊 Ringkasan Status

### Project Status:
| Status | Warna | Arti |
|---|---|---|
| Active | 🟢 Hijau | Sedang dikerjakan |
| On Hold | 🟡 Kuning | Dihentikan sementara |
| Completed | 🔵 Biru | Selesai |
| Archived | ⚫ Abu-abu | Diarsipkan |

### Task Urgency Labels:
| Label | Warna | Arti |
|---|---|---|
| **Urgent** | 🔴 Merah | Harus selesai secepatnya |
| **Normal** | 🟡 Kuning | Prioritas standar |
| **Low** | 🟢 Hijau | Bisa nanti |

### Task Status (Workflow):
| Status | Arti |
|---|---|
| Todo | Belum dikerjakan |
| On Progress | Sedang dikerjakan |
| Done | Selesai |
| Pending | Menunggu |
| Cancel | Dibatalkan |

### Plan/Quota (per Workspace):
| Resource | Free | Pro | Ultimate |
|---|---|---|---|
| Workspace | 1 | 2 | 4 |
| Project per workspace | 3 | Unlimited | Unlimited |
| Task per project | 50 | Unlimited | Unlimited |
| Member per workspace | 1 | 5 | 15 |
| Clients | 5 | Unlimited | Unlimited |
| Invoice per bulan | 10 | Unlimited | Unlimited |

---

## ❓ Kalau Error

| Masalah | Solusi |
|---|---|
| Gagal daftar (email sudah ada) | Gunakan **Lupa Password** |
| Gagal buat workspace baru | Upgrade plan — Free hanya 1 workspace |
| Gagal buat project | Upgrade plan — Free hanya 3 project |
| Gagal invite member | Upgrade plan — Free hanya 1 member |
| Gagal buat invoice | Upgrade plan — Free hanya 10 invoice/bulan |
| Invoice tidak bisa di-paid | Pastikan invoice tidak canceled |
| Token expired | Minta admin resend invitation |

---

## 💰 Pricing

| Plan | Harga/Bulan | Untuk |
|---|---|---|
| **Free** | Rp 0 | Perorangan, coba-coba |
| **Pro** | Rp 79.000 | Freelancer, pekerja lepas |
| **Ultimate** | Rp 199.000 | Tim kecil, agensi |

---

## 🌟 Fitur Premium

Fitur di bawah ini **tidak tersedia** di plan Free:

| Fitur | Free | Pro | Ultimate |
|---|---|---|---|
| Unlimited Projects | ❌ | ✅ | ✅ |
| Unlimited Tasks | ❌ | ✅ | ✅ |
| Unlimited Clients | ❌ | ✅ | ✅ |
| Unlimited Invoices | ❌ | ✅ | ✅ |
| Multiple Workspace | ❌ | ✅ (2) | ✅ (4) |
| Multiple Members | ❌ | ✅ (5) | ✅ (15) |
| Comments di Task | ❌ | ✅ | ✅ |
| Real-time Updates (SSE) | ❌ | ✅ | ✅ |
| Audit Log | ❌ | ❌ | ✅ |

---

*Dokumen ini tidak bersifat teknis. Untuk detail API, lihat Swagger docs di `/swagger/index.html` atau TECHNICAL.md.*