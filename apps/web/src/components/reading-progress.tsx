"use client";

import { useEffect, useState } from "react";
import { readingProgress } from "@/lib/progress";

export function ReadingProgress() {
  const [progress, setProgress] = useState<number | null>(null);

  useEffect(() => {
    const article = document.getElementById("essay-body");
    if (!article) return;
    const update = () => {
      const { top, height } = article.getBoundingClientRect();
      setProgress(
        readingProgress(
          window.scrollY,
          top + window.scrollY,
          height,
          window.innerHeight,
        ),
      );
    };
    update();
    window.addEventListener("scroll", update, { passive: true });
    window.addEventListener("resize", update);
    const observer =
      typeof ResizeObserver !== "undefined" ? new ResizeObserver(update) : null;
    observer?.observe(article);
    return () => {
      window.removeEventListener("scroll", update);
      window.removeEventListener("resize", update);
      observer?.disconnect();
    };
  }, []);

  if (progress === null) return null;
  return (
    <div
      className="reading-progress"
      role="progressbar"
      aria-label="Tiến độ đọc bài"
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={progress}
      aria-valuetext={`${progress}% bài viết đã đi qua`}
    >
      <span style={{ width: `${progress}%` }} />
    </div>
  );
}
