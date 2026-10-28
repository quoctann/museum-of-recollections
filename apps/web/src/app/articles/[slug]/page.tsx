import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import {
  articles,
  findArticle,
  nextArticle,
  readingMinutes,
} from "@/lib/collection";
import { Art } from "@/components/art";
import { ArticleBody } from "@/components/article-body";
import { ReadingProgress } from "@/components/reading-progress";

type Props = { params: Promise<{ slug: string }> };

export function generateStaticParams() {
  return articles.map(({ slug }) => ({ slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const article = findArticle((await params).slug);
  if (!article) return { title: "Không tìm thấy trang" };
  return {
    title: article.title,
    description: article.subtitle,
    openGraph: {
      title: article.title,
      description: article.subtitle,
      type: "article",
      locale: "vi_VN",
    },
  };
}

export default async function ArticlePage({ params }: Props) {
  const article = findArticle((await params).slug);
  if (!article) notFound();
  const next = nextArticle(article.slug)!;
  return (
    <main id="main" tabIndex={-1}>
      <ReadingProgress />
      <article>
        <header className="essay-header shell">
          <Link href="/articles" className="back-link label">
            <span aria-hidden="true">←</span> Bộ sưu tập
          </Link>
          <div className="essay-heading-grid">
            <div className="essay-title">
              <p className="eyebrow label">
                Hiện vật {article.number} / {article.category}
              </p>
              <h1>{article.title}</h1>
              <p className="essay-subtitle">{article.subtitle}</p>
            </div>
            <div className="essay-catalogue label">
              <span>Tản văn / {readingMinutes(article.blocks)} phút đọc</span>
              <span>{article.author}</span>
              <time dateTime={article.date}>
                {article.date.split("-").reverse().join(".")}
              </time>
              <span className="demo-tag">Sáng tác minh họa</span>
            </div>
          </div>
        </header>
        <figure className="essay-figure shell">
          <Art src={article.art} alt={article.artAlt} priority />
          <figcaption className="label">
            <span>
              Nghiên cứu hình khối / {article.category.toLocaleLowerCase("vi")}
            </span>
            <span>Minh họa nguyên bản · HƯ {article.number}</span>
          </figcaption>
        </figure>
        <div className="reading-layout shell">
          <aside className="reading-aside label" aria-hidden="true">
            <span>HƯ / {article.number}</span>
            <span>
              Một khoảng
              <br />
              để đọc chậm.
            </span>
          </aside>
          <div>
            <ArticleBody blocks={article.blocks} />
            <div className="essay-end">
              <span aria-hidden="true">✳</span>
              <p className="label">Hết một trang. Còn một dư âm.</p>
            </div>
            <p className="essay-demo label">
              Tản văn minh họa do Ban biên soạn sáng tác cho bản trưng bày này;
              không phải lời kể hay tư liệu có thật.
            </p>
          </div>
        </div>
      </article>
      <nav className="next-article shell" aria-label="Đọc tiếp">
        <span className="label">Một ký ức khác / {next.number}</span>
        <Link href={`/articles/${next.slug}`}>
          <span>{next.title}</span>
          <span aria-hidden="true">↗</span>
        </Link>
        <p>{next.subtitle}</p>
      </nav>
    </main>
  );
}
