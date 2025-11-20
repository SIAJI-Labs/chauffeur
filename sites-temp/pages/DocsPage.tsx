import React, { useState, useEffect } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { Terminal as TerminalIcon, Menu, X, ChevronRight, AlertTriangle, Github, ArrowLeft } from 'lucide-react';
import { DocSidebar } from '../components/docs/DocSidebar';
import { TableOfContents } from '../components/docs/TableOfContents';
import { CodeBlock } from '../components/docs/CodeBlock';
import { Button } from '../components/ui/Button';

const DocsPage: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();
  const currentSlug = location.pathname.replace('/docs/', '') || 'getting-started/installation';

  // Scroll to top on route change
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [currentSlug]);

  // Content Renderer (Simulated based on slug)
  const renderContent = () => {
    switch (true) {
      case currentSlug.includes('installation'):
        return <InstallationContent />;
      case currentSlug.includes('first-project'):
        return <FirstProjectContent />;
      case currentSlug.includes('commands'):
        return <CommandsContent />;
      default:
        return <InstallationContent />;
    }
  };

  return (
    <div className="min-h-screen bg-background text-slate-300">
      {/* Docs Navbar */}
      <nav className="sticky top-0 z-50 w-full bg-slate-900/95 backdrop-blur border-b border-slate-800">
        <div className="w-full px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button 
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="lg:hidden p-2 text-slate-400 hover:text-white"
            >
              {sidebarOpen ? <X /> : <Menu />}
            </button>
            <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
               <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                 <TerminalIcon className="text-slate-900" size={20} />
               </div>
               <span className="text-xl font-bold tracking-tight text-slate-100 hidden sm:inline">Chauffeur</span>
               <span className="text-xs bg-slate-800 text-slate-400 px-2 py-0.5 rounded-full ml-2 border border-slate-700">Docs</span>
            </Link>
          </div>
          <div className="flex items-center gap-3">
             <a href="https://github.com" target="_blank" rel="noreferrer" className="text-slate-400 hover:text-white transition-colors">
                <Github size={20} />
             </a>
             <Button size="sm" variant="outline" className="hidden sm:flex">v0.1.0-beta</Button>
          </div>
        </div>
      </nav>

      <div className="flex max-w-[1440px] mx-auto">
        {/* Left Sidebar */}
        <DocSidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />

        {/* Main Content Area */}
        <main className="flex-1 min-w-0 lg:pl-[280px] xl:pr-[280px]">
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
            {/* Breadcrumbs */}
            <div className="flex items-center gap-2 text-sm text-slate-500 mb-8 overflow-x-auto whitespace-nowrap pb-2">
              <Link to="/" className="hover:text-primary transition-colors">Home</Link>
              <ChevronRight size={14} />
              <span>Documentation</span>
              <ChevronRight size={14} />
              <span className="text-slate-200 capitalize">{currentSlug.split('/').pop()?.replace('-', ' ')}</span>
            </div>

            {/* Content Injection */}
            <div className="prose prose-invert prose-slate max-w-none">
               {renderContent()}
            </div>

            {/* Page Footer Navigation */}
            <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
               <div className="text-left">
                 <div className="text-xs text-slate-500 mb-1">Previous</div>
                 <Link to="#" className="text-primary hover:underline">Introduction</Link>
               </div>
               <div className="text-right">
                 <div className="text-xs text-slate-500 mb-1">Next</div>
                 <Link to="#" className="text-primary hover:underline">First Project</Link>
               </div>
            </div>
          </div>
        </main>

        {/* Right TOC (Desktop Only) */}
        <TableOfContents />
      </div>
    </div>
  );
};

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

// --- Content Components (Simulated) ---

