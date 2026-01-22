# GitHub Copilot Configuration Template

This is a **template repository** for GitHub Copilot customization assets, designed to be cloned and adapted for new projects. The base collection is sourced from [github/awesome-copilot](https://github.com/github/awesome-copilot), providing a comprehensive starting point for AI-assisted development.

## Using This Template

1. **Clone this repository** as the foundation for your new project
2. **Remove unused instructions** that don't apply to your tech stack
3. **Customize remaining instructions** with project-specific patterns and conventions
4. **Add project-specific context** to agents and prompts
5. **Keep `.github/` structure** intact for proper Copilot integration

## Project Standards

### Design Decision Documentation (REQUIRED)

**All significant design decisions MUST be documented using Architecture Decision Records (ADRs).**

- **Location**: `docs/adrs/ADR-XXX-decision-title.md`
- **Format**: Follow the standard ADR template below
- **Numbering**: Sequential (001, 002, 003, etc.)
- **Trigger**: Create an ADR when:
  - Choosing between multiple technical approaches
  - Selecting languages, frameworks, or major libraries
  - Making architecture or infrastructure decisions
  - Changing core project conventions or patterns
  - Decisions that have long-term impact on the project

#### ADR Template

```markdown
# ADR XXX: [Decision Title]

**Status:** [Proposed | Accepted | Deprecated | Superseded by ADR-YYY]
**Date:** YYYY-MM-DD
**Decision Makers:** [Team/Role]
**Context:** [Project/Component]

## Context and Problem Statement

[Describe the context and the problem requiring a decision]

## Decision Drivers

- [Factor 1]
- [Factor 2]
- [Factor N]

## Options Considered

### Option 1: [Name]

**Pros:**
- [Advantage 1]
- [Advantage 2]

**Cons:**
- [Disadvantage 1]
- [Disadvantage 2]

### Option 2: [Name]

[Repeat for each option]

## Decision Outcome

**Chosen Option:** [Selected option]

### Rationale

[Explain why this option was chosen]

### Trade-offs Accepted

[What was sacrificed to gain the benefits]

## Consequences

### Positive

- [Benefit 1]
- [Benefit 2]

### Negative

- [Cost/Risk 1]
- [Cost/Risk 2]

### Mitigation

[How negative consequences will be addressed]

## References

- [Link to relevant documentation]
- [Link to discussion/RFC]

## Revision History

- **YYYY-MM-DD:** Initial decision
- **YYYY-MM-DD:** [Any updates or amendments]
```

#### Example ADRs

- `docs/adrs/ADR-001-use-go-over-python.md` - Language selection for always-on service
- `docs/adrs/ADR-002-use-rabbitmq-for-queuing.md` - Message queue technology choice
- `docs/adrs/ADR-003-adopt-hexagonal-architecture.md` - Architecture pattern selection

### Diagram Standards (REQUIRED)

**All diagrams MUST use Mermaid format for consistency and version control.**

- **Format**: Mermaid markdown code blocks
- **Location**: Embedded in README.md, ADRs, or separate `.md` files in `docs/diagrams/`
- **Types**: Use appropriate Mermaid diagram types:
  - `flowchart` - Process flows, decision trees
  - `sequenceDiagram` - API interactions, component communication
  - `classDiagram` - Object models, data structures
  - `erDiagram` - Database schemas, entity relationships
  - `stateDiagram` - State machines, lifecycle flows
  - `gitGraph` - Branching strategies
  - `gantt` - Project timelines

#### Mermaid Best Practices

```markdown
## Example Architecture Diagram

\`\`\`mermaid
flowchart LR
    A[Input] --> B{Decision}
    B -->|Yes| C[Process]
    B -->|No| D[Skip]
    C --> E[Output]

    style A fill:#e1f5ff
    style E fill:#d4edda
    style D fill:#fff3cd
\`\`\`

## Example Sequence Diagram

\`\`\`mermaid
sequenceDiagram
    participant Client
    participant API
    participant DB

    Client->>API: POST /users
    API->>DB: INSERT user
    DB-->>API: Success
    API-->>Client: 201 Created
\`\`\`

## Example State Diagram

\`\`\`mermaid
stateDiagram-v2
    [*] --> New
    New --> Processing: Submit
    Processing --> Completed: Success
    Processing --> Failed: Error
    Failed --> Processing: Retry
    Completed --> [*]
\`\`\`
```

#### Why Mermaid?

- ✅ **Version Control**: Text-based diagrams tracked in Git
- ✅ **Collaboration**: Easy to review and update in PRs
- ✅ **Rendering**: Works in GitHub, VS Code, and most documentation tools
- ✅ **No Binary Files**: Avoid binary image files that cause merge conflicts
- ✅ **Consistency**: Standardized syntax across all diagrams
- ✅ **Maintainability**: Update diagrams as code changes

**DO NOT** use:
- ❌ Binary image files (PNG, JPG) for architecture diagrams
- ❌ External diagram tools (draw.io, Visio) unless absolutely necessary
- ❌ ASCII art (hard to read and maintain)
- ❌ External hosting (links break, requires external accounts)

## Repository Structure

```
.github/
├── agents/           # Specialized AI assistants for specific workflows
├── instructions/     # Coding standards and best practices (language/framework-specific)
├── prompts/          # Reusable prompt templates for common tasks
├── skills/           # Domain-specific knowledge modules
└── workflows/        # GitHub Actions for automation
```

## Key Components

### Instructions (`instructions/*.instructions.md`)
Language and framework-specific coding standards that AI applies when generating or reviewing code:
- **Language-specific**: `java.instructions.md`, `nodejs-javascript-vitest.instructions.md`, `python.instructions.md`, `go.instructions.md`, `rust.instructions.md`, `shell.instructions.md`
- **Framework-specific**: `angular.instructions.md`, `nextjs.instructions.md`
- **Cross-cutting**: `security-and-owasp.instructions.md`, `performance-optimization.instructions.md`, `containerization-docker-best-practices.instructions.md`

Each instruction file includes:
- YAML frontmatter with `applyTo` glob patterns (e.g., `**/*.java`)
- Actionable guidelines, not generic advice
- Real code examples demonstrating patterns
- Anti-patterns to avoid

### Agents (`agents/*.agent.md`)
Specialized AI personas for complex workflows:
- **`modernization.agent.md`**: Comprehensive 9-step modernization workflow for legacy systems
- **`openapi-to-application.agent.md`**: Generate production-ready apps from OpenAPI specs
- **`se-security-reviewer.agent.md`**: OWASP Top 10 security analysis
- **`blueprint-mode-codex.agent.md`**: Structured workflow selection (Loop/Debug/Express/Main)

Agents follow a structured format:
```yaml
---
name: 'Agent Name'
description: 'Brief description'
model: 'GPT-4.1' or 'Claude Sonnet 4'
tools: ['codebase', 'edit/editFiles', 'search/codebase']
---
```

### Prompts (`prompts/*.prompt.md`)
Reusable templates for common tasks:
- **Architecture**: `architecture-blueprint-generator.prompt.md`, `create-architectural-decision-record.prompt.md`
- **Code Generation**: `openapi-to-application-code.prompt.md`, `convert-plaintext-to-md.prompt.md`
- **Documentation**: `readme-blueprint-generator.prompt.md`, `create-llms.prompt.md`, `folder-structure-blueprint-generator.prompt.md`
- **Code Quality**: `review-and-refactor.prompt.md`, `sql-code-review.prompt.md`, `postgresql-code-review.prompt.md`
- **Project Management**: `create-github-issues-feature-from-implementation-plan.prompt.md`

Prompts use YAML frontmatter to specify required tools and models.

### Skills (`skills/*/SKILL.md`)
Modular domain knowledge packages that AI can load on-demand:
- **`github-issues/`**: Complete workflow for creating, updating, and managing GitHub issues using MCP tools

## Working with This Repository

### Adding New Instructions
1. Create `.github/instructions/[name].instructions.md`
2. Add YAML frontmatter with `applyTo` glob pattern:
   ```yaml
   ---
   description: 'Brief description'
   applyTo: '**/*.ext'
   ---
   ```
3. Write specific, actionable guidance with code examples
4. Avoid generic advice - focus on YOUR project's patterns

### Adding New Agents
1. Create `.github/agents/[name].agent.md`
2. Define metadata in YAML frontmatter
3. Structure workflow with clear steps using headings
4. Include example inputs/outputs

### Adding New Prompts
1. Create `.github/prompts/[name].prompt.md`
2. Use configuration variables: `${VARIABLE="default|option1|option2"}`
3. Document workflow steps clearly
4. Provide example structures/templates

### File Naming Conventions
- **Instructions**: `[topic].instructions.md` (e.g., `java.instructions.md`)
- **Agents**: `[workflow-name].agent.md` (e.g., `modernization.agent.md`)
- **Prompts**: `[task-name].prompt.md` (e.g., `create-llms.prompt.md`)
- **Skills**: `skills/[skill-name]/SKILL.md` with supporting `references/` folder

## Special Instructions

### Memory Bank System
This repository uses a "Memory Bank" pattern (see `memory-bank.instructions.md`):
- Core files: `projectbrief.md`, `productContext.md`, `activeContext.md`, `systemPatterns.md`, `techContext.md`, `progress.md`
- Task tracking in `tasks/` folder with `_index.md` master list
- AI agents maintain context across sessions using these files

### Workflow Modes
Several instructions define specialized modes:
- **TaskSync V4** (`tasksync.instructions.md`): Terminal-driven continuous task execution
- **Spec-Driven Workflow** (`spec-driven-workflow-v1.instructions.md`): 6-phase loop (Analyze → Design → Implement → Validate → Reflect → Handoff)
- **Blueprint Mode** (`blueprint-mode-codex.agent.md`): Adaptive workflow selection

### Code Review Emphasis
Multiple layers of code review guidance:
- Generic code review checklist: `code-review-generic.instructions.md`
- Language-specific review rules embedded in each language instruction file
- SQL-specific reviews: `sql-code-review.prompt.md`, `postgresql-code-review.prompt.md`
- Security focus: `security-and-owasp.instructions.md`, `se-security-reviewer.agent.md`

## Important Patterns

### Configurable Instructions
Many instruction files use conditional sections:
```markdown
If `apply-doc-file-structure == true`, then apply the following section.
```

### EARS Notation for Requirements
`spec-driven-workflow-v1.instructions.md` uses EARS (Easy Approach to Requirements Syntax):
- `WHEN [trigger] THE SYSTEM SHALL [behavior]`
- Ensures testable, unambiguous requirements

### Self-Explanatory Code
`self-explanatory-code-commenting.instructions.md` emphasizes:
- Code that speaks for itself
- Comments only for WHY, not WHAT
- Specific annotation types: `TODO`, `FIXME`, `HACK`, `NOTE`, `WARNING`, `PERF`, `SECURITY`, `BUG`, `REFACTOR`, `DEPRECATED`

### Documentation Updates
`update-docs-on-code-change.instructions.md` triggers automatic documentation updates when code changes affect:
- Public APIs
- Configuration options
- CLI commands
- Installation/setup steps

### Configuration Consistency (CRITICAL)
When adding, removing, or modifying environment variables:
1. Update `.env.example` with the new variable and sensible default
2. Update `internal/config/config.go` to load and validate the variable
3. **Update `docker-compose.yml`** environment section to include the variable
4. **Update `README.md`** configuration table with the new variable (Input/Parsing/Output/Archive/Logging section)
5. **Update examples** in README.md if the variable affects usage patterns
6. Update ADR-003 if the change affects core behavior principles

**Common mistake**: Forgetting to sync docker-compose.yml or README.md with .env changes, causing container configuration drift and outdated documentation.

### Module Path for GitHub Publishing (IMPORTANT)

**Current State**: Module path is `csv2json` (local development)

**Before pushing to GitHub for the first time:**
1. Update `go.mod`: Change `module csv2json` to `module github.com/techie2000/csv2json`
2. Update `cmd/csv2json/main.go`: Change `csv2json/internal/*` imports to `github.com/techie2000/csv2json/internal/*`
3. Update `internal/processor/processor.go`: Change all `csv2json/internal/*` imports to `github.com/techie2000/csv2json/internal/*`
4. Run `go mod tidy` to update dependencies
5. Run `go test ./... -v` to verify all tests still pass

**Files to update:**
- `go.mod` (1 line)
- `cmd/csv2json/main.go` (2 import lines)
- `internal/processor/processor.go` (5 import lines)

**Why**: Local module name works for development, but GitHub requires full path for remote imports and `go get`.

### Architecture Decision Records (ADRs)
All significant design decisions must be documented in ADR format:
- Located in `docs/adrs/ADR-XXX-decision-title.md`
- Follow standard ADR template (see Project Standards section)
- Include context, options considered, decision rationale, and consequences
- Reference ADRs in README and related documentation

### Mermaid Diagrams
All architecture and technical diagrams must use Mermaid format:
- Embedded directly in Markdown files
- Version-controlled with code
- Use appropriate diagram types (flowchart, sequence, class, ER, state)
- Apply consistent styling and color schemes

## How to Use These Files in VS Code

### Instructions (Automatic Application)

Instructions are **automatically applied** by GitHub Copilot when you work on matching files:

1. **`applyTo` Pattern Matching**: The `applyTo` glob pattern in each instruction file determines when it activates
   ```yaml
   ---
   description: 'Java coding standards'
   applyTo: '**/*.java'
   ---
   ```
   - This instruction activates for ALL `.java` files in the workspace
   - Patterns like `src/**/*.ts` target specific directories
   - Multiple extensions: `**/*.{js,jsx,ts,tsx}`

2. **Automatic Context Loading**: When you:
   - Open a `.java` file → `java.instructions.md` loads automatically
   - Use Copilot chat in that file → Instructions guide the AI's responses
   - Generate code → AI follows the documented patterns

3. **Layered Instructions**: Multiple instructions can apply simultaneously:
   - `java.instructions.md` (language-specific)
   - `security-and-owasp.instructions.md` (applies to all files with `applyTo: '*'`)
   - `performance-optimization.instructions.md` (cross-cutting concerns)

### Agents (Explicit Invocation)

Agents are **invoked explicitly** using the `@` mention syntax in Copilot Chat:

1. **Syntax**: `@[agent-name]` followed by your request
   ```
   @modernization Help me plan the migration of this legacy codebase
   ```

2. **Available Agents**:
   - `@modernization` - 9-step legacy system modernization workflow
   - `@openapi-to-application` - Generate complete apps from OpenAPI specs
   - `@se-security-reviewer` - OWASP Top 10 security analysis
   - `@blueprint-mode-codex` - Adaptive workflow selector
   - `@address-comments` - Process PR review comments systematically

3. **Agent Context**: Agents have access to:
   - The codebase (via `tools: ['codebase']` in frontmatter)
   - File editing capabilities (`'edit/editFiles'`)
   - Search functionality (`'search/codebase'`)

4. **Example Workflow**:
   ```
   User: @modernization Analyze this .NET Framework 4.8 application
   Agent: [Executes 9-step analysis, reads all service files, generates plan]
   ```

### Prompts (Slash Commands)

Prompts are **invoked using slash commands** in Copilot Chat:

1. **Syntax**: `/[prompt-name]` (auto-completes as you type)
   ```
   /review-and-refactor
   /create-llms
   /architecture-blueprint-generator
   ```

2. **Interactive Configuration**: Many prompts have variables:
   ```yaml
   ${PROJECT_TYPE="Auto-detect|.NET|Java|React|Angular"}
   ```
   - Copilot will prompt you for these values
   - Or detect them automatically from context

3. **Common Prompts**:
   - `/review-and-refactor` - Comprehensive code review against all instruction files
   - `/readme-blueprint-generator` - Generate README from existing docs
   - `/sql-code-review` - Specialized SQL performance and security review
   - `/create-architectural-decision-record` - Document design decisions

4. **Prompt Chaining**: Combine prompts for complex workflows:
   ```
   /architecture-blueprint-generator → /create-llms → /readme-blueprint-generator
   ```

### Skills (On-Demand Loading)

Skills are **loaded automatically** when relevant topics are mentioned:

1. **Trigger Words**: Mentioning "GitHub issues", "create issue", "update issue" loads the `github-issues` skill

2. **Skill Structure**:
   - `SKILL.md` - Main instructions and workflow
   - `references/` - Supporting documentation (templates, examples)

3. **Usage Example**:
   ```
   User: Create a bug report issue for the authentication timeout
   Copilot: [Loads github-issues skill, uses templates, creates formatted issue]
   ```

### Practical Usage Patterns

#### For Code Generation
```
1. Open file matching an instruction (e.g., app.java)
2. Use Copilot inline suggestions → Follows java.instructions.md patterns
3. Or ask in chat: "Create a REST controller for User management"
   → Applies java.instructions.md + security-and-owasp.instructions.md
```

#### For Code Review
```
1. Select code block to review
2. Chat: /review-and-refactor
   → Reviews against ALL applicable instruction files
   → Suggests improvements following documented patterns
```

#### For Workflow Automation
```
1. Chat: @modernization Start analysis
   → Executes structured 9-step workflow
   → Generates comprehensive modernization plan
```

#### For Documentation
```
1. Chat: /create-llms
   → Scans repo structure
   → Generates llms.txt with all key documentation links
```

### Best Practices

1. **Start with Instructions**: Customize instruction files FIRST before writing code
2. **Use Agents for Complex Tasks**: Don't try to manually orchestrate multi-step workflows
3. **Chain Prompts**: Use prompts sequentially to build comprehensive documentation
4. **Review Auto-Applied Instructions**: Check which instructions are active with `Ctrl+Shift+P` → "Copilot: Show Instructions"
5. **Test Instructions**: After adding new instruction files, verify they apply correctly by generating code in matching files

### Debugging Copilot Configuration

6. **Document design decisions** in ADR format when making architectural choices
7. **Use Mermaid diagrams** for all architecture and technical documentation

This template repository is designed to make AI coding agents immediately productive by providing comprehensive, project-specific guidance rather than relying on generic training data.

## AI Agent Decision Documentation Workflow

When making significant technical decisions:

1. **Recognize Decision Point**: Identify when multiple valid options exist
2. **Create ADR**: Use `/create-architectural-decision-record` prompt or create manually
3. **Document Options**: List all considered alternatives with pros/cons
4. **Make Decision**: Choose option with clear rationale
5. **Update Diagrams**: Create or update Mermaid diagrams showing the decision's impact
6. **Link Documentation**: Reference ADR in README, related docs, and code comments

## Test Maintenance Workflow (MANDATORY)

**Every functional code change MUST be accompanied by test updates or new tests.**

When modifying Go code in csv2json:

1. **Identify Impact**: Determine which test modules are affected by the change
2. **Update Tests**: Modify existing tests to match new behavior
3. **Add New Tests**: Create new test cases for new functionality
4. **Validate ADR-003**: Ensure tests validate ADR-003 contracts (string values, empty strings, array structure)
5. **Run Tests**: Execute `go test ./... -v` to verify all tests pass
6. **Check Coverage**: Run `go test -cover ./...` to ensure coverage is maintained (>70% per module)
7. **Document**: Update TESTING.md if new test categories are added

**See `.github/instructions/test-driven-maintenance.instructions.md` for complete requirements.**

### ADR Creation Checklist

- [ ] ADR file created in `docs/adrs/` with sequential number
- [ ] All viable options documented with trade-offs
- [ ] Decision rationale clearly explained
- [ ] Consequences (positive and negative) listed
- [ ] Mitigation strategies for negative consequences included
- [ ] References and supporting documentation linked
- [ ] ADR referenced in README or relevant documentation
- [ ] Related Mermaid diagrams created or updated

### Diagram Creation Checklist

- [ ] Diagram uses Mermaid format (not binary images)
- [ ] Appropriate diagram type selected (flowchart, sequence, etc.)
- [ ] Clear labels and descriptions on all nodes/connections
- [ ] Color coding applied for clarity (e.g., green=success, red=error)
- [ ] Diagram embedded in relevant Markdown documentation
- [ ] Diagram explains the "what" and "how" of the system/flow
- [ ] Complex diagrams broken into smaller, focused diagrams
- Verify YAML frontmatter is valid (use `---` delimiters)
- Restart VS Code after adding new instruction files
- Check Copilot output panel: View → Output → Select "GitHub Copilot"

## Quick Start for AI Agents

When working with this repository:
1. **Identify the task type** (code generation, review, modernization, etc.)
2. **Load relevant instructions** from `.github/instructions/` matching the language/framework
3. **Consider using an agent** if the task matches a defined workflow
4. **Use prompts** for common templated tasks
5. **Maintain context** using Memory Bank files if needed for multi-step work

This template repository is designed to make AI coding agents immediately productive by providing comprehensive, project-specific guidance rather than relying on generic training data.
