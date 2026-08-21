import * as React from "react";
import { createFileRoute } from "@tanstack/react-router";
import { Check, Type } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/settings")({ component: SettingsPage });

const fontSizes = [
  { value: "xs", label: "xs", detail: "Compact labels and controls" },
  { value: "sm", label: "sm", detail: "Slightly denser workspace" },
  { value: "m", label: "m", detail: "Balanced default" },
  { value: "lg", label: "lg", detail: "More comfortable reading" },
  { value: "xl", label: "xl", detail: "Largest workspace text" },
] as const;

function SettingsPage() {
  const [fontSize, setFontSize] =
    React.useState<(typeof fontSizes)[number]["value"]>("m");
  const [hydrated, setHydrated] = React.useState(false);
  React.useEffect(() => {
    const stored = window.localStorage.getItem("chauffeur-font-size") as
      | (typeof fontSizes)[number]["value"]
      | null;
    if (stored && fontSizes.some((option) => option.value === stored))
      setFontSize(stored);
    setHydrated(true);
  }, []);
  React.useEffect(() => {
    if (!hydrated) return;
    document.documentElement.dataset.fontSize = fontSize;
    window.localStorage.setItem("chauffeur-font-size", fontSize);
  }, [fontSize, hydrated]);
  const selectFontSize = (index: number) => {
    const option = fontSizes[index];
    setFontSize(option.value);
    requestAnimationFrame(() =>
      document.getElementById(`font-size-${option.value}`)?.focus(),
    );
  };
  return (
    <AppShell
      title="Settings"
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "System", to: "/system" },
        { label: "Settings" },
      ]}
      contentId="settings-content"
      skipLabel="Skip to settings"
    >
      <div className="settings-page" id="settings-content">
        <section className="settings-hero">
          <div>
            <p className="section-kicker">System preferences</p>
            <h2>Make the workspace comfortable to read.</h2>
            <p>
              Font size is the first local preference. It changes the UI without
              sending data anywhere.
            </p>
          </div>
          <Badge variant="outline">Local setting</Badge>
        </section>
        <section className="settings-card" aria-labelledby="font-size-title">
          <div className="settings-card-heading">
            <span className="planned-destination-icon">
              <Type aria-hidden="true" />
            </span>
            <div>
              <p className="section-kicker">Display</p>
              <h3 id="font-size-title">Font size</h3>
              <p>
                Choose the density used across navigation, cards, tables, and
                dialogs.
              </p>
            </div>
            <Badge variant="outline">{fontSize.toUpperCase()}</Badge>
          </div>
          <div
            className="font-size-options"
            role="radiogroup"
            aria-label="Font size"
          >
            <span className="sr-only" aria-live="polite">
              Font size set to {fontSize.toUpperCase()}
            </span>
            {fontSizes.map((option, index) => (
              <Button
                key={option.value}
                id={`font-size-${option.value}`}
                variant={fontSize === option.value ? "default" : "outline"}
                className="font-size-option"
                role="radio"
                aria-checked={fontSize === option.value}
                tabIndex={fontSize === option.value ? 0 : -1}
                onKeyDown={(event) => {
                  if (event.key === "ArrowRight" || event.key === "ArrowDown") {
                    event.preventDefault();
                    selectFontSize((index + 1) % fontSizes.length);
                  } else if (
                    event.key === "ArrowLeft" ||
                    event.key === "ArrowUp"
                  ) {
                    event.preventDefault();
                    selectFontSize(
                      (index - 1 + fontSizes.length) % fontSizes.length,
                    );
                  } else if (event.key === "Home") {
                    event.preventDefault();
                    selectFontSize(0);
                  } else if (event.key === "End") {
                    event.preventDefault();
                    selectFontSize(fontSizes.length - 1);
                  }
                }}
                onClick={() => setFontSize(option.value)}
              >
                <span className={`font-size-sample font-size-${option.value}`}>
                  Aa
                </span>
                <span>
                  <strong>
                    {option.label}
                    {option.value === "m" ? " (default)" : ""}
                  </strong>
                  <small>{option.detail}</small>
                </span>
                {fontSize === option.value ? (
                  <Check aria-hidden="true" />
                ) : null}
              </Button>
            ))}
          </div>
        </section>
        <p className="settings-footnote">
          This preference is stored locally in this browser. Other settings
          remain planned until their supporting runtime controls exist.
        </p>
      </div>
    </AppShell>
  );
}
