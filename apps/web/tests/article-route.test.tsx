import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ArticlePage, {
  generateMetadata,
  generateStaticParams,
} from "@/app/articles/[slug]/page";
import { articles } from "@/lib/collection";

vi.mock("next/navigation", () => ({
  notFound: () => {
    throw new Error("NEXT_NOT_FOUND");
  },
}));

describe("static article route integration", () => {
  it("pre-renders precisely the curated slugs", () => {
    expect(generateStaticParams()).toEqual(
      articles.map(({ slug }) => ({ slug })),
    );
  });

  it.each(articles)(
    "renders $slug with metadata, safe prose and next story",
    async (article) => {
      const params = Promise.resolve({ slug: article.slug });
      render(await ArticlePage({ params }));
      expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
        article.title,
      );
      expect(
        screen.getByRole("navigation", { name: "Đọc tiếp" }),
      ).toBeInTheDocument();
      expect(screen.getByText("Sáng tác minh họa")).toBeVisible();
      expect(document.querySelector("#essay-body")).toHaveTextContent(
        article.blocks[0].type === "divider" ? "" : article.blocks[0].text,
      );
      expect(await generateMetadata({ params })).toMatchObject({
        title: article.title,
        description: article.subtitle,
      });
    },
  );

  it("rejects an unknown slug instead of inventing an article", async () => {
    const params = Promise.resolve({ slug: "does-not-exist" });
    await expect(ArticlePage({ params })).rejects.toThrow("NEXT_NOT_FOUND");
    expect(await generateMetadata({ params })).toEqual({
      title: "Không tìm thấy trang",
    });
  });
});
