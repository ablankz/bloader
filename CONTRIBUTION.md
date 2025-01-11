# Contribution Guide

Thank you for considering contributing to this project! We welcome contributions from the community to help improve the project. This document provides guidelines to ensure that your contributions are effective and aligned with the project's goals.

---

## Table of Contents
- [Contribution Guide](#contribution-guide)
  - [Table of Contents](#table-of-contents)
  - [How to Contribute](#how-to-contribute)
  - [Development Setup](#development-setup)
  - [Submitting a Pull Request](#submitting-a-pull-request)
  - [Code Style Guidelines](#code-style-guidelines)
  - [Issue Reporting](#issue-reporting)
  - [Contact](#contact)

---

## How to Contribute

You can contribute in the following ways:
1. **Reporting Bugs**: Found a bug? Open an issue and provide details.
2. **Suggesting Features**: Have an idea for a new feature? Let us know by opening an issue.
3. **Improving Documentation**: Spotted a typo or an unclear explanation? Submit a pull request.
4. **Fixing Issues**: Pick an issue labeled as `help wanted` or `good first issue` and submit a fix.

---

## Development Setup

To set up the project locally, follow these steps:

1. Fork the repository.
2. Clone the repository to your local machine:
   ```bash
   git clone https://github.com/ablankz/bloader.git
   ```
3. Navigate to the project directory:
   ```bash
   cd bloader
   ```
4. Install dependencies:
   ```bash
   go mod tidy
   ```
5. Run tests to ensure the setup is working:
   ```bash
   go test ./...
   ```

---

## Submitting a Pull Request

Follow these steps to submit a pull request:

1. **Create a new branch**:
   - Use a descriptive branch name with a prefix such as `feat/` for features, `fix/` for bug fixes, etc.:
     ```bash
     git checkout -b feat/your-feature-name
     ```

2. **Make your changes** and commit them:
   ```bash
   git commit -m "Description of the changes made"
   ```

3. **Push your branch** to GitHub:
   ```bash
   git push origin feat/your-feature-name
   ```

4. **Open a pull request**:
   - Go to the original repository.
   - Click the "Compare & pull request" button.
   - Provide a clear description of the changes you made.

---

## Code Style Guidelines

1. Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
2. Use `gofmt` to format your code:
   ```bash
   gofmt -w .
   ```
3. Write tests for new features and bug fixes.
4. Keep functions and methods small and focused.
5. Use clear and descriptive variable and function names.

---

## Issue Reporting

If you find a bug or have a feature request, please:
1. Search the [issue tracker](https://github.com/ablankz/bloader/issues) to ensure the issue hasn’t already been reported.
2. Open a new issue and provide:
   - A clear and descriptive title.
   - Steps to reproduce the issue (if applicable).
   - Expected and actual results.
   - Screenshots or code snippets (if necessary).

---

## Contact

For questions or discussions, feel free to:
- Open an issue.
- Join our community discussions (link to Discord/Slack if applicable).

Thank you for contributing to this project! Together, we can make it even better.

