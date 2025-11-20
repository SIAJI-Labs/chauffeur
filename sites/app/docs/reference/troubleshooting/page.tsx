"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  AlertTriangle,
  Terminal,
  RefreshCw,
  Settings,
  Search,
  HelpCircle,
  CheckCircle2,
  XCircle,
  Info,
  Copy,
  ExternalLink
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function TroubleshootingPage() {
  const currentSlug = 'reference/troubleshooting';

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
        <Link href="/docs/reference" className="hover:text-primary transition-colors">Reference</Link>
        <ChevronRight size={14} />
        <span className="text-slate-200 capitalize">Troubleshooting</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Troubleshooting</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Common issues and solutions for Chauffeur. If you're experiencing problems not covered here, check the logs or ask for help in the community.
          </p>
        </div>

        <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg flex gap-3">
          <HelpCircle className="text-amber-400 shrink-0" />
          <div className="text-sm text-amber-100/80">
            <strong>Quick Help:</strong> Most issues can be resolved by restarting Chauffeur (<code>chauf restart</code>) or checking the logs (<code>chauf logs</code>).
          </div>
        </div>

        <section id="installation-issues">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="installation-issues">
            Installation Issues
            <Link href="#installation-issues" onClick={(e) => scrollToId(e, 'installation-issues')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <XCircle size={16} />
                Command Not Found: chauf command not found
              </h3>
              <p className="text-slate-400 text-sm mb-2">After installation, the <code>chauf</code> command isn't recognized.</p>
              <div className="space-y-2">
                <h4 className="font-medium text-white">Solutions:</h4>
                <CodeBlock code="# Restart your terminal session
# This reloads PATH environment variable

# Or manually source the profile
source ~/.bashrc
# or
source ~/.zshrc

# Check if chauf is installed
which chauf
# Should show: ~/.chauffeur/bin/chauf" />
                <h4 className="font-medium text-white mt-3">If still not found:</h4>
                <CodeBlock code={`# Add to shell manually
echo 'export PATH="$HOME/.chauffeur/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For Zsh
echo 'export PATH="$HOME/.chauffeur/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc`} />
              </div>
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Permission Denied during installation
              </h3>
              <p className="text-slate-400 text-sm mb-2">Installer fails due to insufficient permissions.</p>
              <CodeBlock code="# Run installer with proper permissions
curl -sL https://chauffeur.siaji.com/install | sudo bash

# Or ensure your user has proper permissions
sudo usermod -aG docker $USER  # If using Docker features (not Chauffeur)
sudo usermod -aG www-data $USER  # For web server access" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Info size={16} />
                Network Issues during installation
              </h3>
              <p className="text-slate-400 text-sm mb-2">Cannot download Chauffeur binary.</p>
              <CodeBlock code="# Check internet connection
curl -I https://github.com

# Try alternative download method
wget https://github.com/SIAJI-Labs/chauffeur/releases/latest/download/chauffeur-linux-amd64
chmod +x chauffeur-linux-amd64
sudo mv chauffeur-linux-amd64 /usr/local/bin/chauf" />
            </div>
          </div>
        </section>

        <section id="service-issues">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="service-issues">
            Service Issues
            <Link href="#service-issues" onClick={(e) => scrollToId(e, 'service-issues')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <XCircle size={16} />
                Services Won't Start
              </h3>
              <p className="text-slate-400 text-sm mb-2">Chauffeur services fail to start when running <code>chauf start</code>.</p>
              <div className="space-y-2">
                <h4 className="font-medium text-white">Diagnostic Steps:</h4>
                <CodeBlock code="# Check service status
chauf status

# Check for port conflicts
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :443

# Check logs for errors
chauf logs
chauf logs nginx
chauf logs php" />
                <h4 className="font-medium text-white mt-3">Common Solutions:</h4>
                <CodeBlock code="# Kill conflicting services
sudo systemctl stop nginx
sudo systemctl stop apache2
sudo systemctl stop docker

# Check if ports are available
sudo lsof -i :80
sudo lsof -i :443

# Restart Chauffeur
chauf stop
chauf start" />
              </div>
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Port Already in Use
              </h3>
              <p className="text-slate-400 text-sm mb-2">Another service is using the required ports.</p>
              <CodeBlock code="# Find what's using the ports
sudo lsof -i :80
sudo lsof -i :443

# Option 1: Stop conflicting services
sudo systemctl stop nginx apache2

# Option 2: Change Chauffeur ports
chauf config edit
# Set nginx.http_port and nginx.https_port to different values

# Option 3: Use automatic port resolution
# Chauffeur will automatically find available ports in range 8080-8099
chauf start" />
            </div>

            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <XCircle size={16} />
                Project Not Accessible
              </h3>
              <p className="text-slate-400 text-sm mb-2">Linked project returns 404 or connection refused.</p>
              <CodeBlock code="# Check if project is linked
chauf status

# Check DNS resolution
nslookup my-project.test

# Check if project files exist
ls -la ~/.chauffeur/projects/
cat ~/.chauffeur/projects/my-project/project.yaml

# Check nginx configuration
chauf logs nginx --follow

# Restart services
chauf restart --project=my-project" />
            </div>
          </div>
        </section>

        <section id="php-issues">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="php-issues">
            PHP Issues
            <Link href="#php-issues" onClick={(e) => scrollToId(e, 'php-issues')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <XCircle size={16} />
                PHP Version Not Available
              </h3>
              <p className="text-slate-400 text-sm mb-2">Required PHP version isn't installed.</p>
              <CodeBlock code="# Check available PHP versions
chauf php list

# Install required version
chauf php install 8.3

# Install with specific options
chauf php install 8.3 --force --from source

# Verify installation
~/.chauffeur/php/8.3/bin/php --version" />
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                PHP Errors in Application
              </h3>
              <p className="text-slate-400 text-sm mb-2">PHP application shows errors or blank pages.</p>
              <CodeBlock code="# Enable PHP error reporting
chauf config edit --project=my-app
# Add to php_values section:
# php_values:
#   display_errors: on
#   error_reporting: E_ALL

# Check PHP error log
tail -f ~/.chauffeur/projects/my-app/logs/php/error.log

# Check PHP-FPM errors
chauf logs fpm --project=my-app

# Test PHP configuration
~/.chauffeur/php/8.3/bin/php -i | grep error_reporting" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Info size={16} />
                Slow PHP Performance
              </h3>
              <p className="text-slate-400 text-sm mb-2">PHP applications are loading slowly.</p>
              <CodeBlock code="# Check OPcache status
~/.chauffeur/php/8.3/bin/php -m | grep opcache

# Enable OPcache if not enabled
chauf config edit
# Set opcache.enable: true

# Check PHP-FPM process status
ps aux | grep php-fpm

# Monitor memory usage
ps aux --sort=-%mem | head

# Tune PHP-FPM settings
chauf config edit --project=my-app
# Adjust max_children, start_servers, etc." />
            </div>
          </div>
        </section>

        <section id="ssl-issues">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="ssl-issues">
            SSL Issues
            <Link href="#ssl-issues" onClick={(e) => scrollToId(e, 'ssl-issues')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <XCircle size={16} />
                Certificate Not Trusted
              </h3>
              <p className="text-slate-400 text-sm mb-2">Browser shows SSL security warning.</p>
              <CodeBlock code={`# Check certificate issuer
openssl x509 -in ~/.chauffeur/nginx/certs/my-app.test/my-app.test.crt -text -noout | grep -A5 "Issuer"

# Install or reinstall mkcert
curl -fsSL https://pkg.cloudflare.com/pubkey.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloudflare-archive-keyring.gpg
echo 'deb [signed-by=/usr/share/keyrings/cloudflare-archive-keyring.gpg] https://pkg.cloudflare.com jammy main' | sudo tee /etc/apt/sources.list.d/cloudflare.list
sudo apt-get update && sudo apt-get install mkcert
mkcert -install

# Regenerate certificates
chauf ssl regenerate --force my-app`} />
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Mixed Content Warnings
              </h3>
              <p className="text-slate-400 text-sm mb-2">HTTPS page loads HTTP resources.</p>
              <CodeBlock code="# Check for HTTP resources
# Use browser developer tools > Network tab
# Look for mixed content warnings

# Common fixes:
# 1. Update hardcoded URLs to HTTPS
# 2. Use protocol-relative URLs: //example.com/resource.js
# 3. Ensure all assets load over HTTPS

# Test with curl
curl -I https://my-app.test" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Info size={16} />
                Certificate Expired
              </h3>
              <p className="text-slate-400 text-sm mb-2">SSL certificate has expired.</p>
              <CodeBlock code="# Check certificate expiration
chauf ssl info my-app.test

# Auto-renewal is enabled by default
# Certificates renew 30 days before expiration

# Force renewal
chauf ssl regenerate my-app --force

# Enable auto-renewal if disabled
chauf config edit
# Set ssl.auto_renewal: true" />
            </div>
          </div>
        </section>

        <section id="domain-issues">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="domain-issues">
            Domain Issues
            <Link href="#domain-issues" onClick={(e) => scrollToId(e, 'domain-issues')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <XCircle size={16} />
                Domain Not Resolving
              </h3>
              <p className="text-slate-400 text-sm mb-2"><code>*.test</code> domains aren't resolving.</p>
              <CodeBlock code="# Test DNS resolution
nslookup my-project.test
ping my-project.test

# Check dnsmasq status
sudo systemctl status dnsmasq

# Restart dnsmasq if needed
sudo systemctl restart dnsmasq

# Check dnsmasq configuration
cat /etc/dnsmasq.d/chauffeur.conf

# Flush DNS cache
sudo systemd-resolve --flush-caches" />
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Subdomain Not Working
              </h3>
              <p className="text-slate-400 text-sm mb-2">Subdomains of linked project aren't accessible.</p>
              <CodeBlock code="# Check project configuration
chauf config show --project=my-app

# Add subdomain alias
chauf link --alias=admin.my-app.test --secure

# Or edit project.yaml manually
# domains:
#   aliases:
#     - domain: admin.my-app.test
#       ssl: true

# Restart services
chauf restart --project=my-app" />
            </div>
          </div>
        </section>

        <section id="performance-issues">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="performance-issues">
            Performance Issues
            <Link href="#performance-issues" onClick={(e) => scrollToId(e, 'performance-issues')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                High Memory Usage
              </h3>
              <p className="text-slate-400 text-sm mb-2">Chauffeur is using too much memory.</p>
              <CodeBlock code="# Check memory usage
chauf status --detail
ps aux --sort=-%mem | head

# Enable memory optimization
chauf config edit
# Set performance.memory_optimization: true

# Tune PHP-FPM
chauf config edit --project=my-app
# Adjust max_children based on available memory
# Rule: max_children × memory_limit_per_process < 80% of available RAM

# Restart to apply changes
chauf restart" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Info size={16} />
                Slow Load Times
              </h3>
              <p className="text-slate-400 text-sm mb-2">Websites are loading slowly.</p>
              <CodeBlock code="# Enable gzip compression
chauf config edit
# Set nginx.gzip: true

# Enable caching
chauf config edit
# Set nginx.cache.enabled: true

# Optimize PHP
chauf config edit --project=my-app
# Enable opcache and tune settings

# Monitor performance
chauf logs access --project=my-app --follow
# Look for slow requests (>2s)" />
            </div>
          </div>
        </section>

        <section id="getting-help">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="getting-help">
            Getting Help
            <Link href="#getting-help" onClick={(e) => scrollToId(e, 'getting-help')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Search size={16} />
                Diagnostic Commands
              </h3>
              <CodeBlock code="# Full system diagnostic
chauf doctor

# Detailed status
chauf status --detail

# Check logs
chauf logs --level=error --follow

# Validate configuration
chauf config validate" />
            </div>

            <div className="bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-emerald-300 mb-2 flex items-center gap-2">
                <ExternalLink size={16} />
                Community Resources
              </h3>
              <ul className="text-slate-400 space-y-2">
                <li>• <strong>GitHub Issues:</strong> <a href="https://github.com/SIAJI-Labs/chauffeur/issues" className="text-primary hover:underline">Report bugs and feature requests</a></li>
                <li>• <strong>Documentation:</strong> <a href="https://docs.chauffeur.dev" className="text-primary hover:underline">Full documentation</a></li>
                <li>• <strong>Discussions:</strong> <a href="https://github.com/SIAJI-Labs/chauffeur/discussions" className="text-primary hover:underline">Community discussions</a></li>
                <li>• <strong>Wiki:</strong> <a href="https://github.com/SIAJI-Labs/chauffeur/wiki" className="text-primary hover:underline">Community guides and tips</a></li>
              </ul>
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <HelpCircle size={16} />
                When Reporting Issues
              </h3>
              <p className="text-slate-400 text-sm mb-2">Include the following information when reporting problems:</p>
              <ul className="text-slate-400 space-y-2 text-sm">
                <li>• Chauffeur version: <code>chauf --version</code></li>
                <li>• Operating system: <code>uname -a</code></li>
                <li>• Error messages and logs</li>
                <li>• Steps to reproduce the issue</li>
                <li>• Expected vs actual behavior</li>
              </ul>
            </div>
          </div>
        </section>

        <section id="common-fixes">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="common-fixes">
            Common Quick Fixes
            <Link href="#common-fixes" onClick={(e) => scrollToId(e, 'common-fixes')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <RefreshCw size={16} className="text-blue-400" />
                Restart Everything
              </h4>
              <CodeBlock code="# Restart all Chauffeur services
chauf stop
sleep 2
chauf start" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Settings size={16} className="text-emerald-400" />
                Reset Configuration
              </h4>
              <CodeBlock code="# Reset to defaults (warning: removes custom settings)
chauf config reset

# Reset specific project
chauf config reset --project=my-app" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Terminal size={16} className="text-amber-400" />
                Clean Workspace
              </h4>
              <CodeBlock code="# Clean old files and logs
chauf clean

# Clean specific project
chauf clean --project=my-app" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Copy size={16} className="text-purple-400" />
                Export/Import Config
              </h4>
              <CodeBlock code="# Backup configuration
chauf config export > config-backup.yaml

# Restore configuration
chauf config import config-backup.yaml" />
            </div>
          </div>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
          <div className="text-left">
            <div className="text-xs text-slate-500 mb-1">Previous</div>
            <Link href="/docs/reference/configuration" className="text-primary hover:underline">Configuration</Link>
          </div>
          <div className="text-right">
            <div className="text-xs text-slate-500 mb-1">Next</div>
            <Link href="/docs/reference/commands" className="text-primary hover:underline">CLI Commands</Link>
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