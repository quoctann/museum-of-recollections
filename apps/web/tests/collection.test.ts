import { describe, expect, it } from "vitest";
import {
  articles,
  findArticle,
  nextArticle,
  readingMinutes,
} from "@/lib/collection";
import { readingProgress } from "@/lib/progress";

describe("curated static collection", () => {
  it("contains three distinct, explicitly demo-labelled Vietnamese essays", () => {
    expect(articles).toHaveLength(3);
    expect(new Set(articles.map((article) => article.slug)).size).toBe(3);
    for (const article of articles) {
      expect(article.language).toBe("vi");
      expect(article.demo).toBe(true);
      expect(article.author).toBe("Ban biên soạn");
      expect(article.blocks.length).toBeGreaterThan(5);
      expect(article.art).toMatch(/^\/art\/.+\.svg$/);
      expect(readingMinutes(article.blocks)).toBeGreaterThan(0);
      expect(findArticle(article.slug)).toBe(article);
    }
  });

  it("returns no article for empty, Unicode, hostile or unknown slugs", () => {
    for (const slug of [
      "",
      "không có",
      "<script>",
      "' OR 1=1",
      "constructor",
    ]) {
      expect(findArticle(slug)).toBeUndefined();
      expect(nextArticle(slug)).toBeUndefined();
    }
  });

  it("cycles through the collection without a dead end", () => {
    expect(nextArticle(articles[0].slug)).toBe(articles[1]);
    expect(nextArticle(articles[1].slug)).toBe(articles[2]);
    expect(nextArticle(articles[2].slug)).toBe(articles[0]);
  });

  it("estimates reading time for empty and long collections", () => {
    expect(readingMinutes([])).toBe(1);
    expect(
      readingMinutes([{ type: "paragraph", text: "từ ".repeat(10000) }]),
    ).toBe(40);
  });
});

describe("honest article-only reading progress", () => {
  it("tracks the readable range and clamps outside it", () => {
    expect(readingProgress(100, 200, 1000, 500)).toBe(0);
    expect(readingProgress(200, 200, 1000, 500)).toBe(0);
    expect(readingProgress(450, 200, 1000, 500)).toBe(50);
    expect(readingProgress(700, 200, 1000, 500)).toBe(100);
    expect(readingProgress(2000, 200, 1000, 500)).toBe(100);
  });

  it("does not invent progress for a short article or invalid geometry", () => {
    expect(readingProgress(0, 200, 100, 500)).toBe(0);
    expect(readingProgress(200, 200, 100, 500)).toBe(100);
    expect(readingProgress(NaN, 0, 1000, 500)).toBe(0);
    expect(readingProgress(0, 0, -1, 500)).toBe(0);
    expect(readingProgress(0, 0, 1000, 0)).toBe(0);
    expect(readingProgress(Infinity, 0, 1000, 500)).toBe(0);
  });
});
