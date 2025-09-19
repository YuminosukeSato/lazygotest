# Homebrew Distribution Setup

A simple guide to distribute `lazygotest` via Homebrew.

## Prerequisites

1. **GitHub Personal Access Token (PAT)**
   - Go to [GitHub Settings](https://github.com/settings/tokens) → Personal access tokens → Tokens (classic)
   - Click "Generate new token"
   - Check the `repo` scope
   - Save the token (you'll need it later)

## Step 1: Create Homebrew Tap Repository

1. Create a new repository on GitHub:
   - Repository name: `homebrew-tap`
   - Make it public
   - No README needed

2. Initialize locally:
```bash
mkdir -p ~/repos
cd ~/repos
git clone https://github.com/YuminosukeSato/homebrew-tap.git
cd homebrew-tap
mkdir Formula
echo "# Homebrew Tap for YuminosukeSato tools" > README.md
git add .
git commit -m "Initial commit"
git push origin main
```

## Step 2: Configure GitHub Secrets

1. Go to your `lazygotest` repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add:
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Secret: Your PAT from earlier

## Step 3: Update GitHub Actions Workflow

Uncomment the last line in `.github/workflows/release.yml`:

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}  # ← Uncomment this
```

## Step 4: Create First Release

```bash
# Commit your changes
git add .
git commit -m "Setup Homebrew distribution"
git push

# Create and push a tag
git tag -a v0.1.0 -m "Initial release with Homebrew support"
git push origin v0.1.0
```

GitHub Actions will automatically:
- Build binaries for all platforms
- Create a GitHub Release
- Update the Homebrew Formula in your tap

## Step 5: Test Installation

After the release completes (wait a few minutes):

```bash
# Add your tap
brew tap YuminosukeSato/tap

# Install
brew install lazygotest

# Verify
lazygotest --version
```

## Troubleshooting

### If GitHub Actions Fails
- Check the Actions tab for error logs
- Verify PAT has `repo` scope
- Ensure tap repository exists

### If Homebrew Installation Fails
- Run `brew update`
- Reset tap: `brew untap YuminosukeSato/tap && brew tap YuminosukeSato/tap`

## Future Releases

To release new versions:

```bash
git tag -a v0.2.0 -m "Release notes here"
git push origin v0.2.0
```

The Homebrew formula will update automatically.