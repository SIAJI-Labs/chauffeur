import * as React from "react";
import { Link, createFileRoute } from "@tanstack/react-router";
import { ArrowLeft, Check, Command, Keyboard } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { previewDestinationFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/preview/$slug")({ component: PreviewDestinationPage });

function PreviewDestinationPage() {
  const { slug } = Route.useParams();
  const fixture = previewDestinationFixtures[slug as keyof typeof previewDestinationFixtures];
  const [showActionPreview, setShowActionPreview] = React.useState(false);

  if (!Object.prototype.hasOwnProperty.call(previewDestinationFixtures, slug)) {
    return (
      <AppShell title="Preview not found" breadcrumbs={[{ label: "Workspace", to: "/" }, { label: "Preview" }]} contentId="preview-content" skipLabel="Skip to preview">
        <div className="dashboard-empty-state" id="preview-content"><h2>Preview not found</h2><p>The requested planned destination does not exist.</p><Button render={<Link to="/" />}>Return to overview</Button></div>
      </AppShell>
    );
  }
  const selectedFixture = fixture;

  const Icon = selectedFixture.icon;
  return (
    <AppShell title={selectedFixture.title} breadcrumbs={[{ label: "Workspace", to: "/" }, { label: selectedFixture.title }]} contentId="preview-content" skipLabel={`Skip to ${selectedFixture.title}`} shortcuts={<Button variant="outline" size="sm" className="header-button" render={<Link to="/" />}><ArrowLeft aria-hidden="true" /><span>Overview</span></Button>}>
      <div className="planned-destination-page" id="preview-content">
        <section className="planned-destination-hero" aria-labelledby="preview-title">
          <div>
            <p className="section-kicker">{selectedFixture.eyebrow}</p>
            <div className="planned-destination-title"><span className="planned-destination-icon"><Icon aria-hidden="true" /></span><h2 id="preview-title">{selectedFixture.title}</h2></div>
            <p>{selectedFixture.description}</p>
          </div>
          <Badge variant="outline">{selectedFixture.status}</Badge>
        </section>
        <div className="planned-destination-grid">
          {selectedFixture.sections.map((section) => <Card key={`${section.label}-${section.value}`}><CardHeader><div className="planned-card-label"><span className="planned-state-dot" aria-hidden="true" />{section.label}</div><CardTitle>{section.value}</CardTitle></CardHeader><CardContent><p>{section.detail}</p></CardContent></Card>)}
        </div>
        <section className="planned-destination-notice" aria-live="polite">
          <div><Keyboard aria-hidden="true" /><div><strong>Preview only</strong><p>This surface is static. No files, processes, services, or settings are changed.</p></div></div>
          <Button variant="outline" onClick={() => setShowActionPreview(true)}><Command aria-hidden="true" />Preview an action</Button>
        </section>
      </div>
      <Dialog open={showActionPreview} onOpenChange={setShowActionPreview}>
        <DialogContent><DialogHeader><DialogTitle>Action preview</DialogTitle><DialogDescription>This action is not connected yet. Chauffeur will require backend support before it can make a change.</DialogDescription></DialogHeader><div className="planned-dialog-state"><Check aria-hidden="true" /><span>No changes made</span></div><DialogFooter><Button variant="outline" onClick={() => setShowActionPreview(false)}>Close</Button></DialogFooter></DialogContent>
      </Dialog>
    </AppShell>
  );
}
