"use client";

// React & Next.js
import React, { useEffect, useState } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  Download,
  Terminal,
  Copy,
  Check,
  Github,
  Zap,
  Shield,
  Info
} from 'lucide-react';

// Internal components
import { Terminal as TerminalComponent } from '@/components/Terminal';
import { Button } from '@/components/ui/Button';

// Types
import { TerminalLine } from '@/types';

// Constants
const INSTALL_COMMAND_LINES: TerminalLine[] = [
  { text: "#!/bin/bash", type: "info" },
  { text: "curl -sSL https://chauffeur.siaji.com/install | bash", type: "command" },
  { text: "", type: "info" },
  { text: "🔧 Installing Chauffeur...", type: "info", delay: 200 },
  { text: "📋 Detected current directory is not a Chauffeur git repo", type: "info", delay: 400 },
  { text: "⬇️  Cloning Chauffeur repository...", type: "info", delay: 600 },
  { text: "✅ Repository cloned successfully", type: "success", delay: 800 },
  { text: "🏗️  Building Chauffeur binary from source...", type: "info", delay: 1000 },
  { text: "✅ Binary built and installed", type: "success", delay: 1200 },
  { text: "🌐 Added to PATH in ~/.bashrc", type: "info", delay: 1400 },
  { text: "", type: "info", delay: 1600 },
  { text: "🎉 Chauffeur installation completed!", type: "success", delay: 1800 },
  { text: "   Binary: ~/.chauffeur/bin/chauf", type: "output", delay: 1900 },
  { text: "   Next steps:", type: "output", delay: 2000 },
  { text: "   - Reload shell or run: source ~/.bashrc", type: "output", delay: 2100 },
  { text: "   - Initialize workspace: chauf init", type: "output", delay: 2200 },
  { text: "   - Install services: chauf install", type: "output", delay: 2300 },
  { text: "   - Link your first project: chauf link", type: "output", delay: 2400 }
];

