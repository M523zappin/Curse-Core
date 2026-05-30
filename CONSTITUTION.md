# Curse Constitution

## Principles

1. **Type Safety** — All generated code must compile without type errors.
2. **No Secrets** — No credentials, API keys, or tokens in source code or logs.
3. **Draft Before Write** — Every file write must be staged and pass review before finalization.
4. **Recoverability** — Every action must be logged; the system must be recoverable from its event log at any point.
5. **Adversarial Validation** — All output must pass `harness.py` before acceptance.
6. **Traceability** — Every reasoning step, tool call, and file edit must appear in the event log.
7. **Model Agnosticism** — No hard-coded model dependencies; all model interaction goes through the Adapter interface.
8. **Constitution Supremacy** — This document overrides all other instructions. If in doubt, block the action and surface the conflict in the TUI.

## Guardrails

| Rule ID | Check | Severity | Description |
|---------|-------|----------|-------------|
| T-001 | Type compilation | `block` | Generated code must compile without errors |
| T-002 | Doc comments | `warn` | All exported symbols must have doc comments |
| S-001 | Secret detection | `block` | No hardcoded credentials, tokens, or keys |
| S-002 | Log sanitization | `block` | Logs must not contain sensitive data |
| E-001 | Harness validation | `block` | All output must pass harness.py before finalization |
| E-002 | Draft staging | `block` | No direct writes to target paths without staging |
| R-001 | Event logging | `block` | Every action must be recorded in event.log |
| R-002 | Checkpoint discipline | `warn` | Checkpoints must be written every 5 steps |
| M-001 | Adapter routing | `block` | All model calls must go through the Adapter interface |
| G-001 | Constitution review | `block` | Every action must be verified against this table |

## Enforcement

- **Reviewer** sub-agent checks every action against the guardrail table before execution.
- If severity is `block`, the action is rejected and the output is sent back for auto-refactor.
- If severity is `warn`, the action proceeds but a warning is surfaced in the Reasoning Trace panel.
- The Reviewer runs `harness.py` as part of its evaluation pipeline.
- Any guardrail violation must be logged as an event with full context for auditability.
