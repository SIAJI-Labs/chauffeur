"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Globe,
  Shield,
  Zap,
  Server,
  Settings
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function NginxPage() {
  const currentSlug = 'core/nginx';

  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  // Helper for safe scrolling
  const scrollToId = (e: React.MouseEvent, id: string) => {
    e.preventDefault();
    const element = document.getElementById(id);
    if (element) {
      const yOffset = -100;
      const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset;
      window.scrollTo({ top: y, behavior: 'smooth' });
    }
  };

  return (
    <>
      {/* Breadcrumbs */}
      <div className="flex items-center gap-2 text-sm text-slate-500 mb-8 overflow-x-auto whitespace-nowrap pb-2">
        <Link href="/" className="hover:text-primary transition-colors">Home</Link>
        <ChevronRight size={14} />
        <Link href="/docs" className="hover:text-primary transition-colors">Documentation</Link>
        <ChevronRight size={14} />
        <Link href="/docs/core" className="hover:text-primary transition-colors">Core Concepts</Link>
        <ChevronRight size={14} />
        <span className="text-slate-200 capitalize">Nginx</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Nginx</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Chauffeur manages its own isolated Nginx installation with optimized configurations for local development.
          </p>
        </div>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="self-contained">
            Self-Contained Web Server
            <Link href="#self-contained" onClick={(e) => scrollToId(e, 'self-contained')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Unlike traditional development environments, Chauffeur manages its own Nginx installation that doesn't interfere with your system.
          </p>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-blue-500/10 rounded-lg">
                  <Server className="text-blue-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Workspace Isolation</h3>
              </div>
              <p className="text-slate-400">
                Nginx binary, configurations, and logs are isolated under <code className="bg-slate-800 px-2 py-1 rounded text-blue-300">~/.chauffeur/nginx/</code>, preventing conflicts with system Nginx.
              </p>
            </div>

            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-emerald-500/10 rounded-lg">
                  <Zap className="text-emerald-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Non-Conflicting Service</h3>
              </div>
              <p className="text-slate-400">
                Runs as <code className="bg-slate-800 px-2 py-1 rounded text-emerald-300">chauf-nginx</code> with unique service names, ensuring it never conflicts with existing web servers.
              </p>
            </div>

            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-purple-500/10 rounded-lg">
                  <Settings className="text-purple-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Template-Based Configs</h3>
              </div>
              <p className="text-slate-400">
                Nginx configurations are automatically generated from optimized templates for each project and framework.
              </p>
            </div>

            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-amber-500/10 rounded-lg">
                  <Globe className="text-amber-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Port Flexibility</h3>
              </div>
              <p className="text-slate-400">
                Uses default workspace ports (8080/8443) with optional system port forwarding for seamless development experience.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="port-management">
            Port Management
            <Link href="#port-management" onClick={(e) => scrollToId(e, 'port-management')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur intelligently manages ports to avoid conflicts while providing a seamless development experience.
          </p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Default Workspace Ports</h3>
          <CodeBlock code={`# No sudo required - works out of the box
HTTP:  8080  → Nginx serves your sites
HTTPS: 8443  → Nginx serves SSL sites

# Visit your sites immediately
http://my-project.test    → Port 8080
https://my-project.test   → Port 8443`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Optional System Port Forwarding</h3>
          <CodeBlock code={`# One-time setup with sudo (optional)
sudo iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 8080
sudo iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 8443

# Now use standard ports
http://my-project.test    → Port 80 (forwarded to 8080)
https://my-project.test   → Port 443 (forwarded to 8443)`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Dynamic Port Allocation</h3>
          <CodeBlock code={`# If 8080/8443 are occupied, Chauffeur automatically:
Port Range: 8080-8099

# Example allocation
HTTP:  8081  → My project (8080 occupied)
HTTPS: 8444  → My project (8443 occupied)
HTTP:  8082  → Another project
HTTPS: 8445  → Another project`} />
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="configuration">
            Configuration Management
            <Link href="#configuration" onClick={(e) => scrollToId(e, 'configuration')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Nginx configurations are automatically managed and optimized for different project types and frameworks.
          </p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Directory Structure</h3>
          <CodeBlock code={`~/.chauffeur/nginx/
├── bin/
│   └── nginx                    # Chauffeur's Nginx binary
├── etc/
│   ├── nginx.conf              # Main Nginx configuration
│   └── mime.types              # MIME type mappings
├── sites-available/           # All project configurations
│   ├── my-laravel.test.conf   # Laravel project config
│   ├── my-blog.test.conf      # WordPress project config
│   └── my-api.test.conf       # API project config
├── sites-enabled/             # Active sites (symlinks)
│   ├── my-laravel.test.conf → ../sites-available/my-laravel.test.conf
│   └── my-blog.test.conf → ../sites-available/my-blog.test.conf
├── conf.d/                     # Additional configurations
├── certs/                      # SSL certificates
├── logs/                       # Access and error logs
└── temp/                       # Temporary files`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Automatic Configuration Features</h3>
          <div className="space-y-4">
            <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
              <h4 className="font-semibold text-blue-300 mb-2">🔧 Framework Detection</h4>
              <p className="text-sm text-blue-100/80">
                Automatically detects Laravel, WordPress, Symfony, and other frameworks to apply optimized configurations.
              </p>
            </div>

            <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
              <h4 className="font-semibold text-emerald-300 mb-2">🔐 SSL Integration</h4>
              <p className="text-sm text-emerald-100/80">
                Automatic SSL certificate configuration, HTTPS redirection, and security headers.
              </p>
            </div>

            <div className="p-4 bg-purple-500/10 border border-purple-500/20 rounded-lg">
              <h4 className="font-semibold text-purple-300 mb-2">⚡ Performance Optimizations</h4>
              <p className="text-sm text-purple-100/80">
                Gzip compression, caching headers, fastcgi_cache, and other performance enhancements.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="ssl-configuration">
            SSL Configuration
            <Link href="#ssl-configuration" onClick={(e) => scrollToId(e, 'ssl-configuration')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur seamlessly integrates SSL certificates into Nginx configurations for secure local development.
          </p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Certificate Management</h3>
          <CodeBlock code={`# SSL certificates are stored in:
~/.chauffeur/nginx/certs/
├── my-project.test.crt         # SSL certificate
├── my-project.test.key         # Private key
└── my-project.test-ca-bundle.crt  # CA bundle (if using mkcert)

# Nginx automatically configured with:
server {
    listen 8443 ssl http2;
    ssl_certificate /path/to/cert.crt;
    ssl_certificate_key /path/to/cert.key;
    # Additional SSL security settings...
}`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">SSL Security Features</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>Modern TLS:</strong> TLS 1.2 and TLS 1.3 support with secure cipher suites</li>
            <li><strong>HSTS Headers:</strong> HTTP Strict Transport Security for enhanced security</li>
            <li><strong>Security Headers:</strong> X-Frame-Options, X-Content-Type-Options, and more</li>
            <li><strong>Automatic Renewal:</strong> Certificates regenerated when domains or aliases change</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="vs-system-nginx">
            vs System Nginx
            <Link href="#vs-system-nginx" onClick={(e) => scrollToId(e, 'vs-system-nginx')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur's Nginx installation offers several advantages over using system Nginx for local development.
          </p>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm border-collapse">
              <thead>
                <tr className="border-b border-slate-700 bg-slate-800/50">
                  <th className="p-3 font-mono text-emerald-400">Feature</th>
                  <th className="p-3 text-slate-300">Chauffeur Nginx</th>
                  <th className="p-3 text-slate-300">System Nginx</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                <tr>
                  <td className="p-3 font-mono text-slate-300">Port Conflicts</td>
                  <td className="p-3 text-slate-400">✅ None (8080/8443)</td>
                  <td className="p-3 text-slate-400">❌ Requires root/sudo</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">Configuration</td>
                  <td className="p-3 text-slate-400">✅ Automatic</td>
                  <td className="p-3 text-slate-400">❌ Manual setup</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">SSL Setup</td>
                  <td className="p-3 text-slate-400">✅ One-command</td>
                  <td className="p-3 text-slate-400">❌ Complex manual</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">Multi-Version</td>
                  <td className="p-3 text-slate-400">✅ Isolated</td>
                  <td className="p-3 text-slate-400">❌ Single version</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">Workspace</td>
                  <td className="p-3 text-slate-400">✅ Self-contained</td>
                  <td className="p-3 text-slate-400">❌ System-wide</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <h4 className="font-semibold text-amber-300 mb-2">💡 Key Benefit</h4>
            <p className="text-sm text-amber-100/80">
              Chauffeur's Nginx provides a <strong>zero-configuration experience</strong> for local development while maintaining professional-grade features and security.
            </p>
          </div>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
          <div className="text-left">
            <div className="text-xs text-slate-500 mb-1">Previous</div>
            <Link href="/docs/core/php-versions" className="text-primary hover:underline">PHP Versions</Link>
          </div>
          <div className="text-right">
            <div className="text-xs text-slate-500 mb-1">Next</div>
            <Link href="/docs/core/composer" className="text-primary hover:underline">Composer</Link>
          </div>
        </div>
      </div>

      {/* Right TOC (Desktop Only) */}
      <div className="hidden xl:block">
        <TableOfContents />
      </div>
    </>
  );
}