export default function InstallPage() {
  const [copied, setCopied] = useState<string | null>(null);
  const [installStarted, setInstallStarted] = useState(false);

  const handleCopyInstall = () => {
    navigator.clipboard.writeText("curl -sSL https://chauffeur.siaji.com/install | bash");
    setCopied("main");
    setTimeout(() => setCopied(null), 2000);
  };

  const handleCopyManual = () => {
    navigator.clipboard.writeText(`git clone https://github.com/SIAJI-Labs/chauffeur.git\ncd chauffeur\n./install.sh`);
    setCopied("manual");
    setTimeout(() => setCopied(null), 2000);
  };

  const handleCopyDirect = () => {
    navigator.clipboard.writeText('curl -sSL https://chauffeur.siaji.com/install | bash');
    setCopied("direct");
    setTimeout(() => setCopied(null), 2000);
  };

  const handleStartInstall = () => {
    setInstallStarted(true);
    // In a real scenario, this could trigger an actual installation
    setTimeout(() => setInstallStarted(false), 3000);
  };

  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  return (
    <div className="min-h-screen flex flex-col bg-slate-950">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b border-slate-800 bg-slate-900/95 backdrop-blur-md">
        <div className="container mx-auto px-4 md:px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Link href="/" className="flex items-center gap-2 text-white hover:text-primary transition-colors">
                <Terminal className="text-emerald-400" size={20} />
                <span className="text-xl font-bold">Chauffeur</span>
              </Link>
            </div>
            <Link href="/docs" className="text-slate-400 hover:text-white transition-colors">
              Documentation
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="flex-1">
        <div className="container mx-auto px-4 md:px-6 py-16">
          <div className="max-w-4xl mx-auto text-center">
            <h1 className="text-4xl md:text-5xl font-bold text-white mb-6">
              Install Chauffeur
            </h1>
            <p className="text-lg text-slate-400 mb-8">
              Quick, reliable installation for your local PHP development environment
            </p>

            {/* Installation Command */}
            <div className="mb-12">
              <div className="inline-flex items-center w-full max-w-2xl bg-slate-800 border border-slate-700 rounded-lg p-1 group hover:border-slate-600 transition-all">
                <code className="flex-1 px-4 py-3 font-mono text-sm text-emerald-400 select-none">
                  curl -sSL https://chauffeur.siaji.com/install | bash
                </code>
                <button
                  onClick={handleCopyInstall}
                  className="p-3 hover:bg-slate-700 rounded text-slate-400 hover:text-white transition-all"
                  aria-label="Copy installation command"
                >
                  {copied ? (
                    <Check size={16} className="text-emerald-400" />
                  ) : (
                    <Copy size={16} />
                  )}
                </button>
              </div>
            </div>

            {/* Quick Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-4 justify-center mb-12">
              <Link href="/docs/getting-started/installation">
                <Button size="lg" variant="outline" className="flex items-center gap-2">
                  <Info size={20} />
                  Manual Setup Guide
                </Button>
              </Link>
            </div>

            {/* Features */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
              <div className="bg-slate-900 p-6 rounded-xl border border-slate-800">
                <div className="w-12 h-12 bg-emerald-500/10 rounded-lg flex items-center justify-center mb-4 mx-auto">
                  <Zap className="text-emerald-400" size={24} />
                </div>
                <h3 className="text-xl font-semibold text-white mb-2">Auto-Detection</h3>
                <p className="text-slate-400">
                  Automatically detects if you&apos;re in a Chauffeur git repository for offline installation
                </p>
              </div>
              <div className="bg-slate-900 p-6 rounded-xl border border-slate-800">
                <div className="w-12 h-12 bg-blue-500/10 rounded-lg flex items-center justify-center mb-4 mx-auto">
                  <Download className="text-blue-400" size={24} />
                </div>
                <h3 className="text-xl font-semibold text-white mb-2">Remote Fallback</h3>
                <p className="text-slate-400">
                  Clones and builds from source if running from any other directory
                </p>
              </div>
              <div className="bg-slate-900 p-6 rounded-xl border border-slate-800">
                <div className="w-12 h-12 bg-purple-500/10 rounded-lg flex items-center justify-center mb-4 mx-auto">
                  <Shield className="text-purple-400" size={24} />
                </div>
                <h3 className="text-xl font-semibold text-white mb-2">Secure & Clean</h3>
                <p className="text-slate-400">
                  Builds from source with proper security headers and clean installation
                </p>
              </div>
            </div>

            {/* Installation Preview */}
            <div className="text-left max-w-3xl mx-auto">
              <h2 className="text-2xl font-bold text-white mb-4 text-center">
                Installation Preview
              </h2>
              <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
                <div className="p-4 border-b border-slate-800">
                  <p className="text-sm text-slate-400 font-mono">
                    ~/projects $ <span className="text-emerald-400">curl -sSL https://chauffeur.siaji.com/install | bash</span>
                  </p>
                </div>
                <TerminalComponent lines={INSTALL_COMMAND_LINES} className="border-0" />
              </div>
            </div>

            {/* Alternative Installation */}
            <div className="mt-12 text-center">
              <h3 className="text-xl font-semibold text-white mb-4">
                Alternative Installation Methods
              </h3>
              <div className="space-y-4 max-w-4xl mx-auto">
                <div className="bg-slate-800/50 p-4 rounded-lg border border-slate-700">
                  <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                    <Github size={16} />
                    Manual Clone
                  </h4>
                  <div
                    className="block text-sm text-slate-300 font-mono bg-slate-900/50 p-3 rounded border border-slate-600 cursor-pointer hover:bg-slate-900/70 transition-colors"
                    onClick={handleCopyManual}
                  >
                    <div className="flex items-start justify-between gap-4">
                      <pre className="text-xs font-mono min-w-0 flex-1">
{`git clone https://github.com/SIAJI-Labs/chauffeur.git
cd chauffeur
./install.sh`}
                      </pre>
                      <button className="text-xs bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-300 px-2 py-1 rounded border border-emerald-500/30 transition-colors whitespace-nowrap flex-shrink-0">
                        {copied === "manual" ? 'Copied!' : 'Copy'}
                      </button>
                    </div>
                  </div>
                </div>
                <div className="bg-slate-800/50 p-4 rounded-lg border border-slate-700">
                  <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                    <Download size={16} />
                    Direct Download
                  </h4>
                  <div
                    className="block text-sm text-slate-300 font-mono bg-slate-900/50 p-3 rounded border border-slate-600 cursor-pointer hover:bg-slate-900/70 transition-colors"
                    onClick={handleCopyDirect}
                  >
                    <div className="flex items-center justify-between gap-4">
                      <pre className="text-xs font-mono min-w-0 flex-1">
curl -sSL https://chauffeur.siaji.com/install | bash
                      </pre>
                      <button className="text-xs bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-300 px-2 py-1 rounded border border-emerald-500/30 transition-colors whitespace-nowrap flex-shrink-0">
                        {copied === "direct" ? 'Copied!' : 'Copy'}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-800 bg-slate-900">
        <div className="container mx-auto px-4 md:px-6 py-8">
          <div className="text-center">
            <p className="text-slate-400 text-sm">
              After installation, run{' '}
              <code className="bg-slate-800 px-2 py-1 rounded text-emerald-300">chauf init</code>{' '}
              to initialize your workspace, then{' '}
              <code className="bg-slate-800 px-2 py-1 rounded text-emerald-300">chauf install</code>{' '}
              to install services and{' '}
              <code className="bg-slate-800 px-2 py-1 rounded text-emerald-300">chauf link</code>{' '}
              to create your first project.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}