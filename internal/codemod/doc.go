/*
Package codemod defines the data structures used to represent codification modification
scripts for the contextvibes CLI. These structures allow for a standardized way
to describe a series of automated changes to files within a codebase.

The core types are:
  - Operation: Defines a single modification to be performed on a file, such as
    a regular expression replacement or a file deletion. It includes fields like
    `Type`, `Description`, `FindRegex`, and `ReplaceWith`.
  - FileChangeSet: Groups all `Operation`s intended for a single target file,
    specified by `FilePath`.
  - ChangeScript: Represents the top-level structure of a codemod script, which is
    an array of `FileChangeSet`s, allowing modifications across multiple files.

These types are typically unmarshalled from a JSON file (e.g., the default
`contextvibes-codemod.json` or a user-specified script) by the
`contextvibes codemod` command. The command then interprets these structures
to apply the requested changes to the project's files.

This package itself does not contain the execution logic for applying the
codemods; that logic resides in the `cmd` package (specifically `cmd/codemod.go`).
The primary role of `internal/codemod` is to provide the clear, typed
representation of the modification instructions.
*/
package codemod
