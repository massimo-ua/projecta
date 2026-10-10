# Codebase Exploration Guidelines

## Use the Knowledge Graph (graphify), Not Grep
When exploring code structure, locating definitions, tracing dependencies, or understanding relationships:
- **Do not grep blindly** across the codebase.
- **Use `graphify`** on the prebuilt knowledge graph (`graphify-out/graph.json`):
  - `graphify query <pattern>`: Search symbols (functions, structs, interfaces) by regex.
  - `graphify ask "<question>"`: Natural-language architectural queries.
  - `graphify explain <node>`: Inspect callers, callees, parent files, and imports.
  - `graphify path <from> <to>`: Shortest dependency chain between components.
  - `graphify affected [files]`: Impact analysis of modified files.
- Read source files only after identifying target locations from the graph.
- Run `graphify update .` incrementally after making structural code changes.
