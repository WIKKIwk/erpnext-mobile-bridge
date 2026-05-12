# Accord Mobile Server Rust Porting Daftari

Bu hujjat `accord_mobile_server` Go loyihasini yonida yangi `accord_mobile_server_rs` sifatida Rust'da tartibli va 1:1 mos ishlaydigan qilib qayta yozish uchun ish daftari.

Asosiy qoida: birdan hamma narsani yozmaymiz. Avval Go skeletini tushunamiz, keyin bitta bo'limni Rust'da toza, mayda, review qilish oson fayllarga ajratib yozamiz.

## Maqsad

- Go serverning hozirgi API va biznes xatti-harakatini 1:1 saqlash.
- Rust versiyada katta fayllar va aralash responsibility'larni takrorlamaslik.
- Har bir bo'limni alohida port qilish, tekshirish va hujjatlashtirish.
- Fayllarni imkon qadar 500 qatordan oshirmaslik.
- Eski `data/*.json` formatlari bilan compatibility saqlash.
- `.env` sozlamalarini imkon qadar Go server bilan mos saqlash.

## Ishlash Qoidalari

- Avval audit, keyin kod.
- Har safar bitta kichik scope tanlanadi.
- Yangi Rust papkalar faqat kerak bo'lganda ochiladi.
- Katta domain fayllar bo'linadi: handler, service, model, store, adapter alohida.
- Rust kodda `Arc<AppState>` kabi markaziy state ishlatiladi, lekin domain service'lar bir-biriga haddan tashqari yopishib ketmaydi.
- API response field nomlari o'zgartirilmaydi.
- Har port qilingan bo'lim uchun kamida smoke test yoki unit test yoziladi.

## Go Loyihaning Hozirgi Package Skeleti

```text
cmd/core
cmd/import_acp_products
cmd/import_customer_items

internal/config
internal/core
internal/erpdb
internal/erpnext
internal/importacp
internal/importitems
internal/mobileapi
internal/suplier
```

## Package Ma'nolari

- `cmd/core` - asosiy mobile server entrypoint.
- `cmd/import_acp_products` - ACP product import CLI.
- `cmd/import_customer_items` - customer item import CLI.
- `internal/config` - `.env`, config, env persister.
- `internal/core` - biznes mantiqning asosiy markazi.
- `internal/erpdb` - ERPNext bazasidan direct read.
- `internal/erpnext` - ERPNext HTTP API client/adapters.
- `internal/importacp` - ACP import logikasi.
- `internal/importitems` - customer item import logikasi.
- `internal/mobileapi` - HTTP route/handler layer.
- `internal/suplier` - supplier auth/storage/access code logikasi. Nomi typo, lekin kodda consistent.

## Eng Katta Fayllar va Tartibsizlik Nuqtalari

| Fayl | Qator | Muammo |
| --- | ---: | --- |
| `internal/core/service.go` | 3134 | Auth, profile, supplier, customer, Werka, session va helperlar aralash |
| `internal/mobileapi/server.go` | 2498 | Route va deyarli hamma HTTP handler bitta faylda |
| `internal/erpdb/reader.go` | 1876 | Ko'p direct DB read query va mapperlar bir joyda |
| `internal/core/admin_suppliers.go` | 1174 | Supplier, customer, item admin boshqaruvi aralash |
| `internal/erpnext/purchase_receipt.go` | 1144 | Purchase receipt lifecycle juda katta |
| `internal/erpnext/client.go` | 949 | Generic ERP client, search, stock entry va helperlar aralash |
| `internal/erpnext/delivery_note.go` | 776 | Delivery note lifecycle va remark parserlar bir faylda |
| `internal/mobileapi/werka_ai_search.go` | 572 | AI search handler, Gemini client va parsing bitta joyda |

## Rust'da Takrorlamaydigan Narsalar

- `core/service.go` kabi bitta gigant service fayl qilinmaydi.
- `mobileapi/server.go` kabi hamma handler bir faylga yig'ilmaydi.
- ERPNext client ham bitta universal fayl bo'lib ketmaydi.
- Direct DB read'lar domain bo'yicha bo'linadi.
- Testlar bitta ulkan integration test fayliga yig'ilmaydi.

## Rust Uchun Tavsiya Qilingan Bo'linish

