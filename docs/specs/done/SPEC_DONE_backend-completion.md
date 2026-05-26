# ACTIVE_SPEC: Backend API Completion

## 📌 Overview
- **Fitur:** Completion endpoint backend yang missing (Auth, Profile, Invoice, License)
- **Prioritas:** Must Have
- **Fase:** Phase 1 & 6 & 7 — Auth / Invoice / License
- **Tanggal Dibuat:** 2026-05-25

## 🎯 Goals
Berdasarkan gap analysis antara docs.go / swagger.json (backend Go) dan MIGRATION-FIREBASE-BACKEND-MODEL.md (kebutuhan frontend), terdapat 8 endpoint yang belum diimplementasi. PRD ini mengumpulkan seluruh pekerjaan backend yang missing dalam satu spec agar dapat dikerjakan secara terkoordinasi.

1. Melengkapi alur Auth end-to-end (reset password, logout, update profil, ganti password)
2. Melengkapi fitur Invoice dengan pencatatan pembayaran yang akurat
3. Mengimplementasi seluruh sistem License Key yang belum ada sama sekali

## 👤 User Story
> Sebagai user, saya ingin dapat mereset password, mengubah profil, dan mengaktifkan license key, agar pengalaman saya menggunakan aplikasi lengkap dan aman.
>
> Sebagai freelancer, saya ingin menandai invoice sebagai lunas dengan mencatat tanggal dan jumlah diterima, agar laporan keuangan saya akurat.
>
> Sebagai admin, saya ingin membuat dan mendistribusikan license key secara bulk, agar proses onboarding user berbayar bisa berjalan otomatis.

## 🗂️ Task List (untuk Engineer)

Tabel ringkas semua pekerjaan yang harus diselesaikan engineer, berurutan.
Engineer WAJIB update kolom `Status` per baris setelah setiap task selesai.

| # | Task | Domain | Effort | Status |
|---|------|--------|--------|--------|
| T1 | Buat tabel `password_reset_tokens` (token, user_id, expires_at, used_at) | Backend | 2 jam | ⏳ Backlog |
| T2 | Update `POST /forgot-password` — generate & simpan token ke DB, kirim link via email | Backend | 2 jam | ⏳ Backlog |
| T3 | Endpoint `POST /reset-password` — validasi token, hash & simpan password baru, invalidate token | Backend | 3 jam | ⏳ Backlog |
| T4 | Endpoint `PATCH /users/me` — partial update `name` dan field profil lainnya (tanpa avatar) | Backend | 2 jam | ⏳ Backlog |
| T5 | Endpoint `PATCH /users/me/password` — validasi `current_password`, simpan `new_password` (hashed) | Backend | 2 jam | ⏳ Backlog |
| T6 | Tambah kolom `amount_paid` dan `paid_at` ke tabel `invoices` (migration) | Backend | 1 jam | ⏳ Backlog |
| T7 | Endpoint `PATCH /invoices/{id}/mark-paid` — terima `{ amount_paid, paid_at? }`, set `status='paid'` | Backend | 2 jam | ⏳ Backlog |
| T8 | Update `GET /invoices` dan `GET /invoices/{id}` — sertakan `amount_paid` dan `paid_at` di response | Backend | 1 jam | ⏳ Backlog |
| T9 | Update `total_revenue` di tabel `clients` secara atomik saat invoice ditandai paid | Backend | 1 jam | ⏳ Backlog |
| T10 | Buat tabel `licenses` (id, key, type, status, activated_by, activated_at, expires_at, created_at) | Backend | 1 jam | ⏳ Backlog |
| T11 | Tambah kolom `plan`, `license_key`, `license_status` ke tabel `users` (migration) | Backend | 1 jam | ⏳ Backlog |
| T12 | Endpoint `POST /api/licenses/validate` — cek format & status key, tanpa auth | Backend | 2 jam | ⏳ Backlog |
| T13 | Endpoint `POST /api/licenses/activate` — validasi key, update license + user plan dalam satu transaksi | Backend | 3 jam | ⏳ Backlog |
| T14 | Endpoint `POST /api/licenses` (admin) — bulk insert license keys, require `x-admin-secret` header | Backend | 2 jam | ⏳ Backlog |
| T15 | Update Swagger / docs.go untuk semua endpoint baru | Backend | 1 jam | ⏳ Backlog |

