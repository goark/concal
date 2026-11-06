# Copilot Instructions for today

## Project Overview

- This repository provides a small CLI to show information for today.
- Keep changes practical, incremental, and easy to review.

## Coding Guidelines

- Language: Go.
- Prefer small, focused changes over broad refactors.
- Preserve existing public behavior unless explicitly asked to change it.
- Keep source-code comments in English.
- Follow existing package boundaries and naming style in this repository.

## Validation

- Use Taskfile tasks for local validation.
- Primary check command:

```sh
task test
```

- If needed, run additional checks:

```sh
task govulncheck
```
