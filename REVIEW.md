# PR Review Checklist

This file gives reviewers (and contributors) a lightweight quality gate to run on every PR — human or automated. Tick each item before approving or merging.

If a PR can't tick the relevant items, it is either re-worked or closed.

## Build and tests

- [ ] `gofmt -l .` is empty (zero unformatted files)
- [ ] `go vet -tags testing -unsafeptr=false ./...` reports no issues
- [ ] `make test` passes locally
- [ ] `python3 scripts/test_boot.py` passes (QEMU integration suite)
- [ ] CI on the PR is green

## Scope and shape

- [ ] PR closes a specific issue or has a clear scope statement (no "drive-by" changes)
- [ ] Diff is local — does not touch unrelated files
- [ ] Diff size is justified by the change — no incidental refactors
- [ ] No new third-party Go modules added to kernel code
- [ ] No new standard-library imports in freestanding kernel packages

## Code quality

- [ ] Identifier names explain what each function/variable does
- [ ] Comments explain *why*, not *what*
- [ ] Error handling matches the surrounding file (no new patterns introduced silently)
- [ ] No `TODO` / `FIXME` markers without a linked issue
- [ ] `boot/boot.s`, Multiboot2 header, paging, and IDT setup are untouched (unless the PR is explicitly about them)

## Tests for behavior changes

- [ ] If the PR changes behavior, there is a test that fails on `main` and passes on the branch
- [ ] If the PR changes shell / boot output, the relevant `test_boot.py` case is updated
- [ ] If the PR changes a public API in a Go package, the package's `*_test.go` is updated

## PR description

- [ ] PR body has the three sections from [`pull_request_template.md`](.github/pull_request_template.md): **what**, **why**, **how to test**
- [ ] "How to test" has a concrete command, not a description
- [ ] If AI was used substantially, the PR body includes the disclosure line `AI was used for assistance.`
- [ ] If the PR was opened from a coding agent, the agent followed [`AGENTS.md`](AGENTS.md)

## Final gate

- [ ] Reviewer can answer "what did this PR change?" in one sentence
- [ ] Reviewer can answer "what would break if I revert this PR?" in one sentence

If both answers are easy, the PR is ready to merge.