```text
accord_mobile_server_rs/
  Cargo.toml
  README.md
  PORTING_PLAN.md
  src/
    main.rs
    app.rs
    config.rs
    error.rs
    http/
      mod.rs
      router.rs
      handlers/
        auth.rs
        profile.rs
        supplier.rs
        customer.rs
        werka.rs
        admin.rs
    core/
      auth/
      session/
      profile/
      supplier/
      customer/
      werka/
      admin/
    erpnext/
      client.rs
      suppliers.rs
      customers.rs
      items.rs
      stock_entries.rs
      purchase_receipts.rs
      delivery_notes.rs
    erpdb/
      reader.rs
      credentials.rs
      suppliers.rs
      customers.rs
      notifications.rs
      stock_entries.rs
    store/
      json_file.rs
      profile_store.rs
      session_store.rs
      push_token_store.rs
      admin_state_store.rs
    ai/
      werka_search.rs
```

Bu to'liq skelet birdan yaratilmaydi. Har bo'lim kerak bo'lganda qo'shiladi.

## Dastlabki Rust Scope

Birinchi port qilish uchun eng ma'qul kichik scope:

- `auth`
- `session`
- `profile`
- JSON store bazasi

Sabab:

- Serverga kirish eshigi shu.
- Supplier/customer/Werka/admin flow'larga nisbatan ixchamroq.
- Session va profile JSON compatibility erta tekshiriladi.
- Keyingi endpointlar uchun auth middleware tayyor bo'ladi.

## Birinchi Scope Endpointlari

- `POST /v1/mobile/auth/login`
- `POST /v1/mobile/auth/logout`
- `GET /v1/mobile/auth/me`
- `GET /v1/mobile/profile`
- `PUT /v1/mobile/profile`
- `POST /v1/mobile/profile/avatar`
- `GET /v1/mobile/profile/avatar/view`

Avatar upload va view supplier-specific bo'lgani uchun kerak bo'lsa ikkinchi qadamga qoldiriladi.

## Framework Qarorlari

Hozirgi tavsiya:

- Web framework: `axum`
- Runtime: `tokio`
- HTTP client: `reqwest`
- JSON: `serde`, `serde_json`
- Env: `dotenvy`
- Error: `thiserror`
- Logging: `tracing`, `tracing-subscriber`
- Direct DB: keyinroq `sqlx` yoki `mysql_async` tanlanadi

## Tanlangan Birinchi Scope

Birinchi yoziladigan kichik bo'lim: `auth + session` poydevori.

Bu scope ichida hozircha ERPNext login to'liq port qilinmaydi. Avval quyidagilar tayyorlanadi:

- Rust project skeleton.
- Config loader.
- Axum router.
- Auth endpoint joylari.
- Go session JSON formatiga mos session model.
- Persistent session manager.
- JSON file helper.

`profile` keyingi kichik qadam bo'ladi.

## Qilingan Ishlar

- `accord_mobile_server_rs` papkasi ochildi.
- Minimal Rust project yaratildi.
- `axum`, `tokio`, `serde`, `time`, `tracing` asosiy dependency'lari qo'shildi.
- `MOBILE_API_ADDR` Go uslubidagi `:8081` qiymatini Rust `SocketAddr`ga normalize qiladigan config loader yozildi.
- `/healthz`, `/v1/mobile/auth/login`, `/v1/mobile/auth/logout`, `/v1/mobile/auth/me` route joylari ochildi.
- Session token uzunligi Go bilan mos bo'lishi uchun 24 random byte + base64url no padding usuli tanlandi.
- Session JSON modeli Go field nomlari bilan mos yozildi: `principal`, `created_at`, `updated_at`, `expires_at`.
- `cargo check` va `cargo test` ishga tushirildi.
- Go contract bo'yicha `/v1/mobile/me` route tuzatildi.
- Go contract bo'yicha `MOBILE_API_SESSION_STORE_PATH` env nomi asosiy qilindi.
- Rust auth service'ga ERPNextsiz ishlaydigan admin login qo'shildi.
- Rust auth service'ga code-driven Werka login qo'shildi.
- Session manager'ga `update` qo'shildi.
- Logout bearer token yo'q bo'lsa `401` qaytaradigan qilindi.
- `/v1/mobile/me` route uchun regressiya testi qo'shildi.
- Testlar 8 taga yetdi.
- Supplier login uchun Rust port interface'lari qo'shildi.
- Supplier deterministic access code Go algoritmiga mos portlandi.
- Supplier login `SearchSuppliers(phone, 50)` va fallback `SearchSuppliers("", 500)` tartibida yozildi.
- Supplier `blocked/removed/custom_code` state qoidalari test bilan yopildi.
- Testlar 11 taga yetdi.
- ERPNext supplier search runtime adapter qo'shildi.
- Go formatidagi `mobile_admin_suppliers.json` o'qiydigan admin state store qo'shildi.
- `ERP_URL`, `ERP_API_KEY`, `ERP_API_SECRET`, `ERP_TIMEOUT_SECONDS` config'lari Rust runtime'ga qo'shildi.
- ERP config to'liq bo'lsa supplier login runtime wiring auth service'ga ulanadigan qilindi.
- Testlar 13 taga yetdi.

