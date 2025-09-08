# feedback

Submits feedback, a bug report, or a feature request to a contextvibes repository.

This command provides a low-friction way to file an issue directly from your terminal. It automatically includes diagnostic information like the CLI version and your OS in the issue body.

The target repository can be specified using a short alias defined in your `.contextvibes.yaml` configuration file. If no alias is provided, it defaults to the 'cli' repository.

### Examples

```bash
# File a quick bug report with just a title (interactive prompt for body)
contextvibes feedback "The 'project board list' command is failing"

# File a bug report for a different repository (e.g., 'thea')
contextvibes feedback thea "Typo in the strategic kickoff guide"

# Start a fully interactive session to be guided through filing feedback
contextvibes feedback
```
