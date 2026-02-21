# Contributing to QSGW

Thank you for your interest in contributing to the Quantum-Safe Gateway. QSGW is an open-source project and we welcome contributions from the community, whether they are bug reports, feature requests, documentation improvements, or code changes.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Development Workflow](#development-workflow)
- [Branch Naming](#branch-naming)
- [Commit Message Format](#commit-message-format)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Cryptographic Code Review Requirements](#cryptographic-code-review-requirements)
- [Release Process](#release-process)
- [Getting Help](#getting-help)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code. Please report unacceptable behavior to [conduct@qbitel.dev](mailto:conduct@qbitel.dev).

---

## How to Contribute

### Reporting Bugs

If you find a bug, please open a GitHub issue with the following information:

- A clear and descriptive title.
- Steps to reproduce the problem.
- Expected behavior versus actual behavior.
- Your environment (OS, Rust/Go/Python/Node versions, Docker version).
- Relevant log output or error messages.

For security-related bugs, do **not** open a public issue. Instead, follow the [Security Policy](SECURITY.md).

### Suggesting Features

Feature requests are welcome. Please open a GitHub issue with:

- A clear description of the problem you are trying to solve.
- Your proposed solution or approach.
- Any alternatives you have considered.
- How this feature fits into the QSGW architecture.

### Improving Documentation

Documentation improvements are always appreciated. This includes:

- Fixing typos or unclear wording.
- Adding examples or tutorials.
- Improving API documentation.
- Translating documentation.

For small documentation fixes, feel free to open a pull request directly. For larger changes, open an issue first to discuss the approach.

---

## Development Workflow

### 1. Fork the Repository

Fork the [QSGW repository](https://github.com/yazhsab/qbitel-qsgw) on GitHub and clone your fork locally:

```bash
git clone https://github.com/<your-username>/qbitel-qsgw.git
cd qbitel-qsgw
git remote add upstream https://github.com/yazhsab/qbitel-qsgw.git
```

### 2. Create a Branch

Create a feature branch from `main`:

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

### 3. Set Up Your Development Environment

Follow the [Development Guide](docs/DEVELOPMENT.md) to set up your local environment, including prerequisites, infrastructure services, and database migrations.

### 4. Make Your Changes

- Write clear, well-documented code.
- Follow the coding conventions for the language you are working in (see [Development Guide](docs/DEVELOPMENT.md#code-style-and-conventions)).
- Add or update tests for any new or modified functionality.
- Update documentation if your changes affect the public API or user-facing behavior.

### 5. Run Tests and Linting

Before submitting your pull request, ensure all tests pass and linting is clean:

```bash
# Run all tests
make test

# Run all linters
make lint

# Check formatting
make fmt-check
```

### 6. Commit Your Changes

Write clear commit messages following the [Conventional Commits](#commit-message-format) format:

```bash
git add .
git commit -m "feat(gateway): add connection timeout configuration"
```

### 7. Push and Create a Pull Request

```bash
git push origin feature/your-feature-name
```

Open a pull request on GitHub against the `main` branch of the upstream repository.

---

## Branch Naming

Use the following prefixes for branch names:

| Prefix      | Purpose                                     |
|-------------|---------------------------------------------|
| `feature/`  | New features or enhancements                |
| `fix/`      | Bug fixes                                   |
| `docs/`     | Documentation changes                       |
| `refactor/` | Code refactoring (no functional changes)    |
| `test/`     | Adding or improving tests                   |
| `chore/`    | Build, CI, or tooling changes               |

**Examples:**

```
feature/add-grpc-upstream-support
fix/rate-limiter-cleanup-race
docs/update-api-examples
```

---

## Commit Message Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type       | Description                                    |
|------------|------------------------------------------------|
| `feat`     | A new feature                                  |
| `fix`      | A bug fix                                      |
| `docs`     | Documentation only changes                     |
| `style`    | Formatting, missing semicolons, etc.           |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test`     | Adding or updating tests                       |
| `chore`    | Build process, CI, or auxiliary tool changes    |
| `perf`     | Performance improvements                       |
| `ci`       | CI/CD configuration changes                    |

### Scopes

| Scope           | Description                    |
|-----------------|--------------------------------|
| `gateway`       | Rust gateway engine            |
| `control-plane` | Go control plane API           |
| `ai-engine`     | Python AI threat detection     |
| `admin`         | React admin dashboard          |
| `crypto`        | Cryptography crate             |
| `tls`           | TLS crate                      |
| `types`         | Shared types crate             |
| `docs`          | Documentation                  |

### Examples

```
feat(gateway): add WebSocket proxying support
fix(control-plane): return 404 for unknown gateway IDs
docs(api): document rate limit response headers
test(crypto): add ML-DSA-87 signature verification tests
perf(gateway): reduce memory allocations in proxy path
chore(ci): add Python 3.12 to test matrix
```

---

## Pull Request Guidelines

### Before Submitting

- [ ] All tests pass (`make test`).
- [ ] All linters pass (`make lint`).
- [ ] Code is properly formatted (`make fmt-check`).
- [ ] New functionality includes tests.
- [ ] Documentation is updated if needed.
- [ ] Commit messages follow Conventional Commits format.
- [ ] The branch is up to date with `main`.

### PR Description

Provide a clear description of your changes:

- **What** does this PR do?
- **Why** is this change needed?
- **How** was it tested?
- **Breaking changes**, if any.

### Review Process

1. A maintainer will review your PR within a few business days.
2. CI checks must pass before merging.
3. At least one maintainer approval is required.
4. For cryptographic code changes, the [cryptographic review requirements](#cryptographic-code-review-requirements) apply.
5. The maintainer may request changes. Please address review comments and push updates to the same branch.

### After Merging

Your branch will be deleted after merging. The `main` branch is protected and only accepts merge commits from approved PRs.

---

## Cryptographic Code Review Requirements

Changes to cryptographic code require additional scrutiny due to the security-critical nature of the QSGW gateway. The following rules apply to changes in the `crypto/` and `tls/` crates:

### Mandatory Requirements

- **Two maintainer approvals** are required (instead of the usual one).
- **At least one reviewer** must have cryptographic domain expertise.
- All cryptographic operations must include tests with **known answer test (KAT) vectors** from the relevant NIST standards.
- **Constant-time operations** must be used for all secret-dependent comparisons and computations.
- **Memory zeroization** must be applied to all key material after use.
- No use of `unsafe` Rust without explicit justification and approval.

### Review Checklist

- [ ] Algorithm implementation matches the NIST FIPS specification.
- [ ] KAT vectors pass for all parameter sets.
- [ ] No timing side channels in secret-dependent code paths.
- [ ] Key material is zeroized on drop.
- [ ] Error handling does not leak secret information.
- [ ] No unnecessary copies of key material.
- [ ] Dependencies are audited and from trusted sources.

---

## Release Process

QSGW follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes.
- **MINOR** version for backwards-compatible feature additions.
- **PATCH** version for backwards-compatible bug fixes.

### Release Steps

1. A maintainer creates a release branch: `release/vX.Y.Z`.
2. The [CHANGELOG.md](CHANGELOG.md) is updated with the release notes.
3. Version numbers are updated across all components.
4. CI runs the full test suite.
5. A GitHub Release is created with the tag `vX.Y.Z`.
6. Docker images are published to the container registry.

---

## Getting Help

- **Documentation:** Start with the [docs/](docs/) directory for architecture, API, deployment, and development guides.
- **GitHub Issues:** Search existing issues or open a new one.
- **Discussions:** Use GitHub Discussions for questions and general conversation.
- **Email:** Reach the maintainers at [oss@qbitel.dev](mailto:oss@qbitel.dev).

We appreciate your contributions and look forward to building a quantum-safe future together.
