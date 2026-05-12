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
- Real ERPNext smoke testda vaqtinchalik supplier yaratilib, Rust supplier login `200/supplier/token` bilan tasdiqlandi va cleanup qilindi.
- Customer login core logic qo'shildi.
- Customer login uchun custom code majburiyligi test bilan yopildi.
- ERPNext customer search runtime adapter qo'shildi.
- Real ERPNext smoke testda vaqtinchalik customer va temp admin state bilan Rust customer login `200/customer/token` bilan tasdiqlandi va cleanup qilindi.
- Testlar 16 taga yetdi.
- Profile refresh service qo'shildi.
- Login va `/v1/mobile/me` ichida profile refresh ulanib, refresh muvaffaqiyatli bo'lsa session update qilinadigan bo'ldi.
- ERPNext `GetSupplier` refresh adapteri supplier phone va image fieldlarini o'qiydi.
- ERPNext `GetCustomer` refresh adapteri customer phone fieldini o'qiydi.
- Real ERPNext smoke testda supplier/customer login response profile refresh bilan `200` qaytgani tasdiqlandi va cleanup qilindi.
- Testlar 18 taga yetdi.
- Supplier avatar proxy login va `/v1/mobile/me` response'lariga qo'shildi.
- Supplier avatar proxy token URL encoding bilan test qilindi.
- Customer avatar URL proxy qilinmasligi test bilan yopildi.
- Testlar 20 taga yetdi.
- Supplier avatar view/download endpoint qo'shildi.
- Avatar view query token va Bearer token auth yo'llarini qo'llaydi.
- Supplier bo'lmagan principal uchun avatar view `403` qaytarishi test qilindi.
- Authsiz avatar view `401` qaytarishi test qilindi.
- ERPNext file download adapteri profile lookup orqali ulandi.
- Testlar 23 taga yetdi.
- Werka login `werka_home` preload uchun Rust typed model qo'shildi.
- Go'dagi `WerkaHomeSummary`, `WerkaHomeData` va `DispatchRecord` JSON field shakllari Rust modeliga ko'chirildi.
- Login response ichida `role=werka` bo'lsa `WerkaService.home(20)` preload qilinadigan bo'ldi.
- Go'dagi kabi Werka home preload xato yoki data yo'q bo'lsa login muvaffaqiyatli davom etadi va `werka_home` response'dan omit qilinadi.
- `werka_home` bor/yo'qligi JSON serialization regressiya testlari bilan yopildi.
- WerkaHome data source uchun port/interface ajratildi, lekin direct DB/ERP runtime provider hali keyingi scope sifatida qolmoqda.
- Testlar 27 taga yetdi.
- WerkaHome direct DB runtime provider Rust'da `sqlx/mysql` bilan qo'shildi.
- Go'dagi `ERP_DIRECT_READ_ENABLED=1` va `ERP_DIRECT_SITE_CONFIG_PATH` env kontrakti Rust config'ga ko'chirildi.
- Frappe `site_config.json` dan `db_name`, `db_password`, `db_type` o'qish Go'dagi kabi port qilindi.
- WerkaHome query'lari `tabPurchase Receipt` + `tabPurchase Receipt Item` va `tabDelivery Note` + `tabDelivery Note Item` ustida Go selectlariga mos yozildi.
- Purchase Receipt status klassifikatsiyasi, Telegram marker qty, `Accord Werka Aytilmagan` hide qoidalari va Delivery Note customer state mapping test bilan yopildi.
- Go'dagi pending limit pre-append behavior ham test bilan saqlandi.
- Real ERPNext smoke testda Werka login `200/werka/token` qaytardi va `werka_home` direct DB provider orqali response'da borligi tasdiqlandi.
- Testlar 31 taga yetdi.
- `GET /v1/mobile/werka/home` endpoint Rust'da qo'shildi.
- Endpoint Go'dagi kabi Bearer session auth talab qiladi va faqat `role=werka` uchun `200` qaytaradi.
- Authsiz holat `401 {"error":"unauthorized"}`, boshqa role `403 {"error":"forbidden"}`, provider/data xatosi `500 {"error":"werka home failed"}` bilan test qilindi.
- Go handler method check qilmagani uchun Rust route ham `any` qilib qo'yildi; `POST /v1/mobile/werka/home` ham Go kabi ishlashi regressiya test bilan yopildi.
- Real ERPNext smoke testda `GET` va `POST /v1/mobile/werka/home` ikkalasi ham Werka token bilan `summary/pending_items` response qaytargani tasdiqlandi.
- Testlar 36 taga yetdi.
- `/v1/mobile/werka/summary` endpoint Rust'da qo'shildi.
- Summary direct DB provider Go'dagi status-only querylarga mos alohida yozildi.
- Purchase Receipt status row va Delivery Note status row mapping test bilan yopildi.
- Summary endpoint authsiz `401`, provider yo'q bo'lsa `500 {"error":"werka summary failed"}`, success response esa `pending_count/confirmed_count/returned_count` bo'lishi test qilindi.
- Go handler method check qilmagani uchun summary route ham `any` qilindi; `POST /v1/mobile/werka/summary` regressiya test bilan yopildi.
- Real ERPNext smoke testda `GET` va `POST /v1/mobile/werka/summary` ikkalasi ham Werka token bilan summary response qaytargani tasdiqlandi.
- Testlar 44 taga yetdi.
- `/v1/mobile/werka/pending` endpoint Rust'da qo'shildi.
- Pending direct DB provider Go'dagi `pendingTelegramReceiptRows` va `pendingDeliveryNoteRows` querylariga mos alohida yozildi.
- Pending builder Go'dagi filter/sort/final limit behavior bilan test qilindi.
- Pending endpoint authsiz `401`, provider yo'q bo'lsa `500 {"error":"pending fetch failed"}`, success response array bo'lishi test qilindi.
- Go handler method check qilmagani uchun pending route ham `any` qilindi; `POST /v1/mobile/werka/pending` regressiya test bilan yopildi.
- Real ERPNext smoke testda `GET` va `POST /v1/mobile/werka/pending` ikkalasi ham Werka token bilan array response qaytargani tasdiqlandi.
- Testlar 52 taga yetdi.
- Router fayli 500 qator chegarasiga yetgani uchun router testlari `src/http/router_tests.rs`ga ajratildi.
- Production `src/http/router.rs` 33 qatorga tushirildi, behavior o'zgarmadi.
- Refactordan keyin `cargo test` 52 ta test bilan yana pass bo'ldi.
- `/v1/mobile/werka/history` endpoint Rust'da qo'shildi.
- History direct DB provider Go'dagi 120 recent limit, Telegram Purchase Receipt, supplier ack comment va customer delivery result eventlarini birlashtirish qoidalariga mos yozildi.
- Supplier ack event va customer delivery result event mapping test bilan yopildi.
- History endpoint authsiz `401`, provider yo'q bo'lsa `500 {"error":"history fetch failed"}`, success response array bo'lishi test qilindi.
- Go handler method check qilmagani uchun history route ham `any` qilindi; `POST /v1/mobile/werka/history` regressiya test bilan yopildi.
- Real ERPNext smoke testda `GET` va `POST /v1/mobile/werka/history` ikkalasi ham Werka token bilan array response qaytargani tasdiqlandi.
- Testlar 60 taga yetdi.
- `/v1/mobile/werka/status-breakdown` endpoint Rust'da qo'shildi.
- `kind` query param Go'dagi kabi trim qilinadi va faqat `pending`, `confirmed`, `returned` qiymatlari match qiladi; boshqasi bo'sh array qaytaradi.
- Status breakdown direct DB provider Go'dagi Purchase Receipt va Delivery Note record aggregate qoidalariga mos yozildi.
- Supplier bo'yicha group key, receipt count, sent/accepted/returned totals va sort tartibi test bilan yopildi.
- Go handler method check qilmagani uchun status-breakdown route ham `any` qilindi; `POST /v1/mobile/werka/status-breakdown` regressiya test bilan yopildi.
- Real ERPNext smoke testda `pending/confirmed/returned` kindlari uchun `GET` va `POST` array response qaytargani tasdiqlandi.
- Testlar 68 taga yetdi.
- `/v1/mobile/werka/status-details` endpoint Rust'da qo'shildi.
- `kind` va `supplier_ref` query paramlari Go'dagi kabi trim qilinadi; `supplier_ref` bo'sh bo'lsa barcha supplier/customer ref recordlari ko'riladi.
- Status details direct DB provider Go'dagi Purchase Receipt va Delivery Note record filter qoidalariga mos alohida builder modulda yozildi.
- Supplier ref filter case-insensitive ishlashi, `pending/confirmed/returned` kind mappingi va `created_label` bo'yicha newest-first sort tartibi test bilan yopildi.
- Go handler method check qilmagani uchun status-details route ham `any` qilindi; `POST /v1/mobile/werka/status-details` regressiya test bilan yopildi.
- Real ERPNext smoke testda `pending/confirmed/returned` kindlari uchun `GET /v1/mobile/werka/status-details` array response qaytargani tasdiqlandi.
- Testlar 77 taga yetdi.
- `/v1/mobile/werka/notifications` endpoint Rust'da qo'shildi.
- Go'da notifications handler bevosita history handlerga alias bo'lgani uchun Rust route ham `werka::history` handlerga ulandi.
- `GET` va `POST /v1/mobile/werka/notifications` history payload qaytarishi regressiya test bilan yopildi.
- Real ERPNext smoke testda `GET` va `POST /v1/mobile/werka/notifications` ikkalasi ham array response qaytargani tasdiqlandi.
- Testlar 79 taga yetdi.
- `/v1/mobile/werka/archive` endpoint Rust'da qo'shildi.
- Archive model JSON shakli Go'dagi `kind/period/from/to/summary/items`, `record_count` va `totals_by_uom` contractiga mos yozildi.
- `kind` normalize qoidasi: `received/returned` saqlanadi, boshqa qiymatlar `sent`; `period` normalize qoidasi: `daily/monthly/custom` saqlanadi, boshqa qiymatlar `yearly`.
- Archive direct DB provider Go'dagi conditional query tartibiga mos yozildi: `sent` faqat Delivery Note, `received` faqat Purchase Receipt, `returned` ikkalasini o'qiydi.
- Date filter qoidalari Go bilan mos: Purchase Receipt `posting_date` inclusive, Delivery Note `modified >= from 00:00:00` va `< to+1 day 00:00:00`.
- Archive summary uchun UOM bo'yicha total, empty UOM fallback `Nos`, received/returned/sent metric qty qoidalari test bilan yopildi.
- Invalid archive date Go'dagi kabi `500 {"error":"werka archive failed"}` qaytarishi regressiya test bilan yopildi.
- Archive route ham Go handler kabi `any` qilindi; `POST /v1/mobile/werka/archive` regressiya test bilan yopildi.
- Real ERPNext smoke testda `sent/received/returned` kindlari va `POST` archive array/summary response qaytargani tasdiqlandi.
- Testlar 90 taga yetdi.
- `/v1/mobile/werka/archive/pdf` endpoint Rust'da qo'shildi.
- Go'dagi PDF builder contracti Rust'da alohida `http/archive_pdf.rs` moduliga ko'chirildi: `%PDF-1.4`, Courier font, 46 line/page, xref/trailer yozilishi saqlandi.
- PDF headerlari Go bilan mos: `Content-Type: application/pdf` va `Content-Disposition: attachment; filename="werka-{kind}-{period}.pdf"`.
- PDF ichidagi text escaping Go qoidalariga mos yozildi: `\\`, `(`, `)` escape qilinadi, newline/tab space bo'ladi, non-ASCII `?`ga aylanadi, line 132 belgida kesiladi.
- Quantity format PDF uchun Go'dagi `%.4g` significant-digit qoidalariga mos helper bilan yopildi.
- Archive PDF route ham Go handler kabi `any` qilindi; `POST /v1/mobile/werka/archive/pdf` regressiya test bilan yopildi.
- Real ERPNext smoke testda archive PDF `200`, `application/pdf`, expected attachment filename va `%PDF-1.4` body qaytargani tasdiqlandi.
- Testlar 97 taga yetdi.
- `/v1/mobile/werka/suppliers` endpoint Rust'da qo'shildi.
- Go'dagi query contract ko'chirildi: `q` trim qilinadi, `limit` default/max `200`, invalid/zero limit defaultga qaytadi, invalid/negative offset `0`.
- Supplier directory response JSON shakli Go bilan mos: `ref`, `name`, `phone`.
- Direct DB provider Go querysiga mos yozildi: `tabItem Supplier` orqali supplier itemga bog'langan bo'lishi kerak, supplier/item disabled bo'lmasligi kerak, `modified DESC`, `LIMIT/OFFSET`.
- MySQL LIKE search escaping Go'dagi `likePattern` bilan mos test qilindi.
- Go handler method check qilmagani uchun suppliers route ham `any` qilindi; `POST /v1/mobile/werka/suppliers` regressiya test bilan yopildi.
- Real ERPNext smoke testda `GET` va `POST /v1/mobile/werka/suppliers` ikkalasi ham array response qaytargani tasdiqlandi.
- Testlar 106 taga yetdi.
- `/v1/mobile/werka/customers` endpoint Rust'da qo'shildi.
- Customer directory response JSON shakli Go bilan mos: `ref`, `name`, `phone`.
- Direct DB provider Go querysiga mos yozildi: `tabCustomer` + `tabItem Customer Detail` + `tabItem`; faqat item assignment bor va disabled bo'lmagan customer/itemlar chiqadi.
- Search `customer name/ref/mobile_no` bo'yicha, `modified DESC`, `LIMIT/OFFSET` bilan ishlashi suppliers patterni bilan bir xil port qilindi.
- Go handler method check qilmagani uchun customers route ham `any` qilindi; `POST /v1/mobile/werka/customers` regressiya test bilan yopildi.
- Router testlari 500 qator chegarasidan tushishi uchun avatar route testlari alohida `profile_route_tests.rs`ga ajratildi.
- Real ERPNext smoke testda `GET` va `POST /v1/mobile/werka/customers` ikkalasi ham array response qaytargani tasdiqlandi.
- Testlar 113 taga yetdi.
- `/v1/mobile/werka/supplier-items`, `/v1/mobile/werka/customer-items`, `/v1/mobile/werka/customer-item-options` endpointlari Rust'da qo'shildi.
- Item response modellari Go bilan mos yozildi: supplier/customer item uchun `code/name/uom/warehouse/item_group,omitempty`, customer option uchun `customer_ref/customer_name/customer_phone/item_code/item_name/uom/warehouse`.
- Direct DB provider Go'dagi `SearchWerkaSupplierItemsPage`, `SearchWerkaCustomerItemsPage`, `SearchWerkaCustomerItemOptionsPage` SQL va pagination mantiqlariga mos port qilindi.
- Non-empty `q` qidiruv uchun Go'dagi `SearchQueryScore`, transliteration, compact/skeleton matching, score sort va tie-break qoidalari alohida `werka_item_search.rs` modulida test bilan yopildi.
- Go'dagi `defaultWarehouse` behaviori Rust direct DB configga `ERP_DEFAULT_TARGET_WAREHOUSE` orqali ulandi, aks holda item response'dagi `warehouse` bo'sh qolib ketardi.
- Go handler method check qilmagani uchun uchala item route ham `any` qilindi; `POST` regressiya test bilan yopildi.
- Authsiz `401`, provider yo'q bo'lsa Go error textlari bilan `500`, invalid limit/offset defaultlari va query trim/clamp behaviorlari route testlar bilan yopildi.
- Katta fayl bo'lmasligi uchun direct DB lookup impl `werka_lookup.rs`, item SQL/load `werka_items.rs`, ranking/search `werka_item_search.rs`, route testlar `werka_items_route_tests.rs` qilib ajratildi.
- Real ERPNext smoke testda Werka login token 32 belgi qaytdi; `customer-items`, qidiruvli `customer-items`, `customer-item-options`, va supplier assignment yo'q holatda `supplier-items` 200 response bilan tasdiqlandi.
- Real SQL solishtirishda `customer-items` va `customer-item-options` birinchi natijalari Go SQL contractiga mosligi tekshirildi.
- Testlar 128 taga yetdi.

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
- Direct DB uchun `sqlx` tanlandi.
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
- Customer login service logic.
- Customer login runtime ERPNext adapter wiring.
- Customer custom code state check.
- Login/me profile refresh.
- Supplier/customer ERPNext profile lookup.
- Supplier avatar proxy response rewrite.
- Supplier avatar view/download endpoint.
- Phone normalization.
- Role prefix inference.
- Login response skeleton.
- Session create/get/delete/update.
- Logout bearer validation.
- `/v1/mobile/me` route.
- `/v1/mobile/auth/me` route yo'qligi testi.

