<!-- Reviewer's checklist lives in REVIEW.md. AGENTS.md covers expectations for AI-authored PRs. -->

## What

<!-- One or two lines: what does this change? -->

## Why

<!-- One or two lines: why does it need to change? Link to the issue if there is one. -->

Closes #

## How to test

<!-- A concrete command (or sequence) the reviewer can run to verify the change. Example:
- `make test`
- `python3 scripts/test_boot.py --functional`
- Boot in QEMU and run `<command>` in the shell, expect `<output>`.
-->

## Pre-flight

- [ ] `gofmt -l .` prints nothing
- [ ] `go vet -tags testing -unsafeptr=false ./...` is clean
- [ ] `make test` passes
- [ ] `python3 scripts/test_boot.py` passes
- [ ] PR is a single concern (no drive-by cleanups)

<!-- If AI was used substantially in this PR, leave one line below: "AI was used for assistance." -->