const InstallationContent = () => (
  <div className="space-y-8 animate-fade-in">
    <div>
      <h1 className="text-4xl font-bold text-white mb-4">Installation</h1>
      <p className="text-lg text-slate-400 leading-relaxed">
        Chauffeur is designed to be installed in seconds. It currently supports Debian-based (Ubuntu, Pop!_OS) and Arch-based distributions.
      </p>
    </div>

    <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg flex gap-3">
       <AlertTriangle className="text-amber-400 shrink-0" />
       <div className="text-sm text-amber-100/80">
         <strong>Prerequisite:</strong> Ensure you have <code>curl</code>, <code>git</code>, and <code>unzip</code> installed on your system before proceeding.
       </div>
    </div>

    <section id="quick-install">
      <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="quick-install">
        Quick Install
        <a href="#quick-install" onClick={(e) => scrollToId(e, 'quick-install')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</a>
      </h2>
      <p className="text-slate-400 mb-4">The easiest way to install Chauffeur is via our installer script. This will detect your distribution, install dependencies (Nginx, PHP, dnsmasq), and configure the environment.</p>
      <CodeBlock code="curl -sL chauffeur.dev/get | bash" />
      <p className="text-slate-400 mt-4">Once completed, the installer will ask you to restart your session to add the binary to your path.</p>
    </section>

    <section id="manual-install">
      <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="manual-install">
        Manual Installation (Arch Linux)
        <a href="#manual-install" onClick={(e) => scrollToId(e, 'manual-install')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</a>
      </h2>
      <p className="text-slate-400 mb-4">For Arch users, you can install Chauffeur directly from the AUR.</p>
      <CodeBlock code="yay -S chauffeur-bin" language="bash" />
    </section>

    <section id="verify">
      <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="verify">
        Verification
        <a href="#verify" onClick={(e) => scrollToId(e, 'verify')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</a>
      </h2>
      <p className="text-slate-400 mb-4">Verify the installation by checking the version:</p>
      <CodeBlock code="$ chauf --version\nChauffeur v0.1.0-beta (linux/amd64)" />
    </section>
  </div>
);

const FirstProjectContent = () => (
  <div className="space-y-8 animate-fade-in">
     <div>
      <h1 className="text-4xl font-bold text-white mb-4">Your First Project</h1>
      <p className="text-lg text-slate-400 leading-relaxed">
        Let's get a simple PHP project running with SSL and a custom domain in under a minute.
      </p>
    </div>

    <section id="create-project">
      <h2 className="text-2xl font-bold text-white mb-4" id="create-project">1. Create a Project</h2>
      <p className="text-slate-400 mb-4">Create a directory for your new project and add an index.php file.</p>
      <CodeBlock code={`mkdir my-website\ncd my-website\necho "<?php phpinfo();" > index.php`} />
    </section>

    <section id="link-project">
      <h2 className="text-2xl font-bold text-white mb-4" id="link-project">2. Link the Project</h2>
      <p className="text-slate-400 mb-4">Run the link command to tell Chauffeur to serve this directory.</p>
      <CodeBlock code="chauf link" />
      <p className="text-slate-400">
        By default, this will create a site at <code>http://my-website.test</code>.
      </p>
    </section>

     <section id="add-ssl">
      <h2 className="text-2xl font-bold text-white mb-4" id="add-ssl">3. Add SSL</h2>
      <p className="text-slate-400 mb-4">Secure your local site with a trusted certificate.</p>
      <CodeBlock code="chauf secure" />
      <p className="text-slate-400">
        Now visit <a href="#" onClick={(e) => e.preventDefault()} className="text-primary hover:underline">https://my-website.test</a> in your browser.
      </p>
    </section>
  </div>
);

const CommandsContent = () => (
  <div className="space-y-8 animate-fade-in">
     <div>
      <h1 className="text-4xl font-bold text-white mb-4">Command Reference</h1>
      <p className="text-lg text-slate-400 leading-relaxed">
        Comprehensive guide to the Chauffeur CLI.
      </p>
    </div>

    <section id="link">
      <h2 className="text-2xl font-bold text-white mb-4 font-mono" id="link">chauf link</h2>
      <p className="text-slate-400 mb-4">Registers the current working directory as a Chauffeur site.</p>
      
      <h3 className="text-lg font-semibold text-slate-200 mb-2">Usage</h3>
      <CodeBlock code="chauf link [name] [flags]" />

      <h3 className="text-lg font-semibold text-slate-200 mb-2">Flags</h3>
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm border-collapse">
          <thead>
            <tr className="border-b border-slate-700 bg-slate-800/50">
              <th className="p-3 font-mono text-emerald-400">Flag</th>
              <th className="p-3 text-slate-300">Description</th>
              <th className="p-3 text-slate-300">Default</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            <tr>
              <td className="p-3 font-mono text-slate-300">--secure</td>
              <td className="p-3 text-slate-400">Generate SSL certificate</td>
              <td className="p-3 text-slate-500">false</td>
            </tr>
            <tr>
               <td className="p-3 font-mono text-slate-300">--php</td>
               <td className="p-3 text-slate-400">Specify PHP version (e.g. 8.1)</td>
               <td className="p-3 text-slate-500">global</td>
            </tr>
             <tr>
               <td className="p-3 font-mono text-slate-300">--force</td>
               <td className="p-3 text-slate-400">Overwrite existing config</td>
               <td className="p-3 text-slate-500">false</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section id="isolate">
       <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12" id="isolate">chauf isolate</h2>
       <p className="text-slate-400 mb-4">Isolates the current site to a specific PHP version, separate from the global default.</p>
       <CodeBlock code="chauf isolate 7.4" />
    </section>
  </div>
);

export default DocsPage;