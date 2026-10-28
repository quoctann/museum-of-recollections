import type { Metadata } from "next";
import { articles } from "@/lib/collection";
import { ArticleCard } from "@/components/article-card";

export const metadata: Metadata = {
  title: "Bộ sưu tập",
  description:
    "Ba tản văn minh họa về nơi chốn, đồ vật và những khoảnh khắc ở lại trong trí nhớ.",
};

export default function Collection() {
  return (
    <main id="main" tabIndex={-1} className="shell collection-page">
      <header className="collection-intro">
        <p className="eyebrow label">Phòng đọc / Bộ sưu tập số 01</p>
        <h1>
          Những điều <em>ở lại.</em>
        </h1>
        <div className="collection-description">
          <p>
            Không sắp xếp theo thời gian.
            <br />
            Chỉ theo những gì còn vang lên trong lòng.
          </p>
          <span className="label">03 tản văn · Đọc theo nhịp của bạn</span>
        </div>
      </header>
      <div className="collection-divider label">
        <span>Danh mục trưng bày</span>
        <span>2026 / 01—03</span>
      </div>
      <div className="collection-grid">
        {articles.map((article) => (
          <ArticleCard key={article.slug} article={article} />
        ))}
      </div>
      <p className="demo-note label">
        Ghi chú biên soạn — Ba tản văn do Ban biên soạn sáng tác để minh họa
        trải nghiệm đọc. Không phải tư liệu hay lời kể có thật.
      </p>
      <div className="end-mark" aria-hidden="true">
        ✳
      </div>
    </main>
  );
}
