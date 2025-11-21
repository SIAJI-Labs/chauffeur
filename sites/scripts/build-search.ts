#!/usr/bin/env node

import { generateSearchIndex } from '../utils/content-extractor';
import fs from 'fs';
import path from 'path';

/**
 * Build script to generate search index for production
 * This script scans all documentation pages and creates a JSON file
 * that can be served statically in production
 */
async function buildSearchIndex() {
  console.log('🔍 Building search index...');

  try {
    const searchIndex = await generateSearchIndex();

    // Ensure public directory exists
    const publicDir = path.join(process.cwd(), 'public');
    if (!fs.existsSync(publicDir)) {
      fs.mkdirSync(publicDir, { recursive: true });
    }

    // Write search index to public directory
    const indexPath = path.join(publicDir, 'search-index.json');
    fs.writeFileSync(indexPath, JSON.stringify(searchIndex, null, 2));

    console.log(`✅ Search index built successfully!`);
    console.log(`📄 Generated ${searchIndex.length} searchable pages`);
    console.log(`📍 Index saved to: ${indexPath}`);

    // Show some statistics
    const sections = [...new Set(searchIndex.map(item => item.section))];
    const categories = [...new Set(searchIndex.map(item => item.category))];

    console.log('\n📊 Search Index Statistics:');
    console.log(`   Sections: ${sections.join(', ')}`);
    console.log(`   Categories: ${categories.join(', ')}`);
    console.log(`   Total content length: ${searchIndex.reduce((sum, item) => sum + item.content.length, 0).toLocaleString()} characters`);

    // Show sample of indexed pages
    console.log('\n📋 Sample indexed pages:');
    searchIndex.slice(0, 5).forEach((item, index) => {
      console.log(`   ${index + 1}. ${item.title} (${item.slug})`);
    });

    if (searchIndex.length > 5) {
      console.log(`   ... and ${searchIndex.length - 5} more pages`);
    }

  } catch (error) {
    console.error('❌ Failed to build search index:', error);
    process.exit(1);
  }
}

// Run the build script
if (require.main === module) {
  buildSearchIndex();
}

export { buildSearchIndex };