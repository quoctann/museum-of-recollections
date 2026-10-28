# Handover — Museum of Recollections

Tài liệu bàn giao sau đợt triển khai đầu tiên, ngày 26/10/2026.
Các mục chưa đánh dấu là **chưa hoàn tất**, không phải tính năng đang có đầy đủ.
Đã tiếp tục lát cắt A ở cùng ngày; xem mục 2.1 và checklist A bên dưới.

## 1. Mục tiêu và phạm vi đã thống nhất

Website đọc tản văn theo phong cách tạp chí nghệ thuật/bảo tàng ký ức,
kèm CMS cho **một người viết**. Trải nghiệm đọc mobile quan trọng hơn hiệu ứng.

- Frontend: Next.js, TypeScript, Tailwind, Motion; bổ sung Tiptap và shadcn/ui cho CMS.
- Backend: Go, Gin, PostgreSQL 17, sqlx, sqlc, zap; kiến trúc hexagonal.
- Nội dung rich text dự kiến lưu Tiptap JSONB có `schemaVersion`.
- Hạ tầng k3s và Argo CD đã có, chủ dự án tự triển khai sau khi application ổn định.
- Production sẽ dùng MinIO trong k3s; local dùng MinIO qua Docker Compose.
- **Không làm Kubernetes manifest/Argo CD app-of-apps trong phạm vi hiện tại.**
- Chưa làm bình luận, tài khoản độc giả, nhiều editor, lịch xuất bản, tìm kiếm
  nâng cao hoặc GSAP.

Tham khảo: [ý tưởng gốc](../idea.md), [định hướng thiết kế](design-direction.md),
[hướng dẫn chạy](../README.md), [API foundation](../apps/api/README.md).

## 2. Baseline đã hoàn thành trước lát cắt persistence

- [x] Next.js App Router: `/`, `/articles`, `/articles/[slug]`, trang 404.
- [x] Ba bài tiếng Việt **minh họa**, dữ liệu tĩnh, có nhãn phân biệt rõ.
- [x] Bố cục giấy–mực, tranh SVG local, Noto Serif và Be Vietnam Pro self-hosted.
- [x] Motion nhẹ, reading progress, reduced motion, skip link, nội dung đọc được
  khi JavaScript bị tắt.
- [x] API `/healthz`, `/readyz`, cấu hình có validation, graceful shutdown,
  log đã lọc dữ liệu nhạy cảm và rate limit với budget probe độc lập.
- [x] Kết nối PostgreSQL và kiểm tra bucket MinIO; chưa có schema bài viết.
- [x] Binary `storage-init` riêng, tạo bucket idempotent, không đặt public policy.
- [x] Compose có PostgreSQL 17, MinIO, bootstrap bucket, API và web.
- [x] Dockerfile non-root, Makefile, `.env.example`, sinh credential local ngẫu nhiên.
- [x] CI kiểm tra bootstrap, Go, frontend build và Playwright.

**Ở baseline:** website chưa gọi API bài viết. Chưa có đăng nhập, database
nghiệp vụ, CMS, Tiptap, upload, autosave, preview hoặc publish/unpublish.
Renderer hiện tại chỉ phục vụ cấu trúc bài mẫu, **chưa phải renderer Tiptap**.

### 2.1. Tiếp tục lát cắt A — persistence

- Schema 5 bảng và repository nội bộ đã có: create/get/save draft,
  publish/unpublish/public detail/list. Revision tăng cho cả thay đổi visibility;
  public chỉ đọc snapshot có cờ `published`, unpublish giữ snapshot cuối.
- Slug ASCII khóa sau publish đầu. ID bigint; danh sách summary không tải body,
  keyset theo timestamp xuất bản đầu + ID, mặc định 20 và tối đa 50.
- Content schema v1 có allowlist/budget; image chưa được chấp nhận. Publish kiểm
  tra lại draft đã lưu trước snapshot; renderer an toàn vẫn chưa triển khai.
- Tool migration goose `v3.24.1`, sqlc `v1.29.0` được pin; có `make migrate`,
  `migrate-status`, `migrate-down`, `sqlc`, `sqlc-check`, `test-integration`.
