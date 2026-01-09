"use client";

import React from 'react';
import { LayoutDashboard, Youtube, ListMusic, Subtitles, ShieldCheck } from "lucide-react";
import { cn } from "@/lib/utils";
import QuotaMonitor from "./QuotaMonitor";

interface SidebarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

export default function Sidebar({ activeTab, onTabChange }: SidebarProps) {
  const menuItems = [
    { id: 'video', label: 'Video Info', icon: Youtube },
    { id: 'playlist', label: 'Playlist', icon: ListMusic },
    { id: 'captions', label: 'Captions', icon: Subtitles },
    { id: 'auth', label: 'Authorization', icon: ShieldCheck },
  ];

  return (
    <aside className="w-64 hidden md:flex flex-col h-screen sticky top-0 bg-background border-r border-border/50 p-6">
      <div className="flex items-center gap-3 mb-12">
        <div className="h-8 w-8 rounded-lg bg-primary flex items-center justify-center">
          <span className="text-primary-foreground font-black text-sm">V</span>
        </div>
        <span className="font-bold text-lg tracking-tight">VIBE DATA</span>
      </div>

      <nav className="flex-1 space-y-2">
        {menuItems.map((item) => (
          <button
            key={item.id}
            onClick={() => onTabChange(item.id)}
            className={cn(
              "w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-200",
              activeTab === item.id
                ? "bg-secondary text-primary"
                : "text-muted-foreground hover:bg-secondary/50 hover:text-foreground"
            )}
          >
            <item.icon className="h-5 w-5" />
            {item.label}
          </button>
        ))}
      </nav>

      <div className="mt-auto pt-6 border-t border-border/50">
        <QuotaMonitor />
      </div>
    </aside>
  );
}