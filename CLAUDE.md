
## Todos

This project uses a file-per-todo system in the `todos/` directory.

- Each file is `NNN_slug.md` with YAML frontmatter (`status`, `priority`, `created`, `tags`)
- Valid statuses: `open`, `in-progress`, `done`, `re-opened`, `on-hold`
- When working a todo: set status to `in-progress`, then `done` when complete
- `re-opened` means the user reviewed your work and it needs changes — read their feedback in the file
- Prefer `re-opened` todos over `open` ones
- `on-hold` todos should be skipped — do not work on them unless the user specifically asks
- Do not delete todo files — the user runs `ctd cleanup` or `ctd review` to manage done items
- To see available todos, read files in `todos/` sorted by filename
- When the user asks you to work on todos (e.g., "/todo"), read all files in `todos/`, find the lowest-sequence file with `status: open` or `status: re-opened`, set it to `in-progress`, and begin working on it.
