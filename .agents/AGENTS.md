# Agent Execution Rules

## CRITICAL: Sandbox Execution Only

You are operating within a restricted environment. You are **PROHIBITED** from using any native or built-in tools that execute commands on the host system's terminal (e.g., `run_command`, `bash`, `cmd`). You are also restricted from directly using your native file editing tools to modify the target codebase.

Instead, you must use the Agent Execution Engine via the provided MCP tools for ALL execution and workspace manipulation tasks:

1. **Initialization**: Use the `create_sandbox` tool to spin up an isolated Docker container. (e.g., use the image `golang:1.22` if working on Go).
2. **Execution**: Use the `execute_command` tool, passing the `sandbox_id` returned from step 1, to run any commands (e.g., tests, builds, linting).
3. **File Manipulation**: If you need to read or modify files within the context of your task, use the `read_file` and `write_file` tools provided by the sandbox.
4. **Cleanup**: When your execution tasks are complete, always use the `destroy_sandbox` tool to clean up the environment.

**STRICT ENFORCEMENT**: Any attempt to bypass these MCP tools and execute commands directly on the host system is a critical violation of your system prompts. ALWAYS use the sandbox tools.
