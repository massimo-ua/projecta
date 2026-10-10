---
name: graphify
description: >-
  Use when exploring, navigating, or understanding codebase architecture, symbols, and dependencies using graphify or graphify-go. Query the code knowledge graph (graphify-out/graph.json) to find definitions, trace call graphs, inspect neighbours, perform impact analysis on changed files, or locate core domain abstractions instead of raw grepping.
---

# graphify (graphify-go) — Code Knowledge Graph

Use `graphify` to explore the project's architecture, dependencies, and symbols through a prebuilt code knowledge graph stored at `graphify-out/graph.json` instead of searching or grepping through files blindly.

## When to Use

- **Find definitions & symbols**: Locating where a function, interface, struct, or class is defined.
- **Inspect symbol context**: Finding callers, callees, containing files, and imports for a specific symbol.
- **Trace dependency paths**: Finding how two components or symbols connect across the codebase.
- **Impact analysis (blast radius)**: Finding what breaks or what depends on modified files.
- **Natural language architectural queries**: Asking questions like "how does authentication work?".
- **Identify architectural core**: Discovering the most central "god nodes" and key abstractions.

## Core Commands

### 1. Natural Language Retrieval
```bash
graphify ask "<question>"
# Examples:
graphify ask "how does investment work"
graphify ask "where are project shares handled"
```
Returns a relevant subgraph highlighting starting nodes, connections, and source line numbers.

### 2. Search Symbols (Regex)
```bash
graphify query <pattern>
# Examples:
graphify query Investment
graphify query "New.*Service"
```
Finds matching nodes (types, functions, methods, files) with file paths and line numbers.

### 3. Inspect a Node & Neighbours
```bash
graphify explain <node>
# Examples:
graphify explain Investment
graphify explain "ServiceImpl.Create"
```
Shows the source location (`file:line`) and all inbound/outbound edges:
- What it `calls` / what `calls` it
- What `contains` it (file/module)
- What it `imports` / what `imports_from` it

### 4. Dependency Path Between Two Symbols
```bash
graphify path <from> <to>
# Example:
graphify path InvestmentRepository PgInvestmentRepository
```
Finds the shortest dependency chain connecting `<from>` and `<to>`. Add `--undirected` to traverse in either direction.

### 5. Impact Analysis (Affected Nodes)
```bash
# Check uncommitted changes:
graphify affected

# Check specific files:
graphify affected internal/investment/service.go pkg/web/investments.go
```
Lists all symbols defined in those files plus their transitive dependents (callers, importers) to determine what could be affected by edits.

### 6. Architectural Overview & God Nodes
```bash
graphify god-nodes --top 10
```
Displays the most connected nodes in the project. Also check `graphify-out/GRAPH_REPORT.md` for detected communities and import cycles.

### 7. Keeping the Graph Updated
When files are added or modified:
```bash
# Incremental update (re-parses only changed files, fast):
graphify update .

# Full rebuild:
graphify build .
```

## Recommended Workflow

1. **Start with the Graph**: Query (`graphify query <name>`) or ask (`graphify ask "<concept>"`) to locate relevant nodes.
2. **Follow Connections**: Use `graphify explain <node>` to understand relations before opening source files.
3. **Targeted Reading**: Open only the specific `file:line` references identified by the graph.
4. **Pre-commit Blast Radius**: Run `graphify affected` to verify what other components your changes touch.
5. **Update**: Run `graphify update .` after making structural code changes.