## Progress Checklist

- [x] Go package skeleti sanaldi.
- [x] Eng katta Go fayllar va tartibsizlik nuqtalari belgilandi.
- [x] Rust porting uchun umumiy bo'lish qoidalari yozildi.
- [ ] Go endpoint matrix chiqarish.
- [ ] Auth/session/profile uchun aniq Go behavior xaritasini yozish.
- [x] `accord_mobile_server_rs` papkasini yaratish.
- [x] Rust minimal server skeleton.
- [x] Rust config loader.
- [x] Rust session JSON store.
- [ ] Rust profile JSON store.
- [x] Auth endpointlarining birinchi porti.
- [ ] Profile endpointlarining birinchi porti.
- [x] Cargo test/check.

## Ochiq Savollar

- Rust versiya Go server bilan bir xil port/env nomlarini ishlatadimi yoki `*_RS` prefikslari ham bo'ladimi?
- Direct DB uchun `sqlx` tanlaymizmi yoki `mysql_async`?
- Avatar upload birinchi scope ichida bo'ladimi yoki profile basicdan keyinmi?
- Admin hardcoded identity Rust'da ham aynan shunday qoladimi yoki konfiguratsiyaga chiqariladimi?

## Auth/Session Go Behavior Matrix

Bu bo'lim Go serverdagi auth/session xatti-harakatini Rust port uchun 1:1 contract sifatida yozadi.

### Endpointlar

| Method | Path | Handler | Behavior |
| --- | --- | --- | --- |
| `POST` | `/v1/mobile/auth/login` | `handleLogin` | Login qiladi, session yaratadi, token/profile qaytaradi |
| `POST` | `/v1/mobile/auth/logout` | `handleLogout` | Bearer token bo'lsa sessionni o'chiradi |
| `GET` | `/v1/mobile/me` | `handleMe` | Bearer token orqali principalni qaytaradi |

Muhim: Go serverda `me` endpoint `/v1/mobile/me`. `/v1/mobile/auth/me` emas.

### Login Request/Response

Request:

```json
{
  "phone": "+998901234567",
  "code": "10..."
}
```

Response:

```json
{
  "token": "...",
  "profile": {
    "role": "supplier",
    "display_name": "Abdulloh",
    "legal_name": "Abdulloh",
    "ref": "SUP-001",
    "phone": "+998901234567",
    "avatar_url": "..."
  },
  "werka_home": null
}
```

`werka_home` faqat role `werka` bo'lganda va `WerkaHome` muvaffaqiyatli qaytsa qo'shiladi. Aks holda `omitempty` sabab response'da yo'q.

### Login Handler Tartibi

1. Method `POST` bo'lmasa `405 {"error":"method not allowed"}`.
2. JSON decode xato bo'lsa `400 {"error":"invalid json"}`.
3. `phone` va `code` `strings.TrimSpace` qilinadi.
4. `ERPAuthenticator.Login(ctx, phone, code)` chaqiriladi.
5. `ErrInvalidCredentials` yoki `ErrInvalidRole` bo'lsa `401 {"error":"invalid credentials"}`.
6. Boshqa login xatosi bo'lsa `500 {"error":"internal error"}`.
7. Login muvaffaqiyatli bo'lsa `Profile(ctx, principal)` bilan refresh qilinadi. Bu xato bersa login baribir davom etadi.
8. Session yaratiladi. Xato bo'lsa `500 {"error":"session create failed"}`.
9. Principal role `werka` bo'lsa `WerkaHome(ctx, 20)` preload qilinadi. Xato bo'lsa e'tiborsiz qoldiriladi.
10. `200 LoginResponse` qaytariladi.

### Telefon Normalizatsiya

Login boshida `suplier.NormalizePhone` ishlaydi:

- Bo'sh qiymat xato.
- `+` belgisi olib tashlanadi.
- Qolgan hamma belgi raqam bo'lishi kerak.
- Raqamlar soni kamida 9, ko'pi bilan 12.
- Natija doim `+` bilan qaytadi.

Misol:

- `+998901234567` -> `+998901234567`
- `998901234567` -> `+998901234567`
- `+12345` -> invalid credentials

