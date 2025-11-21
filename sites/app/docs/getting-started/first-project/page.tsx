"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import { ChevronRight } from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

// Import command constants
import { findCommand } from '@/constants';

export default function FirstProjectPage() {
  const currentSlug = 'getting-started/first-project';

  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  // Helper to get command information
  const getCommandInfo = (commandName: string) => {
    const command = findCommand(commandName);
    if (command && command.examples.length > 0) {
      return command.examples[0];
    }
    return null;
  };

  // Get command examples
  const installCmd = getCommandInfo('chauf install');
  const installNginxCmd = findCommand('chauf install nginx');
  const installComposerCmd = findCommand('chauf install composer');
  const installPHPCmd = getCommandInfo('chauf install php');
  const linkCmd = getCommandInfo('chauf link');
  const linkSSLCmd = getCommandInfo('chauf link');
  const startCmd = getCommandInfo('chauf start');
  const statusCmd = getCommandInfo('chauf status');
  const phpInstallCmd = findCommand('chauf php install');

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
        <span className="text-slate-200 capitalize">{currentSlug.split('/').pop()?.replace('-', ' ')}</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Your First Project</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Let's get Chauffeur set up and running a simple PHP project with SSL and a custom domain in just a few minutes.
          </p>
        </div>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="install-services">
            1. Install Services
            <Link href="#install-services" onClick={(e) => scrollToId(e, 'install-services')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Install the core Chauffeur services needed to run your projects. This is a one-time setup.</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Install Core Services</h3>
          <p className="text-slate-400 mb-4">Install each service individually. This is a one-time setup.</p>

          <div className="space-y-4">
            <div>
              <h4 className="text-lg font-semibold text-slate-300 mb-2">Install nginx</h4>
              <CodeBlock code={`$ ${installNginxCmd?.command || 'chauf install nginx'}

📦 Installing nginx...
✓ nginx installed successfully

${installNginxCmd?.examples?.[0]?.output || '✓ nginx installed successfully'}`} />
            </div>

            <div>
              <h4 className="text-lg font-semibold text-slate-300 mb-2">Install PHP</h4>
              <CodeBlock code={`$ ${installPHPCmd?.command || 'chauf install php 8.3'}

📦 Installing PHP 8.3...
✓ PHP 8.3 installed successfully
📦 Setting as global version...
✓ PHP 8.3 set as global default

${installPHPCmd?.output || '✓ PHP 8.3 installed successfully'}`} />
            </div>

            <div>
              <h4 className="text-lg font-semibold text-slate-300 mb-2">Install composer</h4>
              <CodeBlock code={`$ ${installComposerCmd?.command || 'chauf install composer'}

📦 Installing composer...
✓ composer installed successfully

${installComposerCmd?.examples?.[1]?.output || '✓ composer installed successfully'}`} />
            </div>
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Install Specific PHP Version</h3>
          <p className="text-slate-400 mb-4">You can also install specific PHP versions if needed:</p>
          <CodeBlock code={`$ ${installPHPCmd?.command || 'chauf install php 8.2'}

📦 Installing PHP 8.2...
✓ PHP 8.2 installed successfully
📦 Setting as global version...
✓ PHP 8.2 set as global default

${installPHPCmd?.output || '✓ PHP 8.2 installed successfully'}`} />

          <div className="mt-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <h4 className="font-semibold text-blue-300 mb-2">📋 What Gets Installed</h4>
            <div className="text-sm text-blue-100/80 space-y-1">
              <p>• <strong>nginx:</strong> Web server for serving your sites</p>
              <p>• <strong>PHP 8.3:</strong> Default PHP runtime (latest stable)</p>
              <p>• <strong>composer:</strong> PHP dependency manager</p>
              <p>• <strong>dnsmasq:</strong> DNS resolution for .test domains</p>
            </div>
          </div>

          <div className="mt-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <h4 className="font-semibold text-blue-300 mb-2">📋 Available Services</h4>
            <div className="text-sm text-blue-100/80 space-y-1">
              <p><strong>Services:</strong> nginx, php, composer</p>
              <p><strong>Usage:</strong> <code className="bg-slate-800 px-2 py-1 rounded text-blue-300">chauf install &lt;service&gt;</code></p>
              <p><strong>Example:</strong> <code className="bg-slate-800 px-2 py-1 rounded text-blue-300">chauf install nginx</code></p>
            </div>
          </div>

          <div className="mt-4 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <h4 className="font-semibold text-amber-300 mb-2">⚡ Installation Options</h4>
            <div className="text-sm text-amber-100/80">
              <p>Use <code className="bg-slate-800 px-2 py-1 rounded text-amber-300">chauf install --help</code> to see all available options:</p>
              <ul className="mt-2 space-y-1">
                <li>• <code>--force</code>: Reinstall if already installed</li>
                <li>• <code>--local</code>: Use local packages instead of downloading</li>
                <li>• <code>--no-cache</code>: Skip package cache for fresh install</li>
              </ul>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="create-project">
            2. Create a Project
            <Link href="#create-project" onClick={(e) => scrollToId(e, 'create-project')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Create a directory for your new project and add an index.php file.</p>
          <CodeBlock code={`mkdir my-website\ncd my-website\necho "<?php phpinfo();" > index.php`} />
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="link-project">
            3. Link the Project
            <Link href="#link-project" onClick={(e) => scrollToId(e, 'link-project')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Run the link command to tell Chauffeur to serve this directory.</p>
          <CodeBlock code="chauf link" />
          <p className="text-slate-400">
            By default, this will create a site at <code>http://my-website.test</code>.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="add-ssl">
            4. Add SSL
            <Link href="#add-ssl" onClick={(e) => scrollToId(e, 'add-ssl')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Secure your local site with SSL certificate.</p>
          <CodeBlock code="chauf link --ssl" />
          <p className="text-slate-400">
            Now visit <Link href="#" onClick={(e) => e.preventDefault()} className="text-primary hover:underline">https://my-website.test</Link> in your browser.
          </p>

          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <h4 className="font-semibold text-amber-300 mb-2">🔐 SSL Certificate Options</h4>
            <div className="text-sm text-amber-100/80 space-y-2">
              <p>
                <strong>Option 1 - Self-signed (default):</strong> Automatic without sudo.
                Browser shows "Not Secure" warning but connection is encrypted.
              </p>
              <p>
                <strong>Option 2 - Trusted certificates:</strong> Install mkcert for browser-trusted certificates.
              </p>
              <div className="bg-slate-900/50 p-3 rounded mt-2">
                <code className="text-amber-300"># Install mkcert for trusted certificates</code><br/>
                <code>go install -r filippo.io/mkcert@latest</code><br/>
                <code>mkcert -install  # Prompts for sudo password once</code><br/>
                <code># Then relink your project with SSL</code><br/>
                <code>chauf link --ssl --force</code>
              </div>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="start-services">
            5. Start Services
            <Link href="#start-services" onClick={(e) => scrollToId(e, 'start-services')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Start Chauffeur services to serve your project. This needs to be done once after installation or system restart.</p>
          <CodeBlock code={`chauf start

🚀 Starting Chauffeur services...
✓ dnsmasq started (DNS resolution enabled)
✓ nginx started (port 80, 443)
✓ PHP FPM pool started

🌐 All services running successfully!`} />

          <div className="mt-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <h4 className="font-semibold text-blue-300 mb-2">💡 Service Status</h4>
            <p className="text-sm text-blue-100/80">
              You can check if services are running with <code className="bg-slate-800 px-2 py-1 rounded text-blue-300">chauf status</code>.
              Services will automatically start on system boot after the first installation.
            </p>
          </div>

          <div className="mt-4 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <h4 className="font-semibold text-amber-300 mb-2">⚠️ First Time Setup</h4>
            <p className="text-sm text-amber-100/80">
              If this is your first time using Chauffeur, you may need to install PHP first:
              <code className="bg-slate-800 px-2 py-1 rounded text-amber-300">chauf php install 8.3</code>
            </p>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="visit-site">
            6. Visit Your Site
            <Link href="#visit-site" onClick={(e) => scrollToId(e, 'visit-site')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Open your browser and visit your new site with SSL:</p>
          <CodeBlock code="🌐 https://my-website.test" />

          <p className="text-slate-400 mt-4">
            You should see the PHP info page, indicating your site is working perfectly with HTTPS!
          </p>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
           <div className="text-left">
             <div className="text-xs text-slate-500 mb-1">Previous</div>
             <Link href="/docs/getting-started/installation" className="text-primary hover:underline">Installation</Link>
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