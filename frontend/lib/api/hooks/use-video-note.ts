"use client";

import { useState, useCallback } from "react";
import { videoService } from "../services/video.service";
import { VideoNote } from "@/types/video";
import { toast } from "sonner";

interface UseVideoNoteReturn {
  saving: boolean;
  saveNote: (note: Omit<VideoNote, 'id' | 'createdAt' | 'updatedAt'>) => Promise<void>;
}

export function useVideoNote(): UseVideoNoteReturn {
  const [saving, setSaving] = useState(false);

  const saveNote = useCallback(async (note: Omit<VideoNote, 'id' | 'createdAt' | 'updatedAt'>) => {
    try {
      setSaving(true);
      await videoService.saveVideoNote(note);
      toast.success("视频笔记保存成功");
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "保存失败";
      toast.error(errorMessage);
      throw err;
    } finally {
      setSaving(false);
    }
  }, []);

  return {
    saving,
    saveNote,
  };
}