- PostgreSQL test tách riêng: password/project/cổng ngẫu nhiên, tmpfs, không dùng
  development volume hoặc `.env`; cleanup đúng project kể cả test lỗi.
- CI bổ sung generation drift và migration/integration runner.
- [Hướng dẫn persistence/rollback](persistence.md) và
  [hợp đồng HTTP dự kiến](api-contract.md) ghi rõ những route **chưa mở**.
- Website giữ bài mẫu. Chưa có auth/CMS/upload hoặc HTTP bài viết; không nối
  frontend và không thêm Kubernetes/Argo CD trong lượt này.
- Đã chạy lại: `make check`, `make build`, web coverage (18 tests, phạm vi như
  baseline), Playwright (10 lượt desktop/mobile), và `make test-integration`.
  Runner integration kiểm tra goose up → up → down → up trên PostgreSQL tạm;
  không chạy migration trên volume phát triển.
- `make sqlc-check`: đạt. Go race suite integration: đạt; package nghiệp vụ
  `internal/articles` đạt **89,9% statement coverage**, không phải toàn API.
  Web đạt 100% lines/statements/functions, 93,33% branches trong phạm vi cấu hình.
- Đã xử lý các finding review: summary không tải body, keyset phá hòa theo ID,
  surrogate JSON không hợp lệ bị từ chối, publish revalidate draft đã khóa và
  kiểm thử giới hạn 1 MiB không bị sai do JSONB thêm khoảng trắng. Review lại
  không còn finding correctness/security trong phạm vi thay đổi; không thay thế
  security/dependency scan production ở mục F.

### Điểm vào mã nguồn

| Hạng mục | Vị trí hiện tại |
| --- | --- |
| Dữ liệu và helper bài mẫu | `apps/web/src/lib/collection.ts` |
| Trang chủ | `apps/web/src/app/page.tsx` |
| Trang danh sách | `apps/web/src/app/articles/page.tsx` |
| Trang đọc | `apps/web/src/app/articles/[slug]/page.tsx` |
| Renderer bài mẫu | `apps/web/src/components/article-body.tsx` |
| Tranh/chuyển động | `apps/web/src/components/art.tsx`, `apps/web/public/art/` |
| Reading progress | `apps/web/src/components/reading-progress.tsx`, `src/lib/progress.ts` |
| CSS | `apps/web/src/app/{globals,home,reading}.css` |
| Web tests | `apps/web/tests/`, `apps/web/e2e/museum.spec.ts` |
| API composition root | `apps/api/cmd/api/main.go` |
| Bootstrap bucket | `apps/api/cmd/storage-init/`, `internal/storageinit/` |
| Config | `apps/api/internal/config/` |
| Port và nghiệp vụ readiness | `apps/api/internal/health/` |
| Domain/use case/port và PostgreSQL bài viết | `apps/api/internal/articles/` |
| Migration/query/sqlc generated | `apps/api/db/`, `apps/api/sqlc.yaml` |
| Adapters dependency, HTTP | `apps/api/internal/dependencies/`, `internal/httpapi/` |
| Vòng đời HTTP server | `apps/api/internal/server/` |
| Local infrastructure | `compose.yaml`, `deploy/local/minio/` |
| Bootstrap/CI | `scripts/setup_local.py`, `tests/`, `.github/workflows/ci.yml` |
| Tooling migration/integration tách biệt | `scripts/api_tools.py`, `compose.test.yaml` |

Đường dẫn rút gọn trong từng ô là tương đối với application tương ứng.
Baseline chưa có nghiệp vụ bài viết; lát cắt A bổ sung riêng, không thay thế
các port readiness/storage hiện có.

## 3. Khôi phục môi trường và baseline

Yêu cầu: Node.js 22, npm 10, Go ≥1.27.3, Python 3, Docker Compose,
CGO và GCC/Clang cho Go race detector.

```sh
make setup
make up
make check
make build
npm --prefix apps/web run test:coverage
npm --prefix apps/web exec -- playwright install chromium
npm --prefix apps/web run test:e2e
```

E2E dùng production server cổng `3100`, cần `make build` trước.
Linux có thể cần cài browser bằng `playwright install --with-deps chromium`.
Không chạy nhiều Vitest coverage process cùng report directory.

