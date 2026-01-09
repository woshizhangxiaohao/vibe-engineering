"use client";

import React from 'react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Loader2, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";

interface SearchInputGroupProps {
  value: string;
  onChange: (val: string) => void;
  onSearch: () => void;
  loading: boolean;
  placeholder?: string;
  error?: boolean;
}

export default function SearchInputGroup({ value, onChange, onSearch, loading, placeholder, error }: SearchInputGroupProps) {
  return (
    <div className="relative flex items-center w-full group">
      <Search className={cn(
        "absolute left-5 h-5 w-5 transition-colors pointer-events-none z-10",
        error ? "text-destructive" : "text-muted-foreground group-focus-within:text-primary"
      )} />
      <Input
        type="text"
        placeholder={placeholder || "Paste YouTube link or Video ID..."}
        className={cn(
          "pl-14 pr-36 md:pr-44 h-14 md:h-16 text-base md:text-lg rounded-xl border-0 transition-all duration-200",
          error ? "bg-destructive/5 focus:bg-destructive/10" : "bg-muted focus:bg-background",
          loading && "opacity-70"
        )}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && onSearch()}
        readOnly={loading}
      />
      <Button
        onClick={onSearch}
        disabled={loading || !value.trim()}
        className="absolute right-2 rounded-lg h-10 md:h-12 px-5 md:px-8 text-sm md:text-base font-medium border-0 bg-primary text-primary-foreground hover:bg-primary/90 active:scale-[0.98] transition-all duration-200"
      >
        {loading ? (
          <Loader2 className="h-5 w-5 animate-spin" />
        ) : (
          <>
            <span className="hidden sm:inline">Extract</span>
            <ArrowRight className="h-4 w-4 sm:ml-2" />
          </>
        )}
      </Button>
    </div>
  );
}