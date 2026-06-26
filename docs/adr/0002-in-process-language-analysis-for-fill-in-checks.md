# In-Process Language Analysis for Fill-in Checks

First-version language-aware practice value and diversity checks will run inside the Go backend instead of a separate analysis service. This keeps deployment simple for the existing Go API, Vue frontend, and Postgres stack; unsupported languages or parser failures fall back to text checks rather than making fill-in practice unavailable.

**Considered Options**

- Run language analysis inside the Go backend.
- Add a separate Node, Python, or Rust analysis service.
- Skip language-aware analysis and rely on text heuristics plus model scoring.

**Consequences**

The backend owns template generation, validation, and analysis timeouts in one place, but parser dependencies and abstractions must be kept narrow so this does not become a general IDE analysis engine.
