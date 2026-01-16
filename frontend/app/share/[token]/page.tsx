"use client";

import { useState, useEffect } from "react";
import { notFound } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { insightApi } from "@/lib/api/endpoints";
import type { SharedInsightResponse } from "@/lib/api/types";
import {
  Eye,
  Calendar,
  User,
  ExternalLink,
  Lock,
  ArrowRight,
  Youtube,
  Twitter,
  Podcast,
} from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

interface SharePageProps {
  params: { token: string };
}

const sourceIcons = {
  youtube: Youtube,
  twitter: Twitter,
  podcast: Podcast,
};

export default function SharePage({ params }: SharePageProps) {
  const [insight, setInsight] = useState<SharedInsightResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [requiresPassword, setRequiresPassword] = useState(false);
  const [password, setPassword] = useState("");
  const [passwordLoading, setPasswordLoading] = useState(false);

  const loadSharedInsight = async (token: string, pwd?: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await insightApi.getSharedInsight(token, pwd);
      setInsight(response);
      setRequiresPassword(false);
    } catch (err: any) {
      console.error("Failed to load shared insight:", err);

      if (err.status === 401 && err.data?.requires_auth) {
        setRequiresPassword(true);
        setError("请输入访问密码");
      } else if (err.status === 404) {
        setError("分享链接不存在或已过期");
      } else {
        setError(err.message || "加载分享内容失败");
      }
    } finally {
      setLoading(false);
      setPasswordLoading(false);
    }
  };

  useEffect(() => {
    if (params.token) {
      loadSharedInsight(params.token);
    }
  }, [params.token]);

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!password.trim()) {
      toast.error("请输入密码");
      return;
    }

    setPasswordLoading(true);
    await loadSharedInsight(params.token, password);
  };

  if (loading) {
    return (
      <div className="container max-w-4xl mx-auto py-12 px-4">
        <div className="space-y-6">
          <div className="h-8 bg-gray-200 rounded animate-pulse" />
          <div className="space-y-4">
            <div className="h-4 bg-gray-200 rounded animate-pulse w-3/4" />
            <div className="h-4 bg-gray-200 rounded animate-pulse w-1/2" />
          </div>
          <div className="h-64 bg-gray-200 rounded animate-pulse" />
        </div>
      </div>
    );
  }

  if (error && !requiresPassword) {
    return (
      <div className="container max-w-4xl mx-auto py-12 px-4">
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="text-center space-y-4">
              <div className="text-destructive text-lg font-medium">{error}</div>
              <p className="text-muted-foreground">
                请检查分享链接是否正确，或联系分享者确认链接有效性。
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (requiresPassword) {
    return (
      <div className="container max-w-md mx-auto py-12 px-4">
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mb-4">
              <Lock className="w-6 h-6 text-blue-600" />
            </div>
            <CardTitle>受密码保护的分享</CardTitle>
            <p className="text-muted-foreground">
              此分享需要密码才能访问，请输入正确的访问密码。
            </p>
          </CardHeader>
          <CardContent>
            <form onSubmit={handlePasswordSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="password">访问密码</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="请输入访问密码"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="h-11"
                />
                {error && (
                  <p className="text-sm text-destructive">{error}</p>
                )}
              </div>
              <Button
                type="submit"
                className="w-full h-11"
                disabled={passwordLoading}
              >
                {passwordLoading ? "验证中..." : "访问分享"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!insight) {
    notFound();
  }

  const SourceIcon = sourceIcons[insight.source_type] || ExternalLink;

  return (
    <div className="container max-w-4xl mx-auto py-8 px-4">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
          <User className="w-4 h-4" />
          <span>由 {insight.shared_by} 分享</span>
          {insight.shared_at && (
            <>
              <span>•</span>
              <Calendar className="w-4 h-4" />
              <span>
                {new Date(insight.shared_at).toLocaleDateString("zh-CN")}
              </span>
            </>
          )}
        </div>
        <h1 className="text-2xl md:text-3xl font-bold leading-tight mb-4">
          {insight.title}
        </h1>
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <SourceIcon className="w-4 h-4" />
            <span className="capitalize">{insight.source_type}</span>
          </div>
          <span>作者：{insight.author}</span>
          <Button
            variant="outline"
            size="sm"
            asChild
            className="ml-auto"
          >
            <a
              href={insight.source_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2"
            >
              查看原内容
              <ExternalLink className="w-3 h-3" />
            </a>
          </Button>
        </div>
      </div>

      {/* Thumbnail */}
      {insight.thumbnail_url && (
        <div className="mb-8">
          <div className="relative rounded-lg overflow-hidden bg-gray-100">
            <img
              src={insight.thumbnail_url}
              alt={insight.title}
              className="w-full h-64 object-cover"
            />
            <div className="absolute inset-0 bg-black/20" />
          </div>
        </div>
      )}

      <div className="grid gap-8">
        {/* Summary */}
        {insight.content.summary && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <div className="w-6 h-6 bg-blue-500 rounded flex items-center justify-center">
                  <Eye className="w-4 h-4 text-white" />
                </div>
                AI 摘要
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">
                {insight.content.summary}
              </p>
            </CardContent>
          </Card>
        )}

        {/* Key Points */}
        {insight.content.key_points && insight.content.key_points.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <div className="w-6 h-6 bg-green-500 rounded flex items-center justify-center">
                  <span className="text-white text-xs font-bold">💡</span>
                </div>
                关键要点
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3">
                {insight.content.key_points.map((point, index) => (
                  <li key={index} className="flex items-start gap-3">
                    <span className="flex-shrink-0 w-5 h-5 bg-green-100 text-green-600 rounded-full flex items-center justify-center text-xs font-medium mt-0.5">
                      {index + 1}
                    </span>
                    <span className="text-gray-700">{point}</span>
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        )}

        {/* Highlights */}
        {insight.content.highlights && insight.content.highlights.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <div className="w-6 h-6 bg-yellow-500 rounded flex items-center justify-center">
                  <span className="text-white text-xs">🖍</span>
                </div>
                笔记标注
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {insight.content.highlights.map((highlight) => (
                  <div
                    key={highlight.id}
                    className={cn(
                      "p-4 rounded-lg border-l-4",
                      highlight.color === "yellow" && "bg-yellow-50 border-yellow-400",
                      highlight.color === "green" && "bg-green-50 border-green-400",
                      highlight.color === "blue" && "bg-blue-50 border-blue-400",
                      highlight.color === "purple" && "bg-purple-50 border-purple-400",
                      highlight.color === "red" && "bg-red-50 border-red-400"
                    )}
                  >
                    <blockquote className="text-gray-800 font-medium mb-2">
                      "{highlight.text}"
                    </blockquote>
                    {highlight.note && (
                      <p className="text-sm text-gray-600 italic">
                        📝 {highlight.note}
                      </p>
                    )}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Chat Messages */}
        {insight.content.chat && insight.content.chat.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <div className="w-6 h-6 bg-purple-500 rounded flex items-center justify-center">
                  <span className="text-white text-xs">💬</span>
                </div>
                AI 对话记录
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {insight.content.chat.map((message, index) => (
                  <div
                    key={message.id || index}
                    className={cn(
                      "flex gap-3",
                      message.role === "user" ? "justify-end" : "justify-start"
                    )}
                  >
                    <div
                      className={cn(
                        "max-w-[80%] p-4 rounded-lg",
                        message.role === "user"
                          ? "bg-blue-500 text-white"
                          : "bg-gray-100 text-gray-800"
                      )}
                    >
                      <p className="whitespace-pre-wrap">{message.content}</p>
                      <p
                        className={cn(
                          "text-xs mt-2",
                          message.role === "user"
                            ? "text-blue-100"
                            : "text-gray-500"
                        )}
                      >
                        {new Date(message.created_at).toLocaleString("zh-CN")}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* CTA */}
        <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-blue-200">
          <CardContent className="pt-6">
            <div className="text-center space-y-4">
              <div className="text-lg font-medium text-gray-900">
                🚀 想要自己分析内容？试试 InsightFlow！
              </div>
              <p className="text-gray-600">
                轻松解析视频、推文和播客，获得 AI 摘要、关键要点和智能问答。
              </p>
              <Button size="lg" className="rounded-lg">
                立即使用 InsightFlow
                <ArrowRight className="w-4 h-4 ml-2" />
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}