# Syncs the local branch with its remote counterpart.

Workflow:
1. Checks if the working directory is clean. Fails if dirty.
2. Pulls the latest changes from the remote using a rebase strategy.
3. Pushes local changes to the remote if the local branch is ahead.
