import Link from "next/link";

export default function NotFound() {
  return (
    <main id="main" tabIndex={-1} className="not-found shell">
      <p className="label eyebrow">404 / Hiện vật chưa được lưu giữ</p>
      <span className="empty-frame" aria-hidden="true">
        ∅
      </span>
      <h1>
        Một ngăn ký ức <br />
        <em>còn trống.</em>
      </h1>
      <p>
        Trang bạn tìm không có ở đây.
        <br />
        Nhưng vẫn còn những câu chuyện đang chờ.
      </p>
      <Link href="/articles" className="text-link">
        Về bộ sưu tập <span aria-hidden="true">↗</span>
      </Link>
    </main>
  );
}
