import { Metadata } from "next";
import { insightApi } from "@/lib/api/endpoints";

interface ShareLayoutProps {
  children: React.ReactNode;
  params: { token: string };
}

export async function generateMetadata({
  params,
}: {
  params: { token: string };
}): Promise<Metadata> {
  try {
    // 尝试获取分享内容用于SEO
    const insight = await insightApi.getSharedInsight(params.token);

    const title = `${insight.title} - InsightFlow 分享`;
    const description = insight.content.summary
      ? insight.content.summary.slice(0, 160)
      : `查看 ${insight.author} 的内容分析和笔记标注`;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        type: "article",
        images: insight.thumbnail_url
          ? [
              {
                url: insight.thumbnail_url,
                width: 1200,
                height: 630,
                alt: insight.title,
              },
            ]
          : [],
        siteName: "InsightFlow",
      },
      twitter: {
        card: "summary_large_image",
        title,
        description,
        images: insight.thumbnail_url ? [insight.thumbnail_url] : [],
      },
    };
  } catch (error) {
    // 如果获取失败，返回默认元数据
    return {
      title: "分享内容 - InsightFlow",
      description: "查看朋友分享的内容分析和笔记标注",
      openGraph: {
        title: "分享内容 - InsightFlow",
        description: "查看朋友分享的内容分析和笔记标注",
        type: "website",
        siteName: "InsightFlow",
      },
      twitter: {
        card: "summary",
        title: "分享内容 - InsightFlow",
        description: "查看朋友分享的内容分析和笔记标注",
      },
    };
  }
}

export default function ShareLayout({ children }: ShareLayoutProps) {
  return (
    <div className="min-h-screen bg-background">
      {children}
    </div>
  );
}