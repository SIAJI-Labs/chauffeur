"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Download,
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  Terminal,
  Zap
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function UpgradesPage() {
  const currentSlug = 'getting-started/upgrades';

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
        <Link href="/docs/getting-started" className="hover:text-primary transition-colors">Getting Started</Link>
        <ChevronRight size={14} />
        <span className="text-slate-200 capitalize">{currentSlug.split('/').pop()?.replace('-', ' ')}</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Upgrading Chauffeur</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Keep your Chauffeur installation up-to-date with the latest features, bug fixes, and security improvements.
          </p>
        </div>

        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg flex gap-3">
          <Zap className="text-blue-400 shrink-0" />
          <div className="text-sm text-blue-100/80">
            <strong>Why Upgrade?</strong> New versions bring performance improvements, new PHP versions support, enhanced SSL handling, and bug fixes for a more stable development environment.
          </div>
        </div>

        <section id="check-current-version">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="check-current-version">
            Check Current Version
            <Link href="#check-current-version" onClick={(e) => scrollToId(e, 'check-current-version')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur provides dynamic version checking that automatically detects newer versions from GitHub:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Check Version</h3>
          <CodeBlock code={`$ chauf --version
chauf develop-90b3218 (built 2025-11-26T03:36:34Z, commit 90b32186784f4989b7a05d03ae8f62580bdb2e4c)
Latest version: v1.3.6
Update available: Run 'chauf self-update' to upgrade`} />

          <div className="mt-4 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <p className="text-blue-300 text-sm mb-2"><strong>Dynamic Version Detection:</strong></p>
            <ul className="text-blue-200 text-sm space-y-1">
              <li>• Shows your current version with build metadata</li>
              <li>• Fetches latest release from GitHub API</li>
              <li>• Automatically detects if updates are available</li>
              <li>• Works for both development and production builds</li>
            </ul>
          </div>
        </section>

        <section id="upgrade-process">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="upgrade-process">
            Upgrade Process
            <Link href="#upgrade-process" onClick={(e) => scrollToId(e, 'upgrade-process')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Upgrading Chauffeur is a simple, automated process with intelligent version checking:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Standard Upgrade</h3>
          <CodeBlock code={`$ chauf self-update
[INFO] Update available: chauffeur 1.3.5 → 1.3.6
[INFO] Starting self-update process...
[ self-update ] Cloning Chauffeur sources ✓ (1.2s)
[ self-update ] Updating branch main ✓ (2.8s)
[ self-update ] Building Chauffeur CLI ✓ (5.2s)
[ self-update ] binary installed to /home/user/.chauffeur/bin/chauf (5.3s)

[ self-update ] Summary
[ self-update ] Duration: 11.5s
[ self-update ] Previous: fresh installation
[ self-update ] Current: abc1234
[ self-update ] Changes: fresh installation
[ self-update ] ✓ Self-update complete (commit abc1234)

# Verify the upgrade
$ chauf --version
chauf 1.3.6 (built 2025-11-26T04:15:30Z, commit abc1234...)
You're using the latest version (1.3.6)`} />

          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg flex gap-3">
            <AlertTriangle className="text-amber-400 shrink-0" />
            <div className="text-sm text-amber-100/80">
              <strong>Downtime Notice:</strong> Services will be briefly restarted during the upgrade (typically 30-60 seconds). Your projects and configurations remain intact.
            </div>
          </div>
        </section>

        <section id="upgrade-options">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="upgrade-options">
            Upgrade Options
            <Link href="#upgrade-options" onClick={(e) => scrollToId(e, 'upgrade-options')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Dry Run</h3>
          <p className="text-slate-400 mb-4">Preview what would be upgraded without actually doing it:</p>
          <CodeBlock code={`$ chauf update --dry-run

🔍 Checking for updates...
Current version: 2.0.3
Latest version: 2.1.0
📋 Changes that would be applied:
  - Upgrade Chauffeur CLI
  - Update Nginx templates
  - Refresh PHP-FPM configurations
  - Migrate configuration files

💾 Total download size: 15.2 MB
⚠️  Services will be restarted
Run 'chauf update' without --dry-run to proceed`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Force Upgrade</h3>
          <p className="text-slate-400 mb-4">Force upgrade even if already on the latest version:</p>
          <CodeBlock code={`$ chauf update --force

🔄 Forcing Chauffeur upgrade...
📋 Reinstalling current version: 2.1.0
✅ Reinstallation completed`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Skip Service Restart</h3>
          <p className="text-slate-400 mb-4">Upgrade without restarting services (requires manual restart later):</p>
          <CodeBlock code={`$ chauf update --no-restart

🔄 Upgrading Chauffeur (no service restart)...
✅ Upgrade completed
⚠️  Services not restarted. Run 'chauf restart' to apply changes`} />
        </section>

        <section id="what-gets-updated">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="what-gets-updated">
            What Gets Updated
            <Link href="#what-gets-updated" onClick={(e) => scrollToId(e, 'what-gets-updated')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">The upgrade process updates several components:</p>

          <div className="space-y-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">Core Components</h4>
              <ul className="text-sm text-slate-400 space-y-1">
                <li>• Chauffeur CLI binary</li>
                <li>• Nginx configuration templates</li>
                <li>• PHP-FPM pool configurations</li>
                <li>• DNS resolution scripts</li>
                <li>• Default SSL certificates</li>
              </ul>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">Configuration Files</h4>
              <ul className="text-sm text-slate-400 space-y-1">
                <li>• Global configuration (chauffeur.yaml)</li>
                <li>• Default Nginx templates</li>
                <li>• PHP-FPM configuration files</li>
                <li>• Service startup scripts</li>
              </ul>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">What's Preserved</h4>
              <ul className="text-sm text-slate-400 space-y-1">
                <li>• All your project configurations</li>
                <li>• Custom Nginx configurations</li>
                <li>• SSL certificates for your projects</li>
                <li>• PHP installations and extensions</li>
                <li>• Log files and cache</li>
              </ul>
            </div>
          </div>
        </section>

        <section id="post-upgrade">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="post-upgrade">
            Post-Upgrade Steps
            <Link href="#post-upgrade" onClick={(e) => scrollToId(e, 'post-upgrade')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Verify Installation</h3>
          <CodeBlock code={`$ chauf status

┌──────────────────────────────────────────────────────┐
│                  CHAUFFEUR STATUS                    │
├──────────────────────────────────────────────────────┤
│ Version: 2.1.0                                      │
│ Status: ✅ All services running                      │
└──────────────────────────────────────────────────────┘`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Test Your Projects</h3>
          <p className="text-slate-400 mb-4">Verify that your existing projects are working correctly:</p>
          <CodeBlock code={`# Test DNS resolution
$ curl -I http://my-project.test
HTTP/1.1 200 OK

# Test SSL
$ curl -I https://my-project.test
HTTP/1.1 200 OK

# Check project status
$ chauf status --project my-project
✓ Project: my-project (PHP 8.3, SSL: enabled)`} />

          <div className="mt-6 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg flex gap-3">
            <CheckCircle2 className="text-emerald-400 shrink-0" />
            <div className="text-sm text-emerald-100/80">
              <strong>Success!</strong> Your Chauffeur installation has been upgraded successfully. All existing projects should continue to work without any changes.
            </div>
          </div>
        </section>

        <section id="troubleshooting">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="troubleshooting">
            Troubleshooting
            <Link href="#troubleshooting" onClick={(e) => scrollToId(e, 'troubleshooting')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Upgrade Failed
              </h4>
              <CodeBlock code={`# Check logs for detailed error information
$ chauf logs --level error

# Try manual download and install
$ curl -L https://github.com/SIAJI-Labs/chauffeur/releases/latest/download/chauffeur-linux-x86_64 -o chauf
$ chmod +x chauf
$ sudo mv chauf /usr/local/bin/

# Verify installation
$ chauf --version`} />
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Services Not Starting After Upgrade
              </h4>
              <CodeBlock code={`# Check service status
$ chauf status

# Restart all services
$ chauf restart --force

# If issues persist, check logs
$ chauf logs --follow

# Clean cache if needed
$ chauf clean cache --force`} />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Terminal size={16} />
                Configuration Migration Issues
              </h4>
              <CodeBlock code={`# Check configuration compatibility
$ chauf config check

# Reset to default configuration if needed
$ chauf config reset --backup

# Migrate old configuration files
$ chauf config migrate`} />
            </div>
          </div>
        </section>

        <section id="changelog">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="changelog">
            Recent Changes
            <Link href="#changelog" onClick={(e) => scrollToId(e, 'changelog')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">View detailed changelog information:</p>

          <CodeBlock code={`# Show changes since last version
$ chauf update --changelog

📋 Changelog for version 2.1.0:

## 🚀 New Features
- Added PHP 8.3 support
- Enhanced SSL certificate management
- Improved project isolation
- New log filtering options

## 🐛 Bug Fixes
- Fixed DNS resolution issues
- Resolved memory leaks in PHP-FPM
- Improved error handling

## 🔧 Improvements
- Faster startup times (30% improvement)
- Better resource management
- Enhanced security headers

# Show full changelog
$ chauf changelog --all`} />
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
          <div className="text-left">
            <div className="text-xs text-slate-500 mb-1">Previous</div>
            <Link href="/docs/getting-started/first-project" className="text-primary hover:underline">Your First Project</Link>
          </div>
          <div className="text-right">
            <div className="text-xs text-slate-500 mb-1">Next</div>
            <Link href="/docs/getting-started/architecture" className="text-primary hover:underline">Architecture</Link>
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