import * as React from "react";
import { Link } from "@tanstack/react-router";
import { Command, Search } from "lucide-react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { commandPaletteEntries } from "@/data/webui-fixtures";

export function CommandPaletteDialog() {
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");
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
  const entries = commandPaletteEntries.filter((entry) => `${entry.label} ${entry.detail}`.toLowerCase().includes(query.toLowerCase()));
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="command-palette-dialog">
        <DialogHeader><DialogTitle><Command aria-hidden="true" /> Command palette</DialogTitle><DialogDescription>Navigate the workspace without implying unsupported actions.</DialogDescription></DialogHeader>
        <div className="command-palette-input"><Search aria-hidden="true" /><Input autoFocus value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search pages and actions" aria-label="Search command palette" /></div>
        <div className="command-palette-results" role="listbox" aria-label="Command palette results">
          {entries.map((entry) => { const Icon = entry.icon; return entry.group === "Pages" ? <Link key={entry.label} to="/projects" className="command-palette-result" onClick={() => setOpen(false)} role="option"><Icon aria-hidden="true" /><span><strong>{entry.label}</strong><small>{entry.detail}</small></span><kbd>{entry.shortcut}</kbd></Link> : <Link key={entry.label} to="/preview/$slug" params={{ slug: "command-palette" }} className="command-palette-result" onClick={() => setOpen(false)} role="option"><Icon aria-hidden="true" /><span><strong>{entry.label}</strong><small>{entry.detail}</small></span><kbd>{entry.shortcut}</kbd></Link>; })}
          {!entries.length ? <p className="command-palette-empty">No matching workspace destinations.</p> : null}
        </div>
      </DialogContent>
    </Dialog>
  );
}
