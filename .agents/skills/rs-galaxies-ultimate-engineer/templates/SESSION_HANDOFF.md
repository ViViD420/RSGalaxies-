## <Milestone Title> (<YYYY-MM-DD>; <verified in CI / verified locally / pending remote CI>)

- **Workspace & repository state:**
  - Local branch / commit: `<branch>` @ `<sha>`
  - Remote `origin/main`: `<sha>`
- **Files changed:**
  - `<path>`: <concise summary of what changed and why>
- **Defects discovered and fixed (no-hidden-errors log):**
  - `<test or file>`: <exact failure output> -> <root cause> -> <fix and regression test>
- **Verification evidence:**
  - Server unit & PostgreSQL integration tests: `<exact command and pass summary>`
  - `go vet ./...`: `<result>`
  - Godot 4.3 headless smoke tests, editor import, and exports: `<exact command and zero-error log confirmation>`
  - GitHub Actions runs (if pushed): `<workflow run IDs / URLs / status>`
- **Remaining / unverified items & next exact task:**
  - <explicit list of manual device QA or next engineering task>
