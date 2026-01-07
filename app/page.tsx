"use client";

import { useState } from 'react';
import { ContentCard, CardStatus } from '@/components/ContentCard';

interface FeedItem {
  id: string;
  url: string;
  status: CardStatus;
  data?: any;
  error?: string;
}

export default function Home() {
  const [url, setUrl] = useState('');
  const [feed, setFeed] = useState<FeedItem[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validateUrl = (input: string) => {
    const youtubeRegex = /^(https?:\/\/)?(www\.)?(youtube\.com|youtu\.be)\/.*$/;
    const twitterRegex = /^(https?:\/\/)?(www\.)?(twitter\.com|x\.com)\/.*$/;
    return youtubeRegex.test(input) || twitterRegex.test(input);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!url || isSubmitting) return;

    if (!validateUrl(url)) {
      alert("Please enter a valid YouTube or Twitter/X URL");
      return;
    }

    const newItemId = Math.random().toString(36).substring(7);
    const newItem: FeedItem = { id: newItemId, url, status: 'loading' };
    
    setFeed(prev => [newItem, ...prev]);
    setUrl('');
    setIsSubmitting(true);

    try {
      // Mock API call - Replace with actual backend endpoint
      // const response = await fetch('/api/parse', { method: 'POST', body: JSON.stringify({ url }) });
      // const result = await response.json();
      
      await new Promise(resolve => setTimeout(resolve, 1500)); // Simulate network

      const mockData = {
        title: url.includes('youtube') ? "Understanding the Future of AI" : "Thread on Modern Engineering",
        author: url.includes('youtube') ? "Tech Channel" : "@engineer_pro",
        thumbnail: url.includes('youtube') 
          ? "https://images.unsplash.com/photo-1611162617213-7d7a39e9b1d7?w=400&h=225&fit=crop" 
          : "https://images.unsplash.com/photo-1611605698335-8b1569810432?w=400&h=225&fit=crop",
        url: url,
        summary: [
          "Key insight regarding the scalability of current architectures.",
          "The importance of developer experience in high-growth teams.",
          "Strategic shift towards edge computing for lower latency.",
          "Analysis of cost-to-performance ratios in cloud providers."
        ]
      };

      setFeed(prev => prev.map(item => 
        item.id === newItemId ? { ...item, status: 'ready', data: mockData } : item
      ));
    } catch (err) {
      setFeed(prev => prev.map(item => 
        item.id === newItemId ? { ...item, status: 'error', error: 'Failed to fetch metadata' } : item
      ));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-10 bg-white border-b border-gray-200">
        <div className="max-w-2xl mx-auto px-4 h-16 flex items-center">
          <h1 className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-600 to-indigo-600">
            Vibe Digest
          </h1>
        </div>
      </header>

      {/* Input Zone */}
      <div className="max-w-2xl mx-auto px-4 pt-8">
        <form onSubmit={handleSubmit} className="relative">
          <input
            autoFocus
            type="text"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="Paste YouTube or Twitter link..."
            className="w-full h-14 px-5 pr-16 rounded-2xl border border-gray-200 bg-white shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all text-gray-900"
          />
          <button 
            type="submit"
            disabled={!url || isSubmitting}
            className="absolute right-2 top-2 h-10 px-4 bg-gray-900 text-white rounded-xl font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isSubmitting ? '...' : 'Parse'}
          </button>
        </form>
      </div>

      {/* Result Feed */}
      <div className="max-w-2xl mx-auto px-4 py-8">
        {feed.length === 0 && (
          <div className="text-center py-20 text-gray-400">
            <p>Enter a URL above to generate a summary</p>
          </div>
        )}
        <div className="flex flex-col">
          {feed.map((item) => (
            <ContentCard 
              key={item.id} 
              status={item.status} 
              data={item.data} 
              errorMessage={item.error} 
            />
          ))}
        </div>
      </div>
    </main>
  );
}