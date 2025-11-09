# Claude Rules for WAP Bot

## Code Quality Rules

### Error Handling (CRITICAL)

**NEVER use bare `errors.New()` or `fmt.Errorf()` without wrapping**:

```go
// ❌ WRONG
return errors.New("failed to connect")
return fmt.Errorf("failed to connect")

// ✅ CORRECT
return fmt.Errorf("failed to connect to Slack: %w", err)
```

**ALWAYS use sentinel errors for known error conditions**:

- Check `internal/domain/errors_types.go` for existing sentinel errors
- Define new sentinel errors there when needed
- Example: `var ErrNotThreadMessage = errors.New("message is not in a thread")`

**ALWAYS record errors in trace spans with the helper functions**:

```go
ctx, t := telemetry.Tracer.Start(ctx, "slackbot.some_method")
if err != nil {
    return telemetry.WrapErrorWithTrace(t, "some description about the caller context", err)
}
```

**NEVER ignore errors**:

- Do not use `_` to discard error return values, unless you are unable to handle or doesn't makes sense to handle
- Always check and handle every error

### Testing Rules

**ALWAYS write tests when adding new functionality**:

- Use `testify/assert` for non-critical assertions
- Use `testify/require` for critical assertions that should stop the test
- Test naming: `TestFunctionName_Scenario`
- Example: `TestMessageProcessor_ExtractMusicLinks_ValidSpotifyURL`

**ALWAYS run tests with race detection**:

- Use `mise test` (race detection is enabled by default)
- Never disable race detection

**ALWAYS mock external services**:

- Mock Slack API calls in domain tests
- Focus tests on domain logic in `internal/domain/`

### Linting Rules

**BEFORE committing, ALWAYS run**:

```bash
mise lint
```

**Key linting requirements**:

- All switch statements MUST be exhaustive (handle all cases)
- All exported functions/types MUST have doc comments
- All packages MUST have package comments in doc.go or main file
- NEVER ignore context (always pass `context.Context`)
- ALWAYS close HTTP response bodies
- ALWAYS use structured logging with `slog`, NEVER `fmt.Println()`

### Code Style Rules

**Structured logging only**:

```go
// ❌ WRONG
fmt.Println("Processing message:", msgID)

// ✅ CORRECT
slog.Info("processing message", "message_id", msgID)
```

**Use context in logging if available**:

```go
// ❌ WRONG
ctx := context.TODO()
slog.Info("processing message", "message_id", msgID)

// ✅ CORRECT
ctx := context.TODO()
slog.InfoContext(ctx, "processing message", "message_id", msgID)
```

**Context propagation**:

- ALWAYS pass `context.Context` as the first parameter
- NEVER use `context.Background()` or `context.TODO()` in production code (only in main.go)

**Package organization**:

- Private/internal code → `internal/`
- Reusable/public code → `pkg/`
- NEVER export types from `internal/` packages

**Dependency injection**:

- Wire all dependencies in `cmd/bot/main.go`
- NEVER use global variables for services or clients, except for telemetry like logging, tracing or metrics

## Development Workflow Rules

**ALWAYS use mise commands**:

- `mise init` - First-time setup
- `mise test` - Run tests (includes race detection)
- `mise lint` - Run linter with auto-fix
- `mise build` - Build container image
- `mise start` - Start local dev environment

**NEVER commit without**:

1. Running `mise test`
2. Running `mise lint`
3. Ensuring all tests pass
4. Verifying semantic commit message format

## Architecture Rules

**Layered architecture - respect boundaries**:

- `cmd/bot/` → can import from `internal/` and `pkg/`
- `internal/services/` → can import from `internal/domain/`, `internal/config/`, `pkg/`
- `internal/domain/` → can ONLY import from `pkg/` (no services, no config)
- `pkg/` → NEVER import from `internal/` or `cmd/`

**Business logic belongs in `internal/domain/`**:

- Keep `internal/services/` fat every external call or non-business logic should live there
- All message processing logic goes in `internal/domain/` and every business logic goes into it's own file in that folder

## OpenTelemetry Rules

