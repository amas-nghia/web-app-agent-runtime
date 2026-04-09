# Automation

## Issue bot

The issue bot runs on a 15-minute schedule in GitHub Actions (`.github/workflows/issuebot.yml`).
It checks open issues, auto-fixes only recognized safe patterns, runs tests, and closes issues after a successful fix.

### Configuration

- `--repo-root`: repository root
- `--label`: optional issue label filter
- `--once`: run a single cycle and exit
- `--dry-run`: apply fixes but do not commit/push
- `--interval`: polling interval (default 15m)
- `--max-issues`: maximum issues to inspect
- `--auto-close`: close the issue after a successful fix

### Safety

The bot only auto-fixes recognized issue patterns. Unknown issues are skipped.

### Workflow

1. list open issues
2. classify known safe patterns
3. apply a targeted fix
4. run `go test ./...`
5. commit/push the change
6. comment and close the issue
