# Contributing to StatePro

> 🤝 **Welcome Contributors!** - Help us build the future of quantum state machines

Thank you for your interest in contributing to StatePro! This guide outlines how to contribute effectively and ensure your changes are merged smoothly.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Types of Contributions](#types-of-contributions)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Style and Standards](#code-style-and-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)
- [Getting Help](#getting-help)
- [Recognition](#recognition)

## Code of Conduct

StatePro is committed to providing a welcoming and harassment-free experience for everyone. By participating in this project, you agree to abide by our code of conduct.

### Our Standards

- **Be Respectful** - Treat all community members with respect and kindness
- **Be Constructive** - Provide helpful feedback and focus on solutions
- **Stay On-Topic** - Keep discussions relevant to the project
- **Be Collaborative** - Work together to achieve common goals
- **Be Patient** - Help newcomers learn and grow
- **Take Responsibility** - Accept accountability for mistakes and learn from them

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Personal attacks or inflammatory language
- Publishing others' private information
- Trolling, spamming, or disruptive behavior

Violations will be reported to project maintainers. Serious violations may result in temporary or permanent exclusion from the project.

*This code of conduct is adapted from the [Contributor Covenant](https://www.contributor-covenant.org/)*

## Types of Contributions

We welcome all types of contributions to StatePro! Here's how you can help:

### Bug Reports & Issues

- Report bugs with detailed reproduction steps
- Suggest improvements to existing features
- Request new features with clear use cases
- Help triage and respond to existing issues

### Code Contributions

- **Bug Fixes** - Fix reported issues and edge cases
- **New Features** - Implement requested functionality
- **Performance** - Optimize execution speed and memory usage
- **Testing** - Add unit tests, integration tests, and benchmarks
- **Refactoring** - Improve code quality and maintainability

### Documentation

- **API Documentation** - Improve GoDoc comments and examples
- **Guides & Tutorials** - Write step-by-step tutorials
- **Examples** - Create real-world usage examples
- **Troubleshooting** - Document common issues and solutions
- **Translations** - Help translate documentation (future)

### Tooling & Infrastructure

- **Developer Tools** - Improve debugging and development experience
- **CI/CD** - Enhance build and testing pipelines
- **Automation** - Create scripts and tools for maintenance
- **Monitoring** - Add observability and metrics

### Good First Issues

Look for issues labeled:

- `good first issue` - Perfect for newcomers
- `help wanted` - Community contributions needed
- `documentation` - Documentation improvements
- `enhancement` - Feature improvements

## Development Setup

### Prerequisites

Before you start contributing, ensure you have:

- **Go 1.22+** - Required for development
- **Git** - For version control
- **Code Editor** - VS Code, GoLand, or Vim with Go support
- **(Optional) Docker** - For containerized development and testing

### Setup Steps

1. Fork the repository on GitHub
2. Clone your fork locally
3. Add upstream remote for syncing
4. Install dependencies
5. Verify installation by running tests and examples

## Development Workflow

### Getting Started

1. **Find an Issue** - Check open issues or create one
2. **Fork & Clone** - Fork the repo and clone your fork locally
3. **Setup Remotes** - Add upstream remote for syncing
4. **Create Branch** - Always branch from latest main
5. **Make Changes** - Implement changes and verify with examples
6. **Sync & Rebase** - Keep your branch up-to-date with main
7. **Create PR** - Submit pull request with clear description
8. **Code Review** - Collaborate with maintainers on feedback
9. **Merge** - Celebrate your contribution!

### Choosing What to Work On

#### Find an Issue

Browse open issues and filter by labels like `good first issue`, `help wanted`, `bug`, `enhancement`, or `documentation`.

#### Create Your Own Issue

If you have an idea not covered, create a new issue with a clear title, description, use case, and proposed solution.

### Creating a Feature Branch

Use descriptive branch naming conventions:

- Features: `feature/add-custom-validators`
- Bug fixes: `fix/issue-123-memory-leak`
- Documentation: `docs/api-reference-examples`
- Refactoring: `refactor/simplify-observer-logic`

Always start from the latest main branch.

### Developing and Testing

1. Make small, focused changes with one logical change per commit
2. Test early and often
3. Verify examples still work
4. Check code quality
5. Run the complete test suite before committing

### Keeping Your Branch Updated

Regularly sync with upstream by fetching latest changes and rebasing your branch.

## Code Style and Standards

### Go Code Style

- Follow standard Go formatting
- Use gofmt and goimports for consistent formatting
- Follow Go naming conventions
- Write clear, concise comments
- Use meaningful variable and function names

### Commit Messages

Use clear, descriptive commit messages following conventional format:

```plaintext
feat: add support for custom event validators

- Add EventValidator interface
- Implement default validators
- Update documentation

Fixes #123
```

### Code Structure

- Keep packages focused and cohesive
- Use interfaces for extensibility
- Document public APIs thoroughly
- Include examples in package comments

## Testing

### Writing Tests

- Write unit tests for all new functionality
- Include integration tests for complex features
- Use table-driven tests for multiple test cases
- Test error conditions and edge cases
- Test with real JSON state machine definitions
- Aim for >80% test coverage

### Test Structure

- Unit tests for data models
- Integration tests for runtime
- Example-based tests to verify JSON definitions work
- Benchmark tests for performance monitoring

## Documentation Guidelines

### Updating Documentation

- Keep README.md up to date
- Update docs/ for significant changes
- Include code examples in documentation
- Test all code examples

### Documentation Standards

- Use clear, concise language
- Include practical examples
- Explain concepts before implementation details
- Keep cross-references up to date

## Pull Request Process

### Overview

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make changes and test thoroughly
5. Sync with upstream
6. Create pull request
7. Address code review feedback
8. Merge

### Step-by-Step Guide

#### 1. Fork and Clone

- Go to the [StatePro repository](https://github.com/rendis/statepro) on GitHub
- Click "Fork" to create your own copy
- Clone your fork: `git clone https://github.com/YOUR_USERNAME/statepro.git`
- Add upstream remote: `git remote add upstream https://github.com/rendis/statepro.git`

#### 2. Create Feature Branch

- Sync with upstream: `git checkout main && git pull upstream main`
- Create branch: `git checkout -b feature/your-feature-name`

#### 3. Make Changes

- Implement your changes
- Test thoroughly (run tests, verify examples work)
- Commit with clear messages

#### 4. Sync Before PR

- Fetch upstream: `git fetch upstream`
- Rebase your branch: `git rebase upstream/main`
- Push to your fork: `git push origin feature/your-feature-name`

#### 5. Create Pull Request

- Go to your fork on GitHub
- Click "Compare & pull request"
- Fill out the PR template with clear description
- Submit the PR

#### 6. Code Review

- Automated checks will run
- Maintainers will review and provide feedback
- Address any requested changes
- Push updates to your branch

#### 7. Merge

- Once approved, maintainers will merge your PR
- Clean up your local branches

### Prerequisites Checklist

Before creating a PR, ensure:

- Fork is up-to-date with upstream main
- All tests pass
- New functionality has tests (when applicable)
- Examples run successfully
- Code is formatted
- No static analysis issues
- Dependencies are clean
- Documentation updated (if needed)
- Commit messages are descriptive
- No merge conflicts exist

### PR Description Template

Use this template for your PR description:

```plaintext
## Description

Provide a clear and concise description of what this PR does.

### Changes Made

- List key changes in bullet points

### Related Issues

Closes #123
Fixes #456

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update
- [ ] Refactoring
- [ ] Performance improvement
- [ ] Test addition

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed
- [ ] All existing tests pass
- [ ] Examples still work correctly

## Checklist

- [ ] Code follows the project's style guidelines
- [ ] Self-review of code completed
- [ ] Documentation updated (if needed)
- [ ] Tests added for new functionality
- [ ] All tests pass locally
- [ ] No new warnings or errors
- [ ] Commit messages follow convention
- [ ] Breaking changes documented (if any)
```

### Code Review Process

- Automated checks will run (tests, linting)
- Code review by maintainers
- Feedback and requested changes
- Approval and merge

Respond to feedback promptly, make requested changes, and communicate clearly with reviewers.

## Reporting Issues

### Bug Reports

When reporting bugs, include:

- Clear title describing the issue
- Steps to reproduce the problem
- Expected behavior vs actual behavior
- Environment details (Go version, OS, etc.)
- Code samples or minimal reproduction case
- Error messages and stack traces

### Feature Requests

For new features, include:

- Clear description of the proposed feature
- Use case and why it's needed
- Implementation ideas if you have them
- Potential impact on existing code

### Issue Labels

- `bug`: Something isn't working
- `enhancement`: New feature or improvement
- `documentation`: Documentation issues
- `good first issue`: Good for newcomers
- `help wanted`: Community contribution needed
- `question`: Questions and discussions

## Getting Help

- **Issues**: Use GitHub issues for bugs and feature requests
- **Discussions**: Use GitHub discussions for questions and ideas
- **Documentation**: Check docs/ directory for detailed guides

## Recognition

Contributors will be recognized in:

- GitHub's contributor insights
- Release notes for significant contributions
- Project documentation

Thank you for contributing to StatePro! 🚀