Hali portlanmagan:

- Werka login response ichidagi `werka_home` preload.

## Keyingi Qadam

Go endpoint matrix chiqariladi. Har endpoint uchun quyidagilar yoziladi:

- method/path
- Go handler
- core service method
- ERPNext/direct DB dependency
- JSON store dependency
- Rust'dagi kelajak modul joyi

## 2026-01-16: Werka Customer Issue Create Rust Port

Portlangan endpoint:

- `POST /v1/mobile/werka/customer-issue/create`

Go audit qilingan joylar:

- `internal/mobileapi/server.go`
  `handleWerkaCustomerIssueCreate`
- `internal/core/service.go`
  `CreateWerkaCustomerIssueWithSource`
  `normalizeCustomerIssueSource`
  `customerIssueSourceMarker`
  `hasDuplicateCustomerIssueSource`
- `internal/erpnext/delivery_note.go`
  `CreateDraftDeliveryNote`
  `UpdateDeliveryNoteState`
  `SubmitDeliveryNote`
  `DeleteDeliveryNote`
  `EnsureDeliveryNoteStateFields`
- `internal/erpdb/customer_issue.go`
  `CustomerIssueSourceExists`

Rust'da qo'shilgan asosiy qismlar:

- HTTP route:
  `/v1/mobile/werka/customer-issue/create`
