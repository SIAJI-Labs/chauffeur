"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import { ChevronRight } from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function FirstProjectPage() {
  const currentSlug = 'getting-started/first-project';

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
        <span className="text-slate-200 capitalize">{currentSlug.split('/').pop()?.replace('-', ' ')}</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Your First Project</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Let's get a simple PHP project running with SSL and a custom domain in under a minute.
          </p>
        </div>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="create-project">
            1. Create a Project
            <Link href="#create-project" onClick={(e) => scrollToId(e, 'create-project')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Create a directory for your new project and add an index.php file.</p>
          <CodeBlock code={`mkdir my-website\ncd my-website\necho "<?php phpinfo();" > index.php`} />
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="link-project">
            2. Link the Project
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
            3. Add SSL
            <Link href="#add-ssl" onClick={(e) => scrollToId(e, 'add-ssl')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Secure your local site with a trusted certificate.</p>
          <CodeBlock code="chauf secure" />
          <p className="text-slate-400">
            Now visit <Link href="#" onClick={(e) => e.preventDefault()} className="text-primary hover:underline">https://my-website.test</Link> in your browser.
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