- Web: `http://localhost:3000`.
- API: `http://localhost:8080/healthz`, `http://localhost:8080/readyz`.
- PostgreSQL: `localhost:55432`; cổng nội bộ container vẫn là `5432`.
- MinIO: S3 `localhost:9000`, console `localhost:9001`.
- `.env` sinh local, ignored, quyền `0600`; không đưa secret vào tài liệu/commit.
- `make setup` không ghi đè `.env` cũ. Nếu thiếu biến mới, đối chiếu `.env.example`.
- `make down` giữ named volumes; không xóa volume để xử lý lỗi một cách tùy tiện.
- Muốn hot reload: `make down`, `make dev`, rồi `make dev-web` và `make dev-api`
  trong hai terminal riêng.

### Kết quả lịch sử — foundation trước persistence

- `make check`, `make build`, Docker build và `make up`: đạt.
- API race tests + vet: đạt; **85,9% coverage tổng statement**.
- Frontend: **18 unit/integration tests**; 100% lines/statements/functions,
  93,33% branches **trong phạm vi components/lib được cấu hình**, không phải toàn app.
- Playwright: **10 lượt test**, gồm desktop/mobile, reduced motion, keyboard,
  no-JS đúng viewport của từng project và 404.
- Stack local: web/API/PostgreSQL/MinIO healthy; readiness trả `200`.
- Bootstrap bucket chạy hai lần thành công; anonymous bucket access trả `403`.
- Frontend production dependency audit: 0 vulnerability được báo cáo.
- Go scan: không có vulnerability reachable/imported-package; còn advisory
  module-level ở `x/crypto/openpgp` không được import, chưa có bản vá.

Đây là kết quả tại thời điểm bàn giao, cần chạy lại khi tiếp tục. Không coi đó
là chứng nhận bảo mật toàn bộ dependency hoặc image MinIO.

## 4. Các nguyên tắc cần giữ khi phát triển tiếp

1. **Tách draft và publication.** Autosave không được thay đổi nội dung public.
2. **Publish nguyên tử:** snapshot, trạng thái, version phải nhất quán trong transaction.
3. **Version-based concurrency:** request cũ hoặc tab khác không ghi đè draft mới.
4. **Admin đơn lẻ:** tạo tài khoản bằng CLI, không có API đăng ký công khai.
5. **Cùng origin `/api` là hướng dự kiến**, chưa cấu hình proxy/rewrite ở bản hiện tại.
6. **Media draft luôn private.** Không bật public cả bucket để giải quyết hiển thị ảnh.
7. **Giữ native scroll:** không hijack, ép snap hoặc reveal từng đoạn văn.
8. **Production migration chạy riêng**, không tự chạy khi mỗi replica khởi động.
9. **sqlc sinh query có kiểu**, dùng chế độ `database/sql`; sqlx quản lý connection,
   transaction tương thích. Không xây lớp query thủ công trùng với sqlc.
10. TDD cho nghiệp vụ; unit + integration PostgreSQL thật + E2E luồng biên tập.

## 5. Backlog theo thứ tự đề xuất

### A — Persistence và hợp đồng API

- [x] Chọn/pin công cụ migration; thêm `make migrate` và hướng dẫn rollback.
- [x] Thêm `apps/api/db/{migrations,queries,generated}/` và cấu hình sqlc.
- [x] Thiết kế `admins`, `articles`, `article_drafts`, `article_publications`, `media`.
- [x] Tiptap JSONB có `schemaVersion`; draft có optimistic `version`.
- [x] Slug ASCII tường minh, duy nhất, khóa sau publish đầu; chưa hỗ trợ đổi URL cũ.
- [x] Timestamp đầu giữ nguyên; danh sách summary keyset `(first_published_at, id)`
  giảm dần, mặc định 20/tối đa 50; cursor HTTP encoding còn ở adapter tương lai.
- [x] Domain/use case/port và PostgreSQL adapter ở `internal/articles`; publish transaction.
- [ ] Định nghĩa API, validation, error format, giới hạn body và pagination.
- [x] Integration tests PostgreSQL thật, tách DB test khỏi volume phát triển.

