# Diagnostic Logging

Command results go to stdout; errors and actionable warnings are plain text on stderr.
Diagnostic logs are separate JSON-lines files, written at **Info** level by default.
Use these global flags with any command:

```bash
baronial --log-level=debug balance
baronial --log-file=/path/to/baronial.log balance
baronial --log-max-bytes=10485760 balance
baronial --log-level=off balance
```

Levels are `debug`, `info`, `warn`, `error`, and `off`. Info records command lifecycle
and failures; Debug adds index and envelopes persistence diagnostics. Changing the
diagnostic level never hides user-facing errors or safety warnings. Expected
account-versus-budget classification probes are Debug records, not command errors.

Default locations:

| Platform | Active log file |
| --- | --- |
| Linux | `$XDG_STATE_HOME/baronial/baronial.log`, or `~/.local/state/baronial/baronial.log` |
| macOS | `~/Library/Logs/baronial/baronial.log` |
| Windows | `%LOCALAPPDATA%\baronial\logs\baronial.log` |

## Rotation and Coordination

The default file-size limit is `5 * data.Megabyte` from `github.com/marstr/units/data`,
equivalent to 5 MiB. `--log-max-bytes` overrides it with a positive number of bytes;
the example above selects 10 MiB. Invalid values fail before command side effects,
even when diagnostics are disabled.

Each destination holds an active file and three backups (`.1` is newest), bounded
to four times the configured file-size limit: **20 MiB by default**, plus small
`.lock` and `.age` metadata files. Changing the limit prunes oversized managed
files on the next invocation/write. A record exceeding the limit disables
diagnostics for that invocation and emits the same once-only warning as other
logging failures; it does not fail the financial command.

Concurrent commands coordinate size checks, writes, rotation, and retention with
native Unix `flock` or Windows `LockFileEx`, using the existing `golang.org/x/sys`
dependency rather than an additional lock library. Lock acquisition is bounded
to approximately 250 ms. Each write reopens the active log under the lock to avoid
stale handles after another invocation rotates it.

Logs older than seven days are removed on subsequent invocations/writes, including
old content in an active file; there is no background cleanup while Baronial is
not running. The `.age` marker tracks the active file's oldest content rather
than its append-updated modification time. Limits apply separately to each
destination selected with `--log-file`. New log directories and files are
private to the user on Unix.

If logging cannot open, write, rotate, or acquire its lock, Baronial warns once
on stderr and continues the command without diagnostics. `--log-level=off`
creates no log files.

## Adding Diagnostic Records

Use `diagnostics.Logger(ctx)` from `internal/diagnostics` rather than a global
logger. Execution attaches invocation identity, command name, version/revision,
and platform to the logger. Repository opens pass it explicitly through
`filesystem.RepositoryLogger`, including cached repositories.

Keep terminal messages separate from diagnostic events. Return failures for the
execution boundary to report once, and preserve actionable plain-stderr warnings.
Do not add financial data or raw arguments to attributes. JSON `error` attributes
are classified instead of rendering arbitrary error strings that may contain
financial input; this does not sanitize other attributes automatically.

## Bug Reports and Privacy

When reporting a bug, attach the active log and relevant numbered backups to
the GitHub issue; do not attach the metadata files. If more detail is needed,
reproduce with `--log-level=debug` **only when repeating the command is safe**.
Logs are never uploaded automatically.

**Inspect logs before uploading.** Baronial avoids logging financial payloads,
amounts, merchant/comments, bank-record IDs, and raw command arguments.
Filesystem paths, reference names, object IDs, and other diagnostic details can
still be sensitive; logs are not guaranteed to be anonymized.
