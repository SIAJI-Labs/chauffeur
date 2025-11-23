# Site Build Workflow Triggers

This document explains the triggers for the site build workflow.

## 🚀 Triggers

### 1. **Goreleaser Completion** (Primary - Only Trigger)
```bash
# Push a tag to trigger Goreleaser, which then triggers site build
git tag v0.2.0
git push origin v0.2.0
```
**Flow**: `tag push` → `Goreleaser workflow` → `site build workflow`

### 2. **Manual Trigger** (Testing/Debugging Only)
```bash
# Use GitHub Actions UI with optional version override
# Or use GitHub CLI:
gh workflow run build-site.yml -f version_override=v0.2.1 -f release_url_override=https://github.com/SIAJI-Labs/chauffeur/releases/tag/v0.2.1
```

## 🧪 Testing Scenarios

### Test 1: Release Tag Workflow
```bash
# Create and push a tag (triggers Goreleaser then site build)
git tag v0.2.1
git push origin v0.2.1
```
**Expected**:
1. Goreleaser runs first with tag `v0.2.1`
2. Site build runs after Goreleaser completes
3. Site shows version `v0.2.1` with clickable release link

### Test 2: Manual Override
```bash
# Manual trigger with specific version (for testing)
gh workflow run build-site.yml -f version_override=v0.2.2
```
**Expected**:
- Site builds with manual version `v0.2.2`
- Uses auto-generated release URL: `https://github.com/SIAJI-Labs/chauffeur/releases/tag/v0.2.2`

### Test 3: Site Changes Without Release
```bash
# Make site changes and push to main
echo "// Documentation update" >> sites/app/page.tsx
git add sites/app/page.tsx
git commit -m "Update documentation"
git push origin main
```
**Expected**: ❌ **No site build** (prevents spamming wrong version)

## ⚙️ Workflow Logic

The simplified workflow follows clean logic:

1. **Goreleaser completion**: Only builds after successful releases
2. **Success condition**: Only runs if Goreleaser succeeded (`conclusion == 'success'`)
3. **Version extraction**: Gets version directly from the triggering tag
4. **Manual override**: Optional testing/debugging with version override

## 📊 Build Artifact

The workflow creates an artifact named: `site-build-{commit-sha}`

Download with:
```bash
gh run download <run-id> -n site-build-<commit-sha>
```

## 🔍 Debugging

Check workflow run details in GitHub Actions:
- Look for "pre-check" job to see trigger analysis
- Check "build-site" job for build results
- Verify artifact contains correct version

## 🚀 Deployment

The workflow outputs deployment-ready information:
- Build location: `sites/out/`
- Version displayed: From tag or latest release
- Release URL: Links to GitHub release page