# AI Agent Instructions for Beads

## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

## Project Overview

**beads** (command: `bd`) is a Git-backed issue tracker designed for AI-supervised coding workflows.
We use it to track ALL work in this repository.

**Key Features:**
- **Dependency-aware**: Issues can block other issues
- **Git-synced**: Issues are stored in `.beads/issues.jsonl` and versioned with code
- **AI-friendly**: JSON output, clear structure, simple CLI
- **Lightweight**: No external servers, just a local SQLite DB and a JSONL file

## Core Workflow

1. **Find work**:
   - `bd ready --json` - Show unblocked, high-priority issues
   - `bd list --status open --priority 1 --json` - Show all high-priority open issues

2. **Claim a task**:
   - `bd update <id> --status in_progress --json`

3. **Do the work**:
   - Write code, tests, docs
   - If you discover new necessary work, create a linked issue:
     `bd create "New task" --deps discovered-from:<parent-id> --json`

4. **Complete the task**:
   - `bd close <id> --reason "Done" --json`

5. **Sync**:
   - `bd sync` - Flushes changes to `.beads/issues.jsonl`
   - **CRITICAL**: Always commit `.beads/issues.jsonl` along with your code changes!

## Command Reference

Always use `--json` for programmatic interaction to get structured data.

### Reading
- `bd list --json` - List all issues
- `bd show <id> --json` - Show details for an issue
- `bd ready --json` - Show issues ready to work on (not blocked)
- `bd comments <id> --json` - Show comments

### Writing
- `bd create "Title" --body "Description" --priority 0-4 --type bug|feature|task --json`
- `bd update <id> --status in_progress|done|blocked --json`
- `bd close <id> --reason "Fixed" --json`
- `bd comment <id> "Message" --json`

### Dependencies
- `bd create "Child task" --deps <parent-id> --json` - Create a task blocked by parent
- `bd update <id> --deps <blocker-id> --json` - Add a dependency (id is blocked by blocker-id)
- `bd update <id> --deps -<blocker-id> --json` - Remove a dependency

## Best Practices

### 1. Granularity
- Break large tasks into smaller issues
- Use dependencies to order them
- Example: "Implement Feature X" (blocked by) -> "Design Feature X"

### 2. Context
- Add comments to issues with `bd comment` to store findings/decisions
- When stopping work, leave a comment explaining current state

### 3. Discovery
- If you find a bug or refactoring opportunity while working, DO NOT fix it immediately if it's unrelated.
- Create a new issue: `bd create "Fix bug in X" --priority 2 --deps discovered-from:<current-id>`
- Continue with your current task

### 4. Commit Messages
- Reference issue IDs in commit messages: `Fixes #123` or `Updates #123`

## Troubleshooting

- **Database locked?** Run `bd info` to check status.
- **Out of sync?** Run `bd sync` to force synchronization.
- **Weird state?** Run `bd doctor` to check health.

## Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

## Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`
6. **Commit together**: Always commit the `.beads/issues.jsonl` file together with the code changes so issue state stays in sync with code state

## Auto-Sync

bd automatically syncs with git:
- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

## MCP Server (Recommended)

If using Claude or MCP-compatible clients, install the beads MCP server:

```bash
pip install beads-mcp
```

Add to MCP config (e.g., `~/.config/claude/config.json`):
```json
{
  "beads": {
    "command": "beads-mcp",
    "args": []
  }
}
```

Then use `mcp__beads__*` functions instead of CLI commands.

If you are an AI assistant with MCP capabilities, you can use the `beads-mcp` server for native function calling.
This provides tools like `mcp__beads__create_issue`, `mcp__beads__list_issues`, etc.
Check if `mcp__beads__*` tools are available in your context.

## Managing AI-Generated Planning Documents

AI assistants often create planning and design documents during development:
- PLAN.md, IMPLEMENTATION.md, ARCHITECTURE.md
- DESIGN.md, CODEBASE_SUMMARY.md, INTEGRATION_PLAN.md
- TESTING_GUIDE.md, TECHNICAL_DESIGN.md, and similar files

**Best Practice: Use a dedicated directory for these ephemeral files**

**Recommended approach:**
- Create a `history/` directory in the project root
- Store ALL AI-generated planning/design docs in `history/`
- Keep the repository root clean and focused on permanent project files
- Only access `history/` when explicitly asked to review past planning

**Example .gitignore entry (optional):**
```
# AI planning documents (ephemeral)
history/
```

**Benefits:**
- ✅ Clean repository root
- ✅ Clear separation between ephemeral and permanent documentation
- ✅ Easy to exclude from version control if desired
- ✅ Preserves planning history for archeological research
- ✅ Reduces noise when browsing the project

## Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ✅ Store AI planning docs in `history/` directory
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems
- ❌ Do NOT clutter repo root with planning documents
