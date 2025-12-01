# Export Upstream

Exports the source code of upstream dependencies defined in your configuration file.

This is useful for providing AI assistants with context about the libraries your project consumes, allowing them to write better code that utilizes those libraries correctly.

## Configuration

Add the modules you want to export to your `.contextvibes.yaml`:

```yaml
project:
  upstreamModules:
    - github.com/duizendstra-com/flow-sdk
    - github.com/charmbracelet/huh
```
