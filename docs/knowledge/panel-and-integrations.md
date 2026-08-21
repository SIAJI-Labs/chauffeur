# Panel And Integrations

## Panel process

`chauf webui start` starts a local Go HTTP server. By default it runs in the background, records a PID in `panel.pid`, writes output to `panel.log`, prevents duplicate instances, and handles graceful termination with a kill fallback. `-f` keeps it foregrounded. `--dev` runs the API and Vite together from a source checkout, enables hot module replacement at `http://localhost:5173`, and skips embedded asset builds. Frontend development does not require `chauf self-update`; use that command only to update or reinstall the Chauffeur binary.

The default address is `http://panel.test:3083`, while the server itself binds to loopback and does not create DNS for `panel.test`. This needs to be made explicit or implemented consistently.

## Backend API

The embedded server exposes:

```text
GET    /api/health
GET    /api/containers
GET    /api/containers/{name}
POST   /api/containers/{name}/start
POST   /api/containers/{name}/stop
GET    /api/containers/{name}/logs
GET    /api/containers/{name}/databases
GET    /api/backups
POST   /api/backups
DELETE /api/backups/{name}
POST   /api/backups/{name}/restore
GET    /
GET    /assets/
```

The panel primarily manages Podman databases rather than the complete Chauffeur project/service lifecycle.

## Frontend

The React application uses React 19, TanStack Router and Query, Tailwind CSS v4, shadcn-style components, Lucide/Hugeicons, Vite, TypeScript, Vitest, ESLint, and Prettier.

Current routes include the dashboard, container list, container detail, backup, and docs pages. The container list polls every five seconds and supports start/stop actions. The dashboard contains hard-coded zero/unknown values. Sidebar and docs navigation contain placeholder `#` links for logs, sites, config, DNS, SSL, and settings.

## Database persistence

Podman configuration is stored as YAML under `podman/`; persistent bind-mounted data is under `podman/volumes/<name>/`; backups are under `podman/backups/`. Supported container engines have engine-specific commands for console, backup, and restore.

## Integration behavior

- Podman is optional and separate from the PHP/nginx runtime.
- mkcert is optional until SSL is requested.
- dnsmasq is optional but needed for convenient wildcard `.test` resolution.
- systemd is optional for autostart.
- Network port forwarding is optional when users accept 8080/8443 URLs.

## Security and reliability issues to resolve

- Database passwords are stored in plain text in YAML.
- Container detail responses return passwords.
- The panel has no authentication and assumes trusted local users.
- SSE sets wildcard CORS even though the panel is local.
- Backup identifiers and filesystem paths need strict basename/path validation.
- Synchronous backup requests can block for large databases.
- The logs SSE endpoint returns one snapshot rather than a continuous stream.
- Lifecycle actions need idempotent, structured responses for concurrent requests.

## Better panel direction

Choose one product boundary:

1. **Focused database panel:** make container, connection, backup, restore, and logs workflows excellent; remove misleading Chauffeur-wide navigation.
2. **Full control plane:** add project URLs, PHP versions, SSL, DNS/doctor status, service controls, configuration, and project logs with real API-backed data.

Do not retain a broad navigation shell with placeholder pages. It makes the product feel broken and obscures the useful functionality that already exists.
