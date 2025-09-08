# project labels create

Creates a new label in the repository.

This command allows you to define a new label that can be used on issues. You must provide a name, and can optionally provide a description and a hex color code.

### Examples

```bash
# Create a simple label
contextvibes project labels create --name "priority: high"

# Create a label with a description and color
contextvibes project labels create \
  --name "needs-triage" \
  --description "This issue requires review from the team" \
  --color "f29513"```
