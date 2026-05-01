# Grafana DSConfig — Copilot Instructions

This repo contains TypeScript type definitions for Grafana datasource plugin provisioning configurations.

## Skills

### import-datasource-types

Generate detailed TypeScript type files for datasource plugins by exploring their GitHub repos. See [`.claude/skills/import-datasource-types/SKILL.md`](../.claude/skills/import-datasource-types/SKILL.md).

Use when asked to: create a new datasource type, import types from a datasource repo, add a datasource config type.

## Conventions

- Type files live in `src/<camelCaseName>.ts`
- Every field must have JSDoc with requirement level, source permalinks (commit SHA, not branch), backend behavior, and UI hints
- Companion types use `type` aliases (not `enum`) for string unions
- Exports registered in `src/types.ts`
- Reference example: `src/googleSheets.ts`
