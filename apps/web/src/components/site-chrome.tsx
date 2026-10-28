import Link from "next/link";

export function SiteHeader() {
  return (
    <>
      <a className="skip-link" href="#main">
        Đến nội dung chính
      </a>
      <header className="site-header shell">
        <Link
          href="/"
          className="wordmark"
          aria-label="Bảo tàng hồi ức — Trang chủ"
        >
          <span className="museum-mark" aria-hidden="true">
            m<span>r</span>
          </span>
          <span>
            Bảo tàng
            <br />
            hồi ức
            <span className="wordmark-translation">
              Museum of Recollections
            </span>
          </span>
        </Link>
        <nav aria-label="Điều hướng chính">
          <Link href="/articles" className="nav-link">
            Bộ sưu tập <span aria-hidden="true">↗</span>
          </Link>
          <Link href="/#about" className="nav-link about-nav">
            Về bảo tàng
          </Link>
        </nav>
        <span className="header-note label">Một nơi để nhớ.</span>
      </header>
    </>
  );
}

export function SiteFooter() {
  return (
    <footer className="site-footer shell">
      <div className="footer-top">
        <Link href="/" className="footer-name">
          Bảo tàng hồi ức<span aria-hidden="true"> ↗</span>
        </Link>
        <p>
          Giữ lại một chút.
          <br />
          Để ngày mai còn nhớ.
        </p>
      </div>
      <div className="footer-bottom label">
        <span>© 2026 · Một không gian đọc độc lập</span>
        <span>Bản trưng bày tĩnh · không kết nối CMS hay API</span>
        <a href="#top">
          Trở lên đầu <span aria-hidden="true">↑</span>
        </a>
      </div>
    </footer>
  );
}
