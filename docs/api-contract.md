# Article persistence và hợp đồng HTTP dự kiến

## Phạm vi

Lát cắt A bổ sung persistence; **các route bài viết dưới đây chưa được đăng ký**.
API đang chạy chỉ có `/healthz` và `/readyz`. Website vẫn dùng bài mẫu.
Không được nối CMS vào route ghi chưa có authentication/authorization và CSRF.
Các quyết định dưới đây là mặc định triển khai ban đầu, có thể điều chỉnh trước
khi mở HTTP API; chưa phải hợp đồng HTTP đã nghiệm thu.

## Quy tắc dữ liệu

- `articles` định danh bigint dương ổn định. Slug nhập tường minh, tối đa 120 ký tự,
  chỉ `[a-z0-9]+(-[a-z0-9]+)*`; không tự chuyển dấu tiếng Việt. Tiêu đề/nội dung
  vẫn là Unicode. Slug duy nhất, chỉ được sửa trước lần publish đầu tiên.
- Sau lần publish đầu, slug bị khóa cả khi unpublish. Bản đầu không đổi URL,
  không có alias/redirect. Nếu cần thay đổi chính sách, phải thêm migration và
  kiểm thử redirect trước khi mở tính năng đổi slug.
- `article_drafts` là bản đang viết; `article_publications` là snapshot hiện hành.
  Public không đọc draft, không fallback sang draft nếu chưa publish.
- `version` là revision của aggregate biên tập, không chỉ phiên bản nội dung:
  save, publish và unpublish thành công đều tăng version. Client phải nhận version
  mới từ response. Yêu cầu cũ không được tự động ghi đè; giữ nội dung local để
  người viết đối chiếu khi conflict.
- Publish khóa article, kiểm tra version, chụp bản draft đã lưu, cập nhật dấu mốc
  xuất bản và version trong **một transaction**. Unpublish tắt cờ visibility trong
  transaction, giữ draft và snapshot cuối; public query bắt buộc lọc `published`.
  Chỉ lưu draft không ảnh hưởng nội dung public.
- Thời điểm xuất bản đầu giữ nguyên qua unpublish/republish. Lần publish gần nhất
  thể hiện riêng; public sắp theo xuất bản đầu giảm dần và ID giảm dần để phá hòa.
- Nội dung lưu JSONB `{ "schemaVersion": 1, "doc": { "type": "doc", ... } }`.
  Validation phía application không thay thế renderer an toàn. Bộ extension,
  media reference và quy tắc URL cuối cùng cần đồng bộ với Tiptap ở mục C.
- `admins` và `media` mới là schema nền; chưa có tài khoản, session, upload,
  quyền đọc media hoặc vòng đời xóa media. Không bật public bucket.

## Route và payload dự kiến (chưa triển khai)

| Method | Path | Kết quả |
| --- | --- | --- |
| GET | `/api/public/articles?limit=20&cursor=…` | Tóm tắt publication (không thân bài), `nextCursor` |
| GET | `/api/public/articles/{slug}` | Một snapshot hoặc 404 |
| POST | `/api/admin/articles` | Tạo draft, 201 |
| GET | `/api/admin/articles/{id}` | Draft + version + trạng thái publication |
| PUT | `/api/admin/articles/{id}/draft` | Thay toàn bộ draft với expectedVersion |
| POST | `/api/admin/articles/{id}/publish` | Publish đúng expectedVersion |
| POST | `/api/admin/articles/{id}/unpublish` | Unpublish đúng expectedVersion |

Create/save sẽ nhận `slug`, `title`, `subtitle`, `content`; save bắt buộc thêm
`expectedVersion`. Publish/unpublish chỉ nhận `{ "expectedVersion": 3 }`, không
nhận một bản content khác từ client. Publish phải đợi autosave hoàn tất.
Không nhận ID admin/owner từ client để quyết định quyền.

Envelope dự kiến:

```json
{ "success": true, "data": { "version": 4 } }
```

```json
{ "success": false, "error": { "code": "version_conflict", "message": "Bản nháp đã thay đổi. Hãy tải lại để đối chiếu." } }
```

- `400`: JSON/cursor/ID sai định dạng; `413`: body vượt 1 MiB; `422`: nội dung
  không hợp lệ; `404`: không tìm thấy; `409`: version conflict hoặc slug trùng.
- `401/403`: chưa xác thực/không được phép hoặc thiếu CSRF; `429` kèm Retry-After;
  `500/503`: lỗi nội bộ/phụ thuộc, không trả SQL, connection string hay raw error.
- Public list mặc định 20, tối đa 50; HTTP cursor dự kiến opaque mã hóa timestamp + ID bigint,
  kiểm tra định dạng/range, không dùng cursor làm SQL. Keyset không bảo đảm snapshot
  toàn bộ tập kết quả khi bài bị unpublish giữa hai request.
- List chỉ trả ID, slug, title/subtitle và mốc xuất bản; chỉ detail trả `content`.
  Không tải body tối đa 1 MiB của từng bài vào một trang danh sách.
- HTTP adapter phải giới hạn body trước decode, từ chối field không hỗ trợ,
  validate lại dữ liệu và ánh xạ typed error. Giới hạn payload HTTP này chưa có hiệu
  lực vì route chưa tồn tại.

## Biên bảo mật và cache trước khi mở route

Admin đơn lẻ tạo qua CLI; không đăng ký công khai. Mọi route admin (kể cả đọc và
preview) phải qua authentication/authorization. Cookie HttpOnly, Secure ở
production, SameSite và CSRF cho mutation; session/revocation cần chốt ở mục B.
Login có budget riêng; không dùng budget operational probe.

Draft/preview: `Cache-Control: private, no-store`, `noindex`, không lộ qua sitemap.
Public API ban đầu nên `no-store` để unpublish có hiệu lực ở lần đọc tiếp theo;
cache frontend/CDN chỉ được bật khi có invalidation và giới hạn staleness đã kiểm thử.
Proxy `/api` cùng origin chưa được cấu hình. Không mở CORS rộng để né thiết kế này.
