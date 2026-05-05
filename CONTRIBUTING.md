# Contributing

Thanks for even considering contributing 🙌

This is a hobby project and it’s still in a very early stage. The main idea is simple: keep things **minimal but usable**. I’m trying to avoid overcomplicating and overthinking, so small and focused contributions are the best ones.

If you’re not sure where to start, open a Discussion and we’ll align quickly.

## What you can contribute
Anything that helps the project move forward, especially:

- Docs (README improvements, troubleshooting, small diagrams, screenshots/gifs)
- Small features around the shell/monitor (new commands, better help output, small UX)
- Cleanup/refactors (only if they keep the code simpler to read)
- Small kernel building blocks (tiny drivers/utilities, not huge redesigns)
- CI improvements (build checks, automation)

## Setup / how to run
Please follow the README for the full setup (Docker/native toolchain, QEMU, etc.).


## Testing

```bash
# Run all unit tests
make test

# Or run tests for specific packages
go test ./fs ./kernel/scheduler ./kernel/gdt ./kernel/tss

# Run integration tests (requires QEMU)
python3 scripts/test_boot.py

# Iterate on one class of behavior:
python3 scripts/test_boot.py --functional      # only the shell suite
python3 scripts/test_boot.py --fault-probes    # only the kread/kwrite/kpriv probes
```

All tests must pass in CI. The `make test` command automatically discovers and runs all packages containing `*_test.go` files.

## How to open a PR
The workflow is the classic one:

1. Fork the repo
2. Create a branch (`feat/...`, `fix/...`, `docs/...` — nothing strict here)
3. Do the change
4. Make sure it still builds and runs (same commands as in the README)
5. Open the PR

In the PR description, please write 2–3 lines:
- what you changed
- why
- how I can test it (what command to run / what output to expect)

That’s enough.

## A couple of guidelines (to keep things simple)
- Prefer **small PRs** (one thing at a time)
- Try to keep changes **local** (avoid “while I’m here I refactor the whole repo”)
- Readability > cleverness
- Run `gofmt` on Go code (see below)
- Comments are welcome when they explain **why** (not what)

## Code formatting and vet

The CI enforces formatting and runs `go vet` on every PR. Run these locally before pushing to avoid CI failures.

**Fix formatting:**
```bash
gofmt -w .
```

To check which files need formatting without modifying them:
```bash
gofmt -l .
```

**Run go vet:**
```bash
go vet -tags testing -unsafeptr=false ./...

```

Fix any other issues reported before opening a PR. `go vet` catches common mistakes like incorrect format strings, unreachable code, and misuse of sync primitives.

If you want to do a bigger change (design/architecture), open a Discussion first so we don’t waste time.

## Where to start
If you want something easy, check Issues labeled:
- `good first issue`
- `help wanted`

If there aren’t any yet, open a Discussion and tell me what you’d like to work on — I’ll point you to a good first task.

## Bugs / feature ideas
- For bugs: open an Issue with expected vs actual behavior, steps to reproduce, and logs/screenshots if you have them.
- For feature ideas: feel free to open a Discussion (usually better than an Issue at the beginning).

## One last thing
Be kind and respectful. This is a learning project and the goal is to build cool stuff together 👊
