# Agent Protocol

This repository uses an automated coding agent to fulfill issues labeled
`agent-task` or with titles beginning with `[Agent]`.

The agent operates inside a Git repository and is expected to produce
reviewable code changes.

---

## Core Contract

1. **Source of Truth**
   - Treat `ISSUE.md` as the single source of truth for requirements and acceptance criteria.

2. **Plan First**
   - Generate a concrete execution plan.
   - The plan MUST list the exact files to be created or modified.

3. **Mandatory Code Changes**
   - You MUST modify or create at least one source code file.
   - You MUST produce a non-empty `git diff`.
   - If no code changes are produced, the task is considered FAILED.

4. **Deterministic Execution**
   - Implement only what is required by the issue.
   - Prefer minimal, incremental, and reviewable changes.

5. **Verification**
   - Run existing tests or linters when available.
   - If no tests exist, state this explicitly.

6. **PR Preparation**
   - Summarize what changed and why.
   - Link the originating issue.
   - Call out risks, assumptions, or follow-up work.

---

## Safety and Scope Rules

- Do NOT perform speculative refactors.
- Do NOT change unrelated files.
- Prefer clarity over cleverness.
- Keep changes isolated and reversible.

---

## Failure & Escalation Rules

- If the issue does not specify enough information to identify files to change,
  you MUST FAIL rather than silently exiting.
- If requirements are contradictory, FAIL and request clarification in the issue.
- If required secrets (e.g. `OPENAI_API_KEY`) are missing, FAIL fast with a clear error.
- If tests are flaky or unreliable, note this explicitly in the PR body.

---

## Definition of Done

An issue is considered complete ONLY IF:
- At least one source file is modified or created.
- A non-empty `git diff` exists.
- The resulting changes can be reviewed via a Pull Request.