- Models:
  `WerkaCustomerIssueCreateRequest`
  `WerkaCustomerIssueSource`
  `WerkaCustomerIssueCreateInput`
  `WerkaCustomerIssueRecord`
- Core service:
  `WerkaService::create_customer_issue`
- Ports:
  `WerkaCustomerIssueWriter`
  `CustomerIssueSourceLookup`
- ERPNext adapter:
  item lookup
  warehouse/company resolve
  Delivery Note draft create
  Accord custom field ensure
  state update
  submit with one `TimestampMismatchError` retry
  best-effort delete cleanup on update/submit failure
- Direct DB duplicate source check:
  `tabDelivery Note.accord_source_key`

Contract status:

- Non-POST returns `405 {"error":"method not allowed"}`.
- Missing/invalid bearer returns `401 {"error":"unauthorized"}`.
- Non-Werka principal returns `403 {"error":"forbidden"}`.
- Invalid JSON returns `400 {"error":"invalid json"}`.
- Missing writer/internal failure returns
  `500 {"error":"werka customer issue create failed"}`.
- Duplicate source returns
  `409 {"error":"duplicate customer issue source","error_code":"duplicate_customer_issue_source"}`.
- Negative stock / `NegativeStockError` returns
  `409 {"error":"insufficient stock","error_code":"insufficient_stock"}`.
- Success response Go shape bilan mos:
  `entry_id`, `customer_ref`, `customer_name`, `item_code`, `item_name`, `uom`, `qty`, `created_label`.
- Source marker Go tartibida tuziladi:
  `accord_customer_issue_source:source_barcode=...;source_stock_entry=...;source_line_index=...`
- `source_line_index < 0` Go kabi tashlab yuboriladi.

Test holati:

- Rust `cargo test`:
  `137 passed`, `0 failed`.
- Yangi route testlar:
  auth required
  POST-only
  invalid JSON
  provider yo'qligi
  success payload/source metadata
  duplicate source
  insufficient stock
- Yangi service testlar:
  source marker order/trimming
  negative line index normalization

Eslatma:

- Go handler successdan keyin customer push yuboradi va push xatosi response'ni yiqitmaydi. Rust serverda push sender hali port qilinmagan, shu sabab bu slice push yubormaydi. Customer issue yaratishning asosiy ERPNext write contracti portlandi; push subsystem port qilinganda shu endpointga best-effort hook ulanishi kerak.

Keyingi mantiqiy nishon:

- `handleWerkaCustomerIssueBatchCreate`
  yoki push subsystem/notification write contractini audit qilish.

## 2026-01-16: Werka Customer Issue Batch Create Rust Port

Portlangan endpoint:

- `POST /v1/mobile/werka/customer-issue/batch-create`

Go audit qilingan joylar:

- `internal/mobileapi/server.go`
  `handleWerkaCustomerIssueBatchCreate`
- `internal/core/types.go`
  `WerkaCustomerIssueBatchCreateRequest`
  `WerkaCustomerIssueBatchLineResult`
  `WerkaCustomerIssueBatchResult`

Muhim Go behavior:

- Method faqat `POST`; boshqa method `405 {"error":"method not allowed"}`.
- Werka auth required.
- Invalid JSON `400 {"error":"invalid json"}`.
- `lines` bo'sh bo'lsa `400 {"error":"lines are required"}`.
- Handler umumiy response'ni `200` qaytaradi, line xatolari `failed` ichida beriladi.
- Go batch line uchun `CreateWerkaCustomerIssue` chaqiradi, `CreateWerkaCustomerIssueWithSource` emas. Shuning uchun request line ichida source metadata bo'lsa ham batch flow source marker/duplicate source check ishlatmaydi.
- `ErrInsufficientStock` line failure sifatida:
  `{"error":"insufficient stock","error_code":"insufficient_stock"}`.
