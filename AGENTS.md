# AGENTS.md — 2026_NCKU_BIM_Course

Course-materials repo, not a software project. No build/test/lint at repo root.
Working language: Traditional Chinese (zh-TW) for docs and replies.

## Repo shape

- `README.md` — course source of truth (schedule, fork+PR hand-in flow, deadline).
- `20260918/` — intro + Revit basics; `CAD/*.rvt` is the student-edition demo model.
- `20261002/` — GitHub + Revit MCP workflow; includes the hard-won debugging log at the end.
- `20261016/` — RFI board spec (`SPEC.md`) + runnable skeleton (`sample/`).
- `20261030/` — one folder per student group; this is the only place student work lands.

## Rules that prevent real mistakes

- **Main is protected. Never push to it.** All changes go fork → branch → PR → squash merge.
- **Only touch your own hand-in folder** (`20261030/第X組_應用名稱/` or existing `第N組/`).
  Do not rename other groups' folders (note: `第四組` uses a Chinese numeral — leave it).
- **Never commit `.rvt` files.** They are large binaries and stay local / in a shared drive.
  Repo holds scripts, exported JSON/schedules, docs, screenshots only.
- **Revit student vs commercial editions are file-incompatible.** All course files must be
  produced with Revit 2024 student edition.
- **Commit style:** conventional commits, e.g. `feat(1016-sample): ...`, `fix(...): ...`.

## RFI sample (`20261016/sample/`) — run and layering

- Start: `.\start.ps1` from `sample/` (opens backend `:8080` + frontend `:5173`, then browse
  `http://localhost:5173`). Backend must be running before frontend or plugin calls.
- Plugin (classroom = Revit 2024): `cd revit-plugin; dotnet build -c Release.R24; .\install.ps1 -Version 2024`,
  then restart Revit. Instructor machine only: `R26` / `-Version 2026`.
  Build requires that Revit version installed locally (csproj references
  `C:\Program Files\Autodesk\Revit $(RevitVersion)\RevitAPI.dll`).
- Backend binds loopback (`127.0.0.1:8080`) by design — no firewall prompt. SQLite via
  `modernc.org/sqlite` (pure Go, no CGO). DB auto-recreates; delete the db file to reset.
- Layering (enforced in class): plugin extracts (selection + screenshot + POST),
  backend stores (issues/models/elements), frontend presents. New features go in one layer first.
- Issues must bind a model object: Revit `ElementId` or IFC `GlobalId`. Text-only locations rejected.
  Status flow is `open → in_progress → closed` only.

## Revit MCP gotchas (from `20261002/README.md` appendix + log)

- One agent at a time: second connection gets **409** (exclusive lock). Release from the ribbon first.
- Agent only affects the **currently open** Revit project. Back up (save-as copy) before big changes;
  if unsaved and broken, close without saving to revert.
- `port 8964` occupied → `scripts\release-port.ps1` (in `shuotao/REVIT_MCP_study`, not this repo).
- Lesson from the NRE log: when tool calls fail uniformly, check the Revit-side log's
  **parsed field values** first (that case: `CommandName` vs `command` mismatch), not the stack.
