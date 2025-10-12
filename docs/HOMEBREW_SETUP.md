# Homebrew Distribution Setup

This guide explains how to distribute `cct` via Homebrew.

## How It Works

The GitHub Actions workflow automatically generates a Homebrew formula when you create a release tag. The formula includes SHA256 checksums for all platform binaries.

## Setup Steps

### 1. Create a Homebrew Tap Repository

Create a new GitHub repository named `homebrew-cct` (must start with `homebrew-`):

```bash
# Create the repository on GitHub, then:
git clone https://github.com/schlunsen/homebrew-cct.git
cd homebrew-cct
mkdir -p Formula
```

### 2. Get the Generated Formula

After creating a release (e.g., `v0.0.1`), the GitHub Actions workflow generates `cct.rb`. You can find it:

1. Go to the Actions tab in your repository
2. Click on the "Build and Release" workflow run
3. Check the "Upload Homebrew formula to tap" step
4. Copy the printed formula content

Or download it from the release artifacts:
```bash
# Download from release
curl -L https://github.com/schlunsen/claude-control-terminal/releases/download/v0.0.1/cct.rb -o Formula/cct.rb
```

### 3. Commit the Formula to Your Tap

```bash
cd homebrew-cct
git add Formula/cct.rb
git commit -m "Add cct formula v0.0.1"
git push origin main
```

### 4. Users Can Now Install

```bash
brew tap schlunsen/cct
brew install cct
```

Or in one line:
```bash
brew install schlunsen/cct/cct
```

## Updating the Formula

Every time you create a new release:

1. Tag and push: `git tag v0.0.2 && git push origin v0.0.2`
2. GitHub Actions builds and generates new `cct.rb`
3. Download the new formula from the release or Actions output
4. Update `Formula/cct.rb` in your `homebrew-cct` repository
5. Commit and push

Users update with:
```bash
brew update
brew upgrade cct
```

## Automated Formula Updates (Optional)

You can automate formula updates by adding a step to push to the tap repository:

```yaml
- name: Update Homebrew tap
  env:
    TAP_GITHUB_TOKEN: ${{ secrets.TAP_GITHUB_TOKEN }}
  run: |
    git clone https://github.com/schlunsen/homebrew-cct.git tap
    cp dist/cct.rb tap/Formula/cct.rb
    cd tap
    git config user.name "GitHub Actions"
    git config user.email "actions@github.com"
    git add Formula/cct.rb
    git commit -m "Update cct to ${{ steps.version.outputs.version }}"
    git push https://${TAP_GITHUB_TOKEN}@github.com/schlunsen/homebrew-cct.git main
```

You'll need to:
1. Create a Personal Access Token with `repo` scope
2. Add it as a repository secret named `TAP_GITHUB_TOKEN`

## Testing Locally

Before publishing, test the formula locally:

```bash
brew install --build-from-source Formula/cct.rb
brew test cct
brew audit --strict cct
```

## Alternative: Official Homebrew Core (Advanced)

To add `cct` to the official Homebrew repository (requires 75+ GitHub stars and notable usage):

1. Meet requirements: https://docs.brew.sh/Acceptable-Formulae
2. Submit PR to https://github.com/Homebrew/homebrew-core
3. Formula must pass all `brew audit` checks

Users would then install with just:
```bash
brew install cct
```

## Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [How to Create a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