- Boshqa line error:
  `{"error":"werka customer issue create failed"}`.
- Success line `created` ichida `line_index` va `record` bilan qaytadi.
- `client_batch_id` trim qilinadi.
- Go ichida 4 tagacha worker ishlatiladi. Rust port hozir deterministic sequential bajaradi, lekin response contract va line order Go result slice tartibiga mos.

Rust'da qo'shilgan qismlar:

- Models:
  `WerkaCustomerIssueBatchCreateRequest`
  `WerkaCustomerIssueBatchLineResult`
  `WerkaCustomerIssueBatchResult`
- Core service:
  `WerkaService::create_customer_issue_batch`
- HTTP handler:
  `customer_issue_batch_create`
- Route:
  `/v1/mobile/werka/customer-issue/batch-create`

Test holati:

- Rust `cargo test`:
  `142 passed`, `0 failed`.
- Yangi batch route testlar:
  empty lines
  non-POST
  invalid JSON
  created line response
  partial insufficient stock failure

Eslatma:

- Go handler har created line uchun best-effort push yuboradi. Rust push subsystem hali port qilinmagani uchun bu hook hali yo'q. Push port qilinganda single create va batch create ikkalasiga ham response'ni yiqitmaydigan best-effort push ulanishi kerak.

Keyingi mantiqiy nishon:

- Push subsystem/notification write contractini audit qilish yoki navbatdagi Werka write endpointga o'tish.

## 2026-01-16: Werka Unannounced Create Rust Port

Portlangan endpoint:

- `POST /v1/mobile/werka/unannounced/create`

Go audit qilingan joylar:

- `internal/mobileapi/server.go`
  `handleWerkaUnannouncedCreate`
- `internal/core/service.go`
  `CreateWerkaUnannouncedDraft`
  `findSupplierForAdmin`
  `validateSupplierItemAllowed`
  `resolveWarehouse`
- `internal/erpnext/purchase_receipt.go`
  `CreateDraftPurchaseReceipt`
  `UpsertWerkaUnannouncedInRemarks`
  `UpdatePurchaseReceiptRemarks`
  `AddPurchaseReceiptComment`

Muhim Go behavior:

- Method faqat `POST`; boshqa method `405 {"error":"method not allowed"}`.
- Werka auth required.
- Invalid JSON `400 {"error":"invalid json"}`.
- Request body:
  `supplier_ref`, `item_code`, `qty`.
- Failure umumiy ko'rinishda:
  `500 {"error":"werka unannounced create failed"}`.
- Success Purchase Receipt draft yaratadi.
- Remarks ichiga Werka unannounced marker qo'yiladi:
  `Accord Werka Aytilmagan: pending`
- Purchase Receipt comment best-effort qo'shiladi.
- Response dispatch record bo'lib qaytadi:
  `record_type = "purchase_receipt"`
  `event_type = "werka_unannounced_pending"`
  `highlight = "Werka siz qayd etmagan mahsulotni qabul qildi"`

Rust'da qo'shilgan qismlar:

- HTTP route:
  `/v1/mobile/werka/unannounced/create`
- Handler:
  `src/http/handlers/werka/unannounced.rs`
- Models:
  `WerkaUnannouncedCreateRequest`
  `PurchaseReceiptDraft`
  `CreatePurchaseReceiptInput`
  `WerkaSupplierRecord`
- Core service:
  `WerkaService::create_werka_unannounced_draft`
- Core helper:
  `src/core/werka/unannounced.rs`
- Port:
  `WerkaUnannouncedWriter`
  `WerkaSupplierAdminStateLookup`
- ERPNext adapter:
  `src/erpnext/purchase_receipt.rs`
- Admin state store:
  Go JSON shape ichidagi `assigned_item_codes` fallback o'qiladi.

Tartib / file hygiene:

- Go'dagi katta service ichidagi logic Rust'da alohida modullarga bo'lindi.
- Yangi ERPNext Purchase Receipt write logic `erpnext/purchase_receipt.rs` ichida turibdi.
- Dispatch mapping va remarks marker helperlari `core/werka/unannounced.rs` ichida turibdi.
- Rust source/test fayllar tekshirildi: eng katta fayl `484` qatorda, ya'ni `500` limitdan past.
- Supplier-item validation Go tartibiga mos:
  direct DB supplier item lookup
  ERPNext live `Item Supplier` assignment
  `PermissionError` / `status 403` bo'lsa local `assigned_item_codes` fallback

Test holati:

- Rust `cargo test`:
  `150 passed`, `0 failed`.
- Yangi route testlar:
  non-POST
  invalid JSON
  provider yo'qligi
  success dispatch record va pending marker
- Yangi core fallback testlar:
  direct DB lookup validation
  ERP permission error bo'lganda admin assigned code fallback
  non-permission validation error fallback qilinmasligi

Eslatma:

- Go handler successdan keyin supplier push yuboradi va push xatosi response'ni yiqitmaydi. Rust push subsystem hali port qilinmagani uchun bu endpoint hozir push yubormaydi. Push port qilinganda best-effort hook qo'shilishi kerak.
- Supplier-item validation fallback yo'li ham Rustga port qilindi. Real ERP smoke test baribir kerak, lekin diary bo'yicha bu qism endi "known gap" emas.

Keyingi mantiqiy nishon:

- Navbatdagi Werka write endpoint yoki push subsystem/notification write contractini audit qilish.

## 2026-01-16: Supplier Unannounced Respond Rust Port

Portlangan endpoint:

- `POST /v1/mobile/supplier/unannounced/respond`

Go audit qilingan joylar:

- `internal/mobileapi/server.go`
  `handleSupplierUnannouncedRespond`
- `internal/core/types.go`
  `SupplierUnannouncedResponseRequest`
  `NotificationDetail`
  `NotificationComment`
- `internal/core/service.go`
  `RespondWerkaUnannouncedDraft`
  `NotificationDetail`
  `mapPurchaseReceiptToDispatchRecord`
  `parseNotificationComment`
- `internal/erpnext/purchase_receipt.go`
  `GetPurchaseReceipt`
  `UpdatePurchaseReceiptRemarks`
  `ConfirmAndSubmitPurchaseReceipt`
  `ListPurchaseReceiptComments`
  `AddPurchaseReceiptComment`
  `UpsertWerkaUnannouncedInRemarks`
  `ExtractWerkaUnannouncedState`
  `ExtractWerkaUnannouncedReason`

Muhim Go behavior:

- Method faqat `POST`; boshqa method `405 {"error":"method not allowed"}`.
- Auth required.
- Role faqat supplier; boshqa role `403 {"error":"forbidden"}`.
- Invalid JSON `400 {"error":"invalid json"}`.
- Request body:
  `receipt_id`, `approve`, `reason`.
- Core flow faqat pending Werka unannounced Purchase Receipt uchun ishlaydi.
- Purchase Receipt supplieri principal `ref` bilan mos bo'lishi kerak.
- Failure umumiy ko'rinishda:
  `500 {"error":"supplier unannounced response failed"}`.
