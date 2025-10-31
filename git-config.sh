# Always normalize to LF in the repo; convert to CRLF only when checking out on Windows
git config core.autocrlf input        # (Linux/macOS)

# Preserve executable bits correctly (important for scripts)
git config core.fileMode true

# Prevent case-only renames (Windows can't distinguish them)
git config core.ignorecase false

# Make line ending warnings visible
git config core.safecrlf warn

# Prevent Git from choking on long paths on Windows
git config core.longpaths true
