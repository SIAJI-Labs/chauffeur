import { useEffect, useState } from "react";

function readHashTab<T extends string>(tabs: ReadonlyArray<T>, fallback: T): T {
  if (typeof window === "undefined") return fallback;

  const candidate = window.location.hash.slice(1) as T;
  return tabs.includes(candidate) ? candidate : fallback;
}

export function useHashTab<T extends string>(
  tabs: ReadonlyArray<T>,
  fallback: T,
) {
  const [activeTab, setActiveTab] = useState(() => readHashTab(tabs, fallback));

  useEffect(() => {
    const handleHashChange = () => {
      setActiveTab(readHashTab(tabs, fallback));
    };

    window.addEventListener("hashchange", handleHashChange);
    window.addEventListener("popstate", handleHashChange);
    return () => {
      window.removeEventListener("hashchange", handleHashChange);
      window.removeEventListener("popstate", handleHashChange);
    };
  }, [fallback, tabs]);

  const selectTab = (tab: T) => {
    setActiveTab(tab);

    const url = new URL(window.location.href);
    url.hash = tab === fallback ? "" : tab;
    window.history.pushState({}, "", url);
  };

  return [activeTab, selectTab] as const;
}
