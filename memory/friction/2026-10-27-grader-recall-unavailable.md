# Agemem recall unavailable in independent grader

On 2026-10-27 the fresh read-only code grader reported that project memory recall failed with `runtime_unavailable`. The grader could not record details itself because its instructions prohibit writes. Parent-session recall succeeded separately. Impact: the grader inspected source files directly but could not retrieve project memory through the MCP. Reproduction: invoke project memory recall from a fresh independent grader environment; exact backend diagnostics were not supplied.
