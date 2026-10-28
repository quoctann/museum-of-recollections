import { type Block } from "@/lib/collection";

export function ArticleBody({ blocks }: { blocks: readonly Block[] }) {
  return (
    <div id="essay-body" className="essay-body">
      {blocks.map((block, index) => {
        switch (block.type) {
          case "paragraph":
            return <p key={index}>{block.text}</p>;
          case "heading":
            return <h2 key={index}>{block.text}</h2>;
          case "quote":
            return (
              <blockquote key={index}>
                <p>{block.text}</p>
              </blockquote>
            );
          case "divider":
            return <hr key={index} />;
        }
      })}
    </div>
  );
}