Hợp đồng HTTP đã có **đề xuất** tại `docs/api-contract.md`; validation/error
mapping/body limit ở HTTP vẫn chưa triển khai. Không đánh dấu toàn bộ A hoàn tất
vì route chưa tồn tại và chưa nối biên xác thực.

**Nhóm API dự kiến, chưa tồn tại:** `/api/public/articles`,
`/api/admin/articles`, `/api/admin/media`, `/api/admin/auth`.
Cần chốt method/payload/status trong hợp đồng API trước khi nối frontend.

**Nghiệm thu:** migration chạy được trên DB trống; public query chỉ trả snapshot
đã xuất bản; sửa draft không làm public đổi; publish/unpublish đúng transaction;
conflict version trả lỗi có thể xử lý, không mất nội dung.

### B — Xác thực và biên bảo mật

- [ ] CLI `admin-create`, password hash an toàn, không log password.
- [ ] Login/logout, JWT ngắn hạn trong cookie HttpOnly, Secure production, SameSite.
- [ ] Chốt thời hạn phiên, refresh/revocation và đăng xuất phía server nếu có phiên dài.
- [ ] CSRF protection cho thao tác ghi; kiểm tra authorization toàn bộ route admin.
- [ ] Rate limit riêng cho login và nghiệp vụ; giữ budget operational probe độc lập.
- [ ] Cấu hình same-origin proxy/BFF theo kiến trúc chọn, không mở CORS tùy tiện.
- [ ] Tests sai password, hết hạn/thu hồi phiên, thiếu CSRF, truy cập trái phép.

**Nghiệm thu:** không có đường đọc/ghi draft hoặc upload khi chưa xác thực;
logout vô hiệu hóa phiên theo hợp đồng đã chọn; lỗi không lộ secret.

### C — CMS, rich text và media

- [ ] Admin layout, login, danh sách bài, tạo/sửa bài; dùng shadcn/ui cho CMS.
- [ ] Tiptap: paragraph, heading, bold/italic, list, link, blockquote, image.
- [ ] Title, subtitle, slug, cover, alt text/caption; trạng thái loading/empty/error.
- [ ] Schema/extension allowlist dùng chung editor và renderer; không raw HTML/embed tùy ý.
- [ ] Upload kiểm tra byte signature, MIME, dung lượng/kích thước ảnh; chặn SVG ở MVP.
- [ ] Object key ngẫu nhiên; liên kết media với draft, không tin URL tùy ý từ client.
- [ ] Tách endpoint S3 nội bộ khỏi URL ảnh browser truy cập được.
- [ ] Chọn cách phục vụ media private/public (proxy, signed URL hoặc publication copy)
  và quy tắc unpublish/thu hồi URL, cache, dọn media orphan.
- [ ] Production dùng MinIO service account giới hạn quyền, không dùng root local.

**Nghiệm thu:** tạo/sửa bài có ảnh/trích dẫn; ảnh draft không đọc anonymous;
file giả định dạng/quá lớn bị từ chối; renderer không tạo XSS qua link/ảnh.

### D — Autosave, preview và xuất bản

- [ ] Debounce autosave, lưu có version, xử lý conflict/nhiều tab/request sai thứ tự.
- [ ] Hiển thị Đang lưu / Đã lưu / Lỗi; cảnh báo rời trang khi còn thay đổi.
- [ ] Preview có auth, không cache, không index, dùng cùng renderer với public.
- [ ] Publish đợi lưu xong và chốt đúng version; unpublish thay đổi visibility nguyên tử.
- [ ] Thiết kế cache invalidation có xác thực, retry và giới hạn staleness.

**Nghiệm thu:** viết → autosave → preview → publish → sửa draft mà public không đổi
→ publish lại → unpublish. Mất mạng không báo lưu thành công giả; request cũ
không ghi đè; public không tiếp tục hiển thị bài đã unpublish ngoài giới hạn đã chốt.

### E — Kết nối website public và hoàn thiện trải nghiệm

- [ ] Thay nguồn `collection.ts` bằng API publication; demo chỉ giữ làm fixture/test.
- [ ] Renderer Tiptap phía server, validate cấu trúc, sanitize HTML và URL allowlist
  nếu dùng `@tiptap/html`; không chỉ tin dữ liệu đã lưu trong database.
