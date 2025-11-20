"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Shield,
  Globe,
  Lock,
  Key,
  AlertTriangle,
  CheckCircle2,
  Terminal,
  Copy,
  Settings
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function SSLDomainsPage() {
  const currentSlug = 'core/ssl-domains';

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
        <span className="text-slate-200 capitalize">SSL & Domains</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">SSL & Domains</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Chauffeur provides automatic SSL certificate generation and multi-domain support, making HTTPS development seamless and secure.
          </p>
        </div>

        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg flex gap-3">
          <Shield className="text-blue-400 shrink-0" />
          <div className="text-sm text-blue-100/80">
            <strong>Automatic HTTPS:</strong> Chauffeur generates locally-trusted SSL certificates instantly. No manual certificate management, no browser warnings, just production-ready HTTPS on localhost.
          </div>
        </div>

        <section id="ssl-certificates">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="ssl-certificates">
            SSL Certificates
            <Link href="#ssl-certificates" onClick={(e) => scrollToId(e, 'ssl-certificates')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur uses two methods for SSL certificate generation:</p>

          <div className="space-y-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h3 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Key className="text-emerald-400" />
                mkcert (Recommended)
              </h3>
              <p className="text-slate-400 text-sm mb-2">Creates locally-trusted certificates that work in all browsers without security warnings:</p>
              <ul className="text-sm text-slate-400 space-y-1">
                <li>• Trusted by Chrome, Firefox, Safari, Edge</li>
                <li>• No browser security warnings</li>
                <li>• Supports wildcard certificates (*.domain.test)</li>
                <li>• Automatic certificate renewal</li>
              </ul>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h3 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Lock className="text-amber-400" />
                Self-Signed Certificates
              </h3>
              <p className="text-slate-400 text-sm mb-2">Fallback option that generates certificates locally:</p>
              <ul className="text-sm text-slate-400 space-y-1">
                <li>• Generated automatically by Chauffeur</li>
                <li>• Requires manual browser acceptance</li>
                <li>• Good for testing and development</li>
                <li>• Works when mkcert is unavailable</li>
              </ul>
            </div>
          </div>
        </section>

        <section id="enabling-ssl">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="enabling-ssl">
            Enabling SSL
            <Link href="#enabling-ssl" onClick={(e) => scrollToId(e, 'enabling-ssl')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Enable SSL During Linking</h3>
          <CodeBlock code="chauf link --secure" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Add SSL to Existing Project</h3>
          <CodeBlock code="chauf secure my-project" />

          <div className="bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-4 my-4 flex gap-3">
            <CheckCircle2 className="text-emerald-400 shrink-0" />
            <div className="text-sm text-emerald-100/80">
              <strong>Instant HTTPS:</strong> After enabling SSL, your site is immediately available at <code>https://your-project.test</code> with a valid certificate.
            </div>
          </div>
        </section>

        <section id="multi-domain-support">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="multi-domain-support">
            Multi-Domain Support
            <Link href="#multi-domain-support" onClick={(e) => scrollToId(e, 'multi-domain-support')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur supports multiple domains per project with individual SSL certificates:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Domain Aliases</h3>
          <CodeBlock code="chauf link --domain=admin.my-app.test --secure
chauf link --alias=api.my-app.test --secure
chauf link --alias=*.my-app.test --secure" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Project Configuration Example</h3>
          <CodeBlock code="version: 1
path: /home/user/projects/my-app
php: 8.3
site:
  domain: my-app.test
  ssl: true
domains:
  aliases:
    - domain: admin.my-app.test
      ssl: true
    - domain: api.my-app.test
      ssl: false
    - domain: '*.my-app.test'
      ssl: true
runtime:
  php_fpm_socket: ~/.chauffeur/projects/my-app/runtime/php-fpm/php-fpm.sock" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">View Domain Configuration</h3>
          <CodeBlock code="$ chauf status --project=my-app
Project: my-app
┌─────────────────────┬──────────┬──────────┐
│ Domain              │ SSL      │ Status   │
├─────────────────────┼──────────┼──────────┤
│ my-app.test         │ true     │ active   │
│ admin.my-app.test   │ true     │ active   │
│ api.my-app.test     │ false    │ active   │
│ *.my-app.test       │ true     │ active   │
└─────────────────────┴──────────┴──────────┘" />
        </section>

        <section id="certificate-management">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="certificate-management">
            Certificate Management
            <Link href="#certificate-management" onClick={(e) => scrollToId(e, 'certificate-management')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Certificate Storage</h3>
          <p className="text-slate-400 mb-4">SSL certificates are stored in the Chauffeur workspace:</p>
          <CodeBlock code="~/.chauffeur/nginx/certs/
├── my-app.test/
│   ├── my-app.test.crt
│   ├── my-app.test.key
│   └── my-app.test.conf
├── admin.my-app.test/
│   ├── admin.my-app.test.crt
│   └── admin.my-app.test.key
└── wildcard.my-app.test/
    ├── *.my-app.test.crt
    └── *.my-app.test.key" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Certificate Details</h3>
          <CodeBlock code="$ chauf ssl info my-app.test
Certificate: my-app.test
┌─────────────────────┬─────────────────────────────┐
│ Issuer              │ mkcert development CA        │
│ Subject             │ my-app.test                  │
│ SANs                │ my-app.test, *.my-app.test   │
│ Valid From          │ 2025-01-15 10:30:00 UTC       │
│ Valid Until         │ 2026-01-15 10:30:00 UTC       │
│ Type                │ Locally trusted              │
└─────────────────────┴─────────────────────────────┘" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Regenerate Certificates</h3>
          <CodeBlock code="# Regenerate all certificates for a project
chauf ssl regenerate my-app

# Regenerate specific domain
chauf ssl regenerate my-app.test

# Force regeneration (even if not expired)
chauf ssl regenerate --force my-app" />
        </section>

        <section id="dns-resolution">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="dns-resolution">
            DNS Resolution
            <Link href="#dns-resolution" onClick={(e) => scrollToId(e, 'dns-resolution')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur handles local DNS resolution automatically:</p>

          <div className="bg-surface p-4 rounded-lg border border-slate-800 mb-4">
            <h4 className="font-semibold text-white mb-2">Automatic DNS Configuration</h4>
            <ul className="text-slate-400 space-y-2">
              <li>• <code>*.test</code> domains resolve to <code>127.0.0.1</code></li>
              <li>• Wildcard certificates support <code>*.domain.test</code></li>
              <li>• Subdomains work without additional configuration</li>
              <li>• No <code>/etc/hosts</code> modifications needed</li>
            </ul>
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">DNS Troubleshooting</h3>
          <div className="space-y-4">
            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Domain Not Resolving
              </h4>
              <CodeBlock code="# Test DNS resolution
nslookup my-project.test

# Check if dnsmasq is running
systemctl status dnsmasq

# Restart dnsmasq if needed
sudo systemctl restart dnsmasq" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Terminal size={16} />
                Manual DNS Check
              </h4>
              <CodeBlock code="# Flush DNS cache
sudo systemd-resolve --flush-caches

# Test with curl
curl -v https://my-project.test

# Test with browser incognito mode
# (to bypass any browser caching)" />
            </div>
          </div>
        </section>

        <section id="nginx-configuration">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="nginx-configuration">
            Nginx Configuration
            <Link href="#nginx-configuration" onClick={(e) => scrollToId(e, 'nginx-configuration')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur automatically generates Nginx configurations for SSL domains:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Generated Nginx Config</h3>
          <CodeBlock code={`server {
    listen 80;
    listen 443 ssl http2;
    server_name my-app.test *.my-app.test;

    # SSL Configuration
    ssl_certificate ~/.chauffeur/nginx/certs/my-app.test/my-app.test.crt;
    ssl_certificate_key ~/.chauffeur/nginx/certs/my-app.test/my-app.test.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security Headers
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;

    # Redirect HTTP to HTTPS
    if ($scheme != "https") {
        return 301 https://$host$request_uri;
    }

    # Application
    root /home/user/projects/my-app;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \\.php$ {
        fastcgi_pass unix:~/.chauffeur/projects/my-app/runtime/php-fpm/php-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
}`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Custom Nginx Configuration</h3>
          <p className="text-slate-400 mb-4">For advanced use cases, you can add custom Nginx directives:</p>
          <CodeBlock code={`# Create custom config file
~/.chauffeur/projects/my-app/nginx-custom.conf

# Example: Custom headers
add_header Access-Control-Allow-Origin "*";
add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";

# Example: Rewrite rules
location /api/ {
    try_files $uri $uri/ /api/index.php?$query_string;
}

# Restart to apply changes
chauf restart`} />
        </section>

        <section id="ssl-troubleshooting">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="ssl-troubleshooting">
            SSL Troubleshooting
            <Link href="#ssl-troubleshooting" onClick={(e) => scrollToId(e, 'ssl-troubleshooting')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Certificate Not Trusted
              </h4>
              <CodeBlock code="# Check certificate details
openssl x509 -in ~/.chauffeur/nginx/certs/my-app.test/my-app.test.crt -text -noout

# Install mkcert if not available
curl -fsSL https://pkg.cloudflare.com/pubkey.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloudflare-archive-keyring.gpg
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-archive-keyring.gpg] https://pkg.cloudflare.com/ jammy main' | sudo tee /etc/apt/sources.list.d/cloudflare.list
sudo apt-get update && sudo apt-get install mkcert
mkcert -install

# Regenerate certificates
chauf ssl regenerate --force my-app" />
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Mixed Content Warnings
              </h4>
              <CodeBlock code="# Check for HTTP resources in HTTPS page
# Use browser developer tools to find mixed content

# Common fixes:
# 1. Update hardcoded HTTP URLs to HTTPS
# 2. Use protocol-relative URLs: //example.com/resource.js
# 3. Ensure all assets load over HTTPS" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Settings size={16} />
                SSL Certificate Expiration
              </h4>
              <CodeBlock code="# Check certificate expiration
chauf ssl info my-app.test

# Auto-renewal is handled automatically
# Certificates are valid for 1 year by default
# Renewals happen 30 days before expiration" />
            </div>
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
            <Link href="/docs/reference/configuration" className="text-primary hover:underline">Configuration</Link>
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