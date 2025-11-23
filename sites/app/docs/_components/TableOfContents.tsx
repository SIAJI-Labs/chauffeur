"use client";

// React & Next.js
import React, { useEffect, useState } from 'react';
import { usePathname } from 'next/navigation';

interface TOCItem {
  id: string;
  text: string;
  level: number;
}

export const TableOfContents: React.FC = () => {
  const pathname = usePathname();
  const [headings, setHeadings] = useState<TOCItem[]>([]);
  const [activeId, setActiveId] = useState<string>('');

  useEffect(() => {
    // Find the main content area (the prose div)
    const contentElement = document.querySelector('.prose.prose-invert');
    if (!contentElement) return;

    // Find all H2 and H3 elements only within the main content area
    const elements = Array.from(contentElement.querySelectorAll('h2, h3'));
    const items = elements.map((elem, index) => {
      // Handle empty IDs by creating a unique one
      let id = elem.id;
      if (!id) {
        id = `heading-${index}-${elem.textContent?.slice(0, 20).replace(/\s+/g, '-').toLowerCase() || 'unknown'}`;
        elem.id = id; // Set the ID on the element for future reference
      }
      return {
        id,
        text: elem.textContent?.replace('#', '').trim() || '', // Remove the anchor symbol text if present
        level: Number(elem.tagName.charAt(1)),
      };
    });

    // Remove duplicates by keeping only the first occurrence of each ID
    const uniqueItems = items.filter((item, index, self) =>
      index === self.findIndex((t) => t.id === item.id)
    );

    setHeadings(uniqueItems);

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
          }
        });
      },
      { rootMargin: '-100px 0px -66% 0px' }
    );

    // Only observe elements that have unique IDs
    uniqueItems.forEach((item) => {
      const elem = document.getElementById(item.id);
      if (elem) observer.observe(elem);
    });

    return () => observer.disconnect();
  }, [pathname]); // Re-run when path changes (not just hash)

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    e.preventDefault();
    const element = document.getElementById(id);
    if (element) {
      const yOffset = -100; // Navbar height offset
      const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset;
      window.scrollTo({ top: y, behavior: 'smooth' });
      // We avoid pushing state here to not conflict with HashRouter navigation
      setActiveId(id);
    }
  };

  if (headings.length === 0) return null;

  return (
    <div className="hidden xl:block w-64 fixed right-8 top-24 h-[calc(100vh-6rem)] overflow-y-auto custom-scrollbar">
      <h4 className="text-sm font-semibold text-slate-200 mb-4 pl-4">On this page</h4>
      <nav className="relative">
        <div className="absolute left-0 top-0 bottom-0 w-px bg-slate-800" />
        <ul className="space-y-3 text-sm">
          {headings.map((heading, index) => (
            <li key={heading.id || `toc-item-${index}`} className={`pl-4 border-l-2 transition-colors ${
              activeId === heading.id
                ? 'border-emerald-500 text-emerald-400'
                : 'border-transparent text-slate-500 hover:text-slate-300'
            }`}>
              <a
                href={`#${heading.id || `toc-item-${index}`}`}
                onClick={(e) => handleClick(e, heading.id || `toc-item-${index}`)}
                className="block truncate hover:cursor-pointer"
                style={{ paddingLeft: heading.level === 3 ? '12px' : '0px' }}
              >
                {heading.text}
              </a>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
};