- [ ] Danh sách phân trang, bài tiếp theo, empty/error/not-found đúng HTTP status.
- [ ] Metadata động, canonical, Open Graph, sitemap, robots và URL production.
- [ ] Ảnh responsive/tối ưu, kiểm tra layout shift, font và hiệu năng mobile.
- [ ] Kiểm tra lại reduced motion, keyboard, no-JS và reading progress với bài dài/ảnh thật.
- [ ] Chốt với chủ dự án xem nhịp cuộn/thiết kế mẫu đã đúng tinh thần bảo tàng chưa.

**Nghiệm thu:** public dùng dữ liệu thật, không lộ draft; bài đọc thoải mái trên
mobile; thiếu JS không làm mất nội dung; SEO không index admin/preview.

### F — Kiểm thử xuyên suốt và bàn giao production

- [ ] Unit nghiệp vụ, integration DB/storage/auth/cache, E2E đầy đủ luồng biên tập.
- [ ] Test XSS, CSRF, upload sai định dạng, expired session và lỗi dependency.
- [ ] Giữ coverage ≥80% phần application có thể kiểm thử; ghi rõ scope đo.
- [ ] Security/dependency/image scan lại; rà quyền storage, log và production config.
- [x] CI có PostgreSQL disposable, migration check và generation drift check sqlc.
  Integration auth/storage/cache sẽ bổ sung cùng tính năng tương ứng.
- [ ] Tài liệu env, proxy `/api`, health probe, resource requirements và migration job.
- [ ] Quy trình backup/restore cả PostgreSQL và media, thử restore và rollback thực tế.
- [ ] Bàn giao image/tag bất biến và hướng dẫn vận hành; chủ dự án tự làm deployment.

**Nghiệm thu:** có thể deploy application mà không sửa code theo cluster;
không secret trong image/repo; restore dữ liệu và rollback có quy trình đã thử.

## 6. Lưu ý vận hành và rủi ro còn mở

- MinIO local được build từ release `RELEASE.2025-10-15T17-29-55Z`, vì image
  upstream cũ không kéo được và thiếu bản vá CVE-2025-62506. Build đầu tốn vài phút.
  Image này chỉ phục vụ local; không mặc định dùng cho production hoặc coi là đã
  scan sạch. Xem xét AGPLv3 khi sửa/phân phối source/image.
- Local dùng root MinIO và transport không TLS trong môi trường loopback/container;
  production cần credential tối thiểu và TLS phù hợp.
- `storage-init` không sửa policy bucket đã có: bucket public từ trước sẽ không
  tự trở thành private. Hạ tầng phải kiểm tra policy trước khi dùng dữ liệu thật.
- Budget probe độc lập tránh route khác làm cạn quota; chưa ngăn được flood trực
  tiếp probe. Cần network/ingress restriction theo hạ tầng thực tế.
- Session, media lifecycle và cache invalidation còn mở. Migration tool đã pin;
  slug ban đầu bị khóa sau publish thay cho redirect, hợp đồng HTTP mới là đề xuất
  trong `docs/api-contract.md`, chưa phải route đã triển khai/nghiệm thu.
- `idea.md` là tài liệu đầu vào của chủ dự án; không xóa/ghi đè khi dọn working tree.
  Kiểm tra `git status` khi tiếp nhận; tài liệu này không bảo đảm code đã commit/push.

## 7. Prompt gợi ý để tiếp tục ở phiên sau

> Đọc `docs/handover.md`, `README.md`, `docs/design-direction.md` và kiểm tra
> trạng thái repository; đọc thêm `docs/persistence.md` và `docs/api-contract.md`.
> Persistence nội bộ đã có; tiếp tục mục B với CLI tạo admin, session, CSRF và
> kiểm thử biên bảo mật trước khi mở HTTP admin. Giữ frontend mẫu đang chạy và
> không triển khai k3s/Argo CD. Chốt hợp đồng HTTP/validation còn lại ở A trước
> khi nối CMS; không expose draft hoặc route ghi khi chưa có authentication,
> authorization và CSRF. Viết test trước, giữ kiểm thử transaction/version hiện
> có; cập nhật checklist, kết quả test và giới hạn sau mỗi lát cắt.