---

## Grouping Task
**— Auth —** : T1 - T3
**— Profile —** : T4 - T5
**— Invoice —** : T6 - T9
**— License —** : T10 - T14
**— Docs —** : T15

---

## ✅ Acceptance Criteria & 🧪 Test Scenarios

### `POST /reset-password`

**Acceptance Criteria:**
- [ ] Menerima `{ token, new_password }`, mengembalikan 200 jika berhasil
- [ ] Token expired (> 1 jam) → 400
- [ ] Token yang sudah dipakai → 400
- [ ] Password baru tersimpan sebagai bcrypt hash, bukan plaintext
- [ ] Token lama tidak bisa digunakan kembali setelah reset berhasil

**Test Scenarios:**
1. **Happy path** — token valid + password baru valid → 200, password terupdate, token diinvalidasi
2. **Token expired** — token lebih dari 1 jam → 400 `"Token has expired"`
3. **Token sudah dipakai** — `used_at` tidak null → 400 `"Token already used"`
4. **Token tidak dikenal** — random string → 400 `"Invalid token"`
5. **Password terlalu pendek** — kurang dari 8 karakter → 400 validation error

---

### `PATCH /users/me`

**Acceptance Criteria:**
- [ ] Menerima body partial — hanya field yang dikirim yang diupdate
- [ ] Membutuhkan Bearer token valid → 401 jika tidak ada
- [ ] Mengembalikan data user terbaru setelah update
- [ ] Field `email` tidak bisa diubah lewat endpoint ini (diabaikan jika dikirim)
- [ ] Field `avatar` tidak tersedia di endpoint ini
- [ ] `name` string kosong → 400

**Test Scenarios:**
1. **Update name** — `{ name: "Budi Santoso" }` → 200, name terupdate
2. **Tanpa token** — → 401
3. **Body kosong** — `{}` → 200, tidak ada yang berubah, tidak error
4. **Coba ubah email** — field `email` diabaikan, email tidak berubah

---

### `PATCH /users/me/password`

**Acceptance Criteria:**
- [ ] Menerima `{ current_password, new_password }`
- [ ] `current_password` salah → 400 `"Current password is incorrect"`
- [ ] `new_password` minimal 8 karakter
- [ ] `new_password` sama dengan `current_password` → 400
- [ ] Password baru tersimpan sebagai hash
- [ ] Membutuhkan Bearer token valid

**Test Scenarios:**
1. **Happy path** — current_password benar + new_password valid → 200
2. **Password lama salah** — → 400 `"Current password is incorrect"`
3. **Password baru terlalu pendek** — < 8 karakter → 400
4. **Password baru sama dengan lama** — → 400 `"New password must be different"`
5. **Tanpa token** — → 401

---

### `PATCH /invoices/{id}/mark-paid`

**Acceptance Criteria:**
- [ ] Menerima `{ amount_paid, paid_at? }`
- [ ] Status invoice berubah menjadi `paid` otomatis
- [ ] `paid_at` default ke `NOW()` jika tidak dikirim
- [ ] `amount_paid` tidak boleh negatif → 400
- [ ] Invoice berstatus `cancelled` tidak bisa di-mark-paid → 400
- [ ] Field `amount_paid` dan `paid_at` muncul di response `GET /invoices/{id}`

