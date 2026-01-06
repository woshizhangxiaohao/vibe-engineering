"use client";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useEffect, useState } from "react";

interface VideoUrlInputProps {
  value: string;
  onChange: (value: string) => void;
  error?: string;
}

export function VideoUrlInput({ value, onChange, error }: VideoUrlInputProps) {
  const [localValue, setLocalValue] = useState(value);

  useEffect(() => {
    setLocalValue(value);
  }, [value]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    setLocalValue(newValue);
    onChange(newValue);
  };

  useEffect(() => {
    if (!localValue || typeof localValue !== 'string') {
      return;
    }
    
    const trimmedValue = localValue.trim();
    if (trimmedValue !== localValue) {
      setLocalValue(trimmedValue);
      onChange(trimmedValue);
    }
  }, [localValue, onChange]);

  return (
    <div className="space-y-2">
      <Label htmlFor="video-url">视频链接</Label>
      <Input
        id="video-url"
        type="url"
        placeholder="https://www.youtube.com/watch?v=..."
        value={localValue || ''}
        onChange={handleChange}
        className={error ? "border-destructive" : ""}
      />
      {error && (
        <p className="text-sm text-destructive">{error}</p>
      )}
    </div>
  );
}