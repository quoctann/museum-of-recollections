"use client";

import { motion, useReducedMotion } from "motion/react";

export function Art({
  src,
  alt,
  priority = false,
}: {
  src: string;
  alt: string;
  priority?: boolean;
}) {
  const reduced = useReducedMotion();
  return (
    <motion.div
      className="art-frame"
      initial={false}
      whileHover={reduced ? undefined : { y: -4 }}
      transition={{ duration: 0.6, ease: "easeOut" }}
    >
      {/* Original local SVG; deliberately no remote image service. */}
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={src}
        alt={alt}
        width="640"
        height="720"
        loading={priority ? "eager" : "lazy"}
        fetchPriority={priority ? "high" : "auto"}
      />
    </motion.div>
  );
}
