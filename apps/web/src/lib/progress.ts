export function readingProgress(
  scroll: number,
  top: number,
  height: number,
  viewport: number,
): number {
  if (
    ![scroll, top, height, viewport].every(Number.isFinite) ||
    height < 0 ||
    viewport <= 0
  )
    return 0;
  const range = height - viewport;
  if (range <= 0) return scroll >= top ? 100 : 0;
  return Math.round(Math.max(0, Math.min(100, ((scroll - top) / range) * 100)));
}
