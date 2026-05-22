---
name: git-push
description: "Git version control push workflow. Use when the user wants to stage, commit, or push code changes to GitHub, or when running git-push."
user-invocable: true
license: MIT
compatibility: Designed for Antigravity, Claude Code, and other Agent Skills compatible runners.
metadata:
  author: local
  version: "1.0.0"
  openclaw:
    emoji: "🚀"
    homepage: https://github.com/local/git-push
    requires:
      bins:
        - git
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(git:*) Agent
---

Follow these steps to safely push your current work to GitHub:

1. **Check Status**: Run `git status` to review which files have been modified.
2. **Stage Changes**:
   - To stage everything: `git add .`
   - To stage specific files: `git add <file_path>`
3. **Commit**:
   - Propose a clear, concise commit message following Conventional Commits (e.g., `feat: ...`, `fix: ...`, `chore: ...`).
   - Run `git commit -m "<your_message>"`
4. **Push**: Run `git push` to upload your commit to the remote repository.
5. **Verify**: Run `git status` again to ensure your branch is up to date with 'origin/main'.
