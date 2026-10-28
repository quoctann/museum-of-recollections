# Persistence: migration, generation và kiểm thử

## Tool được pin

- goose `v3.24.1` (migration PostgreSQL).
- sqlc `v1.29.0`, `database/sql`; query sinh mã được lưu trong repository.
- `sqlx` quản lý connection/transaction; adapter dùng generated queries với
  transaction, không tự viết lại lớp SQL.

Các lệnh tool chạy bằng `go run module@version`, cần mạng lần đầu. Không tự
chạy migration trong API startup, không cài tool bằng `@latest`.

## Local

```sh
make infra
make migrate-status
make migrate
make sqlc-check
make test-integration
```

`make migrate*` nạp **file `.env` tin cậy** như `make dev-api`. Để chạy với URL
đã cấp từ secret manager mà không nạp `.env`, dùng trực tiếp:

```sh
python3 scripts/api_tools.py status
python3 scripts/api_tools.py up
```

`DATABASE_URL` phải có sẵn trong environment; không đưa secret trực tiếp vào
command line/history. Script truyền connection string qua `GOOSE_DBSTRING`, không
thông qua argv. Tool migration cần role DDL; API production nên dùng role DML
riêng. Migration runner chưa nằm trong runtime Docker image hiện tại.

Sau khi sửa schema/query, chạy `make sqlc`, review SQL và generated diff,
`make sqlc-check` rồi unit/integration tests. Drift check so sánh byte trong thư
mục tạm, không phụ thuộc trạng thái git và không ghi đè generated đang dùng.

## Rollback

**Migration đầu tiên down sẽ xóa bảng và dữ liệu nghiệp vụ.** Không dùng rollback
như cách sửa lỗi thông thường và không xóa development volumes.

1. Dừng ghi nội dung, xác định version hiện tại bằng `migrate-status`.
2. Backup PostgreSQL và media tương ứng; kiểm tra bản backup có thể restore.
3. Review phần `Down`, xác nhận có chấp nhận mất dữ liệu hay không. Production
   ưu tiên forward-fix; restore backup cần quy trình vận hành riêng.
4. Khi đã chấp nhận thao tác phá hủy, `make migrate-down` lùi **một** migration.
5. `make migrate-status`; deploy binary tương thích schema và smoke test.

Test disposable kiểm tra up → up (idempotent) → down → up bằng goose; không phải
chứng nhận backup/restore production đã được diễn tập.

## PostgreSQL thật, không đụng development

`make test-integration` tạo project Compose tên ngẫu nhiên `museum-test-*`,
PostgreSQL 17 với tmpfs, password ngẫu nhiên và cổng loopback được Docker cấp.
Không dùng named volume hay `DATABASE_URL` phát triển. Script luôn dọn đúng project
này khi kết thúc, kể cả test fail. Khi bị kill cưỡng bức, kiểm tra project còn lại
trước khi dọn; không dùng prune toàn bộ Docker.

Integration suite chỉ chạy khi có `TEST_DATABASE_URL` tường minh; không fallback
qua `DATABASE_URL`. Không đặt biến test trỏ vào DB development/production.
Runner mặc định cấp server dùng một lần; test tự tạo database riêng bên trong.
Nếu chạy thủ công với `TEST_DATABASE_URL`, role test cần quyền tạo/xóa database;
chỉ cấp role này trên PostgreSQL disposable, không dùng role API production.
Unit tests bình thường không cần Docker. Coverage integration ghi vào
`apps/api/coverage-integration.out` (ignored); tách generated SQL khỏi báo cáo
coverage nghiệp vụ khi đánh giá mục tiêu 80%.

Xem [hợp đồng dự kiến](api-contract.md). Bước sau là authentication trước khi
mở HTTP admin, không phải bật route ghi ẩn sau giao diện CMS.
