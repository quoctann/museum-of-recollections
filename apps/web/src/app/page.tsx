import Link from "next/link";
import { articles } from "@/lib/collection";
import { Art } from "@/components/art";
import { ArticleCard } from "@/components/article-card";

export default function Home() {
  return (
    <main id="main" tabIndex={-1}>
      <section className="hero shell" aria-labelledby="home-title">
        <div className="hero-copy">
          <p className="eyebrow label">
            <span className="tiny-cross" aria-hidden="true">
              ✳
            </span>{" "}
            Bảo tàng của những điều bình thường
          </p>
          <h1 id="home-title">
            Có những điều
            <br />
            chỉ còn trong
            <br />
            <em>trí nhớ.</em>
          </h1>
          <p className="hero-description">
            Một mái nhà. Một món đồ cũ. Một buổi chiều không tên.
            <br className="desktop-break" /> Chúng tôi giữ lại những điều nhỏ bé
            ấy bằng câu chữ.
          </p>
          <Link href="/articles" className="text-link">
            Bước vào bộ sưu tập <span aria-hidden="true">↗</span>
          </Link>
          <div className="hero-footnote label">
            <span className="status-dot" /> Không cần vội. Ở đây, thời gian đi
            chậm.
          </div>
        </div>
        <figure className="hero-art">
          <Art
            src="/art/house.svg"
            alt="Một mái nhà màu đất dưới cơn mưa, bản khắc hình học nguyên bản"
            priority
          />
          <figcaption className="label">
            <span>Hiện vật 001 — Một nơi để trở về</span>
            <span>2026</span>
          </figcaption>
          <span className="vertical-note label" aria-hidden="true">
            Những gì đã qua · Những gì ở lại
          </span>
        </figure>
      </section>

      <section
        className="collection-preview shell"
        aria-labelledby="collection-title"
      >
        <div className="section-heading">
          <div>
            <p className="label eyebrow">01 / Những trang ký ức</p>
            <h2 id="collection-title">
              Chuyện nhỏ, <em>dư âm dài.</em>
            </h2>
          </div>
          <Link href="/articles" className="text-link small-link">
            Xem tất cả <span className="label">(03)</span>{" "}
            <span aria-hidden="true">↗</span>
          </Link>
        </div>
        <div className="preview-grid">
          {articles.slice(0, 2).map((article) => (
            <ArticleCard key={article.slug} article={article} />
          ))}
        </div>
        <p className="demo-note label">
          Ghi chú biên soạn — Các tản văn là sáng tác minh họa, không phải tư
          liệu hay lời kể có thật.
        </p>
      </section>

      <section
        className="about-section shell"
        id="about"
        aria-labelledby="about-title"
      >
        <div className="about-index label">
          02 / Lời ngỏ<span aria-hidden="true">✳</span>
        </div>
        <div className="about-copy">
          <h2 id="about-title">
            Không phải mọi điều
            <br />
            đáng nhớ đều <em>lớn lao.</em>
          </h2>
          <p>
            Bảo tàng này không có tủ kính. Chỉ có những câu chuyện về điều ta
            từng đi qua, từng chạm vào, từng ngỡ đã quên.
          </p>
          <p>
            Mời bạn dừng lại một chút. Biết đâu, trong ký ức của một người khác,
            bạn nhận ra một phần của mình.
          </p>
          <span className="label about-signature">
            Một không gian nhỏ, dành cho những hồi ức dài.
          </span>
        </div>
      </section>
    </main>
  );
}
