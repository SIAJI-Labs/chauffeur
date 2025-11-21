export interface SearchResult {
  title: string;
  slug: string;
  content: string;
  section: string;
  category: string;
}

export interface SearchIndex {
  pages: SearchResult[];
}

// This will be populated by all documentation pages
export const searchIndex: SearchIndex = {
  pages: []
};

// Function to add content to search index
export function addToSearchIndex(content: SearchResult) {
  searchIndex.pages.push(content);
}

// Function to search content
export function searchContent(query: string): SearchResult[] {
  if (!query.trim()) return [];

  const normalizedQuery = query.toLowerCase().trim();

  return searchIndex.pages
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
}