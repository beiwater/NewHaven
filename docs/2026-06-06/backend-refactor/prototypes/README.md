# Backend Refactor Prototypes

Last updated: 2026-06-06

This folder is for pseudocode and design sketches that come before `backend-next` implementation.

Files here should answer:

- What is the domain model?
- What are the use cases?
- What are the inputs and outputs?
- What state is read?
- What state is mutated?
- What should be tested?
- What old API behavior must remain compatible?

Prototype files should not be imported, compiled, or treated as runtime code.

## Suggested Files

```txt
00-target-architecture.md
01-system-and-config.pseudo.md
02-auth-flow.pseudo.md
03-catalog-domain.pseudo.md
04-company-domain.pseudo.md
05-production-domain.pseudo.md
06-market-domain.pseudo.md
07-finance-domain.pseudo.md
08-storage-boundary.pseudo.md
09-scheduler-lifecycle.pseudo.md
```

## Pseudocode Template

```md
# <Domain> Pseudocode

## Current Behavior

What does old backend A do today?

## Target Behavior

What should backend-next B do?

## Data Read

What state/config/static data is read?

## Data Mutated

What state changes?

## Use Cases

```txt
UseCaseName(input)
  validate input
  load state
  apply domain rules
  persist or update state
  return DTO
```

## Compatibility Notes

What old routes or response quirks must be preserved?

## Tests

What should prove this is safe?
```

