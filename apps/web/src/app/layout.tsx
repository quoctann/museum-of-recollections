import type { Metadata } from "next";
import "@fontsource/noto-serif/latin-400.css";
import "@fontsource/noto-serif/vietnamese-400.css";
import "@fontsource/noto-serif/latin-400-italic.css";
import "@fontsource/noto-serif/vietnamese-400-italic.css";
import "@fontsource/be-vietnam-pro/latin-400.css";
import "@fontsource/be-vietnam-pro/vietnamese-400.css";
import { SiteHeader, SiteFooter } from "@/components/site-chrome";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "Bảo tàng hồi ức — Một nơi để nhớ",
    template: "%s — Bảo tàng hồi ức",
  },
  description:
    "Một không gian đọc chậm bằng tiếng Việt. Những tản văn minh họa về nơi chốn, đồ vật và ký ức của đời thường.",
  robots: { index: true, follow: true },
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="vi">
      <body id="top">
        <SiteHeader />
        {children}
        <SiteFooter />
      </body>
    </html>
  );
}
