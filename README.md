# txtar CLI Tool

`txtar` is a command-line tool for creating, inspecting, extracting, and comparing `txtar` archives.

## Features

- Pack a directory into a `txtar` archive
- Pack files from a Git repository snapshot or changed files
- Unpack archives to the filesystem with overwrite safeguards
- List archive contents
- Diff two archives, or diff a directory against an archive

## Installation

Install with Go:

```bash
go install github.com/phlv/txtar@latest
```

Or build from source:

```bash
git clone https://github.com/phlv/txtar.git
cd txtar
make build
```

The binary will be written to `build/txtar`.

## Usage

```bash
txtar [--config FILE] <command> [flags]
```

Global flag:

- `--config`: use a specific config file instead of the default search path

Config lookup:

- `$HOME/.config/txtar/config.yaml`
- `./config.yaml`

### pack

Create a `txtar` archive from a directory or a Git-backed file set.

```bash
txtar pack [DIR] [flags]
```

If `DIR` is omitted, the current directory is used.

Flags:

- `-o, --output`: output file path, `-` means stdout. Default: `-`
- `-i, --include`: include glob pattern, repeatable
- `-e, --exclude`: exclude glob pattern, repeatable
- `--git`: enable Git-aware packing
- `--diff`: pack files from the current Git diff, including staged, unstaged, and untracked files; requires `--git`
- `--commit`: pack a specific commit snapshot, requires `--git`
- `--since`: pack files changed across the last `N` commits, requires `--git`
- `--staged`: pack staged files from the working tree, requires `--git`
- `--worktree`: pack modified worktree files, requires `--git`
- `--strip-prefix`: remove a path prefix from archived file names
- `--dry-run`: print the files that would be packed instead of writing an archive
- `--ignore-binary`: skip files detected as binary
- `--txtarignore`: ignore file name to load from `DIR`. Default: `.txtarignore`

Behavior notes:

- Without `--git`, `pack` walks the filesystem under `DIR`.
- With `--git` and no extra Git selector flags, `pack` archives the current `HEAD` commit snapshot.
- `--diff`, `--commit`, `--since`, `--staged`, and `--worktree` are mutually exclusive Git selection modes.
- `--git --diff` packs the current working tree diff: staged files, unstaged files, and untracked files.
- `.gitignore` is only loaded in `--git` mode.
- `.txtarignore` is loaded from `DIR` when present.
- In `--dry-run` mode, the file list is printed to stdout.
- Deleted files may appear in Git status but are skipped because there is no file content to archive.

Examples:

```bash
txtar pack . -o archive.txtar
txtar pack . -e '*.log' -e 'node_modules/**' -o archive.txtar
txtar pack . -i 'cmd/**' -i 'internal/**' -o src.txtar
txtar pack . --git --commit HEAD~1 -o head-minus-1.txtar
txtar pack . --git --since 3 -o recent-changes.txtar
txtar pack . --git --diff -o working-tree.txtar
txtar pack . --git --staged -o staged.txtar
txtar pack . --git --worktree --ignore-binary --dry-run
txtar pack . --strip-prefix internal/ -o internal.txtar
```

### unpack

Extract a `txtar` archive to a directory.

```bash
txtar unpack [ARCHIVE] [flags]
```

If `ARCHIVE` is omitted or set to `-`, data is read from stdin.

Flags:

- `-C, --dir`: output directory. Default: `.`
- `--backup`: rename an existing file to `*.bak` before overwriting
- `--dry-run`: print planned writes without creating files
- `--no-overwrite`: fail if a target file already exists

Behavior notes:

- `--backup` and `--no-overwrite` are mutually exclusive.
- Archive entries using absolute paths or `..` path traversal are rejected.
- When `--backup` is enabled and `file.bak` already exists, a timestamped backup name is used.

Examples:

```bash
txtar unpack archive.txtar
txtar unpack archive.txtar -C out
txtar unpack archive.txtar --backup -C out
txtar unpack archive.txtar --no-overwrite -C out
cat archive.txtar | txtar unpack - -C out
txtar unpack archive.txtar --dry-run -C out
```

### list

Print the file paths stored in an archive.

```bash
txtar list [ARCHIVE]
```

If `ARCHIVE` is omitted or set to `-`, data is read from stdin.

Examples:

```bash
txtar list archive.txtar
cat archive.txtar | txtar list
```

### diff

Compare two archives, or compare a directory to an archive.

```bash
txtar diff [LEFT] [RIGHT] [flags]
```

Flags:

- `--dir`: treat `LEFT` as a directory instead of an archive
- `-c, --content`: show content differences for modified files

Output markers:

- `+`: file exists only in `RIGHT`
- `-`: file exists only in `LEFT`
- `M`: file exists in both sides but contents differ

Examples:

```bash
txtar diff left.txtar right.txtar
txtar diff left.txtar right.txtar --content
txtar diff --dir ./workspace archive.txtar
txtar diff --dir ./workspace archive.txtar --content
```

## Configuration

Optional config file: `~/.config/txtar/config.yaml`

Supported keys:

```yaml
pack:
  output: "-"
  default_exclude:
    - "*.log"
    - "node_modules/**"
  ignore_binary: true

unpack:
  backup: false
  dir: "./out"
```

Notes:

- `pack.default_exclude` is prepended to CLI `--exclude` values.
- `pack.ignore_binary` is used only when `--ignore-binary` is not set explicitly.
- `unpack.backup` and `unpack.dir` are used only when the matching CLI flags are not set explicitly.

## Development

Available `make` targets:

- `make build`
- `make test`
- `make fmt`
- `make vet`
- `make lint`
- `make install`

## License

MIT
