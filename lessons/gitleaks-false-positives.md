# Gitleaks False Positives

## The Problem
When using Gitleaks in CI/CD pipelines (e.g., via `gitleaks-action`), it may flag mock keys, test secrets, or strings with high entropy (like test idempotency keys) as real secrets (e.g., `generic-api-key`).

Even if you change the value to a simpler string, Gitleaks might still flag it if the variable name implies a secret (e.g., `idempotencyKey`) and the string literal is attached to it, or if it has matching patterns.

## The Solution
To fix false positives, especially in test files where mock secrets are required:

1. **Inline Ignore**: Append `// gitleaks:allow` (for Go/JS/C++) or `# gitleaks:allow` (for Python/Ruby/Bash) to the exact line where the mock secret is defined. This tells Gitleaks to ignore that specific line.

    ```go
    // Example in Go
    idempotencyKey := "test-idempotency-key-1" // gitleaks:allow
    ```

2. **Configuration File**: If an entire directory (like `tests/`) or file type should be ignored by Gitleaks, you can create a `.gitleaks.toml` file at the root of the repository and configure the `allowlist` paths:

    ```toml
    [allowlist]
    paths = [
        '''src/tests/.*''',
        '''_test\.go$'''
    ]
    ```

## Lesson Learned
Always add the `// gitleaks:allow` inline comment for dummy keys or mock tokens in test files to prevent the CI pipeline from failing on false positives, rather than just trying to change the mock string to something simpler.