**ALWAYS add tracing to new functions**:

```go
ctx, span := tracer.Start(ctx, "struct_name.method_name")
defer span.End()
```

**ALWAYS record errors in spans**:

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
// Or use the above mentioned helper function
```

## Configuration Rules

**All config from environment variables**:

- Check `.env.example` for required variables
- Use `internal/config/` for config structs
- NEVER hardcode tokens, URLs, or environment-specific values

**Required environment variables**:

- `SLACK_BOT_TOKEN` - Slack bot OAuth token
- `SLACK_APP_TOKEN` - Slack app-level token for Socket Mode
- `DEBUG` - Enable debug mode
- `OTEL_*` - OpenTelemetry configuration

## Git Commit Rules

**ALWAYS use semantic commit format**:

- `feat:` - New features
- `fix:` - Bug fixes
- `refactor:` - Code refactoring
- `test:` - Test additions/changes
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks

**CI will reject non-semantic commits**.

## Documentation Rules

### When to Update CLAUDE.md

**ALWAYS update CLAUDE.md when**:

- Adding new coding standards or best practices
- Introducing new architecture patterns or layers
- Adding new linting rules or changing existing ones
- Modifying development workflows or mise commands
- Adding required environment variables
- Changing error handling patterns
- Updating OpenTelemetry/tracing practices
- Adding new testing requirements or patterns
- Changing package organization rules
- Adding new "What NOT to Do" items based on common mistakes

**Format for CLAUDE.md**:

- Keep rules concise and actionable
- Use ❌/✅ examples for clarity
- Add code examples for patterns
- Update "Key Files Reference" when adding important new files

### When to Update README.md

**ALWAYS update README.md when**:

- Adding new features or bot commands
- Changing setup instructions or prerequisites
- Adding, removing, or modifying environment variables
- Updating Go version requirements
- Adding new dependencies (Docker, tools, etc.)
- Adding new mise commands or tasks
- Changing the project structure (new packages/directories)
- Changing how to run, test, or deploy the application

**Format for README.md**:

- **Maintain brevity** - if information is already present concisely, don't expand it
- Avoid duplicating information that's already clearly documented
- Only add what's truly missing or unclear
- Respect the existing tone and style
- Keep user-facing and focused on "how to use"
- Update Features section when adding user-visible functionality
- Keep Development Workflow up-to-date with actual commands
- Update Project Structure when adding new packages
- Add examples for new bot commands or features

### Documentation Update Workflow

**When making changes**:

1. Make your code changes
2. Update CLAUDE.md if coding standards/patterns changed
3. Update README.md if user-facing functionality changed
4. Run `mise test` and `mise lint`
5. Commit with `docs:` prefix if documentation-only, or appropriate prefix if code + docs

**Documentation is part of the feature**:

- A feature is NOT complete without documentation updates
- PRs that add features MUST update README.md
- PRs that establish new patterns MUST update CLAUDE.md

## What NOT to Do

- ❌ Do NOT use `fmt.Println()` or `log.Println()` - use `slog`
- ❌ Do NOT use bare `errors.New()` without wrapping
- ❌ Do NOT skip writing tests for new functionality
- ❌ Do NOT commit without running `mise test` and `mise lint`
- ❌ Do NOT add global variables for services or clients
- ❌ Do NOT import `internal/` packages from `pkg/`
- ❌ Do NOT use `context.Background()` or `context.TODO()` except in main.go
- ❌ Do NOT disable race detection in tests
- ❌ Do NOT bypass exhaustive switch statement checks
- ❌ Do NOT add features without updating README.md
- ❌ Do NOT establish new patterns without updating CLAUDE.md
- ❌ Do NOT add redundant or duplicate documentation that's already clearly present

## Key Files Reference

- `cmd/bot/main.go` - Component wiring, start here for architecture understanding
- `internal/domain/slack.go` - Core business logic
- `internal/domain/errors_types.go` - Sentinel errors
- `internal/services/bot.go` - Slack Socket Mode event handling
- `pkg/musicextractors/` - Music URL extraction logic
- `mise.toml` - Development task definitions
