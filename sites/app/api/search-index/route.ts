import { NextResponse } from 'next/server';
import { generateSearchIndex } from '../../../utils/content-extractor';

export async function GET() {
  try {
    console.log('🔍 API: Generating dynamic search index...');
    const searchIndex = await generateSearchIndex();
    console.log(`✅ API: Generated ${searchIndex.length} searchable pages`);

    return NextResponse.json(searchIndex);
  } catch (error) {
    console.error('❌ API: Failed to generate search index:', error);

    return NextResponse.json(
      { error: 'Failed to generate search index' },
      { status: 500 }
    );
  }
}

export async function POST() {
  // Allow POST to force refresh the index
  return GET();
}