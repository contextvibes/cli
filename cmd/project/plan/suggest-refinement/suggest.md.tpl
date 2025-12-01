# Generate a prompt for an AI to classify untyped issues.

Scans the project's issue tracker for open issues that lack classification labels (e.g., `epic`, `story`, `task`, `bug`, `chore`).
It then generates a comprehensive prompt designed for an external AI (like Gemini or ChatGPT).

This prompt instructs the AI to:
1.  Analyze the title and body of each unclassified issue.
2.  Determine the most appropriate type/label.
3.  Generate a `bash` script using the `gh` CLI to apply these labels in bulk.
