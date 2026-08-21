import { createFileRoute } from "@tanstack/react-router";
import * as React from "react";
import { Activity, CircleAlert, Filter } from "lucide-react";
import { AppShell } from "@/components/app-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { activityPageFixtures } from "@/data/webui-fixtures";

export const Route = createFileRoute("/activity")({ component: ActivityPage });

function ActivityPage() {
  const [filter, setFilter] = React.useState<"all" | "success" | "warning">(
    "all",
  );
  const filteredEvents = activityPageFixtures.filter(
    (event) => filter === "all" || event.tone === filter,
  );
  return (
    <AppShell
      title="Recent activity"
      breadcrumbs={[
        { label: "Workspace", to: "/" },
        { label: "Recent activity" },
      ]}
      contentId="activity-content"
      skipLabel="Skip to recent activity"
    >
      <div className="initialized-page activity-page" id="activity-content">
        <section
          className="initialized-page-hero"
          aria-labelledby="activity-title"
        >
          <div>
            <p className="section-kicker">Workspace signal</p>
            <div className="initialized-page-title">
              <span className="initialized-page-icon">
                <Activity aria-hidden="true" />
              </span>
              <h2 id="activity-title">Recent activity</h2>
            </div>
            <p>
              Review project, service, runtime, and backup events in one
              timeline.
            </p>
          </div>
          <Badge variant="outline">Disconnected feed</Badge>
        </section>
        <section className="activity-toolbar" aria-label="Activity filters">
          <Filter aria-hidden="true" />
          {(["all", "success", "warning"] as const).map((value) => (
            <Button
              key={value}
              variant={filter === value ? "secondary" : "ghost"}
              size="sm"
              aria-pressed={filter === value}
              onClick={() => setFilter(value)}
            >
              {value === "all"
                ? "All activity"
                : value === "success"
                  ? "Healthy"
                  : "Attention"}
            </Button>
          ))}
          <Badge variant="outline">
            {filteredEvents.length} fixture events
          </Badge>
          <span>Live updates require backend support.</span>
        </section>
        <section
          className="activity-timeline"
          aria-labelledby="activity-timeline-title"
          aria-live="polite"
        >
          <div className="initialized-section-heading">
            <div>
              <p className="section-kicker">Timeline</p>
              <h3 id="activity-timeline-title">Workspace events</h3>
            </div>
            <Badge variant="outline">Static data</Badge>
          </div>
          {filteredEvents.map((event) => {
            const Icon = event.icon;
            return (
              <article
                className="activity-timeline-row"
                key={`${event.title}-${event.time}`}
              >
                <span className={`activity-timeline-icon ${event.tone}`}>
                  <Icon aria-hidden="true" />
                </span>
                <div>
                  <strong>{event.title}</strong>
                  <p>{event.detail}</p>
                </div>
                <time>{event.time}</time>
              </article>
            );
          })}
          {!filteredEvents.length ? (
            <div className="dashboard-empty-state">
              <h3>No events in this filter</h3>
              <p>Choose another fixture state to continue.</p>
            </div>
          ) : null}
        </section>
        <section className="initialized-actions">
          <div>
            <p className="section-kicker">Unavailable state</p>
            <h3>
              <CircleAlert aria-hidden="true" /> Activity streaming is
              disconnected
            </h3>
            <p>
              No stream is simulated. Retry and live subscription controls will
              require backend support.
            </p>
          </div>
          <Button variant="outline" disabled>
            Reconnect <Badge variant="outline">Requires backend</Badge>
          </Button>
        </section>
      </div>
    </AppShell>
  );
}