- `approve=true`:
  remarks `approved` markerga o'tadi
  Purchase Receipt accepted qty bilan submit qilinadi
  best-effort comment qo'shiladi
  response `NotificationDetail`
  response record override:
  `accepted_qty = result.accepted_qty`
  `status = "accepted"`
  `event_type`, `highlight`, `note` bo'shatiladi
- `approve=false`:
  remarks `rejected` markerga o'tadi
  reason remarks ichida saqlanadi
  best-effort comment qo'shiladi
  response `NotificationDetail`
  dispatch note:
  `Supplier aytilmagan molni rad etdi.`
  reason bo'lsa `Sabab: ...`

Rust'da qo'shilgan qismlar:

- HTTP route:
  `/v1/mobile/supplier/unannounced/respond`
- Handler:
  `src/http/handlers/supplier/unannounced.rs`
- Models:
  `SupplierUnannouncedResponseRequest`
  `NotificationDetail`
  `NotificationComment`
- Core service extension:
  `WerkaService::respond_supplier_unannounced`
- Core helper:
  `src/core/werka/supplier_unannounced.rs`
  `src/core/werka/unannounced.rs`
- Port:
  `SupplierUnannouncedWriter`
- ERPNext adapter:
  `src/erpnext/purchase_receipt/response.rs`

Tartib / file hygiene:

- Supplier respond logic alohida `core/werka/supplier_unannounced.rs` fayliga chiqarildi.
- ERPNext submit/comment/detail logic `erpnext/purchase_receipt/response.rs` submodule ichida turibdi.
- Asosiy `service.rs` 500 qatordan past saqlandi.
- Rust source/test fayllar tekshirildi: eng katta fayl `484` qatorda, ya'ni `500` limitdan past.

Test holati:

- Rust `cargo test`:
  `156 passed`, `0 failed`.
- Yangi route testlar:
  non-POST
  non-supplier forbidden
  invalid JSON
  provider yo'qligi
  approve success response
  reject success response reason bilan

Eslatma:

- Go handler response'dan keyin Werka uchun best-effort push yuboradi. Rust push subsystem hali port qilinmagani uchun bu endpoint hozir push yubormaydi. Push port qilinganda `Supplier javob berdi` hook response'ni yiqitmaydigan qilib ulanishi kerak.
- `ConfirmAndSubmitPurchaseReceipt` approve path uchun port qilindi. Partial/returned qty flowlari bu endpointda ishlatilmaydi; ular keyingi supplier/customer receipt response flowlarda alohida 1:1 audit qilinadi.

Keyingi mantiqiy nishon:

- Push subsystem/notification write contracti yoki navbatdagi supplier/customer operational endpoint.

## 2026-01-16: Notification Detail Rust Port

Portlangan endpoint:

- `GET /v1/mobile/notifications/detail`

Go audit qilingan joylar:

- `internal/mobileapi/server.go`
  `handleNotificationDetail`
- `internal/core/service.go`
  `NotificationDetail`
  `resolveNotificationTarget`
  `mapPurchaseReceiptToDispatchRecord`
  `mapDeliveryNoteToDispatchRecord`
  `buildCustomerDeliveryResultEvent`
  `parseNotificationComment`
- `internal/erpdb/notification_detail.go`
  `NotificationDetailByReceiptID`
  `purchaseReceiptNotificationDetail`
  `deliveryNoteNotificationDetail`
  `notificationComments`
- `internal/erpnext/purchase_receipt.go`
  `GetPurchaseReceipt`
  `ListPurchaseReceiptComments`
- `internal/erpnext/delivery_note.go`
  `GetDeliveryNote`
  `ListDeliveryNoteComments`

Muhim Go behavior:

- Method faqat `GET`; boshqa method `405 {"error":"method not allowed"}`.
- Auth required.
- Role faqat supplier, werka, customer; admin va boshqa role `403 {"error":"forbidden"}`.
- Query `receipt_id` bo'sh bo'lsa `400 {"error":"receipt_id is required"}`.
- Unauthorized detail access `403 {"error":"forbidden"}`.
- Boshqa failure:
  `500 {"error":"notification detail failed"}`.
- Purchase Receipt detail:
  customer role ocholmaydi
  supplier faqat o'z `supplier_ref` recordini ochadi
  supplier display name response recordga override qilinadi
  `supplier_ack:` id event sifatida taniladi
- Delivery Note detail:
  `customer_delivery_result:` id delivery note targetga resolve qilinadi
  customer faqat o'z delivery note recordini ochadi
  accepted/partial/rejected result event highlight bilan qaytadi
- Go avval direct DB `NotificationDetailByReceiptID` ishlatadi; DB xato bersa ERPNext fallbackga o'tadi.

Rust'da qo'shilgan qismlar:

- HTTP route:
  `/v1/mobile/notifications/detail`
- Handler:
  `src/http/handlers/notifications.rs`
- Core helper:
  `src/core/werka/notification.rs`
- Ports:
  `NotificationDetailWriter`
  `NotificationDetailLookup`
- ERPNext adapter:
  `src/erpnext/notification.rs`
- Direct DB adapter:
  `src/erpdb/notification_detail.rs`

Tartib / file hygiene:

- Notification target resolve, auth visibility, PR/DN mapping `core/werka/notification.rs` ichida alohida turibdi.
- ERPNext detail adapteri `erpnext/notification.rs` ichida.
- Direct DB detail lookup `erpdb/notification_detail.rs` ichida.
- Rust source/test fayllar tekshirildi: eng katta fayl `493` qatorda, ya'ni `500` limitdan past.
- Keyingi slice oldidan `service.rs` va `service_tests.rs`ni yana bo'lish kerak, chunki ikkalasi ham limitga yaqinlashdi.

Test holati:

- Rust `cargo test`:
  `163 passed`, `0 failed`.
- Yangi route testlar:
  non-GET
  missing `receipt_id`
  admin forbidden
  provider yo'qligi
  supplier Purchase Receipt detail
  customer Purchase Receipt forbidden
  customer delivery result event detail

Eslatma:

- `POST /v1/mobile/notifications/comments` hali port qilinmagan. U write flow bo'lgani uchun alohida slice bo'ladi.
- Push subsystem hali port qilinmagan; notification comment ichidagi supplier acknowledgment push hook keyingi comment/push slice'da yopiladi.

Keyingi mantiqiy nishon:

- Avval file hygiene: `service.rs` va `service_tests.rs`ni kichikroq modullarga bo'lish.
- Keyin `POST /v1/mobile/notifications/comments` route'ini 1:1 port qilish.

## 2026-01-16: Notification Comment Rust Port

Portlangan endpoint:

- `POST /v1/mobile/notifications/comments`

Go audit qilingan joylar:

- `internal/mobileapi/server.go`
  `handleNotificationComment`
- `internal/core/types.go`
  `NotificationCommentCreateRequest`
- `internal/core/service.go`
  `AddNotificationComment`
  `NotificationDetail`
  `resolveNotificationTarget`
  `formatNotificationComment`
  `isSupplierAcknowledgmentMessage`
- `internal/erpnext/purchase_receipt.go`
  `AddPurchaseReceiptComment`
  `UpdatePurchaseReceiptRemarks`
  `UpsertSupplierAcknowledgmentInRemarks`
  `GetPurchaseReceipt`
- `internal/erpnext/delivery_note.go`
  `AddDeliveryNoteComment`

Muhim Go behavior:

