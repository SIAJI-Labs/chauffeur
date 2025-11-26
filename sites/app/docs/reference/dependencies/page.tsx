"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Terminal,
  Package,
  Shield,
  Globe,
  Zap,
  Wrench,
  AlertTriangle,
  CheckCircle2,
  Info,
  Copy,
  Settings,
  Cpu,
  Lock,
  Layers,
  HelpCircle
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function DependenciesPage() {
  const currentSlug = 'reference/dependencies';

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
        <span className="text-slate-200 capitalize">Dependencies</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Dependencies & System Requirements</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Chauffeur operates with minimal system requirements but needs specific dependencies for PHP compilation and SSL certificate management. Learn what's required for both the CLI and the services it manages.
          </p>
        </div>

        <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg flex gap-3">
          <CheckCircle2 className="text-emerald-400 shrink-0" />
          <div className="text-sm text-emerald-100/80">
            <strong>Quick Start:</strong> Run <code>chauf doctor</code> to check your system and get installation commands for your specific distribution.
          </div>
        </div>

        <section id="overview">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="overview">
            Overview
            <Link href="#overview" onClick={(e) => scrollToId(e, 'overview')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-6">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <Terminal className="text-primary" />
                CLI Dependencies
              </h3>
              <p className="text-slate-400 mb-3">Required for installing and running Chauffeur itself:</p>
              <ul className="text-slate-400 space-y-2">
                <li className="flex items-start gap-2">
                  <CheckCircle2 size={16} className="text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>Go 1.22+</strong> - For building from source
                    <CodeBlock code="# Verify Go installation
go version
# Expected output: go version go1.22.x linux/amd64" />
                  </div>
                </li>
                <li className="flex items-start gap-2">
                  <CheckCircle2 size={16} className="text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>Core tools</strong> - git, curl, tar
                    <CodeBlock code="# Install on Ubuntu/Debian
sudo apt update && sudo apt install -y git curl tar

# Install on Fedora
sudo dnf install -y git curl tar

# Install on Arch
sudo pacman -S git curl tar" />
                  </div>
                </li>
                <li className="flex items-start gap-2">
                  <CheckCircle2 size={16} className="text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>Build tools</strong> - gcc, make, pkg-config
                    <CodeBlock code="# Install on Ubuntu/Debian
sudo apt install -y build-essential pkg-config

# Install on Fedora
sudo dnf groupinstall -y 'Development Tools'
sudo dnf install -y pkgconf

# Install on Arch
sudo pacman -S base-devel pkgconf" />
                  </div>
                </li>
              </ul>
            </div>

            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-6">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <Cpu className="text-blue-400" />
                Service Dependencies
              </h3>
              <p className="text-slate-400 mb-3">Required for PHP compilation and SSL certificates:</p>
              <ul className="text-slate-400 space-y-2">
                <li className="flex items-start gap-2">
                  <CheckCircle2 size={16} className="text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>PHP Build Libraries</strong> - For compiling PHP from source
                    <CodeBlock code="# Ubuntu/Debian (complete setup)
sudo apt install -y libzip-dev libjpeg-dev libpng-dev libfreetype6-dev \
  libxml2-dev libcurl4-openssl-dev libbz2-dev zlib1g-dev libxslt1-dev \
  libreadline-dev libmagickwand-dev libgmp-dev libsodium-dev

# Fedora
sudo dnf install -y libzip-devel libjpeg-turbo-devel libpng-devel \
  freetype-devel libxml2-devel libcurl-devel bzip2-devel zlib-devel \
  libxslt-devel readline-devel ImageMagick-devel gmp-devel libsodium-devel

# Arch
sudo pacman -S libzip libjpeg-turbo libpng freetype2 libxml2 curl \
  bzip2 zlib libxslt readline imagemagick gmp libsodium" />
                  </div>
                </li>
                <li className="flex items-start gap-2">
                  <CheckCircle2 size={16} className="text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>SSL Certificate Tools</strong> - For trusted certificates
                    <CodeBlock code="# OpenSSL (usually pre-installed)
sudo apt install -y openssl  # Ubuntu/Debian
sudo dnf install -y openssl-devel  # Fedora
sudo pacman -S openssl  # Arch

# mkcert for trusted certificates (recommended)
sudo apt install -y mkcert  # Ubuntu/Debian
sudo dnf install -y mkcert  # Fedora
sudo pacman -S mkcert  # Arch
mkcert -install  # Initialize local CA" />
                  </div>
                </li>
                <li className="flex items-start gap-2">
                  <CheckCircle2 size={16} className="text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>Network Tools</strong> - For port forwarding and DNS
                    <CodeBlock code="# dnsmasq for .test domain resolution
sudo apt install -y dnsmasq  # Ubuntu/Debian
sudo dnf install -y dnsmasq  # Fedora
sudo pacman -S dnsmasq  # Arch

# iptables for port forwarding (usually pre-installed)
sudo apt install -y iptables  # Ubuntu/Debian
sudo dnf install -y iptables  # Fedora
sudo pacman -S iptables  # Arch" />
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </section>

        <section id="php-dependencies">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="php-dependencies">
            PHP Build Dependencies
            <Link href="#php-dependencies" onClick={(e) => scrollToId(e, 'php-dependencies')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4 mb-6">
            <div className="flex items-center gap-2 mb-2">
              <Info className="text-blue-400" />
              <span className="font-semibold text-blue-300">Why these dependencies?</span>
            </div>
            <p className="text-slate-400 text-sm">
              Chauffeur compiles PHP from source to ensure compatibility and include specific extensions. Each library enables different PHP functionality commonly needed by web applications.
            </p>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr className="border-b border-slate-700">
                  <th className="text-left p-3 text-white">Library</th>
                  <th className="text-left p-3 text-white">Purpose</th>
                  <th className="text-left p-3 text-white">PHP Extension</th>
                  <th className="text-left p-3 text-white">Required For</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">libzip</td>
                  <td className="p-3 text-slate-400">ZIP archive support</td>
                  <td className="p-3 text-slate-300"><code>zip</code></td>
                  <td className="p-3 text-slate-400">Archive creation/extraction, Composer packages</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">libjpeg</td>
                  <td className="p-3 text-slate-400">JPEG image processing</td>
                  <td className="p-3 text-slate-300"><code>gd</code></td>
                  <td className="p-3 text-slate-400">Image upload/thumbnail processing</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">libpng</td>
                  <td className="p-3 text-slate-400">PNG image support</td>
                  <td className="p-3 text-slate-300"><code>gd</code></td>
                  <td className="p-3 text-slate-400">PNG image processing</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">freetype</td>
                  <td className="p-3 text-slate-400">Font rendering</td>
                  <td className="p-3 text-slate-300"><code>gd</code></td>
                  <td className="p-3 text-slate-400">Text rendering in images, CAPTCHAs</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">libxml2</td>
                  <td className="p-3 text-slate-400">XML processing</td>
                  <td className="p-3 text-slate-300"><code>xml</code>, <code>dom</code>, <code>simplexml</code></td>
                  <td className="p-3 text-slate-400">XML parsing, web services, RSS feeds</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">libcurl</td>
                  <td className="p-3 text-slate-400">HTTP client library</td>
                  <td className="p-3 text-slate-300"><code>curl</code></td>
                  <td className="p-3 text-slate-400">API calls, HTTP requests, file downloads</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">zlib</td>
                  <td className="p-3 text-slate-400">Compression library</td>
                  <td className="p-3 text-slate-300"><code>zlib</code></td>
                  <td className="p-3 text-slate-400">Output compression, archive handling</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">libxslt</td>
                  <td className="p-3 text-slate-400">XSLT transformation</td>
                  <td className="p-3 text-slate-300"><code>xsl</code></td>
                  <td className="p-3 text-slate-400">XML transformation, stylesheet processing</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">readline</td>
                  <td className="p-3 text-slate-400">Command line editing</td>
                  <td className="p-3 text-slate-300">Built-in</td>
                  <td className="p-3 text-slate-400">Interactive PHP shell, arrow keys in tinker</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">MagickWand</td>
                  <td className="p-3 text-slate-400">ImageMagick binding</td>
                  <td className="p-3 text-slate-300"><code>imagick</code></td>
                  <td className="p-3 text-slate-400">Advanced image processing, photo manipulation</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">gmp</td>
                  <td className="p-3 text-slate-400">Arbitrary precision math</td>
                  <td className="p-3 text-slate-300"><code>gmp</code></td>
                  <td className="p-3 text-slate-400">Cryptography, large number calculations</td>
                </tr>
                <tr className="hover:bg-slate-800/50">
                  <td className="p-3 font-mono text-primary">libsodium</td>
                  <td className="p-3 text-slate-400">Modern cryptography</td>
                  <td className="p-3 text-slate-300"><code>sodium</code></td>
                  <td className="p-3 text-slate-400">Laravel encryption, secure hashing</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <AlertTriangle className="text-amber-400" />
              <span className="font-semibold text-amber-300">Build Time Impact</span>
            </div>
            <p className="text-slate-400 text-sm">
              Installing all these libraries adds significant time to PHP compilation. However, they ensure that common PHP applications work out-of-the-box without additional configuration. Use <code>chauf doctor --auto-fix</code> to install only what's missing.
            </p>
          </div>
        </section>

        <section id="ssl-dependencies">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="ssl-dependencies">
            SSL Certificate Dependencies
            <Link href="#ssl-dependencies" onClick={(e) => scrollToId(e, 'ssl-dependencies')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-6">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <Shield className="text-emerald-400" />
                SSL Certificate Options
              </h3>

              <div className="space-y-4">
                <div className="border border-slate-600 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <CheckCircle2 size={16} className="text-emerald-400" />
                    <span className="font-semibold text-emerald-300">mkcert (Recommended)</span>
                  </div>
                  <p className="text-slate-400 text-sm mb-2">Creates locally-trusted SSL certificates that browsers accept without warnings.</p>
                  <CodeBlock code="# Install mkcert
sudo dnf install -y mkcert  # Fedora
sudo apt install -y mkcert  # Ubuntu/Debian
sudo pacman -S mkcert  # Arch

# Create local certificate authority
mkcert -install

# Create certificates (handled automatically by Chauffeur)
mkcert myapp.test" />
                  <div className="flex gap-2 mt-2">
                    <span className="text-xs bg-emerald-500/20 text-emerald-300 px-2 py-1 rounded">Trusted</span>
                    <span className="text-xs bg-blue-500/20 text-blue-300 px-2 py-1 rounded">No browser warnings</span>
                    <span className="text-xs bg-purple-500/20 text-purple-300 px-2 py-1 rounded">Easy setup</span>
                  </div>
                </div>

                <div className="border border-slate-600 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <AlertTriangle size={16} className="text-amber-400" />
                    <span className="font-semibold text-amber-300">OpenSSL (Fallback)</span>
                  </div>
                  <p className="text-slate-400 text-sm mb-2">Chauffeur can create self-signed certificates if mkcert isn't available.</p>
                  <CodeBlock code="# Create self-signed certificate (handled by Chauffeur)
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes" />
                  <div className="flex gap-2 mt-2">
                    <span className="text-xs bg-amber-500/20 text-amber-300 px-2 py-1 rounded">Built-in</span>
                    <span className="text-xs bg-red-500/20 text-red-300 px-2 py-1 rounded">Browser warnings</span>
                    <span className="text-xs bg-blue-500/20 text-blue-300 px-2 py-1 rounded">No external deps</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-6">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <Lock className="text-blue-400" />
                SSL Security Considerations
              </h3>

              <div className="space-y-3 text-sm">
                <div className="flex items-start gap-3">
                  <CheckCircle2 size={16} className="text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>Development environments</strong> - mkcert provides perfect SSL for local development with trusted certificates
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <AlertTriangle size={16} className="text-amber-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>Production deployments</strong> - Use proper CA-issued certificates, not mkcert or self-signed
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <Info size={16} className="text-blue-400 mt-0.5 shrink-0" />
                  <div>
                    <strong>Multi-domain certificates</strong> - Chauffeur automatically creates SAN certificates covering primary domain + aliases
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="network-dependencies">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="network-dependencies">
            Network & DNS Dependencies
            <Link href="#network-dependencies" onClick={(e) => scrollToId(e, 'network-dependencies')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-6">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <Globe className="text-blue-400" />
                DNS Resolution (.test domains)
              </h3>

              <div className="space-y-3">
                <div>
                  <h4 className="font-medium text-white mb-2">dnsmasq Configuration</h4>
                  <p className="text-slate-400 text-sm mb-3">Chauffeur requires dnsmasq to resolve <code>*.test</code> domains to localhost (127.0.0.1).</p>
                  <CodeBlock code="# Install dnsmasq
sudo apt install -y dnsmasq  # Ubuntu/Debian
sudo dnf install -y dnsmasq  # Fedora
sudo pacman -S dnsmasq  # Arch

# Configure dnsmasq for Chauffeur (manual setup)
sudo tee /etc/dnsmasq.d/chauffeur.conf > /dev/null <<EOF
# Chauffeur .test domain resolution
address=/test/127.0.0.1
listen-address=127.0.0.1
EOF

# Restart dnsmasq
sudo systemctl restart dnsmasq

# Test resolution
nslookup myapp.test" />
                </div>

                <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-3">
                  <div className="flex items-center gap-2 mb-1">
                    <Info className="text-blue-400 text-sm" />
                    <span className="text-blue-300 text-sm font-medium">Automatic DNS Setup</span>
                  </div>
                  <p className="text-slate-400 text-sm">
                    Chauffeur automatically detects dnsmasq configuration and provides exact commands for setup. Use <code>chauf doctor --auto-fix</code> for guided configuration.
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-6">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <Zap className="text-purple-400" />
                Port Forwarding
              </h3>

              <div className="space-y-3">
                <p className="text-slate-400 text-sm">
                  Chauffeur uses iptables to forward privileged ports (80, 443) to user-space ports (8080, 8443) so nginx can run without root privileges.
                </p>

                <div className="bg-slate-900/50 border border-slate-600 rounded p-3">
                  <h4 className="font-medium text-white mb-2">Default Port Configuration:</h4>
                  <div className="space-y-1 text-sm font-mono">
                    <div><span className="text-slate-400">HTTP:</span> 80 → 8080 (user-space nginx)</div>
                    <div><span className="text-slate-400">HTTPS:</span> 443 → 8443 (user-space nginx)</div>
                    <div><span className="text-slate-400">PHP-FPM:</span> 9000 (default, configurable)</div>
                  </div>
                </div>

                <CodeBlock code="# Check iptables rules (managed by Chauffeur)
sudo iptables -t nat -L -n | grep '80 -> 8080'
sudo iptables -t nat -L -n | grep '443 -> 8443'

# Check available ports
netstat -tlnp | grep ':8080'
netstat -tlnp | grep ':8443'" />
              </div>
            </div>
          </div>
        </section>

        <section id="distribution-specific">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="distribution-specific">
            Distribution-Specific Setup
            <Link href="#distribution-specific" onClick={(e) => scrollToId(e, 'distribution-specific')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-4">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <span className="bg-orange-500 text-white text-xs px-2 py-1 rounded">Ubuntu/Debian</span>
              </h3>
              <CodeBlock code="# Complete setup for Ubuntu/Debian
sudo apt update

# Basic CLI dependencies
sudo apt install -y git curl tar build-essential pkg-config

# PHP build dependencies (complete)
sudo apt install -y libzip-dev libjpeg-dev libpng-dev libfreetype6-dev \
  libxml2-dev libcurl4-openssl-dev libbz2-dev zlib1g-dev libxslt1-dev \
  libreadline-dev libmagickwand-dev libgmp-dev libsodium-dev

# SSL and network tools
sudo apt install -y openssl mkcert dnsmasq iptables

# Initialize mkcert
mkcert -install" />
            </div>

            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-4">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <span className="bg-blue-500 text-white text-xs px-2 py-1 rounded">Fedora</span>
              </h3>
              <CodeBlock code="# Complete setup for Fedora
sudo dnf update

# Basic CLI dependencies
sudo dnf groupinstall -y 'Development Tools'
sudo dnf install -y pkgconf

# PHP build dependencies (complete)
sudo dnf install -y libzip-devel libjpeg-turbo-devel libpng-devel \
  freetype-devel libxml2-devel libcurl-devel bzip2-devel zlib-devel \
  libxslt-devel readline-devel ImageMagick-devel gmp-devel libsodium-devel

# SSL and network tools
sudo dnf install -y openssl-devel mkcert dnsmasq iptables

# Initialize mkcert
mkcert -install" />
            </div>

            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-4">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <span className="bg-indigo-500 text-white text-xs px-2 py-1 rounded">Arch Linux</span>
              </h3>
              <CodeBlock code="# Complete setup for Arch Linux
sudo pacman -Syu

# Basic CLI dependencies
sudo pacman -S base-devel pkgconf

# PHP build dependencies (complete)
sudo pacman -S libzip libjpeg-turbo libpng freetype2 libxml2 curl \
  bzip2 zlib libxslt readline imagemagick gmp libsodium

# SSL and network tools
sudo pacman -S openssl mkcert dnsmasq iptables

# Initialize mkcert
mkcert -install" />
            </div>

            </div>

          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <Wrench className="text-amber-400" />
              <span className="font-semibold text-amber-300">Automatic Setup with Doctor</span>
            </div>
            <p className="text-slate-400 text-sm mb-3">
              Use <code>chauf doctor --auto-fix</code> to automatically detect your distribution and install only the missing dependencies.
            </p>
            <CodeBlock code="# Check system and get fix commands
chauf doctor --fix

# Automatically install missing dependencies
chauf doctor --auto-fix" />
          </div>
        </section>

        <section id="troubleshooting">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="troubleshooting">
            Common Dependency Issues
            <Link href="#troubleshooting" onClick={(e) => scrollToId(e, 'troubleshooting')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <Terminal className="text-red-400" />
                PHP Build Fails with "configure: error"
              </h3>
              <p className="text-slate-400 text-sm mb-2">PHP compilation stops due to missing libraries.</p>
              <div className="space-y-2">
                <h4 className="font-medium text-white">Solution:</h4>
                <CodeBlock code="# Check what's missing
chauf doctor --check-php --verbose

# Auto-install missing dependencies
chauf doctor --auto-fix

# Manual verification
pkg-config --modversion libzip
pkg-config --modversion libjpeg
# ... check each library" />
              </div>
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <Globe className="text-amber-400" />
                DNS Resolution Fails
              </h3>
              <p className="text-slate-400 text-sm mb-2"><code>*.test</code> domains don't resolve.</p>
              <CodeBlock code="# Check DNS resolution
nslookup myapp.test

# Check dnsmasq status
sudo systemctl status dnsmasq

# Restart dnsmasq
sudo systemctl restart dnsmasq

# Flush DNS cache
sudo systemd-resolve --flush-caches

# Test again
nslookup myapp.test" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Shield className="text-blue-400" />
                SSL Certificate Warnings
              </h3>
              <p className="text-slate-400 text-sm mb-2">Browser shows SSL security warnings.</p>
              <CodeBlock code="# Install mkcert for trusted certificates
chauf doctor --check-ssl --fix

# Or reinstall mkcert
sudo apt install -y mkcert  # Ubuntu/Debian
sudo dnf install -y mkcert  # Fedora
mkcert -install

# Relink project with SSL
chauf unlink my-project
chauf link --secure my-project" />
            </div>
          </div>
        </section>

        <section id="validation">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="validation">
            Validation & Testing
            <Link href="#validation" onClick={(e) => scrollToId(e, 'validation')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-4">
              <h3 className="font-semibold text-emerald-300 mb-2 flex items-center gap-2">
                <CheckCircle2 className="text-emerald-400" />
                Comprehensive System Check
              </h3>
              <p className="text-slate-400 text-sm mb-3">Run the complete health check to validate all dependencies:</p>
              <CodeBlock code="# Full system validation
chauf doctor

# Include detailed output
chauf doctor --verbose

# Show fix suggestions
chauf doctor --fix

# Auto-fix issues
chauf doctor --auto-fix" />
            </div>

            <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-4">
              <h3 className="font-semibold text-white mb-3 flex items-center gap-2">
                <Settings className="text-blue-400" />
                Manual Validation Steps
              </h3>
              <div className="space-y-3 text-sm">
                <div className="flex items-start gap-3">
                  <span className="font-mono text-primary">1.</span>
                  <div className="text-slate-400">
                    <strong>Build Tools:</strong> Verify compilers and build tools
                    <CodeBlock code="gcc --version
make --version
pkg-config --version" />
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <span className="font-mono text-primary">2.</span>
                  <div className="text-slate-400">
                    <strong>PHP Libraries:</strong> Check via pkg-config
                    <CodeBlock code="pkg-config --modversion libzip
pkg-config --modversion libjpeg
pkg-config --modversion libpng" />
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <span className="font-mono text-primary">3.</span>
                  <div className="text-slate-400">
                    <strong>Network Setup:</strong> Test DNS and ports
                    <CodeBlock code="nslookup myapp.test
sudo lsof -i :80
sudo lsof -i :443" />
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <span className="font-mono text-primary">4.</span>
                  <div className="text-slate-400">
                    <strong>SSL Setup:</strong> Verify certificate tools
                    <CodeBlock code="openssl version
mkcert -version
mkcert -CAROOT" />
                  </div>
                </div>
              </div>
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
            <Link href="/docs/reference/troubleshooting" className="text-primary hover:underline">Troubleshooting</Link>
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