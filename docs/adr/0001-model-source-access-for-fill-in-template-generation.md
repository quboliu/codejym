# Model Source Access for Fill-in Template Generation

Fill-in templates need meaningful blanks, so model-assisted generation may send the current source file version's content to the model from the server. This is limited to the file being practiced, excludes credentials and unrelated repository context, and must be configurable off so local fallback templates remain available when source egress is not acceptable.

**Considered Options**

- Send no source to the model and rely only on local heuristics.
- Send the whole asset or repository to give the model broader context.
- Send only the current source file version with an opt-out.

**Consequences**

The model can judge practice value better than pure heuristics, but deployments that cannot allow source egress must disable model source access and use fallback template generation.