- Method faqat `POST`; boshqa method `405 {"error":"method not allowed"}`.
- Auth required.
- Role faqat supplier, werka, customer; admin va boshqa role `403 {"error":"forbidden"}`.
- Query `receipt_id` bo'sh bo'lsa `400 {"error":"receipt_id is required"}`.
- Invalid JSON:
  `400 {"error":"invalid json"}`.
- Empty/whitespace `message` core'da xato bo'ladi va handler Go kabi `500 {"error":"notification comment failed"}` qaytaradi.
- Comment yozishdan oldin `NotificationDetail` chaqiriladi, shu orqali access/existence tekshiriladi.
- Purchase Receipt target:
  `Comment` doctype orqali `reference_doctype = "Purchase Receipt"` yoziladi.
- Delivery Note target:
  `Comment` doctype orqali `reference_doctype = "Delivery Note"` yoziladi.
- Comment content formati:
  `Supplier|Werka|Customer|Admin • DisplayName\nmessage`.
- Supplier message `tasdiqlayman` bilan boshlansa:
  PR qayta olinadi
  remarks ichida `Supplier tasdiqladi:` display line olib tashlanadi
  `Accord Supplier Tasdiq: <message>` qo'shiladi
  remarks update xatosi Go kabi best-effort, response'ni yiqitmaydi.
- Handler supplier acknowledgment'dan keyin Werka'ga best-effort push yuboradi. Rust push subsystem hali port qilinmagani uchun push hook hozircha pending parity gap.

Rust'da qo'shilgan qismlar:

- HTTP route:
  `/v1/mobile/notifications/comments`
- Handler:
  `src/http/handlers/notifications.rs`
- Request model:
  `NotificationCommentCreateRequest`
- Core flow:
  `src/core/werka/notification_comment.rs`
- Ports:
  `NotificationDetailWriter` comment/remarks write metodlari bilan kengaydi.
- ERPNext adapter:
  `src/erpnext/notification.rs`
- Route tests:
  `src/http/notification_comment_route_tests.rs`

Tartib / file hygiene:

- Notification detail logic `notification.rs` ichida qoldi.
- Notification comment write flow alohida `notification_comment.rs` fayliga ajratildi.
- Route tests detail va comment bo'yicha alohida fayllarga bo'lindi.
- Rust source/test fayllar tekshirildi: eng katta fayl `459` qatorda, ya'ni `500` limitdan past.

Test holati:

- Rust `cargo test`:
  `172 passed`, `0 failed`.
- `git diff --check` o'tdi.
- Yangi route testlar:
  non-POST
  missing `receipt_id`
  invalid JSON
  admin forbidden
  whitespace message Go kabi `500`
  supplier Purchase Receipt comment
  supplier acknowledgment remarks update
  customer Delivery Note comment

Eslatma:

- Cargo global cache permission warning chiqadi:
  `/home/wikki/.cargo/registry/... Permission denied`
  lekin test/build yiqilmadi.
- Push subsystem hali port qilinmagani uchun `supplier_ack` push hook bu slice'da bajarilmadi. Keyingi push slice'da handler response'ni yiqitmaydigan best-effort send sifatida ulanishi kerak.

Keyingi mantiqiy nishon:

- Push subsystemni port qilish yoki navbatdagi notification/customer operational endpointni audit qilish.

## 2026-01-26: Werka Domain Lock Checklist

Foydalanuvchi qarori bo'yicha porting endi har safar turli domainlarga sakramaydi. Hozirgi fokus:

- Avval `Werka` domain 100% yopiladi.
- `Werka` tugamaguncha supplier/customer/admin domainlarga o'tilmaydi.
- Faqat blocker dependency bo'lsa, u alohida "pending parity gap" sifatida belgilanadi.

Go `Werka` route checklist:

- `[x] /v1/mobile/werka/summary`
- `[x] /v1/mobile/werka/home`
- `[x] /v1/mobile/werka/customers`
- `[x] /v1/mobile/werka/suppliers`
- `[x] /v1/mobile/werka/ai-search-suggestion`
- `[x] /v1/mobile/werka/supplier-items`
- `[x] /v1/mobile/werka/customer-items`
- `[x] /v1/mobile/werka/customer-item-options`
- `[x] /v1/mobile/werka/customer-issue/create`
- `[x] /v1/mobile/werka/customer-issue/batch-create`
- `[x] /v1/mobile/werka/unannounced/create`
- `[x] /v1/mobile/werka/status-breakdown`
- `[x] /v1/mobile/werka/status-details`
- `[x] /v1/mobile/werka/pending`
- `[x] /v1/mobile/werka/history`
- `[x] /v1/mobile/werka/notifications`
- `[x] /v1/mobile/werka/archive`
- `[x] /v1/mobile/werka/archive/pdf`
- `[x] /v1/mobile/werka/confirm`

Hozir Werka route coverage:

- `19/19`
- Werka mobile API route'lari Rustda yopildi.

## 2026-01-26: Supplier Domain Lock Checklist

Fokus endi `Supplier` domain. Qoida:

- Supplier domain 100% bo'lmaguncha boshqa domainlarga sakralmaydi.
- Har safar bitta kichik endpoint/slice olinadi.
- Go xatti-harakati 1:1 saqlanadi, lekin Go'dagi katta fayl tartibsizligi Rustda takrorlanmaydi.
- Direct DB read va ERPNext fallback alohida portlar orqali ulanadi.

Go `Supplier` route checklist:

- `[x] /v1/mobile/supplier/unannounced/respond`
- `[x] /v1/mobile/supplier/summary`
- `[x] /v1/mobile/supplier/history`
- `[x] /v1/mobile/supplier/status-breakdown`
- `[ ] /v1/mobile/supplier/status-details`
- `[ ] /v1/mobile/supplier/items`
- `[ ] /v1/mobile/supplier/dispatch`

Hozir Supplier route coverage:

- `4/7`
- Keyingi tavsiya qilingan slice:
  `/v1/mobile/supplier/status-details`
  chunki `status-breakdown` bilan bir xil kind mapping va receipt source ustida item drill-down qiladi.

## 2026-01-26: Supplier Status Breakdown Rust Port

Portlangan endpoint:

- `GET/POST /v1/mobile/supplier/status-breakdown`

Go source audit:

- Handler:
  `internal/mobileapi/server.go`
  `handleSupplierStatusBreakdown`
- Core:
  `internal/core/service.go`
  `SupplierStatusBreakdown`
  `recordMatchesSupplierBreakdown`
- Types:
  `internal/core/types.go`
  `SupplierStatusBreakdownEntry`

Go contract:

- Auth required.
- Role faqat Supplier; boshqa role:
  `403 {"error":"forbidden"}`.
- Go handler method check qilmaydi; Rust route ham `any` qilib qo'yildi va `POST` regressiya test bilan yopildi.
- Provider xatosi yoki provider yo'q bo'lsa:
  `500 {"error":"supplier status breakdown failed"}`.
- Query:
  `kind`
- Supplier kind mapping:
  `pending` -> `pending` yoki `draft`
  `submitted` -> `accepted`
  `returned` -> `partial`, `rejected`, `cancelled`
  boshqa kind -> empty array
- Response JSON:
  `item_code`
  `item_name`
  `receipt_count`
  `total_sent_qty`
  `total_accepted_qty`
  `total_returned_qty`
  `uom`
- Group key:
  `item_code`, bo'sh bo'lsa `item_name`