**Test Scenarios:**
1. **Happy path** — invoice status `sent`, kirim `{ amount_paid: 500000 }` → 200, status jadi `paid`
2. **Dengan paid_at eksplisit** — `{ amount_paid: 500000, paid_at: "2026-05-20T..." }` → 200, paid_at tersimpan
3. **Invoice cancelled** — → 400 `"Cannot mark a cancelled invoice as paid"`
4. **amount_paid negatif** — → 400 validation error
5. **Invoice tidak ditemukan** — → 404
6. **Invoice milik user lain** — → 404 (bukan 403, untuk keamanan)

---

### `POST /api/licenses/validate`

**Acceptance Criteria:**
- [ ] Mengembalikan `{ valid, message, license: { id, type } }` tanpa auth
- [ ] Format key salah → 400 tanpa hit database
- [ ] Key tersedia → `{ valid: true }`
- [ ] Key sudah activated / revoked / expired → `{ valid: false, message: "..." }`

**Test Scenarios:**
1. **Key valid tersedia** — → 200 `{ valid: true, license: { type: "pro" } }`
2. **Key sudah aktif** — → 200 `{ valid: false, message: "Already activated" }`
3. **Key revoked** — → 200 `{ valid: false, message: "License has been revoked" }`
4. **Format salah** — → 400 format error

---

### `POST /api/licenses/activate`

**Acceptance Criteria:**
- [ ] Membutuhkan Bearer token valid → 401 jika tidak ada
- [ ] Update `licenses` dan `users.plan` dalam satu DB transaction
- [ ] Key yang sudah activated / revoked / expired → 400
- [ ] Setelah aktivasi, `GET /api/auth/me` mengembalikan plan terbaru

**Test Scenarios:**
1. **Happy path** — key available + user login → 200, user plan jadi `pro`
2. **Key sudah dipakai** — → 400 `"License already activated"`
3. **Key tidak ditemukan** — → 400 `"License key not found"`
4. **Tanpa token** — → 401

---

### `POST /api/licenses` (admin)

**Acceptance Criteria:**
- [ ] Membutuhkan header `x-admin-secret` yang valid → 401 jika salah
- [ ] Menerima array of `{ key, type }` untuk bulk insert
- [ ] Mengembalikan hasil per key: `{ key, status: "created" | "error", id? }`
- [ ] Duplicate key tidak crash, dikembalikan sebagai error per baris

**Test Scenarios:**
1. **Happy path** — secret valid + 10 keys → 200, semua terbuat
2. **Ada duplicate key** — salah satu key sudah ada → baris itu status `"error"`, sisanya tetap created
3. **Secret salah** — → 401
4. **Tanpa secret header** — → 401

---

## 🎨 UX Guidelines
*(Pure Backend — skip)*

## ⚠️ Catatan Teknis & Batasan
- **Reset token** — generate dengan `crypto/rand`, bukan `math/rand`. Expiry 1 jam, single-use
- **Logout** — untuk MVP, logout cukup dilakukan di sisi client (hapus token dari storage). Blacklist token di server tidak diimplementasi di fase ini
- **License activation** — wajib atomic transaction (update `licenses` + update `users` dalam satu TX)
- **License key normalization** — selalu `toUpperCase().trim()` sebelum simpan dan query
- **`x-admin-secret`** — simpan di env var, jangan di-log, jangan masuk response error
- **Invoice mark-paid** — update `clients.total_revenue` dalam transaksi yang sama dengan update invoice
- **Partial payment** — konfirmasi ke product owner: apakah `amount_paid < amount` perlu status `partial`? Jika ya, tambah enum baru di kolom `status`
- **Dependency order** — T1→T2→T3 (reset password) harus selesai sebelum frontend bisa test; T10→T11 (license tables) harus selesai sebelum T12–T14

## 🔜 Next Steps
- Setelah spec ini selesai, lanjut ke **Frontend Integration** — update semua hooks dan service layer untuk consume endpoint baru ini
- Koordinasi dengan frontend soal URL halaman reset password (`/reset-password?token=xxx`) agar link di email mengarah ke tempat yang benar
EOF