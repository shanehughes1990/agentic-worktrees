# GraphQL Resolvers: Generated File Rule

`gqlgen` resolver files in this package (for example `*-resolver.go`) are regeneration-prone.

## Required practice

- Keep only resolver method implementations inside generated resolver files.
- Do not add package-level helpers, mappers, utils, constants, or extra types in generated resolver files.
- Put all helper logic in separate non-generated files in this same package (for example `resolver_helpers.go`, `todo_mapper.go`).

## Why

Running `gqlgen generate` can overwrite or remove non-method code added directly in generated resolver files.

## Safe pattern

1. Keep resolver method body thin.
2. Call helper/mapping logic from a separate stable file.
3. Regenerate safely without losing helper code.
