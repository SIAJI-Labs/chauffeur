import * as React from "react";
import { createFileRoute } from "@tanstack/react-router";
import { Check, Type } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/settings")({ component: SettingsPage });

const fontSizes = [
  { value: "xs", label: "XS", detail: "Compact labels and controls" },
  { value: "sm", label: "SM", detail: "Slightly denser workspace" },
  { value: "m", label: "M", detail: "Balanced default" },
  { value: "lg", label: "LG", detail: "More comfortable reading" },
  { value: "xl", label: "XL", detail: "Largest workspace text" },
] as const;

function SettingsPage() {
  const [fontSize, setFontSize] = React.useState<(typeof fontSizes)[number]["value"]>("m");
  const [hydrated, setHydrated] = React.useState(false);
  React.useEffect(() => {
    const stored = window.localStorage.getItem("chauffeur-font-size") as typeof fontSizes[number]["value"] | null;
    if (stored && fontSizes.some((option) => option.value === stored)) setFontSize(stored);
    setHydrated(true);
  }, []);
  React.useEffect(() => {
    if (!hydrated) return;
    document.documentElement.dataset.fontSize = fontSize;
    window.localStorage.setItem("chauffeur-font-size", fontSize);
  }, [fontSize, hydrated]);
  return (
    <AppShell title="Settings" breadcrumbs={[{ label: "Workspace", to: "/" }, { label: "System", to: "/system" }, { label: "Settings" }]} contentId="settings-content" skipLabel="Skip to settings">
      <div className="settings-page" id="settings-content">
        <section className="settings-hero"><div><p className="section-kicker">System preferences</p><h2>Make the workspace comfortable to read.</h2><p>Font size is the first local preference. It changes the UI without sending data anywhere.</p></div><Badge variant="outline">Local setting</Badge></section>
        <section className="settings-card" aria-labelledby="font-size-title"><div className="settings-card-heading"><span className="planned-destination-icon"><Type aria-hidden="true" /></span><div><p className="section-kicker">Display</p><h3 id="font-size-title">Font size</h3><p>Choose the density used across navigation, cards, tables, and dialogs.</p></div><Badge variant="outline">{fontSize.toUpperCase()}</Badge></div><div className="font-size-options" role="radiogroup" aria-label="Font size"><span className="sr-only" aria-live="polite">Font size set to {fontSize.toUpperCase()}</span>{fontSizes.map((option) => <Button key={option.value} variant={fontSize === option.value ? "default" : "outline"} className="font-size-option" role="radio" aria-checked={fontSize === option.value} onClick={() => setFontSize(option.value)}><span className={`font-size-sample font-size-${option.value}`}>Aa</span><span><strong>{option.label}{option.value === "m" ? " (default)" : ""}</strong><small>{option.detail}</small></span>{fontSize === option.value ? <Check aria-hidden="true" /> : null}</Button>)}</div></section>
        <p className="settings-footnote">This preference is stored locally in this browser. Other settings remain planned until their supporting runtime controls exist.</p>
      </div>
    </AppShell>
  );
}
