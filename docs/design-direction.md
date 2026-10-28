# Một cuộc dạo qua ký ức

## Mục tiêu

Trang chủ giống một phòng trưng bày; trang bài viết giống một chỗ ngồi yên
để đọc. Hiệu ứng giúp chuyển nhịp, không trở thành điều kiện để đọc nội dung.

## Hệ thống thị giác

- Giấy trắng ngà `#F5F1E9`, mực than `#25231F`, điểm nhấn nâu đỏ `#895442`.
- Noto Serif cho văn bản và tiêu đề; Be Vietnam Pro cho nhãn, điều hướng.
  Self-host, không yêu cầu truy cập Google Fonts lúc build hoặc lúc đọc.
- Ảnh hoặc tranh đứng riêng như hiện vật, nhãn nhỏ và đường kẻ mảnh.
- Khoảng trắng rộng; không card đổ bóng, không giao diện dashboard ở public.
- Mobile dùng một luồng dọc; desktop cho phép nhịp bố cục lệch trái/phải.

## Trải nghiệm đọc

- Body khoảng 18px trên mobile, line-height 1.8, độ rộng tối đa khoảng 65ch.
- Thân bài không reveal từng đoạn hoặc từng chữ.
- Skip link, focus nhìn rõ, heading đúng thứ bậc; nhãn tiếng Việt.
- Có trạng thái bài không tồn tại và đường về danh sách.
- Bài mẫu là nội dung minh họa, không giả lập publication đã lưu vào database.

## Chuyển động

- Giữ native scroll; không smooth-scroll engine, scroll hijacking hoặc snap.
- Chỉ nhấn chuyển động ở ảnh, đầu bài và chuyển tiếp cuối bài.
- Nội dung có sẵn trong HTML server, không opacity 0 chờ JavaScript.
- `prefers-reduced-motion` tắt chuyển động không cần thiết.
- Progress dựa trên vùng đọc, không phải thước đo thời gian hay buộc người đọc
  phải tương tác để hoàn thành.

## Những việc cần đánh giá ở vòng thiết kế

1. Dấu tiếng Việt ở tiêu đề lớn và cỡ chữ nhỏ.
2. Đọc một bài dài trên mobile, không chỉ xem hero desktop.
3. Cuộn bằng bàn phím, reduced motion và JavaScript bị tắt.
4. Tỷ lệ tranh/ảnh với văn bản; khoảng thở không thành khoảng trống vô nghĩa.
5. Sau này thay tranh minh họa bằng ảnh có quyền sử dụng mà không phá bố cục.
