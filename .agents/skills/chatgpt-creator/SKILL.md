```markdown
# chatgpt-creator Development Patterns

> Auto-generated skill from repository analysis

## Overview

This skill teaches you how to contribute effectively to the `chatgpt-creator` codebase, a Go-based project with a modular structure and a Cloudflare worker submodule (TypeScript). You'll learn the project's coding conventions, commit patterns, testing strategies, and the main development workflows for features, bugfixes, documentation, and Cloudflare worker enhancements. This guide also introduces the custom `/commands` used to streamline common tasks.

## Coding Conventions

### File Naming

- **Go files:** Use `camelCase` for file names.
  - Example: `userHandler.go`, `emailService.go`
- **TypeScript files (Cloudflare worker):** Use `camelCase`.
  - Example: `tempMailWorker.ts`

### Import Style

- **Go:** Use relative imports within modules.
  ```go
  import (
      "internal/user"
      "cmd/server"
  )
  ```
- **TypeScript:** Use relative imports.
  ```ts
  import { sendMail } from './mailService'
  ```

### Export Style

- **Go:** Use named exports (exported identifiers start with uppercase).
  ```go
  func SendEmail(to string, body string) error {
      // ...
  }
  ```
- **TypeScript:** Use named exports.
  ```ts
  export function sendMail(to: string, body: string) { /* ... */ }
  ```

### Commit Patterns

- **Conventional commits** with these prefixes: `feat`, `fix`, `docs`, `refactor`, `chore`
- Example commit message:
  ```
  feat: add support for multiple email domains in temp mail worker
  ```

## Workflows

### Feature Development with Tests and Docs

**Trigger:** When adding a new capability or major improvement  
**Command:** `/new-feature`

1. Implement the feature in the appropriate source files:
    - Go: `internal/**/*.go`, `cmd/**/*.go`
    - Cloudflare worker: `cloudflare-temp-mail/src/**/*.ts`
2. Add or update corresponding tests:
    - Go: `*_test.go`
    - TypeScript: `cloudflare-temp-mail/tests/**/*.ts`, `cloudflare-temp-mail/tests/**/*.spec.ts`
3. Update or add relevant documentation:
    - `docs/*.md`, `README.md`, `TODO.md`, `docs/project-changelog.md`, `docs/project-roadmap.md`, `docs/system-architecture.md`

**Example:**
```go
// internal/emailSender.go
func SendEmail(to string, body string) error {
    // implementation
}
```
```go
// internal/emailSender_test.go
func TestSendEmail(t *testing.T) {
    // test cases
}
```
```markdown
# docs/project-changelog.md
- Added support for multiple email domains in temp mail worker
```

---

### Post-Review Hardening and Bugfix

**Trigger:** When addressing code review findings or bug reports  
**Command:** `/fix-after-review`

1. Update code to address review findings or bugs:
    - `internal/**/*.go`, `cmd/**/*.go`, `cloudflare-temp-mail/src/**/*.ts`
2. Update or add regression/unit tests to cover fixes:
    - Go: `*_test.go`
    - TypeScript: `cloudflare-temp-mail/tests/**/*.ts`, `cloudflare-temp-mail/tests/**/*.spec.ts`
3. Synchronize documentation and changelogs:
    - `docs/project-changelog.md`, `docs/deployment-guide.md`

**Example:**
```go
// internal/userHandler.go
// Fix: check for nil pointer before accessing user
```
```go
// docs/project-changelog.md
- Fixed bug in userHandler nil pointer dereference
```

---

### Documentation Synchronization

**Trigger:** When updating project documentation, standards, or integration guides  
**Command:** `/sync-docs`

1. Edit or add documentation files:
    - `docs/*.md`, `AGENTS.md`, `CLAUDE.md`, `README.md`
2. Synchronize index/metadata files if needed:
    - `AGENTS.md`, `CLAUDE.md`
3. Optionally update codebase summary or roadmap

**Example:**
```markdown
# docs/system-architecture.md
- Updated architecture diagram to reflect new mail worker integration
```

---

### Cloudflare Temp Mail Worker App Enhancement

**Trigger:** When adding or fixing features in the Cloudflare temp mail worker app  
**Command:** `/cloudflare-mail-update`

1. Implement or update features in:
    - `cloudflare-temp-mail/src/**/*.ts`
2. Add or update database migrations:
    - `cloudflare-temp-mail/migrations/*.sql`
3. Update or add tests:
    - `cloudflare-temp-mail/tests/**/*.ts`, `cloudflare-temp-mail/tests/**/*.spec.ts`
4. Update documentation:
    - `cloudflare-temp-mail/docs/*.md`, `docs/project-changelog.md`, `docs/deployment-guide.md`

**Example:**
```ts
// cloudflare-temp-mail/src/emailHandler.ts
export function handleIncomingMail(event: MailEvent) { /* ... */ }
```
```sql
-- cloudflare-temp-mail/migrations/002_add_mail_index.sql
ALTER TABLE mails ADD INDEX (recipient);
```
```markdown
# cloudflare-temp-mail/docs/usage.md
- Documented new mail filtering feature
```

## Testing Patterns

- **Go:** Use standard Go testing with files named `*_test.go`.
    ```go
    // internal/emailSender_test.go
    func TestSendEmail(t *testing.T) {
        // ...
    }
    ```
- **TypeScript (Cloudflare worker):** Use [Vitest](https://vitest.dev/) with test files matching `*.test.ts` or `*.spec.ts`.
    ```ts
    // cloudflare-temp-mail/tests/emailHandler.test.ts
    import { describe, it, expect } from 'vitest'
    import { handleIncomingMail } from '../src/emailHandler'

    describe('handleIncomingMail', () => {
      it('should process valid mail', () => {
        // ...
      })
    })
    ```

## Commands

| Command                | Purpose                                                        |
|------------------------|----------------------------------------------------------------|
| /new-feature           | Start a new feature with code, tests, and docs                 |
| /fix-after-review      | Address review findings or bug reports with code and tests      |
| /sync-docs             | Synchronize or update documentation files                      |
| /cloudflare-mail-update| Enhance or fix Cloudflare temp mail worker app                 |
```
