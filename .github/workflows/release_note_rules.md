# Release Note Generation Rules

This document outlines the rules for generating release notes from git commit logs.

## 1. Source Data

- **Command**: `git log <start_commit>..HEAD --pretty=format:"%h - %s (%an)" --stat`
- **Range**: From the last release tag (or specified commit) to the current HEAD.

## 2. Categorization Rules

Commits should be grouped based on their **Conventional Commits** type prefix:

| Type       | Header in Release Notes | Description                                     |
| ---------- | ----------------------- | ----------------------------------------------- |
| `feat`     | 🚀 Features             | New features or significant additions.          |
| `fix`      | 🐛 Bug Fixes            | Bug fixes, error handling, and corrections.     |
| `refactor` | ♻️ Refactor             | Code restructuring without behavioral changes.  |
| `perf`     | ⚡ Performance          | Performance improvements.                       |
| `docs`     | 📚 Documentation        | Documentation only changes.                     |
| `chore`    | 🔧 Chore                | Build process, dependency updates, assets, etc. |

_Note: If a commit message does not strictly follow Conventional Commits (e.g., "debug tools"), infer the category based on the content (e.g., "Features")._

## 3. Formatting Standards

- **File Format**: Markdown (`.md`)
- **Structure**:

  ```markdown
  # Release Notes

  ## [Emoji] [Category Name]

  - **[Scope/Summary]**: [Detailed Description] ([Commit Hash])
  ```

- **Merging**: Group related commits (e.g., multiple commits for the same feature) into a single bullet point to keep notes concise.
- **Language**: Use English for headers. Use the original language of the commit message for descriptions, or translate to English/Chinese if requested.

## 4. Exclusion Criteria

- Exclude "Merge branch..." commits unless they represent a significant feature merge without squashed children.
- Exclude trivial "version bump" commits.

## 5. Output Location

- **Path**: `.github/workflows/release.md`
