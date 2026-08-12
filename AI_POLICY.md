# AI Contribution Policy

AI tools are optional in go-dav-os. Contributions are judged by the same
technical and review standards whether or not AI was used.

This is a lightweight policy for a hobby project. It establishes human
responsibility, useful disclosure, and safe handling of project data without
adding a separate approval process for AI-assisted work.

## Human responsibility

The human contributor remains responsible for every submitted change. Before
submitting AI-assisted work, the contributor must:

- understand the change well enough to explain and maintain it;
- review the complete diff, including generated code, tests, and documentation;
- verify factual claims and run the checks required by
  [CONTRIBUTING.md](CONTRIBUTING.md);
- confirm that the contribution is correctly licensed and does not include
  secrets, personal data, private code, or other material that may not be
  shared.

Do not submit bulk-generated, unreviewed, or poorly understood changes.
AI-generated output and AI review comments, including automated reviewers such
as CodeRabbit, are suggestions rather than evidence that a change is correct.

## Disclosure

Every commit must include the human contributor's `Signed-off-by` trailer as
described in [CONTRIBUTING.md](CONTRIBUTING.md). This identifies the human who
accepts responsibility under the Developer Certificate of Origin; never put an
AI tool in that trailer.

Add an `AI-Assisted-by` trailer when AI assistance materially affected the
content of a commit. Assistance is material when generated or transformed
content remains in the commit, or when AI analysis materially shaped the
implementation, tests, documentation, or design. Routine completion, spelling
help, search, formatting, or an unused suggestion does not require the trailer
when it does not materially affect the committed content.

Use a product or service name, not a person:

```text
docs: clarify contribution workflow

Signed-off-by: Human Name <human@example.com>
AI-Assisted-by: AI tool or service
```

When AI assistance was not material:

```text
docs: fix broken link

Signed-off-by: Human Name <human@example.com>
```

Do not use `Co-authored-by` for an AI tool. The human contributor is the
author and accountable party. The existing sign-off CI verifies
`Signed-off-by`; `AI-Assisted-by` is disclosure metadata and does not create
a separate CI approval gate.

When a change is substantially agent-written, also retain the short disclosure
required by the pull request template: `AI was used for assistance.`

## Review

Review AI-assisted contributions against [REVIEW.md](REVIEW.md), using the same
scope, testing, and quality requirements as any other contribution. Reviewers
should verify the submitted change itself rather than relying on the tool's
explanation or an automated review result.

This policy is informed by the project discussion in
[#191](https://github.com/dmarro89/go-dav-os/issues/191), including the proposal
to keep process metadata with commits.
