---
name: git-push
description: "Commit changes with Conventional Commits, push them on a new branch, and open a GitHub pull request. Use when the user asks to stage, commit, push, create a PR, or run git-push."
user-invocable: true
license: MIT
compatibility: Designed for Antigravity, Claude Code, and other Agent Skills compatible runners.
metadata:
  author: local
  version: "1.1.1"
  openclaw:
    emoji: "🚀"
    homepage: https://github.com/local/git-push
    requires:
      bins:
        - git
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(git:*) Bash(gh:*) Agent
---

Safely turn the requested local changes into a reviewable GitHub pull request.

1. Inspect `git status`, the relevant diff, the current branch, remotes, and the
   remote default branch. Preserve unrelated user changes and never stage them
   merely because they are present.
2. Choose a short branch name derived from the change, such as
   `fix/gemini-auth-key-validation`. Create and switch to that new branch before
   committing so the commit is not placed on the default branch. If the name
   already exists, choose a clear non-conflicting variant; do not overwrite it.
3. Stage only files belonging to the requested change. Review the staged diff
   and stop if it contains credentials, generated user data, or unrelated work.
4. Create one clear Conventional Commit based on the staged diff, for example
   `fix: accept Gemini authorization keys`. Do not amend or rewrite existing
   commits unless the user explicitly requests it.
5. Push the new branch and establish its upstream with
   `git push -u origin <branch>`. Never force-push unless explicitly requested.
6. If GitHub CLI is available and authenticated, create the pull request with
   `gh pr create`, targeting the remote default branch. Use a concise title and
   a body that summarizes the change and its verification. Return the PR URL.
7. If PR creation is unavailable or fails, the pushed branch is still a useful
   result. Return a GitHub compare URL that opens the PR form for the pushed
   branch, using the repository URL, base branch, and head branch discovered
   from Git rather than guessing them. Also provide a ready-to-paste PR title
   and Markdown body. The title should describe the committed change concisely.
   The body should summarize the change and list only verification that actually
   ran; do not invent test results. Present both in clearly labeled code blocks.
8. Finish by reporting the branch, commit, push result, PR URL or fallback URL,
   fallback PR title and body when applicable, and `git status --short`. Do not
   claim that the default branch is up to date merely because the feature branch
   was pushed.

An explicit request to run `git-push` authorizes this complete commit, new-branch
push, and pull-request workflow. If the user asks for only one part, such as a
local commit, stop at that boundary. Automatic skill selection alone does not
authorize an external mutation.
