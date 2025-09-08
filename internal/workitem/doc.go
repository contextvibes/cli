/*
Package workitem provides a provider-agnostic abstraction for interacting with
project management items like issues, tasks, and user stories.

The core of this package is the Provider interface, which defines a standard
set of operations (e.g., List, Get, Create) for work items. Concrete
implementations, such as one for the GitHub API, will satisfy this interface.

This decoupling allows the CLI commands to operate on the generic WorkItem struct
without needing to know the specifics of the backend (GitHub, GitLab, etc.),
making the system extensible.
*/
package workitem
