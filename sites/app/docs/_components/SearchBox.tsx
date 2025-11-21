"use client";

// React & Next.js
import React, { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

// Third-party libraries
import {
  Search,
  X,
  FileText,
  ArrowRight,
  Command
} from 'lucide-react';

// Search utilities
import { searchContent, SearchResult } from './SearchData';
import { getSearchIndex } from '../../../utils/search-adapter';

interface SearchBoxProps {
  placeholder?: string;
  className?: string;
}

export const SearchBox: React.FC<SearchBoxProps> = ({
  placeholder = "Search docs...",
  className = ""
}) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [searchIndex, setSearchIndex] = useState<SearchResult[]>([]);
  const [isIndexLoaded, setIsIndexLoaded] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const resultsRef = useRef<HTMLDivElement>(null);
  const router = useRouter();

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Open search on Cmd/Ctrl + K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        inputRef.current?.focus();
        setIsOpen(true);
        return;
      }

      // Close on Escape
      if (e.key === 'Escape') {
        setIsOpen(false);
        inputRef.current?.blur();
        return;
      }

      // Navigate results if open
      if (isOpen && results.length > 0) {
        if (e.key === 'ArrowDown') {
          e.preventDefault();
          setSelectedIndex(prev => (prev + 1) % results.length);
        } else if (e.key === 'ArrowUp') {
          e.preventDefault();
          setSelectedIndex(prev => (prev - 1 + results.length) % results.length);
        } else if (e.key === 'Enter') {
          e.preventDefault();
          if (results[selectedIndex]) {
            router.push(`/docs/${results[selectedIndex].slug}`);
            setIsOpen(false);
            setQuery('');
          }
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, results, selectedIndex, router]);

  // Initialize search index
  useEffect(() => {
    async function initializeIndex() {
      try {
        const index = await getSearchIndex();
        setSearchIndex(index);
        setIsIndexLoaded(true);
      } catch (error) {
        console.error('Failed to initialize search index:', error);
        setIsIndexLoaded(true);
      }
    }

    if (!isIndexLoaded) {
      initializeIndex();
    }
  }, [isIndexLoaded]);

  // Dynamic search function
  const performDynamicSearch = (query: string): SearchResult[] => {
    if (!query.trim()) return [];

    const normalizedQuery = query.toLowerCase().trim();
    return searchIndex
      .filter(page => {
        const titleMatch = page.title.toLowerCase().includes(normalizedQuery);
        const contentMatch = page.content.toLowerCase().includes(normalizedQuery);
        const sectionMatch = page.section.toLowerCase().includes(normalizedQuery);
        const categoryMatch = page.category.toLowerCase().includes(normalizedQuery);

        return titleMatch || contentMatch || sectionMatch || categoryMatch;
      })
      .sort((a, b) => {
        // Prioritize title matches
        const aTitleMatch = a.title.toLowerCase().includes(normalizedQuery);
        const bTitleMatch = b.title.toLowerCase().includes(normalizedQuery);

        if (aTitleMatch && !bTitleMatch) return -1;
        if (!aTitleMatch && bTitleMatch) return 1;

        // Then prioritize exact title matches
        const aExactMatch = a.title.toLowerCase() === normalizedQuery;
        const bExactMatch = b.title.toLowerCase() === normalizedQuery;

        if (aExactMatch && !bExactMatch) return -1;
        if (!aExactMatch && bExactMatch) return 1;

        // Finally, alphabetize
        return a.title.localeCompare(b.title);
      });
  };

  // Update search results
  useEffect(() => {
    if (query.trim() && isIndexLoaded) {
      // Use dynamic search if index is available, otherwise fall back to static search
      const searchResults = searchIndex.length > 0
        ? performDynamicSearch(query)
        : searchContent(query);

      setResults(searchResults.slice(0, 8)); // Limit to 8 results
      setSelectedIndex(0);
    } else {
      setResults([]);
    }
  }, [query, searchIndex, isIndexLoaded]);

  // Close search when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        inputRef.current &&
        !inputRef.current.contains(e.target as Node) &&
        resultsRef.current &&
        !resultsRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const highlightMatch = (text: string, query: string) => {
    if (!query.trim()) return text;

    const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    const parts = text.split(regex);

    return parts.map((part, index) =>
      regex.test(part) ? (
        <mark key={index} className="bg-emerald-500/30 text-emerald-300 rounded px-0.5">
          {part}
        </mark>
      ) : part
    );
  };

  const getCategoryColor = (category: string) => {
    switch (category.toLowerCase()) {
      case 'getting started':
        return 'text-blue-400 bg-blue-500/10 border-blue-500/20';
      case 'core concepts':
        return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20';
      case 'reference':
        return 'text-purple-400 bg-purple-500/10 border-purple-500/20';
      default:
        return 'text-slate-400 bg-slate-500/10 border-slate-500/20';
    }
  };

  return (
    <div className={`relative ${className}`}>
      {/* Search Input */}
      <div className="relative">
        <Search className="absolute left-3 top-2.5 text-slate-500" size={16} />
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setIsOpen(true);
          }}
          onFocus={() => setIsOpen(true)}
          placeholder={placeholder}
          className="w-full bg-slate-800 border border-slate-700 text-slate-200 text-sm rounded-lg pl-9 pr-20 py-2 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary/50"
        />
        <div className="absolute right-2 top-2.5 flex items-center gap-1">
          {query && (
            <button
              onClick={() => {
                setQuery('');
                setResults([]);
                inputRef.current?.focus();
              }}
              className="p-0.5 hover:bg-slate-700 rounded text-slate-400 hover:text-slate-200 transition-colors"
            >
              <X size={12} />
            </button>
          )}
          <div className="flex items-center gap-1 px-1.5 py-0.5 rounded bg-slate-700 text-[10px] text-slate-400 font-medium">
            <Command size={10} />
            K
          </div>
        </div>
      </div>

      {/* Search Results Dropdown */}
      {isOpen && results.length > 0 && (
        <div
          ref={resultsRef}
          className="absolute top-full left-0 right-0 mt-2 bg-slate-900 border border-slate-700 rounded-lg shadow-2xl z-50 max-h-96 overflow-y-auto"
        >
          <div className="p-2">
            {results.map((result, index) => (
              <Link
                key={`${result.slug}-${index}`}
                href={`/docs/${result.slug}`}
                className={`
                  block p-3 rounded-md transition-colors group
                  ${index === selectedIndex
                    ? 'bg-primary/10 border border-primary/30'
                    : 'hover:bg-slate-800 border border-transparent'
                  }
                `}
                onClick={() => {
                  setIsOpen(false);
                  setQuery('');
                }}
              >
                <div className="flex items-start gap-3">
                  <FileText size={16} className="text-slate-400 mt-0.5 shrink-0" />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h4 className="font-medium text-slate-200 text-sm">
                        {highlightMatch(result.title, query)}
                      </h4>
                      <span className={`
                        text-[10px] px-1.5 py-0.5 rounded border font-medium
                        ${getCategoryColor(result.category)}
                      `}>
                        {result.category}
                      </span>
                    </div>
                    {result.section && (
                      <p className="text-xs text-slate-400 mb-1">
                        in {highlightMatch(result.section, query)}
                      </p>
                    )}
                    <p className="text-xs text-slate-500 truncate">
                      {highlightMatch(
                        result.content.replace(/[#*`]/g, '').substring(0, 120),
                        query
                      )}
                    </p>
                  </div>
                  <ArrowRight
                    size={14}
                    className={`
                      text-slate-600 mt-0.5 transition-all
                      group-hover:text-primary group-hover:translate-x-0.5
                    `}
                  />
                </div>
              </Link>
            ))}
          </div>

          {results.length === 0 && query && (
            <div className="p-8 text-center">
              <Search size={32} className="text-slate-600 mx-auto mb-3" />
              <p className="text-slate-400 text-sm">No results found for "{query}"</p>
              <p className="text-slate-500 text-xs mt-1">Try different keywords or check spelling</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};