### Role Aniqlash

Admin login prefix orqali emas, birinchi alohida tekshiriladi.

| Role | Signal |
| --- | --- |
| `admin` | normalized phone admin phone bilan teng va code admin code bilan teng |
| `supplier` | code supplier prefix bilan boshlanadi, default `10` |
| `werka` | code Werka prefix bilan boshlanadi, default `20` |
| `customer` | code `30` bilan boshlanadi |

Role aniqlanmasa `ErrInvalidRole`, HTTP response esa `401 {"error":"invalid credentials"}`.

### Admin Login

`cmd/core/main.go` hozir admin identity'ni hardcode qiladi:

- phone: `+998880000000`
- name: `Admin`
- code: `19621978`

Admin login muvaffaqiyatli bo'lsa principal:

```json
{
  "role": "admin",
  "display_name": "Admin",
  "legal_name": "Admin",
  "ref": "admin",
  "phone": "+998880000000"
}
```

Admin login ERPNext yoki JSON state'ga bormaydi.

### Supplier Login

Supplier login tartibi:

1. `SearchSuppliers(normalizedPhone, 50)` chaqiriladi.
2. Natija bo'lmasa `SearchSuppliers("", 500)` fallback.
3. `AdminSupplierStore.List()` orqali state map olinadi.
4. Har supplier uchun state tekshiriladi.
5. `Removed` yoki `Blocked` bo'lsa o'tkazib yuboriladi.
6. Access code aniqlanadi.
7. Code teng bo'lsa va supplier phone normalized phone bilan exact case-insensitive teng bo'lsa login muvaffaqiyatli.
8. Principal profile prefs bilan merge qilinadi.

Supplier access code:

- Agar `AdminSupplierState.CustomCode` bo'lsa, shu ishlatiladi.
- Aks holda `suplier.GenerateAccessCredentials` deterministic code yaratadi.
- Deterministic supplier code har doim `10` prefix bilan boshlanadi.
- Seed tartibi: supplier `Ref`, bo'lmasa normalized phone, bo'lmasa name.

Supplier principal:

```json
{
  "role": "supplier",
  "display_name": "Supplier Name",
  "legal_name": "Supplier Name",
  "ref": "SUP-001",
  "phone": "+998901234567"
}
```

### Customer Login

Customer login tartibi:

1. `SearchCustomers(normalizedPhone, 50)` chaqiriladi.
2. Natija bo'lmasa `SearchCustomers("", 500)` fallback.
3. `AdminSupplierStore.List()` orqali state map olinadi.
4. Customer ID bilan state topiladi.
5. `state.CustomCode` bo'sh bo'lsa customer o'tkazib yuboriladi.
6. Code teng bo'lsa va customer phone normalized phone bilan exact case-insensitive teng bo'lsa login muvaffaqiyatli.
7. Principal profile prefs bilan merge qilinadi.

Muhim: customer login deterministic code yaratmaydi. Customer kirishi uchun `AdminSupplierState.CustomCode` oldindan bor bo'lishi kerak.

Customer principal:

```json
{
  "role": "customer",
  "display_name": "Customer Name",
  "legal_name": "Customer Name",
  "ref": "CUST-001",
  "phone": "+998901234567"
}
```

### Werka Login

Werka login tartibi:

1. Code `werkaPrefix` bilan boshlansa role `werka` bo'ladi.
2. Keyin code aynan `a.werkaCode` ga teng va bo'sh emasligi tekshiriladi.
3. Phone faqat normalize qilinadi va principalga yoziladi.
4. Werka account code-driven. Phone configured Werka phone bilan solishtirilmaydi.

Test bilan tasdiqlangan behavior: `+99888862440` va `+123456789` kabi boshqa valid phone formatlar ham to'g'ri Werka code bilan kira oladi.

Werka principal:

```json
{
  "role": "werka",
  "display_name": "Werka",
  "legal_name": "Werka",
  "ref": "werka",
  "phone": "+998..."
}
```

### Profile Refresh Login/Me Ichida

Login va `me` ikkalasi ham `Profile(ctx, principal)` chaqiradi.

- Supplier profile refresh `GetSupplier` qiladi, phone va avatar URL yangilaydi.
- Customer profile refresh `GetCustomer` qiladi, phone yangilaydi.
- Admin/Werka uchun ERP refresh yo'q.
- Profile prefs doim `role:ref` key bilan merge qilinadi.
- Nickname bo'lsa `display_name` ustiga yoziladi.
- AvatarURL prefs'da bo'lsa `avatar_url` ustiga yoziladi.

