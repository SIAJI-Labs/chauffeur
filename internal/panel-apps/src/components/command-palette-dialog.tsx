import * as React from "react";
import { Link } from "@tanstack/react-router";
import { Command, Search } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { commandPaletteEntries } from "@/data/webui-fixtures";

export function CommandPaletteDialog() {
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");
  const [activeIndex, setActiveIndex] = React.useState(0);
  const resultRefs = React.useRef<Array<HTMLAnchorElement | null>>([]);
  React.useEffect(() => {
    const openPalette = () => setOpen(true);
    window.addEventListener("chauffeur:open-command-palette", openPalette);
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setOpen(true);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("chauffeur:open-command-palette", openPalette);
      window.removeEventListener("keydown", onKeyDown);
    };
  }, []);
  const entries = commandPaletteEntries.filter((entry) =>
    `${entry.label} ${entry.detail}`
      .toLowerCase()
      .includes(query.toLowerCase()),
  );
  const navigableEntries = entries.filter((entry) => entry.to);
  React.useEffect(() => {
    setActiveIndex(0);
    resultRefs.current = [];
  }, [open, query]);
  const moveSelection = (delta: number) => {
    if (!navigableEntries.length) return;
    const nextIndex =
      (activeIndex + delta + navigableEntries.length) % navigableEntries.length;
    setActiveIndex(nextIndex);
    requestAnimationFrame(() => resultRefs.current[nextIndex]?.focus());
  };
  const handleInputKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      moveSelection(1);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      moveSelection(-1);
    } else if (event.key === "Enter") {
      event.preventDefault();
      resultRefs.current[activeIndex]?.click();
    }
  };
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="command-palette-dialog">
        <DialogHeader>
          <DialogTitle>
            <Command aria-hidden="true" /> Command palette
          </DialogTitle>
          <DialogDescription>
            Navigate the workspace without implying unsupported actions.
          </DialogDescription>
        </DialogHeader>
        <div className="command-palette-input">
          <Search aria-hidden="true" />
          <Input
            autoFocus
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            onKeyDown={handleInputKeyDown}
            placeholder="Search pages and actions"
            aria-label="Search command palette"
            aria-activedescendant={
              navigableEntries[activeIndex]
                ? `command-result-${activeIndex}`
                : undefined
            }
          />
        </div>
        <div
          className="command-palette-results"
          role="listbox"
          aria-label="Command palette results"
        >
          {entries.map((entry) => {
            const Icon = entry.icon;
            const navigableIndex = entry.to
              ? navigableEntries.findIndex((item) => item.label === entry.label)
              : -1;
            const unsupported =
              entry.group === "Actions" || entry.group === "Toggles";
            return unsupported ? (
              <div
                key={entry.label}
                className="command-palette-result command-palette-result-disabled"
                aria-disabled="true"
                role="option"
              >
                <Icon aria-hidden="true" />
                <span>
                  <strong>{entry.label}</strong>
                  <small>{entry.detail}</small>
                  <em>Requires backend</em>
                </span>
                <kbd>{entry.shortcut}</kbd>
              </div>
            ) : entry.to ? (
              <Link
                key={entry.label}
                to={entry.to}
                id={`command-result-${navigableIndex}`}
                className="command-palette-result"
                onClick={() => setOpen(false)}
                onKeyDown={(event) => {
                  if (event.key === "ArrowDown") {
                    event.preventDefault();
                    moveSelection(1);
                  } else if (event.key === "ArrowUp") {
                    event.preventDefault();
                    moveSelection(-1);
                  } else if (event.key === "Escape") {
                    event.preventDefault();
                    setOpen(false);
                  }
                }}
                role="option"
                aria-selected={navigableIndex === activeIndex}
                tabIndex={navigableIndex === activeIndex ? 0 : -1}
                ref={(element) => {
                  resultRefs.current[navigableIndex] = element;
                }}
              >
                <Icon aria-hidden="true" />
                <span>
                  <strong>{entry.label}</strong>
                  <small>{entry.detail}</small>
                </span>
                <kbd>{entry.shortcut}</kbd>
              </Link>
            ) : null;
          })}
          {!entries.length ? (
            <p className="command-palette-empty">
              No matching workspace destinations.
            </p>
          ) : null}
        </div>
      </DialogContent>
    </Dialog>
  );
}
