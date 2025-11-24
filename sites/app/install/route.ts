import { NextRequest, NextResponse } from 'next/server';
import { readFileSync } from 'fs';
import { join } from 'path';

export const dynamic = 'force-static';

export async function GET(request: NextRequest) {
  try {
    // Read the install.sh file from the public directory (symlink to root install.sh)
    const installScriptPath = join(process.cwd(), 'public', 'install.sh');
    const installScript = readFileSync(installScriptPath, 'utf8');

    // Always return as plain text (both for curl and browsers)
    return new NextResponse(installScript, {
      status: 200,
      headers: {
        'Content-Type': 'text/plain; charset=utf-8',
        'Cache-Control': 'no-cache, no-store, must-revalidate',
        'Pragma': 'no-cache',
        'Expires': '0',
        // Security headers for script serving
        'X-Content-Type-Options': 'nosniff',
        'Content-Disposition': 'inline; filename="install.sh"',
      },
    });
  } catch (error) {
    console.error('Error serving install script:', error);
    return new NextResponse('Error: Install script not found', {
      status: 404,
      headers: {
        'Content-Type': 'text/plain',
      },
    });
  }
}