import React from 'react';
import { ExternalLink, AlertCircle } from 'lucide-react';

export type CardStatus = 'loading' | 'ready' | 'error';

interface ContentData {
  title: string;
  author: string;
  summary: string[];
  thumbnail: string;
  url: string;
}

interface ContentCardProps {
  status: CardStatus;
  data?: ContentData;
  errorMessage?: string;
}

export const ContentCard: React.FC<ContentCardProps> = ({ status, data, errorMessage }) => {
  if (status === 'loading') {
    return (
      <div className="w-full border border-gray-200 rounded-xl p-4 mb-4 animate-pulse bg-white">
        <div className="flex gap-4">
          <div className="w-32 h-20 bg-gray-200 rounded-lg"></div>
          <div className="flex-1 space-y-3">
            <div className="h-4 bg-gray-200 rounded w-3/4"></div>
            <div className="h-3 bg-gray-200 rounded w-1/2"></div>
          </div>
        </div>
        <div className="mt-4 space-y-2">
          <div className="h-3 bg-gray-100 rounded w-full"></div>
          <div className="h-3 bg-gray-100 rounded w-full"></div>
          <div className="h-3 bg-gray-100 rounded w-2/3"></div>
        </div>
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="w-full border border-red-100 bg-red-50 rounded-xl p-4 mb-4 flex items-center gap-3 text-red-600">
        <AlertCircle size={20} />
        <p className="text-sm font-medium">{errorMessage || 'Failed to parse content. Please try again.'}</p>
      </div>
    );
  }

  if (!data) return null;

  return (
    <div className="w-full border border-gray-200 bg-white rounded-xl overflow-hidden mb-4 hover:shadow-md transition-shadow">
      <div className="p-4">
        <div className="flex gap-4">
          <a 
            href={data.url} 
            target="_blank" 
            rel="noopener noreferrer" 
            className="shrink-0 w-32 h-20 bg-gray-100 rounded-lg overflow-hidden relative group"
          >
            <img 
              src={data.thumbnail} 
              alt={data.title} 
              className="w-full h-full object-cover group-hover:scale-105 transition-transform"
              onError={(e) => { (e.target as HTMLImageElement).src = 'https://placehold.co/400x225?text=No+Image'; }}
            />
            <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors flex items-center justify-center">
              <ExternalLink className="text-white opacity-0 group-hover:opacity-100" size={18} />
            </div>
          </a>
          <div className="flex-1 min-w-0">
            <a 
              href={data.url} 
              target="_blank" 
              rel="noopener noreferrer" 
              className="text-lg font-bold text-gray-900 leading-tight hover:text-blue-600 transition-colors line-clamp-2 block"
            >
              {data.title}
            </a>
            <p className="text-sm text-gray-500 mt-1">{data.author}</p>
          </div>
        </div>
        
        <div className="mt-4">
          <ul className="space-y-2">
            {data.summary.slice(0, 5).map((item, idx) => (
              <li key={idx} className="flex gap-2 text-sm text-gray-700">
                <span className="text-blue-500 shrink-0">•</span>
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
};