- Sort:
  `receipt_count DESC`
  keyin `strings.ToLower(item_name) ASC`
- Go 1:1 muhim detal:
  bu endpoint `a.reader` direct DB reader'dan foydalanmaydi.
  Rustda ham ataylab direct DB shortcut qo'shilmadi; faqat ERPNext purchase receipt lookup ishlatiladi.

Rust'da qo'shilgan qismlar:

- Model:
  `SupplierStatusBreakdownEntry`
- Service:
  `supplier_status_breakdown`
- HTTP handler:
  `src/http/handlers/supplier/read.rs`
- Route tests:
  `src/http/supplier_read_route_tests.rs`
- Core tests:
  grouping, `submitted` kind mapping, sent/accepted/returned totals

Tartib / file hygiene:

- Status breakdown logic existing supplier read service ichida kichik helperlar bilan qoldi.
- Direct DB modulega qo'shilmadi, chunki Go parity shuni talab qiladi.
- Rust source/test fayllar tekshirildi: eng katta fayl `466` qatorda, ya'ni `500` limitdan past.

Test holati:

- Rust `cargo test`:
  `203 passed`, `0 failed`.
- `cargo fmt --check` o'tdi.
- `git diff --check` o'tdi.

Natija:

- Supplier domain route coverage:
  `4/7`.
- Supplier `status-breakdown` endpointi ERPNext fallback yo'li bilan Go contractga mos port qilindi.

## 2026-01-26: Supplier History Rust Port

Portlangan endpoint:

- `GET/POST /v1/mobile/supplier/history`

Go source audit:

- Handler:
  `internal/mobileapi/server.go`
  `handleSupplierHistory`
- Core:
  `internal/core/service.go`
  `SupplierHistory`
  `purchaseReceiptCommentsByName`
  `dispatchRecordNeedsCommentScan`
  `isSupplierAcknowledgmentComment`
- Direct DB:
  `internal/erpdb/reader.go`
  `SupplierHistory`

Go contract:

- Auth required.
- Role faqat Supplier; boshqa role:
  `403 {"error":"forbidden"}`.
- Go handler method check qilmaydi; Rust route ham `any` qilib qo'yildi va `POST` regressiya test bilan yopildi.
- Provider xatosi yoki provider yo'q bo'lsa:
  `500 {"error":"supplier history failed"}`.
- Response JSON:
  `[]DispatchRecord`
- Direct DB yo'li:
  `supplier_delivery_note LIKE 'TG:%'`
  `supplier = principal.Ref`
  hidden Werka unannounced receiptlar chiqarilmaydi
  `created_label` bo'yicha newest-first sort qilinadi.
- ERPNext fallback:
  `collectSupplierPurchaseReceipts`
  page size `200`
  duplicate `name` skip
  Go kabi natija ERP list tartibida qoladi, alohida sort qilinmaydi.
- Comment batch:
  faqat `partial/rejected/cancelled` yoki `note` bor recordlar uchun comment scan qilinadi.
  `Supplier ...\nTasdiqlayman...` comment topilsa note ichiga:
  `Supplier tasdiqladi: Tasdiqlayman, shu holat bo‘lganini ko‘rdim.`
  qo'shiladi.

Rust'da qo'shilgan qismlar:

- Service:
  `supplier_history`
- Direct DB adapter:
  `src/erpdb/supplier_read.rs`
- ERPNext adapter:
  `src/erpnext/purchase_receipt/supplier_read.rs`
- HTTP handler:
  `src/http/handlers/supplier/read.rs`
- Route tests:
  `src/http/supplier_read_route_tests.rs`
- Core tests:
  comment scan skip/ack note path
- Direct DB tests:
  hidden unannounced filter va newest-first sort

Qo'shimcha parity tuzatish:

- ERPNext fallback mapping Go'dagi `mapPurchaseReceiptToDispatchRecord`ga yaqinlashtirildi:
  `Accord Qabul:`
  `Accord Qaytarildi:`
  `Accord Sabab:`
  `Accord Izoh:`
  `Accord Supplier Tasdiq:`
  remarklari note/status hisobida o'qiladi.

Tartib / file hygiene:

- ERPNext supplier read adapter `purchase_receipt.rs` ichiga tiqilmadi, alohida `purchase_receipt/supplier_read.rs` fayliga chiqarildi.
- Direct DB supplier read summary+history bitta kichik `supplier_read.rs` modulida.
- Rust source/test fayllar tekshirildi: eng katta fayl `466` qatorda, ya'ni `500` limitdan past.

Test holati:

- Rust `cargo test`:
  `199 passed`, `0 failed`.
- `cargo fmt --check` o'tdi.
- `git diff --check` o'tdi.

Natija:

- Supplier domain route coverage:
  `3/7`.
- Supplier `history` endpointi direct DB va ERPNext fallback yo'llari bilan port qilindi.

## 2026-01-26: Supplier Summary Rust Port

Portlangan endpoint:

- `GET/POST /v1/mobile/supplier/summary`

Go source audit:

- Handler:
  `internal/mobileapi/server.go`
  `handleSupplierSummary`
- Core:
  `internal/core/service.go`
  `SupplierSummary`
  `collectSupplierPurchaseReceipts`
- Direct DB:
  `internal/erpdb/reader.go`
  `SupplierSummary`

Go contract:

- Auth required.
- Role faqat Supplier; boshqa role:
  `403 {"error":"forbidden"}`.
- Go handler method check qilmaydi; Rust route ham `any` qilib qo'yildi va `POST` regressiya test bilan yopildi.
- Provider xatosi yoki provider yo'q bo'lsa:
  `500 {"error":"supplier summary failed"}`.
- Response JSON:
  `pending_count`
  `submitted_count`
  `returned_count`
- Direct DB yo'li `tabPurchase Receipt` ichida:
  `supplier_delivery_note LIKE 'TG:%'`
  `supplier = principal.Ref`
- ERPNext fallback Go kabi page bilan o'qiydi:
  `ListSupplierPurchaseReceiptsPage`
  page size `200`
  duplicate `name` skip

Rust'da qo'shilgan qismlar:

- Model:
  `SupplierHomeSummary`
- Ports:
  `SupplierReadLookup`
  `SupplierPurchaseReceiptLookup`
- Core service:
  `src/core/werka/supplier_read.rs`
- Direct DB adapter:
  `src/erpdb/supplier_read.rs`
- ERPNext adapter:
  `SupplierPurchaseReceiptLookup for ErpnextClient`
- HTTP handler:
  `src/http/handlers/supplier/read.rs`
- Supplier shared auth helper:
  `src/http/handlers/supplier/authz.rs`
- Route tests:
  `src/http/supplier_read_route_tests.rs`

Tartib / file hygiene:

- Supplier HTTP auth `authz.rs`ga ajratildi; `unannounced.rs` ichidagi duplicate auth helperlar olib tashlandi.
- Direct DB supplier summary `reader.rs`ga qo'shilmadi, alohida modulga ajratildi.
- Core summary collection/count logic alohida `supplier_read.rs`da.
- Rust source/test fayllar tekshirildi: eng katta fayl `461` qatorda, ya'ni `500` limitdan past.

Test holati:

- Rust `cargo test`:
  `194 passed`, `0 failed`.
- `cargo fmt --check` o'tdi.
- `git diff --check` o'tdi.

