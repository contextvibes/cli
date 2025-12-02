# Bootstraps the secure environment (GPG, Pass, GitHub).

Configures the "Chain of Trust" workflow:
1.  **Plumbing:** Sets up GPG Agent, Git signing config, and shell aliases.
2.  **Identity:** Imports your GPG Key and applies "Ultimate Trust".
3.  **Vault:** Initializes the `pass` password store.
4.  **Auth:** Securely injects your GitHub PAT into the vault and authenticates the CLI.
