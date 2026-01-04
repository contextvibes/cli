### **The `contextvibes` CLI: A Design Manifesto**

#### **The `Why`: Our Guiding Principle**

The primary goal of the `contextvibes` CLI is to serve as an ergonomic and intuitive workbench for the modern software craftsman. Every design decision is driven by the need to reduce extraneous cognitive load, allowing the developer to focus their mental energy on the creative act of building software, not on wrestling with their tools.

To achieve this, we have abandoned a traditional, flat command structure in favor of a two-tiered model that serves both novice and expert users. This model is explicitly organized around the **W5H1 framework (Why, What, When/How, Where, Who/How)**, ensuring that the CLI's structure directly maps to a developer's natural thought process.

---

#### **The `What`: The Two-Tiered Menu Structure**

The CLI is divided into two distinct but complementary tiers:

1.  **The Guided Path:** A single, intelligent `library` command that acts as an interactive wizard. It is the entry point for discovery and learning.
2.  **The Expert Path:** A set of powerful, deterministic pillars for direct, scriptable access to the CLI's full capabilities.

---

### **The `How`: A Detailed Breakdown of the Menu**

This section details the complete command menu, explaining the purpose and rationale for each component.

#### **Tier 1: The Guided Path (`library` - The "Asker")**

This is the conversational interface to the entire system. It is designed to be the first command a new user learns. It answers the user's questions by intelligently coordinating the expert commands.

| Command | The User's Question | What it Does |
| :--- | :--- | :--- |
| **`library why`** | "Why are we doing this work?" | Consults the `project` pillar to show project goals and guide the user through selecting a work item. |
| **`library what`** | "What am I building?" | Consults the `product` pillar to provide a holistic description of the codebase and its current state. |
| **`library how`** | "How do I get my work done?" | Inspects the current state and consults the `factory` and `craft` pillars to provide context-aware, step-by-step guidance on the workflow. |
| **`library where`** | "Where are the key resources?" | Consults the `library` and `factory` pillars to provide pointers to important locations like the Git remote, configuration files, and knowledge bases. |

#### **Tier 2: The Expert Path (The "Doers")**

This is the core toolset for the experienced craftsman. It is organized into five pillars that directly answer the W5H1 questions.

##### **1. `project` — The "Why?" (The Front Office)**
*This pillar is for defining, planning, and managing the work itself.*

| Command | Rationale: The "Why" of the Action |
| :--- | :--- |
| **`describe`** | To define the "Why" for a specific task for an AI. |
| **`issues`** | To manage the "Why" of the project by creating and tracking work tickets. |

##### **2. `product` — The "What?" (The Workshop)**
*This pillar is for creating, manipulating, and analyzing the tangible source code.*

| Command | Rationale: The "What" of the Action |
| :--- | :--- |
| **`bootstrap`** | Creates the initial "What" (the source code). |
| **`build`** | Compiles the "What" into a runnable artifact. |
| **`test`** | Validates the correctness of the "What." |
| **`quality`** | Checks the quality of the "What." |
| **`format`** | Formats the "What." |
| **`clean`** | Cleans build artifacts related to the "What." |
| **`run`** | Runs the "What." |
| **`codemod`** | Programmatically modifies the "What." |

##### **3. `factory` — The "When & How (Mechanical)?" (The Assembly Line)**
*This pillar is for the process, sequence, and mechanical execution of the development workflow.*

| Command | Rationale: The "When & How" of the Action |
| :--- | :--- |
| **`init`** | **How** to set up the factory's configuration. |
| **`kickoff`** | **When** to start a new task on the assembly line. |
| **`commit`** | **How** to save a discrete step of work. |
| **`status`** | **How** to check the current state of the assembly line. |
| **`diff`** | **How** to inspect the work-in-progress. |
| **`sync`** | **How** to synchronize with the central factory. |
| **`finish`** | **When** a piece of work is finished and ready for review. |
| **`tidy`** | **How** to clean up the factory floor after a job is done. |
| **`plan`, `apply`, `deploy`** | The sequence (**When & How**) of deploying infrastructure. |

##### **4. `library` — The "Where?" (The Reference Room)**
*This pillar is for managing and accessing the knowledge, standards, and reusable assets.*

| Command | Rationale: The "Where" of the Action |
| :--- | :--- |
| **`index`** | Manages the index of **where** all knowledge is located. |
| **`thea`** | Interacts with an external library located at a specific **where**. |
| **`system-prompt`** | Manages the prompt templates stored **where** the knowledge lives. |
| **`add`** | Adds a new document to a specific **where** in the library. |

##### **5. `craft` — The "Who & How (Creative)?" (The AI Co-Pilot)**
*This pillar is for the craftsman (**Who**) to apply their skill and judgment in partnership with the AI to creatively solve problems (**How**).*

| Command | Rationale: The "Who & Creative How" of the Action |
| :--- | :--- |
| **`message`** | **How** the craftsman creatively formulates a commit message with AI help. |
| **`pr-description`** | **How** the craftsman creatively writes a PR description with AI help. |
| **`kickoff`** | **How** the craftsman creatively plans a new project in a strategic session. |

---

#### **The `Who` and `Where`: The Final Context**

*   **Who is this for?** The Software Craftsman. The entire design, from the guided `library` to the powerful expert pillars, is built to empower them.
*   **Where is this happening?** In the command-line interface—the craftsman's chosen workbench.


---

### **The Future: The Convergence Strategy**

Our North Star is the dissolution of the `craft` pillar.

Currently, `craft` exists as a bridge between **Intent** (what you want to do) and **Execution** (the mechanical command). It isolates the "AI Prompt Generation" into a separate step.

However, as the system matures, the "Creative" and the "Mechanical" will merge. The AI will become a modifier on the deterministic action, not a separate command.

#### **The "Orchestrator as API" Protocol**

We do not need to build complex, brittle integrations with specific LLM providers to achieve this. We embrace the **"Orchestrator as API"** pattern.

In this model, the **User (The Orchestrator)** acts as the secure, intelligent data transport layer between the CLI and the AI.

1.  **The CLI Prepares:** The CLI generates the perfect context and prompt (the "Request").
2.  **The Orchestrator Transports:** The user copies the request to the AI and copies the result back.
3.  **The CLI Executes:** The CLI parses the AI's structured response (the "Response") and performs the action.

#### **The Roadmap to Convergence**

Eventually, `craft` commands will be absorbed into the deterministic pillars as flags, streamlining the mental model:

| Current Workflow (Two Steps) | Future Workflow (One Concept) |
| :--- | :--- |
| `craft message` $\rightarrow$ `factory commit` | `factory commit --ai` |
| `craft pr-description` $\rightarrow$ `factory finish` | `factory finish --ai` |
| `craft refactor` $\rightarrow$ `product codemod` | `product codemod --ai` |
| `craft kickoff` $\rightarrow$ `project plan` | `project plan --ai` |

**The Goal:** One command, zero context switching, with the Orchestrator remaining in full control of the intelligence loop.
