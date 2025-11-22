import fs from 'fs';
import path from 'path';

export interface SearchResult {
  title: string;
  slug: string;
  content: string;
  section: string;
  category: string;
}

export interface PageMetadata {
  title?: string;
  description?: string;
  category?: string;
  section?: string;
}

interface PageInfo {
  path: string;
  slug: string;
  metadata: PageMetadata;
  section: string;
}

/**
 * Extract text content from a TSX/MD file by removing imports, exports, and JSX tags
 */
export function extractTextContent(fileContent: string): string {
  return fileContent
    // Remove imports
    .replace(/import.*from.*['"];/g, '')
    // Remove export statements
    .replace(/export.*function.*\{[\s\S]*?\}/g, '')
    .replace(/export.*default.*[\s\S]*?;/g, '')
    // Remove React component definitions
    .replace(/function.*\([^)]*\).*\{[\s\S]*?\}/g, '')
    // Remove JSX opening tags
    .replace(/<[^>]*>/g, ' ')
    // Remove JSX closing tags
    .replace(/<\/[^>]*>/g, ' ')
    // Remove code blocks but keep the code
    .replace(/<CodeBlock[^>]*>([\s\S]*?)<\/CodeBlock>/g, '$1')
    // Remove inline code markers but keep content
    .replace(/`([^`]+)`/g, '$1')
    // Normalize whitespace
    .replace(/\s+/g, ' ')
    // Clean up extra spaces
    .replace(/\s+/g, ' ')
    .trim();
}

/**
 * Extract metadata from file content (title, description, etc.)
 */
export function extractMetadata(fileContent: string): PageMetadata {
  const metadata: PageMetadata = {};

  // Extract title from various patterns
  const titleMatch = fileContent.match(/title:\s*['"]([^'"]+)['"]/);
  if (titleMatch) metadata.title = titleMatch[1];

  // Extract description
  const descMatch = fileContent.match(/description:\s*['"]([^'"]+)['"]/);
  if (descMatch) metadata.description = descMatch[1];

  // Extract category
  const categoryMatch = fileContent.match(/category:\s*['"]([^'"]+)['"]/);
  if (categoryMatch) metadata.category = categoryMatch[1];

  // Extract section
  const sectionMatch = fileContent.match(/section:\s*['"]([^'"]+)['"]/);
  if (sectionMatch) metadata.section = sectionMatch[1];

  // If no title found, try to extract from h1 tags
  if (!metadata.title) {
    const h1Match = fileContent.match(/<h1[^>]*>([^<]+)<\/h1>/);
    if (h1Match) metadata.title = h1Match[1].trim();
  }

  return metadata;
}

/**
 * Extract section from slug path
 */
export function extractSection(slug: string): string {
  const parts = slug.split('/');
  return parts[0] || 'Other';
}

/**
 * Extract category from slug path
 */
export function extractCategory(slug: string): string {
  const parts = slug.split('/');
  if (parts.length >= 2) {
    return parts[0].charAt(0).toUpperCase() + parts[0].slice(1);
  }
  return 'Other';
}

/**
 * Recursively scan the docs directory for page files
 */
export async function scanDocumentsDirectory(docsPath: string): Promise<PageInfo[]> {
  const pages: PageInfo[] = [];

  function scanDirectory(dir: string, basePath: string = ''): void {
    try {
      const items = fs.readdirSync(dir);

      for (const item of items) {
        const fullPath = path.join(dir, item);
        const relativePath = path.join(basePath, item);

        const stat = fs.statSync(fullPath);

        if (stat.isDirectory()) {
          scanDirectory(fullPath, relativePath);
        } else if (item === 'page.tsx' || item === 'page.md') {
          const slug = basePath.replace(/\\/g, '/');
          const fileContent = fs.readFileSync(fullPath, 'utf8');
          const metadata = extractMetadata(fileContent);
          const section = extractSection(slug);

          pages.push({
            path: fullPath,
            slug: slug,
            metadata: metadata,
            section: section
          });
        }
      }
    } catch (error) {
      console.warn(`Error scanning directory ${dir}:`, error);
    }
  }

  scanDirectory(docsPath);
  return pages;
}

/**
 * Extract content from a specific page file
 */
export async function extractContentFromPage(pagePath: string): Promise<SearchResult> {
  const fileContent = fs.readFileSync(pagePath, 'utf8');
  const metadata = extractMetadata(fileContent);
  const textContent = extractTextContent(fileContent);

  // Generate slug from path
  const relativePath = path.relative(path.join(process.cwd(), 'app/docs'), pagePath);
  const slug = relativePath.replace(/\\/g, '/').replace(/\/page\.(tsx|md)$/, '');

  return {
    title: metadata.title || extractTitleFromContent(fileContent) || 'Untitled',
    slug: slug,
    content: metadata.description ? `${metadata.description} ${textContent}` : textContent,
    section: metadata.section || extractSection(slug),
    category: metadata.category || extractCategory(slug)
  };
}

/**
 * Generate search index for all documentation pages
 */
export async function generateSearchIndex(): Promise<SearchResult[]> {
  const docsPath = path.join(process.cwd(), 'app', 'docs');

  if (!fs.existsSync(docsPath)) {
    console.warn('Docs directory not found:', docsPath);
    return [];
  }

  try {
    const pages = await scanDocumentsDirectory(docsPath);
    const searchIndex: SearchResult[] = [];

    for (const page of pages) {
      const result = await extractContentFromPage(page.path);
      searchIndex.push(result);
    }

    return searchIndex;
  } catch (error) {
    console.error('Error generating search index:', error);
    return [];
  }
}

/**
 * Extract title from content using various patterns
 */
function extractTitleFromContent(fileContent: string): string | null {
  // Look for title in metadata
  const metadataTitle = fileContent.match(/title:\s*['"]([^'"]+)['"]/);
  if (metadataTitle) return metadataTitle[1];

  // Look for h1 tags
  const h1Match = fileContent.match(/<h1[^>]*>([^<]+)<\/h1>/);
  if (h1Match) return h1Match[1].trim();

  // Look for const title = "..."
  const constTitleMatch = fileContent.match(/const\s+title\s*=\s*['"]([^'"]+)['"]/);
  if (constTitleMatch) return constTitleMatch[1];

  return null;
}