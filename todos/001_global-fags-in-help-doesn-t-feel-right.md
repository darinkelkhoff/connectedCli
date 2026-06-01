---
status: done
priority:
created: 2026-06-01 06:45
tags: []
---
# global fags in help doesn t feel right
it feels to me like --json might be the only actual "global" flag.

--exit and --intro are actually subcommands.

--short only applies to --exit, right?

--say and --play, only apply to some commands

so - this needs cleaned up, and individual command helps need to specify what 
flags *actually* apply to them.

2026-06-01 07:30 Resolved by: Moved the whole intro/exit flag group (--intro,
--exit, --play, --say, --llm, --short, --prompt, --episode) off the root's
PersistentFlags (global) and onto the root's local Flags. Now `--json` is the
only global flag, and it's the only one shown under "Global Flags" in every
subcommand's help. The intro/exit flags now appear only in `conctl --help`
(root), and modifier flags are tagged "[--intro/--exit]" so it's clear they only
apply to those. `conctl search --exit` (etc.) now correctly errors with "unknown
flag". Kept --intro/--exit as flags rather than converting to subcommands, to
preserve the `connected --exit` joke that the project is built around.
