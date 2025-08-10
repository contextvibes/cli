# AI Prompt: Refactor Prompt-Loading Logic in factory-cli

## Your Role & Task
You are an expert Go developer. Your task is to refactor the prompt-loading mechanism within the `factory-cli` tool. You will be given the complete codebase for `factory-cli` after this prompt.

## The Goal
The current implementation loads prompt templates from a hardcoded path. You must replace this with a more robust, layered search mechanism.

## New Prompt-Finding Logic (Implement in Order of Priority)
1.  **Project-Specific Path:** First, search for the prompt file in `./docs/prompts/` (within the project being analyzed).
2.  **Generic THEA Path:** If not found, search in `../thea/building-blocks/prompts/`. (The `thea` directory is a sibling to the `factory` and the project being analyzed).
3.  **Hardcoded Default:** If still not found, return a simple, hardcoded default prompt as a string.

## Refactoring Directives
1.  **Target the `generateReportHeader` function** in the file `factory/cli/context_helpers.go`. This is the function that needs to be completely rewritten.
2.  **The function signature must remain the same:** `func generateReportHeader(promptFile, defaultTitle, defaultTask string) (string, error)`. The `promptFile` argument should now just be the filename (e.g., `export-project-context.md`).
3.  **Implement the three-step search logic** described above inside this function.
4.  Ensure robust file existence checks and error handling.
5.  **Produce the complete, final Go code for the `factory/cli/context_helpers.go` file.** Do not use snippets or diffs.

---