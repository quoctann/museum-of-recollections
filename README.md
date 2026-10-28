# Museum of Recollections

Một bảo tàng nhỏ của những ký ức: website đọc tản văn với nền giấy, chữ mực
và nhịp cuộn tự nhiên. Ưu tiên khả năng đọc trên mobile hơn hiệu ứng.

## Trạng thái

Đây là **lát cắt đầu tiên**, chưa phải toàn bộ MVP:

- Website công khai với bài mẫu, danh sách bài và trang đọc.
- Typography tiếng Việt self-hosted, chuyển động nhẹ, hỗ trợ reduced motion.
- Go API: liveness, readiness và kết nối PostgreSQL/MinIO.
- Lệnh bootstrap riêng tạo bucket private local, không tạo bucket khi API khởi động.
- Môi trường local, kiểm thử và Dockerfile cho hai application.
- Persistence bài viết: migration/sqlc, draft tách publication, optimistic version
  và publish/unpublish trong transaction; chưa expose qua HTTP.

**Chưa có:** CMS, đăng nhập, editor/renderer Tiptap, autosave, upload hoặc luồng
biên tập qua HTTP. Nội dung mẫu được phục vụ trực tiếp bởi frontend; API chưa
cung cấp route bài viết. Xem [persistence](docs/persistence.md) và
[hợp đồng API dự kiến](docs/api-contract.md).

## Chạy local

Yêu cầu: Node.js 22, npm 10, Go (xem `apps/api/go.mod`), Python 3,
Docker và Docker Compose v2 trở lên.
`make test`/`make check` có Go race detector, cần CGO và C compiler (GCC/Clang).

```sh
make setup
make up
```

`make setup` tạo `.env` với credential ngẫu nhiên, quyền `0600`, không ghi đè
cấu hình hiện có, sau đó cài frontend theo lockfile. `.env.example` chỉ mô tả
cấu hình, không chứa credential dùng thật.
Nếu dùng credential riêng, đồng bộ `POSTGRES_*`, `DATABASE_URL` (host local)
và `DATABASE_URL_CONTAINER` (host `postgres`); percent-encode ký tự đặc biệt
trong username/password của hai URL, không encode giá trị `POSTGRES_PASSWORD`.

MinIO local được build từ source release `RELEASE.2025-10-15T17-29-55Z`
(đã vá CVE-2025-62506), không dựa vào image MinIO cũ không còn khả dụng.
Lần build đầu cần mạng và có thể mất vài phút. Source MinIO dùng AGPLv3;
xem xét điều khoản license nếu phân phối image hoặc sửa đổi source.

| Dịch vụ | Địa chỉ local |
| --- | --- |
| Website | http://localhost:3000 |
| API liveness | http://localhost:8080/healthz |
| API readiness | http://localhost:8080/readyz |
| MinIO S3 | http://localhost:9000 |
| MinIO Console | http://localhost:9001 |
| PostgreSQL | localhost:55432 |

Đổi `POSTGRES_PORT` và cổng trong `DATABASE_URL` nếu `55432` đã được sử dụng.
Cổng PostgreSQL nội bộ container và `DATABASE_URL_CONTAINER` vẫn là `5432`.

Đăng nhập MinIO Console bằng `MINIO_ROOT_USER` và `MINIO_ROOT_PASSWORD` trong
`.env`. Các cổng Docker chỉ bind loopback. Credential root MinIO chỉ dành cho
local; production cần tài khoản service giới hạn quyền.

### Phát triển với hot reload

```sh
make dev
```

Sau đó mở hai terminal riêng:

```sh
make dev-web
```

```sh
make dev-api
```

Không chạy đồng thời `make up` và các dev server trên cùng cổng. Nếu đang chạy
toàn bộ stack, dùng `make down` trước rồi `make dev`.
Chỉ dùng `.env` tin cậy: `make dev-api` nạp file bằng shell.

`make logs` xem log; `make down` dừng stack nhưng giữ nguyên dữ liệu.
PostgreSQL và MinIO dùng named volume. Không xóa volume nếu cần giữ dữ liệu.

## Kiểm thử

```sh
make check
make test-integration
make sqlc-check
make build
npm --prefix apps/web run test:coverage
npm --prefix apps/web exec -- playwright install chromium
npm --prefix apps/web run test:e2e
```

Unit test không cần dịch vụ Docker. Readiness thực tế có thể kiểm tra sau
`make up` bằng `curl -f http://localhost:8080/readyz`.
E2E cần Chromium và thư viện hệ thống tương ứng; trên Linux có thể cần
`playwright install --with-deps chromium`.

## Cấu trúc

```text
apps/web/       Next.js App Router, TypeScript, Tailwind, Motion
apps/api/       Go, Gin, application ports và infrastructure adapters
scripts/        Tạo cấu hình local an toàn
tests/          Kiểm thử bootstrap local
compose.yaml    PostgreSQL 17, MinIO và application profile
```

## Hợp đồng vận hành

- Web lắng nghe `PORT` (mặc định `3000`), bind `HOSTNAME=0.0.0.0` trong container.
- API cấu hình bằng `API_PORT`, `DATABASE_URL`, `S3_ENDPOINT`, `S3_ACCESS_KEY`,
  `S3_SECRET_KEY`, `S3_BUCKET`, `S3_USE_SSL`. Endpoint S3 là `host:port`, không
  có protocol; TLS điều khiển bằng `S3_USE_SSL`.
- `GET /healthz`: tiến trình đang sống; không phụ thuộc database/storage.
- `GET /readyz`: kiểm tra dependency với deadline; lỗi trả `503`, không lộ secret.
- Container chạy non-root; log ra stdout/stderr. API xử lý graceful shutdown.
- Credential phải cấp qua secret manager của hạ tầng, không nhúng vào image.
- Không tự chạy migration khi replica khởi động. Migration chạy riêng bằng
  `make migrate` ở local; CLI tạo tài khoản admin chưa được triển khai.
- Bucket và chính sách media production sẽ được hoàn thiện ở phase upload;
  không tự bật public bucket hoặc cấp quyền public cho draft.
- `storage-init` là tác vụ local riêng, idempotent; Compose chạy xong mới mở API.
  Production có thể provision bucket qua hạ tầng hoặc chạy binary này có chủ đích.
- Docker Compose chỉ dành cho local, **không phải** cấu hình production.

Bạn quản lý k3s, Argo CD, PostgreSQL và MinIO production riêng. Repository
chưa cung cấp Kubernetes manifest hay app-of-apps. Các bước sau sẽ tách endpoint
S3 nội bộ khỏi URL media công khai và giữ credential storage ở phía server.

## Hướng phát triển tiếp theo

Xem **[tài liệu handover](docs/handover.md)** để có trạng thái chi tiết,
backlog theo thứ tự, tiêu chí nghiệm thu và prompt tiếp tục ở phiên sau.

1. Chốt giao diện từ trang chủ và trang đọc mẫu.
2. Persistence nội bộ đã có; tiếp theo mở HTTP có validation và biên bảo mật.
3. Đăng nhập một admin, CSRF, CMS/Tiptap và upload ảnh riêng tư.
4. Nối version/publish vào autosave và preview; thêm cache invalidation.
5. Kiểm thử xuyên suốt luồng biên tập, backup/restore và bàn giao vận hành.
