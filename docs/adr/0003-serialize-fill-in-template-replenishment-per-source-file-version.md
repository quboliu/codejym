# Serialize Fill-in Template Replenishment per Source File Version

Template replenishment is serialized per source file version so concurrent learners do not trigger duplicate model calls or overfill the template pool. Requests can use existing active templates immediately, wait briefly for the first template when the pool is empty, or fall back when generation cannot produce an accepted template.

**Considered Options**

- Let every request independently replenish the pool.
- Serialize replenishment per source file version.
- Run a global background generator for all files.

**Consequences**

The template limit and diversity checks stay consistent, and model cost is bounded, but the backend needs a per-pool generation lock or job state around replenishment and acceptance.