### Avatar Proxy

Response'da supplier `avatar_url` bo'lsa URL proxy qilinadi:

```text
{scheme}://{host}/v1/mobile/profile/avatar/view?token={token}
```

Faqat supplier va `ref` bo'lsa proxy qilinadi. Boshqa role uchun original avatar URL qoladi.

### Session Behavior

Session manager:

- Token: 24 random byte, `base64.RawURLEncoding`, padding yo'q.
- Token uzunligi amalda 32 belgi.
- Persistent store path default: `data/mobile_sessions.json`.
- Env nomi: `MOBILE_API_SESSION_STORE_PATH`.
- TTL default: `30 * 24h`.
- TTL env: `MOBILE_API_SESSION_TTL_HOURS`.
- TTL `0` bo'lsa expiry yozilmaydi.
- Manfiy TTL startup'da fatal error.

Session JSON record:

```json
{
  "principal": {},
  "created_at": "2026-01-16T12:00:00Z",
  "updated_at": "2026-01-16T12:00:00Z",
  "expires_at": "2026-02-15T12:00:00Z"
}
```

Session file format:

```json
{
  "TOKEN": {
    "principal": {},
    "created_at": "...",
    "updated_at": "...",
    "expires_at": "..."
  }
}
```

Session operationlar:

- `Create` file'ni lazy load qiladi, expired sessionlarni tozalaydi, yangi token yozadi.
- `Get` token topilmasa false. Expired bo'lsa o'chirib false qaytaradi.
- `Delete` token bor bo'lsa o'chiradi, yo'q bo'lsa silent.
- `Update` token bor bo'lsa principalni yangilaydi, `created_at` saqlanadi, `updated_at` va `expires_at` yangilanadi.
- Write atomic: temp file yozib, keyin rename.

### Logout Behavior

Logout:

1. Method `POST` bo'lmasa `405`.
2. `Authorization: Bearer TOKEN` bo'lmasa yoki token bo'sh bo'lsa `401 {"error":"unauthorized"}`.
3. Token sessionda bo'lmasa ham `Delete` silent ishlaydi.
4. Response doim `200 {"ok":true}` agar bearer token format valid bo'lsa.

### Me Behavior

`GET /v1/mobile/me`:

1. Bearer token talab qiladi.
2. Token yo'q, bo'sh yoki invalid bo'lsa `401 {"error":"unauthorized"}`.
3. Session topilsa `Profile` refresh qiladi.
4. Profile refresh muvaffaqiyatli bo'lsa session `Update` qilinadi.
5. Principal avatar proxy bilan `200` qaytariladi.

### Rust Port Uchun Darhol Tuzatiladigan Nuqtalar

- Rust skeletonda `GET /v1/mobile/auth/me` ochilgan. Go contract bo'yicha bu `/v1/mobile/me` bo'lishi kerak.
- Rust skeletonda env `MOBILE_API_SESSION_STORE` ishlatilgan. Go contract bo'yicha `MOBILE_API_SESSION_STORE_PATH` bo'lishi kerak.
- Rust `SessionRecord::new` hozir `created_at`ni har doim yangi vaqt qiladi. `Update` qo'shilganda Go kabi eski `created_at` saqlanishi kerak.
- Rust `logout` hozir bearer token bo'lmasa ham `200` qaytaradi. Go contract bo'yicha bearer format yo'q bo'lsa `401`.

Status: yuqoridagi 4 ta nuqta Rust kodda tuzatildi.

### Auth/Session Rust Port Status

Portlangan:

- Admin login.
- Werka login.
- Supplier login service logic.
- Supplier deterministic access code.
- Supplier lookup/admin state port interface'lari.
- Supplier login runtime ERPNext adapter wiring.
- Supplier admin state JSON store.
- Phone normalization.
- Role prefix inference.
- Login response skeleton.
- Session create/get/delete/update.
- Logout bearer validation.
- `/v1/mobile/me` route.
- `/v1/mobile/auth/me` route yo'qligi testi.

Hali portlanmagan:

- Supplier login real ERPNext smoke test.
- Customer login ERPNext search + custom code state.
- Login ichidagi `Profile` refresh.
- Supplier avatar proxy.
- Werka login response ichidagi `werka_home` preload.

## Keyingi Qadam

Go endpoint matrix chiqariladi. Har endpoint uchun quyidagilar yoziladi:

- method/path
- Go handler
- core service method
- ERPNext/direct DB dependency
- JSON store dependency
- Rust'dagi kelajak modul joyi
