# Guide for Coding Agents

This file gives coding agents (Claude Code, Codex, Copilot, Cursor, automated PR bots, etc.) the rules and context they need to ship good changes to **go-dav-os**.

If you are a human, you do not need to read this — start with [`README.md`](README.md) and [`CONTRIBUTING.md`](CONTRIBUTING.md). The same expectations apply to agent-authored PRs.

## Project shape

go-dav-os is a 64-bit freestanding hobby kernel written in Go (`gccgo`, x86_64 long mode), booted via GRUB/Multiboot2. Kernel-only — bootloader and BIOS are delegated to existing tools. Top-level packages: `boot/`, `kernel/`, `terminal/`, `keyboard/`, `mem/`, `fs/`, `drivers/ata`, `fs/fat16`, `shell/`, `user/`. Build via `make`/`make iso`/`make run`. CI runs `gofmt -l .`, `go vet -tags testing -unsafeptr=false ./...`, and `make test` plus `python3 scripts/test_boot.py`.

## Hard rules for agents

1. **Stay minimal.** This is a hobby kernel. Don't refactor adjacent code, don't add abstractions you don't need, don't introduce dependencies. The maintainer's stated bar is "minimal but usable". Three repeated lines beat a premature helper.
2. **One concern per PR.** No "while I'm here I cleaned up X". Tangential cleanups belong in their own PR.
3. **Run the gates the human reviewer runs.** Before opening a PR, confirm:
   - `gofmt -l .` prints nothing
   - `go vet -tags testing -unsafeptr=false ./...` is clean
   - `make test` passes
   - `python3 scripts/test_boot.py` passes (the QEMU integration suite)
4. **Don't touch `boot/boot.s`, the Multiboot2 header layout, or paging/IDT setup unless the issue is explicitly about them.** Those are load-bearing.
5. **No standard library imports in kernel code.** This is a freestanding build (`gccgo`, no stdlib). Stick to packages already imported in the file you're editing.
6. **Comments explain *why*, not *what*.** Identifier names already say what; only comment when the reasoning would otherwise be lost.
7. **Disclose AI authorship in the PR body** when the change was substantially agent-written. One line is enough: `AI was used for assistance.` Do not advertise the engine name or describe verification — keep it short.

## Recommended workflow

1. Read the issue all the way through. If it lists numbered acceptance criteria, your PR must close every one of them or you should not open the PR yet.
2. Open the relevant file before changing it. Match the existing style of that file (indentation, comment density, error-handling shape).
3. Keep diffs small. Aim for under 100 added lines on a first contribution.
4. Run the full gate stack locally. If a gate fails, fix it locally — do not push and let CI fail.
5. Follow the [PR template](.github/pull_request_template.md). Three lines (what / why / how to test) is the minimum, not a suggestion.

## Things that get auto-rejected

- PRs that fail `gofmt` or `go vet`.
- PRs that bundle unrelated cleanups with the actual fix.
- PRs that import new third-party Go modules into kernel code.
- PRs that add `TODO`/`FIXME` comments without an issue link.
- PRs whose body is generic AI-style boilerplate ("This PR adds robust support for...") instead of a concrete what/why/how-to-test.

## Review

Every PR — agent-authored or human-authored — is judged against [`REVIEW.md`](REVIEW.md). Read it before opening the PR; if your change can't pass that checklist, rework it before pushing.
