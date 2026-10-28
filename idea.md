Ý tưởng:
Mình đang muốn triển khai app dạng Literary Journal / Digital Literature Experience — một website đọc tản văn mang phong cách tạp chí nghệ thuật, kết hợp hiệu ứng cuộn trang (scrollytelling), hình ảnh, typography và một hệ thống CMS nhỏ để tự viết và xuất bản bài.

Phong cách (styling): Minimal Editorial / Literary Magazine: nền giấy, chữ có chân, khoảng trắng rộng, hình ảnh giàu cảm xúc, animation nhẹ nhàng, font chữ tương thích tốt tiếng Việt

Một số hiệu ứng mình thấy phù hợp với chủ đề này:
- Trang chủ	Hero với typography lớn, background parallax rất nhẹ
- Danh sách bài	Card và tiêu đề xuất hiện lần lượt khi scroll
- Đầu bài viết	Ảnh cover reveal, tiêu đề fade-in
- Giữa bài	Ảnh minh họa hiện dần, đôi khi một câu trích dẫn nổi bật
- Trang đọc	Reading progress bar ở phía trên
- Cuối bài	Chuyển tiếp nhẹ sang bài tiếp theo

Techstack ý tưởng:
FE: NextJS+TS, styling Tailwindcss, animation với Motion.dev (hiện chưa có hiệu ứng từng chương nên chưa sử dụng GSAP, nhưng hãy tạo dư địa để tích hợp sau này), editor với Tiptap, admin ui với shadcn/ui
BE: golang, gin, postgres 17, sqlx, sqlc, zap logger
Deployment: self-hosted k3s argocd app of apps
Mobile first, ì đây là website đọc văn, khả năng đọc dễ chịu phải được ưu tiên hơn hiệu ứng. Mình đề xuất body text khoảng 17–19px trên mobile, line-height 1.7–1.9, bề rộng nội dung tối đa khoảng 65–70 ký tự và khoảng trắng thoáng giữa các đoạn
Tránh hiệu ứng ép cuộn từng màn hình hoặc liên tục làm mờ từng chữ trong nội dung dài. Trên mobile nên giảm parallax, hạn chế các section pin, và hỗ trợ prefers-reduced-motion để người đọc có thể tắt hiệu ứng chuyển động
Các chức năng MVP nên có: title, subtitle, nội dung rich text, trích dẫn, chèn ảnh, autosave draft, preview và nút Publish/Unpublish
Về lưu trữ nội dung, mình đề xuất dùng JSONB trong PostgreSQL cho cấu trúc Tiptap. Khi hiển thị public, Next.js có thể chuyển JSON thành HTML với các extension tương ứng qua @tiptap/html, đồng thời xử lý an toàn nội dung HTML
Chỉ 1 tài khoản admin/editor nên không cần api register, tối giản với JWT

Điểm mình muốn nhấn mạnh: Website văn học đẹp không nhất thiết phải có quá nhiều animation. Sự kết hợp giữa typography tốt, hình ảnh có chiều sâu, khoảng trắng và một vài hiệu ứng cuộn tinh tế thường mang lại trải nghiệm đọc tốt hơn rất nhiều so với việc mọi đoạn văn đều chuyển động.

Hãy thiết kế hệ thống clean code, dùng hexagonal cho backend và nextjs practice cho frontend, cung cấp 1 makefile để start local




