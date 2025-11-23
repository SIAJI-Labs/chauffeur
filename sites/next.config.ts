import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  // Custom output directory for GitHub Actions artifacts
  distDir: 'out',
  // Add empty turbopack config to silence the warning
  turbopack: {},
  webpack: (config, { isServer, dev }) => {
    // Generate search index during production build
    if (isServer && !dev) {
      // Add custom plugin to run search index generation
      config.plugins.push({
        apply: (compiler: any) => {
          compiler.hooks.afterDone.tap('GenerateSearchIndex', async () => {
            try {
              console.log('🔍 Generating search index...');
              const { buildSearchIndex } = require('./scripts/build-search');
              await buildSearchIndex();
              console.log('✅ Search index generated successfully');
            } catch (error) {
              console.error('❌ Failed to generate search index:', error);
            }
          });
        },
      });
    }

    return config;
  },
};

export default nextConfig;
