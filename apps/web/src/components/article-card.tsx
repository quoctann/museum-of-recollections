import Link from "next/link";
import { type Article, readingMinutes } from "@/lib/collection";
import { Art } from "./art";

export function ArticleCard({ article }: { article: Article }) {
  return (
    <article className={`article-card specimen-${article.number}`}>
      <Link href={`/articles/${article.slug}`} className="card-link">
        <div className="card-art">
          <Art src={article.art} alt={article.artAlt} />
          <span className="art-catalogue label" aria-hidden="true">
            HƯ / {article.number}
          </span>
        </div>
        <div className="card-meta label">
          <span>{article.category}</span>
          <span>Tản văn · {readingMinutes(article.blocks)} phút đọc</span>
        </div>
        <h2>
          {article.title}
          <span className="card-arrow" aria-hidden="true">
            ↗
          </span>
        </h2>
        <p>{article.subtitle}</p>
      </Link>
    </article>
  );
}
