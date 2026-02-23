# AGENTS.md

Guidance for coding agents working in `terraform-provider-wormly`.

## Scope

- Applies to the whole repository.
- Prioritize consistency with existing code over introducing new patterns.
- Keep changes minimal, test-backed, and focused.

## Repository Facts

- Language: Go (`go 1.24.7`)
- Main module: `github.com/radarnex/terraform-provider-wormly`
- Provider framework: HashiCorp Terraform Plugin Framework
- Main packages:
  - `internal/provider` (Terraform provider/resources/data sources)
  - `internal/client` (Wormly API client + API domain logic)
  - `tools` (docs/code generation helpers)

## Rule Files Discovery

- Checked for Cursor rules:
  - `.cursorrules`
  - `.cursor/rules/**`
- Checked for Copilot rules:
  - `.github/copilot-instructions.md`
- Result: none found in this repository.

If these files are later added, treat them as authoritative additions to this guide.

---

## Build, Lint, Test Commands

Use `make` targets first when available to match CI/local conventions.

### Common Commands

- Install dependencies:
  - `go mod download`
- Build all packages:
  - `make build`
  - Equivalent: `go build -v ./...`
- Install provider locally:
  - `make install`
- Format code:
  - `make fmt`
  - Equivalent: `gofmt -s -w -e .`
- Lint:
  - `make lint`
  - Equivalent: `golangci-lint run`
- Generate docs/examples formatting:
  - `make generate`

### Test Commands

- Run all tests:
  - `make test`
  - Equivalent: `go test -v -cover -timeout=120s -parallel=10 ./...`
- Run all tests in one package:
  - `go test -v ./internal/client`
  - `go test -v ./internal/provider`
- Run a single test by name:
  - `go test -v ./internal/client -run '^TestParseHTTPSensorParams$'`
  - `go test -v ./internal/provider -run '^TestHostResource_Metadata$'`
- Run a specific subtest:
  - `go test -v ./internal/client -run 'TestConvertBasicSensorToHTTP_EnabledField/enabled_TRUE'`
- Re-run without test cache:
  - `go test -v -count=1 ./internal/provider -run '^TestProvider_Configure$'`

### Acceptance Tests

Requires real Wormly credentials and `TF_ACC=1`.

- Run all acceptance tests:
  - `make testacc`
  - Equivalent: `TF_ACC=1 go test -v -cover -timeout 120m ./...`
- Run a single acceptance test:
  - `TF_ACC=1 WORMLY_API_KEY=... go test -v -timeout 120m ./internal/provider -run '^TestAccHostResource_basic$'`

### CI Parity Checks

CI runs:
- Build + lint
- `make generate` and verifies clean diff
- `make test`
- Separate workflow for acceptance tests

Before finishing larger changes, run:
1. `make fmt`
2. `make lint`
3. `make test`
4. `make generate` (if docs/schema/examples may change)

---

## Code Style Guidelines

### Formatting and File Hygiene

- Always use `gofmt` formatting (`make fmt`).
- Keep files ASCII unless existing file requires Unicode.
- Preserve existing comments unless they are demonstrably incorrect.
- Do not add temporal comments (for example: "new", "old", "refactored").

### Imports

- Let `gofmt` organize imports.
- Keep import groups in standard Go order:
  1. standard library
  2. third-party dependencies
  3. internal module imports
- Avoid unused imports; lint fails on these.

### Naming Conventions

- Exported identifiers: `PascalCase`.
- Unexported identifiers: `camelCase`.
- Interface names describe behavior (`HostAPI`, `SensorHTTPAPI`).
- Test names:
  - Unit: `TestXxx`
  - Acceptance: `TestAccXxx`
- Prefer explicit, domain-oriented names (`ScheduledDowntimePeriod`, `GlobalAlertMute`).

### Types and Data Modeling

- Use strong structs for API payloads and Terraform models.
- Keep Terraform schema models separate from API/domain models.
- Use Terraform framework `types.*` in provider models.
- Convert boundaries explicitly (for example: `types.Int64` to/from `int`).
- Keep interfaces where mocking/test seams are needed.

### Error Handling

- Return errors, do not panic.
- Wrap underlying errors with context using `%w` when propagating:
  - `fmt.Errorf("failed to ...: %w", err)`
- In Terraform provider/resource/data source code:
  - report user-facing issues via `resp.Diagnostics.AddError(...)`
  - return early when `resp.Diagnostics.HasError()`
- Handle "not found" paths explicitly and remove resource state when appropriate.

### Context, Networking, and Retries

- Thread `context.Context` through API boundaries.
- In tests, use `t.Context()` where available.
- Respect existing client patterns for:
  - rate limiting
  - retry/backoff
  - transient network/http error handling

### Terraform Provider Patterns

- Follow framework lifecycle methods cleanly:
  - `Metadata`, `Schema`, `Configure`, CRUD, `ImportState`
- Validate `ProviderData` types in `Configure`.
- Set computed/default fields consistently in state.
- Keep schema descriptions concise and user-focused.

### Testing Practices

- Prefer table-driven tests for multi-scenario behavior.
- Use `t.Run(...)` for subtests.
- Use `testify/assert` and `testify/mock` where already established.
- Keep acceptance tests isolated and credential-gated via `testAccPreCheck`.
- For HTTP client behavior, prefer `httptest.NewServer`.

### Linting Expectations

Configured linters include (non-exhaustive):
- `errcheck`
- `staticcheck`
- `unused`
- `ineffassign`
- `misspell`
- `godot`
- `unparam`
- `usetesting`
- `copyloopvar`
- `durationcheck`

Write code to satisfy these proactively instead of relying on later lint fixes.

### Documentation and Generation

- `make generate` runs:
  - `terraform fmt -recursive ../examples/`
  - `tfplugindocs generate --provider-dir .. -provider-name wormly`
- If schema/resource/data-source behavior changes, regenerate docs.
- Keep generated outputs committed when changed.

### Security and Secrets

- Never hardcode API keys or credentials.
- Acceptance tests must read `WORMLY_API_KEY` from env.
- Avoid logging secrets (especially request auth values).

### Change Checklist (Agent)

Before handing off:
1. `make fmt`
2. `make lint`
3. `make test`
4. `make generate` (when relevant)
5. Re-run targeted single tests for changed logic
6. Ensure no accidental secret/test-state artifacts are introduced
