import { VideoNoteForm } from "@/components/video/video-note-form";

export default function VideoNotePage() {
  return (
    <div className="container mx-auto max-w-3xl py-8">
      <div className="mb-8 space-y-2">
        <h1 className="text-3xl font-bold text-foreground">视频笔记</h1>
        <p className="text-muted-foreground">
          粘贴 YouTube 链接，自动获取视频信息并保存笔记
        </p>
      </div>

      <VideoNoteForm />
    </div>
  );
}