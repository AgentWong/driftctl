---
applyTo: "**/*.go"
---

# Inline Code Comments for Go

When writing or modifying Go code, include brief inline comments where they
add genuine value for a competent Go developer reading the code for the first
time. Follow the principle: **if the "why" isn't obvious from the code itself,
add a comment; if it is, don't.**

## When to comment

- Non-obvious business logic or domain rules ("Config API returns resources
  across all regions, so we deduplicate by ARN")
- Workarounds, quirks, or SDK/API edge cases ("Route53 returns a trailing dot
  on zone names that Terraform strips")
- The reason behind a particular approach when alternatives seem simpler
  ("We normalize remote resources here because IaC resources are already
  normalized during creation via CreateAbstractResource")
- Concurrency patterns: why a goroutine, channel, or mutex exists
- Non-trivial type assertions or interface satisfaction

## When NOT to comment

- Self-explanatory code: variable declarations, simple assignments, standard
  error handling, struct literals with clear field names
- Restating what the code does ("increment counter", "return error")
- Godoc-style comments on unexported helpers that are only called in one place
- Obvious loop or conditional logic

## Style

- Use `//` comments on the line above or at end-of-line, whichever reads
  more naturally. Prefer the line above for anything longer than a few words.
- Start with a lowercase letter unless it begins a sentence after a blank line.
- Keep comments to one or two lines. If you need more, it may warrant a
  function-level doc comment instead.
- Match the terse, direct tone of the existing codebase.
