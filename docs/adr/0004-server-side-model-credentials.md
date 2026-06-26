# Server-side Model Credentials

Model provider credentials are never committed to the repository or exposed to the browser. A development or deployment default key may be supplied only through external secret configuration, while a user's own model key can override that default for their model-assisted template generation.

**Considered Options**

- Commit a default provider key for development convenience.
- Require every user to provide their own key before fill-in practice can use model-assisted generation.
- Use server-side secret configuration for development or deployment defaults, with optional user-provided keys.

**Consequences**

Local and demo environments need explicit secret setup when model-assisted generation is enabled, but the repository remains safe to publish and the frontend never becomes a model credential boundary.
