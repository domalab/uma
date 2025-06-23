# UMA Automated Version Management System

This document describes the automated version management system that ensures version consistency across all UMA project files.

## Overview

The UMA project uses an automated version management system that eliminates manual version updates and prevents version inconsistencies that can cause plugin update status issues in Unraid.

## Single Source of Truth

**File**: `VERSION` (project root)
- Contains the authoritative version number in `YYYY.MM.DD` format
- All other version references are automatically synchronized from this file

## Automated Synchronization

The system automatically updates version numbers in:

- `uma.go` - Runtime version variable
- `uma.plg` - Plugin definition version entity
- `package/uma/VERSION` - Web interface version file
- `package/uma/uma.page` - PHP fallback version
- `package/create-plugin-package.sh` - Package script version
- `package/test-plugin-installation.sh` - Test script version

## Usage

### Setting a New Version

```bash
# Set new version and sync all files
make version-set VERSION=2025.06.24

# Or use the script directly
./scripts/update-version.sh set 2025.06.24
```

### Synchronizing Existing Version

```bash
# Sync current version across all files
make version-sync

# Or use the script directly
./scripts/update-version.sh sync
```

### Verifying Version Consistency

```bash
# Check if all files have consistent versions
make version-verify

# Or use the script directly
./scripts/update-version.sh verify
```

### Getting Current Version

```bash
# Display current version
make version-current

# Or use the script directly
./scripts/update-version.sh current
```

## Build Integration

The version management is integrated into the build process:

```bash
# Local build (automatically syncs version)
make local

# Release build (automatically syncs version)
make release
```

## Release Automation

Create GitHub releases with consistent versioning:

```bash
# Create complete GitHub release
./scripts/release.sh create

# Verify existing release
./scripts/release.sh verify
```

## Version Format

- **Format**: `YYYY.MM.DD` (e.g., `2025.06.23`)
- **No Prefixes**: No "v" prefix to avoid inconsistencies
- **Consistent**: Same format across all files and GitHub releases

## Benefits

1. **Single Update**: Change version in one place only
2. **Consistency**: Eliminates version mismatches
3. **Automation**: No manual file editing required
4. **Validation**: Built-in consistency checks
5. **Unraid Compatibility**: Fixes plugin update status issues

## Troubleshooting

### Version Inconsistency Errors

If you see version inconsistency errors:

```bash
# Fix automatically
make version-sync
```

### Plugin Update Status Issues

If Unraid shows "Unknown" plugin status:

1. Ensure all versions are consistent: `make version-verify`
2. Rebuild plugin package: `cd package && ./create-plugin-package.sh`
3. Upload to GitHub release with correct version format

### Manual Version Updates

**❌ Don't do this**: Manually edit version numbers in individual files

**✅ Do this**: Use the automated system:
```bash
make version-set VERSION=2025.06.24
```

## Files Managed

| File | Purpose | Auto-Updated |
|------|---------|--------------|
| `VERSION` | Single source of truth | ✅ Manual |
| `uma.go` | Runtime version | ✅ Auto |
| `uma.plg` | Plugin definition | ✅ Auto |
| `package/uma/VERSION` | Web interface | ✅ Auto |
| `package/uma/uma.page` | PHP fallback | ✅ Auto |
| `package/create-plugin-package.sh` | Package script | ✅ Auto |
| `package/test-plugin-installation.sh` | Test script | ✅ Auto |

## Integration with CI/CD

The version management system integrates with:

- **Makefile**: Build targets automatically sync versions
- **Package Creation**: Reads version from single source
- **GitHub Releases**: Consistent version format
- **Unraid Plugin System**: Proper update status detection
