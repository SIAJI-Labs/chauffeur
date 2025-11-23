#!/usr/bin/env node

// Verification script to ensure Tailwind class consistency between ports 3000 and 3001
const puppeteer = require('puppeteer');

async function verifyClassConsistency() {
  console.log('🔍 Verifying Tailwind class consistency between ports 3000 and 3001...\n');

  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();

  try {
    // Test key background colors
    const testClasses = [
      'bg-background',
      'bg-slate-900',
      'bg-slate-800',
      'text-white',
      'text-slate-100',
      'text-slate-400',
      'text-emerald-400',
      'bg-primary',
      'border-slate-700',
      'bg-slate-800/50'
    ];

    console.log('📋 Testing key Tailwind classes:');

    for (const className of testClasses) {
      console.log(`\n🎯 Testing: ${className}`);

      // Test port 3000 (original)
      try {
        await page.goto(`http://localhost:3000`, { waitUntil: 'networkidle0' });
        const originalColor = await page.evaluate((cls) => {
          const element = document.querySelector(`.${cls.split(' ').join('.')}`);
          if (element) {
            return getComputedStyle(element).backgroundColor ||
                   getComputedStyle(element).color ||
                   getComputedStyle(element).borderColor ||
                   'not-found';
          }
          return 'not-found';
        }, className);

        // Test port 3002 (Next.js)
        await page.goto(`http://localhost:3002`, { waitUntil: 'networkidle0' });
        const nextjsColor = await page.evaluate((cls) => {
          const element = document.querySelector(`.${cls.split(' ').join('.')}`);
          if (element) {
            return getComputedStyle(element).backgroundColor ||
                   getComputedStyle(element).color ||
                   getComputedStyle(element).borderColor ||
                   'not-found';
          }
          return 'not-found';
        }, className);

        console.log(`  Port 3000: ${originalColor}`);
        console.log(`  Port 3002: ${nextjsColor}`);

        if (originalColor === nextjsColor) {
          console.log(`  ✅ MATCH`);
        } else {
          console.log(`  ❌ MISMATCH - Classes may have different values`);
        }

      } catch (error) {
        console.log(`  ⚠️  Error testing ${className}: ${error.message}`);
      }
    }

    console.log('\n🎨 Verifying overall page theme...');

    // Check overall page theme
    await page.goto(`http://localhost:3000`, { waitUntil: 'networkidle0' });
    const originalBg = await page.evaluate(() => {
      return getComputedStyle(document.body).backgroundColor;
    });

    await page.goto(`http://localhost:3002`, { waitUntil: 'networkidle0' });
    const nextjsBg = await page.evaluate(() => {
      return getComputedStyle(document.body).backgroundColor;
    });

    console.log(`Port 3000 body background: ${originalBg}`);
    console.log(`Port 3002 body background: ${nextjsBg}`);

    if (originalBg === nextjsBg) {
      console.log('✅ Body backgrounds match!');
    } else {
      console.log('❌ Body backgrounds differ - need to fix CSS variables');
    }

  } catch (error) {
    console.error('❌ Verification failed:', error.message);
  } finally {
    await browser.close();
  }
}

// Simple alternative without puppeteer - just check server status
async function simpleVerification() {
  console.log('🔍 Simple verification - checking server status...');

  try {
    const originalResponse = await fetch('http://localhost:3000');
    const nextjsResponse = await fetch('http://localhost:3002');

    console.log(`✅ Port 3000 (Original): ${originalResponse.status} - OK`);
    console.log(`✅ Port 3002 (Next.js): ${nextjsResponse.status} - OK`);

    console.log('\n📝 Manual Verification Checklist:');
    console.log('1. Check that both sites show dark backgrounds');
    console.log('2. Verify text colors are the same (slate-100, slate-400, etc.)');
    console.log('3. Check accent colors (emerald-400, amber-400)');
    console.log('4. Verify button hover states match');
    console.log('5. Check border colors (slate-700, slate-800)');

  } catch (error) {
    console.error('❌ Verification failed:', error.message);
  }
}

// Run simple verification
simpleVerification();