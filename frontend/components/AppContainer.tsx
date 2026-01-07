"use client";

import React, { useState } from 'react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Loader2 } from "lucide-react";
import ContentCard from './ContentCard';
import { contentApi } from '@/lib/api/endpoints';
import { CardData } from '@/types';
import { toast } from '@/lib/utils/toast';

export default function AppContainer() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [cards, setCards] = useState<CardData[]>([]);

  const handleParse = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url.trim()) return;

    const tempId = Math.random().toString(36).substring(7);
    setLoading(true);

    try {
      const response = await contentApi.parseUrl(url);
      
      const newCard: CardData = {
        id: response.id,
        url: response.url,
        status: response.status === 'SUCCESS' ? 'DISPLAY_READY' : 'PARSING_FAILED',
        timestamp: new Date().toISOString(),
        title: response.title,
        author: response.author,
        summary: response.summary,
      };

      setCards(prev => [newCard, ...prev]);
      setUrl('');
      toast.success("Content parsed successfully");
    } catch (error: any) {
      console.error(error);
      const errorCard: CardData = {
        id: tempId,
        url: url,
        status: 'PARSING_FAILED',
        timestamp: new Date().toISOString(),
      };
      setCards(prev => [errorCard, ...prev]);
      toast.error(error.message || "Failed to parse URL");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-12">
      <div className="text-center mb-12 space-y-4">
        <h1 className="text-4xl font-bold tracking-tight text-slate-900">
          Vibe Summarizer
        </h1>
        <p className="text-lg text-slate-600">
          Paste a YouTube or Twitter link to get an instant AI summary.
        </p>
      </div>

      <form onSubmit={handleParse} className="relative mb-12">
        <div className="relative flex items-center">
          <Search className="absolute left-4 text-slate-400 h-5 w-5" />
          <Input
            type="url"
            placeholder="https://www.youtube.com/watch?v=..."
            className="pl-12 pr-32 h-14 text-lg rounded-full border-slate-200 shadow-sm focus:ring-2 focus:ring-primary"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            disabled={loading}
          />
          <Button 
            type="submit" 
            className="absolute right-2 rounded-full h-10 px-6"
            disabled={loading || !url}
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Summarize'}
          </Button>
        </div>
      </form>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {cards.map((card) => (
          <ContentCard key={card.id} data={card} />
        ))}
        {loading && <ContentCard data={{} as any} loading={true} />}
      </div>

      {cards.length === 0 && !loading && (
        <div className="text-center py-20 border-2 border-dashed border-slate-200 rounded-3xl">
          <p className="text-slate-400">No summaries yet. Start by pasting a link above!</p>
        </div>
      )}
    </div>
  );
}