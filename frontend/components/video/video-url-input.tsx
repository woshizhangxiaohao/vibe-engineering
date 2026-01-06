"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { parseYouTubeUrl } from "@/lib/utils/video";
import { useDebounce } from "@/hooks/use-debounce";

interface VideoUrlInputProps {
  value: string;
  onChange: (value: string) => void;
  onValidUrlDetected?: (videoId: string) => void;
}

export function VideoUrlInput({
  value,
  onChange,
  onValidUrlDetected,
}: VideoUrlInputProps) {
  const [validationError, setValidationError] = useState<string | null>(null);
  const debouncedUrl = useDebounce(value, 500);

  useEffect(() => {
    if (!debouncedUrl.trim()) {
      setValidationError(null);
      return;
    }

    const { isValid, videoId } = parseYouTubeUrl(debouncedUrl);

    if (!isValid) {
      setValidationError("请输入有效的 YouTube 链接");
      return;
    }

    setValidationError(null);
    if (videoId && onValidUrlDetected) {
      onValidUrlDetected(videoId);
    }
  }, [debouncedUrl, onValidUrlDetected]);

  return (
    <div className="space-y-2">
      <Label htmlFor="video-url">YouTube 视频链接</Label>
      <Input
        id="video-url"
        type="url"
        placeholder="粘贴 YouTube 链接（如：https://www.youtube.com/watch?v=xxxxx）"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="font-mono text-sm"
      />
      {validationError && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{validationError}</AlertDescription>
        </Alert>
      )}
    </div>
  );
}