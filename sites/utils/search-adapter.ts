import { SearchResult } from './content-extractor';
import { searchIndex as staticSearchIndex, addToSearchIndex } from '../app/docs/_components/SearchData';

/**
 * Search adapter that provides dynamic search in development
 * and static search in production
 */
export class SearchAdapter {
  private static instance: SearchAdapter;
  private dynamicIndex: SearchResult[] | null = null;
  private isInitialized = false;

  private constructor() {}

  public static getInstance(): SearchAdapter {
    if (!SearchAdapter.instance) {
      SearchAdapter.instance = new SearchAdapter();
    }
    return SearchAdapter.instance;
  }

  /**
   * Check if we're in development mode
   */
  private isDevelopment(): boolean {
    return process.env.NODE_ENV === 'development';
  }

  /**
   * Initialize the search index based on environment
   */
  public async initializeSearchIndex(): Promise<SearchResult[]> {
    if (this.isInitialized && this.dynamicIndex) {
      return this.dynamicIndex;
    }

    if (this.isDevelopment()) {
      // Development: Use static index with dynamic updates capability
      console.log('🔍 Initializing search index for development...');
      try {
        // Try to fetch from search API endpoint first (if available)
        const response = await fetch('/api/search-index');
        if (response.ok) {
          const indexData = await response.json() as SearchResult[];
          this.dynamicIndex = indexData || [];
          console.log(`✅ Dynamic search index loaded with ${this.dynamicIndex.length} pages`);
        } else {
          throw new Error(`API not available: ${response.status}`);
        }
      } catch (error) {
        console.warn('⚠️ API not available, using static index:', error);
        this.dynamicIndex = staticSearchIndex.pages;
      }
    } else {
      // Production: Use pre-built static index
      console.log('🔍 Loading static search index for production...');
      try {
        const response = await fetch('/search-index.json');
        if (response.ok) {
          const indexData = await response.json() as SearchResult[];
          this.dynamicIndex = indexData || [];
          console.log(`✅ Static search index loaded with ${this.dynamicIndex.length} pages`);
        } else {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
      } catch (error) {
        console.warn('⚠️ Failed to load static search index, falling back to inline static:', error);
        this.dynamicIndex = staticSearchIndex.pages;
      }
    }

    this.isInitialized = true;
    return this.dynamicIndex;
  }

  /**
   * Populate the static search index with dynamic content
   * This maintains compatibility with existing SearchData
   */
  private populateStaticIndex(dynamicIndex: SearchResult[]): void {
    // Clear existing static index
    staticSearchIndex.pages.length = 0;

    // Add dynamic entries to static index
    dynamicIndex.forEach(item => {
      addToSearchIndex(item);
    });
  }

  /**
   * Get the current search index
   */
  public async getSearchIndex(): Promise<SearchResult[]> {
    if (!this.isInitialized) {
      await this.initializeSearchIndex();
    }
    return this.dynamicIndex || staticSearchIndex.pages;
  }

  /**
   * Force reinitialization of the search index
   * Useful for hot reload during development
   */
  public async reinitialize(): Promise<SearchResult[]> {
    this.isInitialized = false;
    this.dynamicIndex = null;
    return this.initializeSearchIndex();
  }

  /**
   * Get search statistics
   */
  public async getSearchStats(): Promise<{
    totalPages: number;
    sections: string[];
    categories: string[];
    totalContentLength: number;
  }> {
    const index = await this.getSearchIndex();

    return {
      totalPages: index.length,
      sections: [...new Set(index.map(item => item.section))],
      categories: [...new Set(index.map(item => item.category))],
      totalContentLength: index.reduce((sum, item) => sum + item.content.length, 0)
    };
  }
}

/**
 * Convenience function to get search index
 */
export async function getSearchIndex(): Promise<SearchResult[]> {
  const adapter = SearchAdapter.getInstance();
  return adapter.getSearchIndex();
}

/**
 * Convenience function to reinitialize search index
 */
export async function reinitializeSearchIndex(): Promise<SearchResult[]> {
  const adapter = SearchAdapter.getInstance();
  return adapter.reinitialize();
}