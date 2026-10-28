export type Block =
  | { readonly type: "paragraph" | "heading" | "quote"; readonly text: string }
  | { readonly type: "divider" };

export type Article = {
  readonly slug: string;
  readonly number: string;
  readonly title: string;
  readonly subtitle: string;
  readonly category: string;
  readonly date: string;
  readonly art: string;
  readonly artAlt: string;
  readonly language: "vi";
  readonly demo: true;
  readonly author: string;
  readonly blocks: readonly Block[];
};

export const articles: readonly Article[] = [
  {
    slug: "ngoi-nha-trong-tieng-mua",
    number: "01",
    title: "Ngôi nhà trong tiếng mưa",
    subtitle: "Có những mái nhà ta đã rời đi, nhưng tiếng mưa vẫn còn ở lại.",
    category: "Nơi chốn",
    date: "2026-06-12",
    art: "/art/house.svg",
    artAlt: "Hình khối một mái nhà màu đất dưới những nét mưa mảnh",
    language: "vi",
    demo: true,
    author: "Ban biên soạn",
    blocks: [
      {
        type: "paragraph",
        text: "Chiều nay, một cơn mưa ngang qua thành phố. Tôi đang đứng dưới mái hiên của một hiệu sách thì nghe thấy tiếng nước rơi xuống tấm tôn bên cạnh. Chỉ một âm thanh ấy thôi, con đường trước mặt bỗng lùi xa. Tôi lại ngồi trên bậc cửa của ngôi nhà cũ, hai bàn chân không chạm đất.",
      },
      {
        type: "paragraph",
        text: "Ngôi nhà chẳng có gì đặc biệt. Một khoảng sân nhỏ, một cây khế không năm nào sai quả, một cánh cửa phải nhấc lên một chút mới khép được. Mỗi lần trời chuyển mưa, mẹ lại gọi tôi mang áo quần vào. Tiếng gọi đi qua gian bếp, qua chiếc quạt đang quay chậm, rồi mới đến được chỗ tôi.",
      },
      { type: "heading", text: "Bản đồ của những âm thanh" },
      {
        type: "paragraph",
        text: "Tôi không nhớ chính xác màu sơn của căn nhà. Trong trí nhớ, nó lúc vàng nhạt, lúc gần như trắng. Nhưng tôi nhớ tiếng thìa chạm vào miệng cốc của cha mỗi sáng. Nhớ tiếng dép kéo trên sân. Nhớ cả khoảng im lặng ngắn ngủi sau khi mất điện, trước khi ai đó tìm được cây nến trong ngăn kéo.",
      },
      {
        type: "quote",
        text: "Có lẽ một ngôi nhà không nằm trong địa chỉ, mà nằm trong cách ta nhận ra một tiếng động.",
      },
      {
        type: "paragraph",
        text: "Ngày chuyển đi, chúng tôi mang theo bàn ghế, sách vở và mấy chậu cây. Chẳng ai nghĩ phải mang theo tiếng mưa. Những điều không có hình dạng thường bị bỏ quên trong một cuộc dọn nhà. Vậy mà nhiều năm sau, chúng lại tự tìm đến, nguyên vẹn hơn bất cứ đồ vật nào.",
      },
      { type: "divider" },
      {
        type: "paragraph",
        text: "Cơn mưa trước hiệu sách đã ngớt. Tôi bước ra đường, chậm hơn lúc đến một chút. Ngôi nhà cũ có lẽ đã được sửa sang, cây khế có lẽ cũng không còn. Nhưng trong một khoảnh khắc rất ngắn, tôi đã được trở về. Không cần mở cửa. Không cần ai đợi.",
      },
      {
        type: "paragraph",
        text: "Tối nay, khi về đến căn phòng của mình, tôi sẽ mở hé cửa sổ. Biết đâu, giữa tiếng xe và những âm thanh xa lạ, lại có một giọt mưa biết đường về nhà.",
      },
    ],
  },
  {
    slug: "chiec-tach-me-de-lai",
    number: "02",
    title: "Chiếc tách mẹ để lại",
    subtitle:
      "Một vết sứt rất nhỏ, đủ để giữ nguyên hình dáng của một buổi sáng.",
    category: "Đồ vật",
    date: "2026-06-08",
    art: "/art/cup.svg",
    artAlt:
      "Một chiếc tách trừu tượng màu mực, quai tròn và chiếc bóng màu đất",
    language: "vi",
    demo: true,
    author: "Ban biên soạn",
    blocks: [
      {
        type: "paragraph",
        text: "Trong tủ bếp nhà tôi có một chiếc tách không cùng bộ với bất cứ thứ gì. Nó màu trắng ngà, quai hơi nhỏ, miệng sứt một chấm bằng hạt gạo. Mẹ để lại sau lần lên thăm gần nhất, cùng một túi chanh và lời dặn đừng bỏ bữa sáng.",
      },
      {
        type: "paragraph",
        text: "Ban đầu tôi định cất nó vào sâu trong tủ. Một người sống một mình không cần nhiều tách đến thế. Nhưng sáng hôm sau, tay tôi lại với đúng chiếc tách ấy. Có lẽ vì nó ở gần. Cũng có lẽ vì một lý do mà lúc đó tôi chưa biết cách gọi tên.",
      },
      { type: "heading", text: "Một chỗ ngồi cho điều bình thường" },
      {
        type: "paragraph",
        text: "Ở nhà, mẹ không bao giờ uống hết trà khi còn nóng. Mẹ pha một tách, đặt xuống bàn, rồi đi tưới cây, mở cửa, nhặt một chiếc lá rơi. Đến khi quay lại, trà chỉ còn âm ấm. Tôi từng nghĩ mẹ hay quên. Bây giờ tôi nghĩ có những người dành sự chú ý của mình cho mọi thứ xung quanh, chỉ giữ lại cho bản thân một chút ấm vừa đủ.",
      },
      {
        type: "quote",
        text: "Đồ vật không nhớ thay chúng ta. Chúng chỉ kiên nhẫn ở đó, chờ một lần ta chịu nhìn kỹ.",
      },
      {
        type: "paragraph",
        text: "Chiếc tách đã ở đây được ba mùa mưa. Vết sứt không lớn thêm. Tôi học cách xoay nó về phía bên kia khi uống, như học cách chung sống với một điều không hoàn hảo mà không nhất thiết phải sửa chữa. Không phải thứ gì cũ đi cũng cần được thay mới.",
      },
      { type: "divider" },
      {
        type: "paragraph",
        text: "Sáng chủ nhật, tôi gọi điện cho mẹ. Chúng tôi nói về giá rau, một người hàng xóm, chậu hoa vừa ra nụ. Không có chuyện gì quan trọng. Trong lúc nghe, tôi dùng ngón tay lần theo quai tách. Nắng đi qua cửa sổ, đặt một vệt sáng nhỏ lên mặt bàn.",
      },
      {
        type: "paragraph",
        text: "Tôi không kể với mẹ về chiếc tách. Có những lời cảm ơn không thành tiếng. Chúng được rửa sạch mỗi tối, úp lên giá, rồi sáng mai lại được nâng lên bằng cả hai tay.",
      },
    ],
  },
  {
    slug: "khoang-lang-giua-hai-chuyen-tau",
    number: "03",
    title: "Khoảng lặng giữa hai chuyến tàu",
    subtitle:
      "Đôi khi, phần đáng nhớ nhất của một hành trình là lúc ta chưa đi đâu cả.",
    category: "Khoảnh khắc",
    date: "2026-06-02",
    art: "/art/window.svg",
    artAlt: "Một ô cửa vòm màu đất nhìn ra đường chân trời và mặt trời nhỏ",
    language: "vi",
    demo: true,
    author: "Ban biên soạn",
    blocks: [
      {
        type: "paragraph",
        text: "Chuyến tàu bị chậm bốn mươi phút. Người bán nước kéo chiếc ghế nhựa vào chỗ râm, còn tôi ngồi xuống băng ghế cuối sân ga. Chiếc túi bên cạnh nhẹ tênh. Lần đầu sau nhiều tuần, tôi không thể làm gì để đến nơi sớm hơn.",
      },
      {
        type: "paragraph",
        text: "Thành phố nhỏ này chỉ là một điểm đổi tàu. Tôi không biết tên con phố bên ngoài nhà ga, không biết quán nào bán bữa sáng, không quen ai trong số những người đang chờ. Sự xa lạ ấy, thay vì làm tôi lo lắng, lại khiến mọi thứ trở nên dễ chịu.",
      },
      { type: "heading", text: "Thời gian không cần được lấp đầy" },
      {
        type: "paragraph",
        text: "Ở băng ghế đối diện, một cậu bé đếm những viên gạch lát sân. Cứ đến một chỗ nứt, cậu lại dừng, cúi xuống nhìn, rồi bắt đầu đếm lại. Người bà ngồi bên không giục. Bà chỉ giữ chiếc mũ trong lòng, thỉnh thoảng nhìn lên bảng giờ tàu.",
      },
      {
        type: "quote",
        text: "Một khoảng chờ cũng có thể là một nơi chốn, nếu ta thôi tìm cách bước ra khỏi nó.",
      },
      {
        type: "paragraph",
        text: "Tôi cất điện thoại. Trên đường ray, một vệt cỏ mọc giữa hai thanh sắt. Có tiếng chim rất gần mà tôi không tìm thấy. Ly nước trên tay đã hết lạnh. Những điều nhỏ ấy không làm hành trình tiến thêm, nhưng làm buổi sáng trở nên có thật.",
      },
      {
        type: "paragraph",
        text: "Tôi nghĩ đến những lần mình đã chờ: một lá thư, một cuộc gọi, một người bước ra khỏi cửa. Khi nhớ lại, tôi thường chỉ nhớ điều xảy đến sau đó. Phần thời gian ở giữa bị gấp vào, như tờ giấy lót trong một cuốn sách. Hôm nay, tôi muốn giữ nó phẳng ra.",
      },
      { type: "divider" },
      {
        type: "paragraph",
        text: "Khi tàu đến, cậu bé vẫn chưa đếm hết sân ga. Người bà đội mũ cho cậu, phủi nhẹ vai áo. Tôi xách túi lên. Bốn mươi phút đã trôi qua mà không mang theo một thành tựu nào. Chỉ có một khoảng lặng được đặt ngay ngắn vào trí nhớ, đủ rộng để lâu lâu tôi quay lại ngồi.",
      },
    ],
  },
];

export function findArticle(slug: string): Article | undefined {
  return articles.find((article) => article.slug === slug);
}

export function nextArticle(slug: string): Article | undefined {
  const index = articles.findIndex((article) => article.slug === slug);
  return index < 0 ? undefined : articles[(index + 1) % articles.length];
}

export function readingMinutes(blocks: readonly Block[]): number {
  const words = blocks.flatMap((block) =>
    block.type === "divider"
      ? []
      : block.text.trim().split(/\s+/).filter(Boolean),
  ).length;
  return Math.max(1, Math.ceil(words / 250));
}
