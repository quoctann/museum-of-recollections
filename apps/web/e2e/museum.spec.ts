import { test, expect } from "@playwright/test";

test("a visitor enters, reads, and finds another memory", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto("/");
  await expect(page.locator("html")).toHaveAttribute("lang", "vi");
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Có những điều",
  );
  await page.getByRole("link", { name: /Bước vào bộ sưu tập/ }).click();
  await expect(page).toHaveURL("/articles");
  await expect(page.locator("main article")).toHaveCount(3);
  await page.getByRole("link", { name: /Ngôi nhà trong tiếng mưa/ }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    "Ngôi nhà trong tiếng mưa",
  );
  await expect(
    page.getByText("Sáng tác minh họa", { exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("progressbar")).toHaveAttribute(
    "aria-valuenow",
    "0",
  );
  const body = page.locator("#essay-body");
  await expect(body.locator("p").first()).toBeAttached();
  await page
    .getByRole("navigation", { name: "Đọc tiếp" })
    .scrollIntoViewIfNeeded();
  await expect(page.getByRole("progressbar")).toHaveAttribute(
    "aria-valuenow",
    "100",
  );
  await page.getByRole("link", { name: "Chiếc tách mẹ để lại" }).click();
  await expect(page).toHaveURL("/articles/chiec-tach-me-de-lai");
  expect(errors).toEqual([]);
});

test("unknown paths have an honest 404 and a way back", async ({ page }) => {
  const response = await page.goto("/articles/khong-co-trong-bo-suu-tap");
  expect(response?.status()).toBe(404);
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Một ngăn ký ức",
  );
  await page.getByRole("link", { name: /Về bộ sưu tập/ }).click();
  await expect(page).toHaveURL("/articles");
  const unknownRoute = await page.goto("/mot-phong-khong-ton-tai");
  expect(unknownRoute?.status()).toBe(404);
});

test("reduced motion keeps all reading content visible and viewport readable", async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("/articles/ngoi-nha-trong-tieng-mua");
  await expect(page.locator("#essay-body p").first()).toBeVisible();
  const geometry = await page.evaluate(() => {
    const body = document.querySelector("#essay-body")!;
    const computed = getComputedStyle(body);
    return {
      overflow: document.documentElement.scrollWidth > innerWidth,
      font: computed.fontSize,
      lineHeight: computed.lineHeight,
      hidden: computed.opacity === "0",
      reduced: matchMedia("(prefers-reduced-motion: reduce)").matches,
    };
  });
  expect(geometry).toEqual({
    overflow: false,
    font: "18px",
    lineHeight: "32.4px",
    hidden: false,
    reduced: true,
  });
  for (const path of ["/", "/articles"]) {
    await page.goto(path);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
  }
});

test("keyboard users can skip repeated navigation", async ({ page }) => {
  await page.goto("/");
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("link", { name: "Đến nội dung chính" }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page.locator("main")).toBeFocused();
});

test.describe("without JavaScript", () => {
  test.use({ javaScriptEnabled: false });

  test("the collection is useful with JavaScript disabled", async ({
    page,
  }, testInfo) => {
    if (testInfo.project.name === "mobile") {
      expect(page.viewportSize()?.width).toBe(390);
    }
    await page.goto("/");
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await expect(page.locator(".hero-art img")).toBeVisible();
    await page.getByRole("link", { name: /Bước vào bộ sưu tập/ }).click();
    await page.getByRole("link", { name: /Ngôi nhà trong tiếng mưa/ }).click();
    await expect(page.getByRole("heading", { level: 1 })).toHaveText(
      "Ngôi nhà trong tiếng mưa",
    );
    await expect(page.locator("#essay-body p").first()).toBeVisible();
    await expect(page.getByRole("progressbar")).toHaveCount(0);
    await page.getByRole("link", { name: "Chiếc tách mẹ để lại" }).click();
    await expect(page).toHaveURL("/articles/chiec-tach-me-de-lai");
  });
});