Natija:

- Supplier domain route coverage:
  `2/7` o'sha slice yakunidagi holat edi; hozirgi umumiy coverage yuqoridagi lock checklistda `3/7`.
- Supplier `summary` endpointi direct DB va ERPNext fallback yo'llari bilan port qilindi.

## 2026-01-16: Werka Confirm Rust Port

Portlangan endpoint:

- `POST /v1/mobile/werka/confirm`

Go audit qilingan joylar:

- `internal/mobileapi/server.go`
  `handleWerkaConfirm`
- `internal/core/types.go`
  `ConfirmReceiptRequest`
- `internal/core/service.go`
  `ConfirmReceipt`
  `dispatchStatusFromQuantities`
- `internal/erpnext/purchase_receipt.go`
  `ConfirmAndSubmitPurchaseReceipt`
  `buildAccordDecisionNote`
  `upsertAccordDecisionInRemarks`
  `ExtractAccordDecisionNote`
  `findAlternateWarehouse`

Muhim Go behavior:

- Method faqat `POST`; boshqa method `405 {"error":"method not allowed"}`.
- Auth required.
- Role faqat Werka; boshqa role `403 {"error":"forbidden"}`.
- Invalid JSON:
  `400 {"error":"invalid json"}`.
- Missing JSON fieldlar Go decoder kabi zero/default bo'ladi.
- Confirm/submit xatosi:
  `500 {"error":"receipt confirm failed"}`.
- Response `DispatchRecord`.
- Status sent/accepted qty bo'yicha:
  `accepted`, `partial`, `rejected`.
- Partial/return bo'lsa `Accord Qabul`, `Accord Qaytarildi`, optional `Accord Sabab`, optional `Accord Izoh` decision note yaratiladi.
- Decision note `remarks` ichiga upsert qilinadi va eski decision/supplier ack line'lari olib tashlanadi.
- Decision note comment sifatida best-effort yoziladi.
- Submit xato bersa Go kabi original Purchase Receipt doc rollback qilinadi.
- Full return holatida Go kabi submit qilinmaydi, faqat remarks/comment yoziladi va response qaytadi.

Rust'da qo'shilgan qismlar:

- HTTP route:
  `/v1/mobile/werka/confirm`
- Handler:
  `src/http/handlers/werka/confirm.rs`
- Request model:
  `ConfirmReceiptRequest`
- Core flow:
  `src/core/werka/confirm.rs`
- Port:
  `WerkaConfirmWriter`
- ERPNext adapter:
  `src/erpnext/purchase_receipt/response.rs`
- Decision helper submodule:
  `src/erpnext/purchase_receipt/response/decision.rs`
- Route tests:
  `src/http/werka_confirm_route_tests.rs`

Tartib / file hygiene:

- Confirm core logic alohida `core/werka/confirm.rs` ichida.
- HTTP handler alohida `http/handlers/werka/confirm.rs` ichida.
- Purchase Receipt decision-note helperlari `response/decision.rs`ga ajratildi.
- Rust source/test fayllar tekshirildi: eng katta fayl `459` qatorda, ya'ni `500` limitdan past.

Test holati:

- Rust `cargo test`:
  `179 passed`, `0 failed`.
- `git diff --check` o'tdi.
- Yangi route testlar:
  non-POST
  non-Werka forbidden
  invalid JSON
  provider yo'qligi
  success response va decision field forwarding
- Yangi ERPNext helper testlar:
  decision note Go format
  old decision/supplier ack line'larini upsertda olib tashlash

Eslatma:

- Go handler confirmdan keyin supplier push yuboradi. Rust push subsystem hali port qilinmagani uchun bu hook pending parity gap.
- Keyingi Werka nishon:
  `/v1/mobile/werka/ai-search-suggestion`.

## 2026-01-26: Werka AI Search Suggestion Rust Port

Portlangan endpoint:

- `POST /v1/mobile/werka/ai-search-suggestion`

Go audit qilingan joylar:

- `internal/mobileapi/werka_ai_search.go`
  `handleWerkaAISearchSuggestion`
  `werkaAISearchService`
  `inferSuggestion`
  `decodeWerkaAISearchPayload`
  `sanitizeSearchQuery`
  `normalizeServerFriendlyQuery`
  `rankQueries`
  `detectImageMIMEType`
- `internal/mobileapi/werka_ai_search_test.go`
- `cmd/core/main.go`
  `SetWerkaAISearchConfig`
- `README.md`
  `GEMINI_API_KEY`
  `GEMINI_VISION_MODEL`

Muhim Go behavior:

- Method faqat `POST`; boshqa method:
  `405 {"error":"method not allowed","code":"method_not_allowed"}`.
- Auth required.
- Role faqat Werka; boshqa role:
  `403 {"error":"forbidden"}`.
- AI config yo'q bo'lsa upload parse qilmasdan oldin:
  `503 {"error":"werka ai search is not configured","code":"not_configured"}`.
- Upload multipart `image` field orqali keladi.
- Upload limit:
  `8 MiB`.
- Multipart parse xatosi:
  `400 {"error":"invalid image upload","code":"invalid_image"}`.
- `image` field yo'q yoki bo'sh bo'lsa:
  `400 {"error":"image is required","code":"invalid_image"}`.
- Gemini request:
  `generateContent`
  `temperature: 0`
  `responseMimeType: application/json`
  inline base64 image
- Default model:
  `gemini-flash-lite-latest`.
- No result bo'lsa Go kabi `200` va empty suggestion shape.
- Upstream xatolar:
  `502` va `code: upstream_failed`.
- Query normalization:
  Nivea, Musaffo, Hot Lunch/Xot Lanch, Simba Chips, Mini Rulet kabi Go special-case'lari port qilindi.

Rust'da qo'shilgan qismlar:

- AI adapter:
  `src/ai/werka_search.rs`
- Core service:
  `src/core/werka/ai_search.rs`
- Models:
  `WerkaAiSearchSuggestion`
- Ports:
  `WerkaAiSearch`
  `WerkaAiSearchImage`
  `WerkaAiSearchError`
- HTTP handler:
  `src/http/handlers/werka/ai_search.rs`
- Route tests:
  `src/http/werka_ai_search_route_tests.rs`
- Runtime wiring:
  `GEMINI_API_KEY`
  `GEMINI_VISION_MODEL`

Tartib / file hygiene:

- AI Gemini client `src/ai/werka_search.rs` ichida.
- HTTP multipart parsing/status mapping handler modulida.
- Core service faqat port chaqiradi.
- Rust source/test fayllar tekshirildi: eng katta fayl `459` qatorda, ya'ni `500` limitdan past.

Test holati:

- Rust `cargo test`:
  `189 passed`, `0 failed`.
- `git diff --check` o'tdi.
- Yangi route testlar:
  non-POST
  non-Werka forbidden
  config yo'q bo'lsa parse'dan oldin `503`
  image required
  success suggestion va MIME detect
  no result empty suggestion
- Yangi AI helper testlar:
  Nivea normalization
  wrapped JSON extraction
  MIME detection
  multipart boundary extraction

Natija:

- Werka domain route coverage:
  `19/19`.
- Werka mobile API endpointlari Rustda 100% yopildi.

Keyingi domain tanlash:

- Supplier domainni 100% yopish tavsiya qilinadi.
