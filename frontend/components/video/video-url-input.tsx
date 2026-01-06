"use client";

import { useState, useCallback, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { validateYouTubeUrl } from "@/lib/utils/video";
import { useDebounce } from "@/hooks/use-debounce";

interface VideoUrlInputProps {
  onUrlValidated: (url: string) => void;
  disabled?: boolean;
}

export function VideoUrlInput({ onUrlValidated, disabled }: VideoUrlInputProps) {
  const [url, setUrl] = useState("");
  const [error, setError] = useState<string | null>(null);
  const debouncedUrl = useDebounce(url, 500);

  useEffect(() => {
    if (!debouncedUrl.trim()) {
      setError(null);
      return;
    }

    const validation = validateYouTubeUrl(debouncedUrl);
    
    if (!validation.isValid) {
      setError("请输入有效的 YouTube 链接");
      return;
    }

    setError(null);
    onUrlValidated(validation.url!);
  }, [debouncedUrl, onUrlValidated]);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setUrl(e.target.value);
  }, []);

  const handlePaste = useCallback((e: React.ClipboardEvent<HTMLInputElement>) => {
    const pastedUrl = e.clipboardData.getData('text');
    setUrl(pastedUrl);
  }, []);

  return (
    <div className="space-y-2">
      <Input
        type="text"
        placeholder="粘贴 YouTube 链接 (支持 youtube.com 和 youtu.be)"
        value={url}
        onChange={handleChange}
        onPaste={handlePaste}
        disabled={disabled}
        className="w-full"
      />
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
    </div>
  );
}