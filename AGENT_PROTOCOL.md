# Agent Protocol

This repository uses an automated agent to fulfill issues labeled `agent-task` or with titles beginning with `[Agent]`. The workflow below clarifies how the agent should behave when triggered by GitHub Actions or a local runner.

## Core Contract
1. **Read the issue**: Treat `ISSUE.md` as the single source of truth for requirements and acceptance criteria.
2. **Plan before coding**: Generate a brief execution plan that maps requirements to concrete steps.
3. **Execute deterministically**: Implement the plan with minimal scope creep, preferring incremental, reviewable changes.
4. **Verify**: Run the project’s automated tests or linters. Surface failures explicitly.
5. **Prepare the PR**: Summarize the change, link the originating issue, and highlight remaining risks or follow-ups.

## Safety and Clarity Guidelines
- Avoid speculative refactors unrelated to the issue.
- Document key assumptions in commit messages or PR descriptions.
- Prefer readability over cleverness; future maintainers should understand decisions quickly.
- Keep changes isolated so they are easy to revert if necessary.

## Escalation Rules
- If requirements conflict or are ambiguous, stop and request clarification in the issue before proceeding.
- If a required secret (e.g., `OPENAI_API_KEY`) is unavailable, fail fast with a clear error message.
- When tests are unreliable or flaky, note the limitation in the PR body and include rerun instructions if applicable.
