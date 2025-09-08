# Applies a structured Change Plan or executes a shell script.

This command is the primary executor for AI-generated solutions. It can operate
in two modes:
1. Structured Plan (JSON): This is the preferred and safer mode of operation.
2. Fallback Script (Shell): For simple, imperative scripts.

Input can be read from a file with --script or piped from standard input.
