# txtar CLI Tool

A complete command-line tool for working with txtar format archives, supporting packing, unpacking, listing, and comparing operations with Git integration.

## Features

- **pack**: Pack directories or Git changesets into txtar format
- **unpack**: Safely unpack txtar archives to filesystem
- **list**: List files in a txtar archive
- **diff**: Compare two archives or directory with archive

## Installation

```bash
go install github.com/phlv/txtar@latest
```

Or build from source:

```bash
git clone https://github.com/phlv/txtar.git
cd txtar
go build -o txtar
```

## Usage

### Pack

Pack a directory into txtar format:

```bash
txtar pack --exclude "*.log" -o archive.txtar
```

Pack with Git integration:

```bash
txtar pack --git --since=1 -o delta.txtar
```

Pack staged changes:

```bash
txtar pack --git --staged -o staged.txtar
```

### Unpack

Unpack with backup:

```bash
txtar unpack --backup --dir /output archive.txtar
```

Dry-run to preview operations:

```bash
txtar unpack --dry-run archive.txtar
```

### List

List archive contents:

```bash
txtar list archive.txtar
```

From stdin:

```bash
cat archive.txtar | txtar list
```

### Diff

Compare two archives:

```bash
txtar diff left.txtar right.txtar
```

Compare directory with archive:

```bash
txtar diff --dir ./src archive.txtar
```

## Configuration

Create `~/.config/txtar/config.yaml`:

```yaml
pack:
  default_exclude:
    - "*.log"
    - "node_modules/"
    - ".git/"
  ignore_binary: true

unpack:
  backup: false
  dir: "./out"
```

## Examples

See [.config.yaml.example](.config.yaml.example) and [.txtarignore.example](.txtarignore.example) for configuration examples.

## Testing

```bash
go test ./...
```

## License

MIT
