# PHP PostgreSQL Extension Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add PostgreSQL PHP extensions (`pgsql` and `pdo_pgsql`) to the standard PHP build so installed versions can connect to Postgres databases.

**Architecture:** This is a narrow installer-and-docs change. The PHP source build already owns the extension list, so the implementation should add the two PostgreSQL configure flags there and then update the project’s PHP support docs so the supported extension set stays accurate. No runtime behavior or UI changes are needed unless later verification reveals an installer summary that also hardcodes the list.

**Tech Stack:** Go, PHP source builds, Markdown docs

---

### Task 1: Update the PHP configure flags

**Files:**
- Modify: `internal/installers/php.go:223-253`

**Step 1: Write the failing test**

Add or update a unit test in the PHP installer test suite that asserts the configure args include PostgreSQL support, specifically `--with-pgsql` and `--with-pdo-pgsql`.

```go
func TestConfigureArgsIncludesPostgresExtensions(t *testing.T) {
    // Arrange a PHPInstaller for a supported version.
    // Act: call configureArgs().
    // Assert: the returned args contain both PostgreSQL flags.
}
```

**Step 2: Run the test to verify it fails**

Run: `go test ./internal/installers -run TestConfigureArgsIncludesPostgresExtensions -v`

Expected: FAIL because the PostgreSQL flags are not present yet.

**Step 3: Write the minimal implementation**

Add the following configure flags next to the existing MySQL flags in `configureArgs()`:

```go
"--with-pgsql",
"--with-pdo-pgsql",
```

**Step 4: Run the test to verify it passes**

Run: `go test ./internal/installers -run TestConfigureArgsIncludesPostgresExtensions -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/installers/php.go internal/installers/php_test.go
git commit -m "feat: add PostgreSQL PHP extension support"
```

---

### Task 2: Update PHP integration docs

**Files:**
- Modify: `.agent/integrations/php.md:22-50`

**Step 1: Write the failing test**

Treat the documentation as the contract: create a small docs check or manual checklist that confirms the build configuration snippet shows PostgreSQL support.

If the repo already has docs validation, add a test/assertion for the configure block; otherwise, record the expectation in the implementation notes and verify by diff review.

**Step 2: Run the verification**

Run the relevant docs or snapshot check if available; otherwise, inspect the rendered markdown diff after editing.

Expected: the PHP build configuration section does not yet mention PostgreSQL extensions.

**Step 3: Write the minimal documentation update**

Add PostgreSQL to the configuration snippet in `.agent/integrations/php.md`:

```bash
  --with-pgsql \
  --with-pdo-pgsql \
```

**Step 4: Verify the doc update**

Re-read the file or run the docs check again.

Expected: the build config snippet now lists PostgreSQL support.

**Step 5: Commit**

```bash
git add .agent/integrations/php.md
git commit -m "docs: document PostgreSQL PHP extensions"
```

---

### Task 3: Update tech stack conventions

**Files:**
- Modify: `.agent/conventions/tech-stack.md:60-74`

**Step 1: Write the failing test**

Add a lightweight documentation expectation in the form of a review checklist: the extension list under PHP must include PostgreSQL support.

**Step 2: Run the verification**

Open the file and confirm the current list only mentions MySQL-related DB extensions.

Expected: the list is outdated for PostgreSQL support.

**Step 3: Write the minimal documentation update**

Update the database extension bullet to include PostgreSQL:

```md
- Database: `mysqli`, `pdo_mysql`, `pdo_pgsql`, `pgsql`, `mysqlnd`
```

**Step 4: Verify the update**

Re-read the section to confirm the supported extensions are accurate.

**Step 5: Commit**

```bash
git add .agent/conventions/tech-stack.md
git commit -m "docs: update PHP database extension list"
```

---

### Task 4: Run verification across code and docs

**Files:**
- No code changes; verification only

**Step 1: Run the installer test suite**

Run: `go test ./internal/installers ./...`

Expected: all tests pass.

**Step 2: Review the doc diffs**

Confirm the PostgreSQL support appears in both PHP docs and the tech-stack conventions.

Expected: no leftover references that say PHP only ships MySQL database extensions.

**Step 3: Commit any final cleanup**

```bash
git status --short
git add -A
git commit -m "chore: finalize PostgreSQL PHP support docs"
```
