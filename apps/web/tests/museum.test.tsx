import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ArticleCard } from "@/components/article-card";
import { ArticleBody } from "@/components/article-body";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import { Art } from "@/components/art";
import { ReadingProgress } from "@/components/reading-progress";
import { articles } from "@/lib/collection";
import Home from "@/app/page";
import Collection from "@/app/articles/page";
import NotFound from "@/app/not-found";

describe("museum semantic surfaces", () => {
  it("offers a skip link and labelled primary navigation", () => {
    render(
      <>
        <SiteHeader />
        <SiteFooter />
      </>,
    );
    expect(
      screen.getByRole("link", { name: "Đến nội dung chính" }),
    ).toHaveAttribute("href", "#main");
    expect(
      screen.getByRole("navigation", { name: "Điều hướng chính" }),
    ).toBeVisible();
    expect(
      screen.getAllByRole("link", { name: /Bộ sưu tập/ })[0],
    ).toHaveAttribute("href", "/articles");
    expect(screen.getByText(/không kết nối CMS/i)).toBeVisible();
  });

  it("renders a meaningful card without nested links", () => {
    render(<ArticleCard article={articles[0]} />);
    expect(screen.getByRole("link")).toHaveAttribute(
      "href",
      `/articles/${articles[0].slug}`,
    );
    expect(screen.getByRole("heading", { level: 2 })).toHaveTextContent(
      articles[0].title,
    );
    expect(screen.getByText(/Tản văn/)).toBeVisible();
  });

  it("renders every safe block type and never interprets text as HTML", () => {
    render(
      <ArticleBody
        blocks={[
          { type: "paragraph", text: "<script>alert(1)</script>" },
          { type: "heading", text: "Một khoảng lặng" },
          { type: "quote", text: "Đi chậm lại." },
          { type: "divider" },
        ]}
      />,
    );
    expect(screen.getByText("<script>alert(1)</script>")).toBeVisible();
    expect(document.querySelector("script")).toBeNull();
    expect(screen.getByRole("heading", { level: 2 })).toHaveTextContent(
      "Một khoảng lặng",
    );
    expect(screen.getByRole("separator")).toBeInTheDocument();
    expect(document.querySelector("blockquote")).toHaveTextContent(
      "Đi chậm lại.",
    );
  });

  it("keeps art visible without an opacity-zero reveal", () => {
    render(
      <Art src={articles[0].art} alt="Nghiên cứu về một mái nhà" priority />,
    );
    expect(
      screen.getByRole("img", { name: "Nghiên cứu về một mái nhà" }),
    ).toBeVisible();
    expect(screen.getByRole("img").parentElement).not.toHaveStyle({
      opacity: "0",
    });
  });

  it("has one main heading on home, collection and error pages", () => {
    const home = render(<Home />);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      "Có những điều",
    );
    expect(
      screen.getByRole("link", { name: /Bước vào bộ sưu tập/ }),
    ).toHaveAttribute("href", "/articles");
    home.unmount();
    const collection = render(<Collection />);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      "Những điều ở lại",
    );
    expect(screen.getAllByRole("heading", { level: 2 })).toHaveLength(3);
    collection.unmount();
    render(<NotFound />);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      "Một ngăn ký ức còn trống",
    );
    expect(screen.getByRole("link", { name: /Về bộ sưu tập/ })).toHaveAttribute(
      "href",
      "/articles",
    );
  });
});

describe("progress enhancement", () => {
  it("stays absent when no article is present", () => {
    render(<ReadingProgress />);
    expect(screen.queryByRole("progressbar")).not.toBeInTheDocument();
  });

  it("measures the article, updates on scroll, and cleans up listeners", () => {
    const article = document.createElement("div");
    article.id = "essay-body";
    article.getBoundingClientRect = () =>
      ({ top: 200 - window.scrollY, height: 2000 }) as DOMRect;
    document.body.append(article);
    Object.defineProperty(window, "scrollY", {
      configurable: true,
      value: 0,
      writable: true,
    });
    Object.defineProperty(window, "innerHeight", {
      configurable: true,
      value: 1000,
      writable: true,
    });
    const removed = vi.spyOn(window, "removeEventListener");
    const view = render(<ReadingProgress />);
    expect(
      screen.getByRole("progressbar", { name: "Tiến độ đọc bài" }),
    ).toHaveAttribute("aria-valuenow", "0");
    window.scrollY = 700;
    fireEvent.scroll(window);
    expect(screen.getByRole("progressbar")).toHaveAttribute(
      "aria-valuenow",
      "50",
    );
    fireEvent.resize(window);
    view.unmount();
    expect(removed).toHaveBeenCalledWith("scroll", expect.any(Function));
    article.remove();
    window.scrollY = 0;
  });
});
