"use client";

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { Terminal } from '../components/Terminal';
import { CommandExplorer } from '../components/CommandExplorer';
import { HowItWorks } from '../components/HowItWorks';
import { SystemImpactComparison } from '../components/SystemImpactComparison';
import { Button } from '../components/ui/Button';
import { FEATURES, HERO_TERMINAL_LINES } from '../constants';
import { Github, Copy, ChevronRight, Terminal as TerminalIcon, Heart, AlertTriangle, Check } from 'lucide-react';

const LandingPage: React.FC = () => {
  const [scrolled, setScrolled] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 10);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const handleCopyInstall = () => {
    navigator.clipboard.writeText("curl -sL chauffeur.dev/get | bash");
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const scrollToSection = (id: string) => {
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  };

  return (
    <div className="min-h-screen flex flex-col relative">

      {/* Dev Warning Banner */}
      <div className="bg-amber-500/10 border-b border-amber-500/20 px-4 py-2 text-amber-200/90 text-xs md:text-sm font-medium flex items-center justify-center gap-2 relative z-[60]">
        <AlertTriangle size={14} className="text-amber-400" />
        <span>Development Preview: Chauffeur is currently tested on Arch Linux. Support for other distributions is experimental.</span>
      </div>

      {/* Navbar */}
      <nav className={`sticky top-0 z-50 w-full transition-all duration-300 py-4 -mb-20 ${
        scrolled
          ? 'bg-slate-900/95 backdrop-blur-md border-b border-slate-800/50 shadow-2xl'
          : 'bg-transparent'
      }`}>
        <div className="container mx-auto px-4 md:px-6 flex items-center justify-between">
          <div className="flex items-center gap-2 cursor-pointer" onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}>
            <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center shadow-lg shadow-primary/20">
               <TerminalIcon className="text-slate-900" size={20} />
            </div>
            <span className="text-xl font-bold tracking-tight text-slate-100">Chauffeur</span>
          </div>
          <div className="hidden md:flex items-center gap-8 text-sm font-medium text-slate-400">
            <button onClick={() => scrollToSection('features')} className="hover:text-white transition-colors focus:outline-none">Features</button>
            <button onClick={() => scrollToSection('how-it-works')} className="hover:text-white transition-colors focus:outline-none">How it works</button>
            <Link href="/docs" className="hover:text-white transition-colors">Documentation</Link>
          </div>
          <div className="flex items-center gap-3">
            <Button variant="ghost" size="sm" className="hidden sm:flex gap-2">
              <Github size={18} />
              <span className="hidden lg:inline">Star on GitHub</span>
            </Button>
            <Button size="sm" className="bg-slate-100 text-slate-900 hover:bg-white">Download</Button>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="relative pt-32 pb-20 lg:pt-48 lg:pb-32 overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-indigo-900/20 via-slate-900 to-slate-900 z-0" />

        <div className="container mx-auto px-4 md:px-6 relative z-10">
          <div className="flex flex-col lg:flex-row items-center gap-12 lg:gap-20">

            {/* Hero Content */}
            <div className="flex-1 text-center lg:text-left">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm font-medium mb-6">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                </span>
                v0.1.0-beta: Native Linux Support
              </div>

              <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold tracking-tight text-white mb-6 leading-[1.1]">
                The Linux Development Environment <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary to-emerald-300">You've Been Waiting For.</span>
              </h1>

              <p className="text-lg md:text-xl text-slate-400 mb-8 max-w-2xl mx-auto lg:mx-0 leading-relaxed">
                Zero configuration. Isolated PHP versions per project. Automatic SSL and DNS.
                Stop fighting Docker configurations and start coding in seconds.
              </p>

              <div className="flex flex-col sm:flex-row items-center gap-4 justify-center lg:justify-start">
                 <div
                   className="flex items-center w-full sm:w-auto bg-slate-800/50 border border-slate-700 rounded-lg p-1 pr-2 group cursor-pointer hover:border-slate-600 transition-all"
                   onClick={handleCopyInstall}
                 >
                    <code className="px-4 py-2 font-mono text-sm text-emerald-400 select-all">
                      curl -sL chauffeur.dev/get | bash
                    </code>
                    <button className="p-2 hover:bg-slate-700 rounded text-slate-400 hover:text-white transition-colors">
                       {copied ? <Check size={16} className="text-emerald-400" /> : <Copy size={16} />}
                    </button>
                 </div>
                 <Link href="/docs" className="w-full sm:w-auto">
                   <Button variant="secondary" className="w-full">
                     Read the Docs <ChevronRight size={16} className="ml-2" />
                   </Button>
                 </Link>
              </div>

              <p className="mt-6 text-sm text-slate-500">
                Supports Ubuntu, Debian, Arch, and Fedora.
              </p>
            </div>

            {/* Hero Terminal */}
            <div className="flex-1 w-full max-w-xl lg:max-w-2xl perspective-1000">
               <div className="transform rotate-y-[-5deg] rotate-x-[5deg] hover:rotate-0 transition-transform duration-500">
                  <Terminal lines={HERO_TERMINAL_LINES} className="shadow-2xl shadow-indigo-500/10" />
               </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section id="features" className="py-24 bg-slate-900 border-y border-slate-800/50">
        <div className="container mx-auto px-4 md:px-6">
          <div className="text-center max-w-3xl mx-auto mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              Why Developers Choose Chauffeur
            </h2>
            <p className="text-slate-400 text-lg">
              Native performance without the container complexity. Designed for modern PHP workflows.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {FEATURES.map((feature, index) => (
              <div key={index} className="bg-surface p-6 rounded-2xl border border-slate-800 hover:border-primary/30 transition-all hover:shadow-lg hover:shadow-primary/5 group">
                <div className="w-12 h-12 bg-slate-900 rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform border border-slate-800 group-hover:border-primary/30">
                  <feature.icon className="text-emerald-400" size={24} />
                </div>
                <h3 className="text-xl font-semibold text-white mb-2">
                  {feature.title}
                </h3>
                <p className="text-slate-400 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section id="how-it-works" className="py-24 bg-background relative overflow-hidden border-t border-slate-800/50">
        <div className="container mx-auto px-4 md:px-6">
          <HowItWorks />
        </div>
      </section>

      {/* Command Explorer */}
      <section className="py-20">
        <div className="container mx-auto px-4 md:px-6">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              Master the CLI
            </h2>
            <p className="text-xl text-slate-400 max-w-2xl mx-auto">
              A clean, intuitive command-line interface designed for speed.
            </p>
          </div>
          <CommandExplorer />
        </div>
      </section>

      {/* System Impact Comparison */}
      <SystemImpactComparison />

      {/* CTA Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-emerald-500">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
            Ready to Transform Your PHP Development?
          </h2>
          <p className="text-xl text-white/80 mb-8">
            Join developers who are already building amazing PHP applications with Chauffeur.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button
              onClick={handleCopyInstall}
              size="lg"
              variant="secondary"
              className="flex items-center gap-2"
            >
              <TerminalIcon size={20} />
              {copied ? 'Copied!' : 'curl -sL chauffeur.dev/get | bash'}
              {copied ? <Check size={16} /> : <Copy size={16} />}
            </Button>
            <Link href="/docs" className="w-full sm:w-auto">
              <Button
                size="lg"
                variant="outline"
                className="flex items-center gap-2 border-white text-white hover:bg-white hover:text-emerald-500"
              >
                Read the Docs
                <ChevronRight size={16} />
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-slate-900 dark:bg-slate-950 border-t border-slate-800 dark:border-slate-700">
        <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <div className="flex flex-col md:flex-row justify-between items-center gap-4">
            <div className="text-sm text-slate-400">
              © 2024 Chauffeur. Open Source (MIT).
            </div>
            <div className="flex items-center gap-1 text-sm text-slate-400">
              Made with <Heart size={14} className="text-red-500 fill-red-500" /> for the Linux Community
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default function Home() {
  return <LandingPage />;
}