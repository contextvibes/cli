# Project Kickoff Guide (Using ContextVibes CLI)

The `contextvibes kickoff --strategic` command (or `contextvibes kickoff` when run in a project for the first time) facilitates a comprehensive, AI-assisted strategic project kickoff. This guide explains its purpose, workflow, and how to get the most out of it.

## Purpose & Value

Starting a new project or a significant new phase of an existing project requires clear alignment on goals, scope, technical direction, and working practices. A thorough kickoff ensures everyone is on the same page, reduces ambiguity, and sets the project up for success.

The strategic kickoff feature in ContextVibes aims to:

*   **Standardize Kickoffs:** Provide a consistent and structured process for initiating projects.
*   **Deepen Understanding:** Guide you through critical questions covering project definition, technical foundations, development workflows, and governance.
*   **Enhance AI Collaboration:** Prepare a rich, detailed prompt for your preferred external AI assistant (like Gemini, Claude, ChatGPT, etc.), enabling it to act as an expert facilitator for your strategic planning session.
*   **Improve Outcomes:** Lead to better-defined projects, clearer objectives, proactive risk identification, and a more robust plan.
*   **Capture Knowledge:** Result in a summary document (that you create post-AI session) which serves as a valuable reference for the project team.

## Workflow Overview

The strategic kickoff process involves a few key phases:

1.  **Initiation with `contextvibes kickoff --strategic`:**
    *   You run `contextvibes kickoff --strategic` (or `contextvibes kickoff` if it's the first time in a project where a strategic kickoff hasn't been marked complete).
    *   ContextVibes CLI will engage you in a brief interactive session to:
        *   Confirm/set your preferences for how the ContextVibes CLI itself should provide outputs (e.g., code snippets, Markdown files) if it were to generate them based on future AI interactions or summaries. These are saved in `.contextvibes.yaml`.
        *   Gather essential high-level details about the project you are kicking off (e.g., project name, primary application type like "Go API" or "Python CLI", whether it's a new project or a new phase of an existing one).
        *   Run a quick technical readiness check regarding your `contextvibes` CLI setup.

2.  **Master Prompt Generation by ContextVibes:**
    *   Based on your initial input, `contextvibes` generates a detailed Markdown file named `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md` in your project root.
    *   This file contains a comprehensive "Master Kickoff Prompt" designed to be given to an advanced AI assistant. It includes:
        *   Instructions for the AI on how to act as an expert project kickoff facilitator.
        *   A detailed, multi-module checklist (Project Definition, Technical Foundations, Development Workflow, Governance, Next Steps) with specific questions for the AI to ask you.
        *   Instructions for the AI to prompt you to run other `contextvibes` commands (like `contextvibes describe`, `contextvibes status`) at relevant points to gather live project data.
        *   Instructions for the AI to help you synthesize certain decisions into structured YAML format (e.g., for your `.contextvibes.yaml` AI collaboration preferences).

3.  **AI-Guided Strategic Kickoff Session (You + Your External AI):**
    *   You **copy the entire content** of `STRATEGIC_KICKOFF_PROTOCOL_FOR_AI.md`.
    *   You **paste this content as the initial prompt** into your preferred external AI chat interface (e.g., Gemini, Claude, ChatGPT).
    *   The AI, now equipped with the protocol, will guide you interactively through all the checklist modules and questions.
    *   When the AI needs specific details about your current project's state (e.g., file structure, Git status, specific file contents), it will instruct you to run commands like `contextvibes describe -o temp_context.md` or `contextvibes status` in your project terminal.
    *   You then **copy the output from these `contextvibes` commands and paste it back into your chat with the AI facilitator.** This provides the AI with the concrete data it needs to help you plan effectively.
    *   If the AI generates YAML snippets for configuration (like `ai.collaborationPreferences`), it will instruct you to save these into your project's `.contextvibes.yaml` file.

4.  **User Actions: Summarize & Mark Complete:**
    *   **Create a Summary:** After your interactive session with the AI is complete, it's highly recommended that you synthesize the key decisions, action items, and outcomes into your own project document (e.g., `PROJECT_KICKOFF_SUMMARY.md`). The AI can help you draft this summary.
    *   **Save AI-Generated Configs:** Manually update your `.contextvibes.yaml` (or other relevant files) with any structured YAML data the AI helped you generate.
    *   **Inform ContextVibes:** Run `contextvibes kickoff --mark-strategic-complete` in your project terminal. This updates your project's `.contextvibes.yaml` to record that the strategic kickoff has been performed. Future runs of `contextvibes kickoff` (without the `--strategic` flag) in this project will then default to the daily Git workflow.

## Tips for a Successful Strategic Kickoff

*   **Be Prepared:** Before running `contextvibes kickoff --strategic`, have a general idea of the project or feature you want to kick off.
*   **Allocate Time:** The AI-guided session can be comprehensive. Allocate sufficient uninterrupted time for it (e.g., 1-2 hours depending on project complexity).
*   **Be Specific with the AI:** When the AI asks questions from the protocol, provide clear, specific, and thoughtful answers. The more context you give it, the better its guidance will be.
*   **Use `contextvibes describe` Actively:** When the AI facilitator suggests running `contextvibes describe` to gather more project context for it, do so. The output of `describe` is designed to be very helpful for an AI.
*   **Iterate:** Don't expect the first pass with the AI to be perfect. You might go back and forth on certain checklist items.
*   **Review AI Outputs:** Carefully review any code, configurations (like YAML), or architectural suggestions provided by the AI before implementing them. The AI is an assistant; you are the final decision-maker.
*   **Document Outcomes:** The summary document you create is crucial for team alignment and future reference.

By following this process, you can leverage ContextVibes CLI and your AI assistant to conduct highly effective and thorough project kickoffs.
