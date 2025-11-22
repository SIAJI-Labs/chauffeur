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
  Command,
  Loader2
} from 'lucide-react';

// Search utilities
import { searchContent, SearchResult } from './SearchData';
import { getSearchIndex } from '../../../utils/search-adapter';

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const SearchModal: React.FC<SearchModalProps> = ({ isOpen, onClose }) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResult[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [isSearching, setIsSearching] = useState(false);
  const [searchIndex, setSearchIndex] = useState<SearchResult[]>([]);
  const [isIndexLoaded, setIsIndexLoaded] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();

  // Focus input when modal opens
  useEffect(() => {
    if (isOpen) {
      inputRef.current?.focus();
      setQuery('');
      setResults([]);
      setSelectedIndex(0);
    }
  }, [isOpen]);

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isOpen) return;

      switch (e.key) {
        case 'Escape':
          e.preventDefault();
          onClose();
          break;
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev =>
            prev < results.length - 1 ? prev + 1 : 0
          );
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev =>
            prev > 0 ? prev - 1 : results.length - 1
          );
          break;
        case 'Enter':
          e.preventDefault();
          if (results[selectedIndex]) {
            router.push(`/docs/${results[selectedIndex].slug}`);
            onClose();
          }
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, results, selectedIndex, onClose, router]);

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
    if (!query.trim()) {
      setResults([]);
      setIsSearching(false);
      return;
    }

    setIsSearching(true);

    // Simulate search delay for UX
    const timer = setTimeout(() => {
      if (isIndexLoaded) {
        // Use dynamic search if index is available, otherwise fall back to static search
        const searchResults = searchIndex.length > 0
          ? performDynamicSearch(query)
          : searchContent(query);

        setResults(searchResults.slice(0, 8)); // Limit to 8 results
        setSelectedIndex(0);
      }
      setIsSearching(false);
    }, 150);

    return () => clearTimeout(timer);
  }, [query, searchIndex, isIndexLoaded]);

  const highlightMatch = (text: string, query: string) => {
    if (!query.trim()) return text;

    const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    const parts = text.split(regex);

    return parts.map((part, index) =>
      regex.test(part) ? (
        <mark key={index} className="bg-emerald-500/30 text-emerald-300 rounded px-0.5 font-medium">
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

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] overflow-y-auto">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal Container */}
      <div className="relative min-h-screen flex items-start justify-center pt-[20vh] p-4">
        <div className="relative w-full max-w-2xl bg-slate-900 border border-slate-700 rounded-xl shadow-2xl overflow-hidden">
          {/* Search Header */}
          <div className="flex items-center gap-3 p-4 border-b border-slate-700">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400" size={18} />
              <input
                ref={inputRef}
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search documentation..."
                className="w-full bg-slate-800 border border-slate-600 text-slate-200 text-base rounded-lg pl-10 pr-4 py-2.5 focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-transparent placeholder:text-slate-500"
                autoComplete="off"
                spellCheck={false}
              />
              {query && (
                <button
                  onClick={() => setQuery('')}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 p-1 text-slate-400 hover:text-slate-200 transition-colors"
                >
                  <X size={14} />
                </button>
              )}
            </div>

            {/* Close button */}
            <button
              onClick={onClose}
              className="p-2 text-slate-400 hover:text-slate-200 transition-colors rounded-lg hover:bg-slate-800"
            >
              <X size={18} />
            </button>
          </div>

          {/* Search Results */}
          <div className="max-h-[60vh] overflow-y-auto">
            {isSearching ? (
              <div className="flex items-center justify-center py-16">
                <Loader2 className="animate-spin text-slate-400" size={24} />
                <span className="ml-3 text-slate-400">Searching...</span>
              </div>
            ) : query && results.length === 0 ? (
              <div className="text-center py-16">
                <Search size={48} className="text-slate-600 mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-slate-200 mb-2">No results found</h3>
                <p className="text-slate-400 mb-4">Try different keywords or check spelling</p>
                <div className="flex flex-wrap gap-2 justify-center">
                  <span className="text-xs bg-slate-800 text-slate-400 px-2 py-1 rounded">commands</span>
                  <span className="text-xs bg-slate-800 text-slate-400 px-2 py-1 rounded">installation</span>
                  <span className="text-xs bg-slate-800 text-slate-400 px-2 py-1 rounded">php</span>
                  <span className="text-xs bg-slate-800 text-slate-400 px-2 py-1 rounded">nginx</span>
                </div>
              </div>
            ) : results.length > 0 ? (
              <div className="p-2">
                {/* Results header */}
                <div className="px-2 py-1.5 text-xs text-slate-500 font-medium">
                  {results.length} result{results.length !== 1 ? 's' : ''}
                </div>

                {/* Results list */}
                {results.map((result, index) => (
                  <Link
                    key={`${result.slug}-${index}`}
                    href={`/docs/${result.slug}`}
                    onClick={() => onClose()}
                    className={`
                      block p-3 mx-1 mb-1 rounded-lg transition-all group
                      ${index === selectedIndex
                        ? 'bg-primary/20 border border-primary/40'
                        : 'hover:bg-slate-800 border border-transparent'
                      }
                    `}
                  >
                    <div className="flex items-start gap-3">
                      <div className="shrink-0 mt-1">
                        <FileText
                          size={16}
                          className={`
                            transition-colors
                            ${index === selectedIndex ? 'text-primary' : 'text-slate-500 group-hover:text-slate-400'}
                          `}
                        />
                      </div>

                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <h4 className="font-medium text-slate-200 text-sm leading-tight">
                            {highlightMatch(result.title, query)}
                          </h4>
                          <span className={`
                            text-[10px] px-1.5 py-0.5 rounded border font-medium shrink-0
                            ${getCategoryColor(result.category)}
                          `}>
                            {result.category}
                          </span>
                        </div>

                        {result.section && (
                          <p className="text-xs text-slate-500 mb-1">
                            in {highlightMatch(result.section, query)}
                          </p>
                        )}

                        <p className="text-xs text-slate-400 leading-relaxed">
                          {highlightMatch(
                            result.content.replace(/[#*`]/g, '').substring(0, 120),
                            query
                          )}
                        </p>
                      </div>

                      <ArrowRight
                        size={14}
                        className={`
                          shrink-0 mt-1 transition-all
                          ${index === selectedIndex
                            ? 'text-primary translate-x-0.5'
                            : 'text-slate-600 group-hover:text-slate-500 group-hover:translate-x-0.5'
                          }
                        `}
                      />
                    </div>
                  </Link>
                ))}
              </div>
            ) : query ? null : (
              <div className="p-8">
                <div className="text-center mb-6">
                  <h3 className="text-lg font-semibold text-slate-200 mb-2">Quick search tips</h3>
                  <p className="text-sm text-slate-400">
                    Search across all documentation pages, commands, and configuration options
                  </p>
                </div>

                {/* Popular searches */}
                <div className="space-y-4">
                  <h4 className="text-xs font-medium text-slate-500 uppercase tracking-wider">Popular searches</h4>
                  <div className="grid grid-cols-2 gap-2">
                    {[
                      'installation',
                      'php versions',
                      'link command',
                      'nginx config',
                      'troubleshooting',
                      'composer'
                    ].map((term) => (
                      <button
                        key={term}
                        onClick={() => setQuery(term)}
                        className="text-left px-3 py-2 text-sm text-slate-300 hover:text-slate-100 hover:bg-slate-800 rounded-lg transition-colors"
                      >
                        {term}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Keyboard shortcuts */}
                <div className="mt-6 pt-6 border-t border-slate-800">
                  <h4 className="text-xs font-medium text-slate-500 uppercase tracking-wider mb-3">Keyboard shortcuts</h4>
                  <div className="space-y-2 text-xs text-slate-400">
                    <div className="flex items-center justify-between">
                      <span>Navigate results</span>
                      <div className="flex gap-1">
                        <kbd className="px-2 py-0.5 bg-slate-800 border border-slate-600 rounded">↑</kbd>
                        <kbd className="px-2 py-0.5 bg-slate-800 border border-slate-600 rounded">↓</kbd>
                      </div>
                    </div>
                    <div className="flex items-center justify-between">
                      <span>Select result</span>
                      <kbd className="px-2 py-0.5 bg-slate-800 border border-slate-600 rounded">Enter</kbd>
                    </div>
                    <div className="flex items-center justify-between">
                      <span>Close search</span>
                      <kbd className="px-2 py-0.5 bg-slate-800 border border-slate-600 rounded">Esc</kbd>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};