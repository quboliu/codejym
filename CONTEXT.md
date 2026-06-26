# CodeJYM Practice

CodeJYM is a code practice product where a learner works against source files using different practice modes. This context captures the product language around practice modes, reusable exercises, and learner progress.

## Language

**Practice Mode**:
A top-level way of practicing a source file, with its own validation rules and progress model.
_Avoid_: display mode, toggle

**Source File**:
A user-provided code file inside a training asset.
_Avoid_: document, exercise

**Source File Version**:
A source file identified by its content, not only by its path.
_Avoid_: file path, latest file

**Tracing Practice**:
The practice mode where the learner reproduces the source file character by character in order.
_Avoid_: default mode, normal mode

**Fill-in Practice**:
The practice mode where the learner completes only the hidden parts of a source file according to a fill-in template.
_Avoid_: tracing option, blank toggle

**Fill-in Template**:
A reusable blanking plan for exactly one source file version.
_Avoid_: prompt result, LLM answer, exercise instance

**Template Pool**:
The reusable set of fill-in templates available for one source file version.
_Avoid_: generation cache, prompt history

**Template Replenishment**:
The background growth of a template pool until it reaches its template limit.
_Avoid_: eager generation, batch precomputation

**Candidate Attempt**:
One attempt to produce and validate a candidate template during template replenishment.
_Avoid_: retry loop, prompt call

**Template Limit**:
The maximum number of fill-in templates kept for one source file version.
_Avoid_: LLM quota, session limit

**Candidate Template**:
A proposed fill-in template that has not yet been accepted into the template pool.
_Avoid_: saved template, practice template

**Structured Candidate Template**:
A candidate template expressed as machine-validated structured data.
_Avoid_: natural-language plan, prose answer

**Template Diversity**:
The degree to which a fill-in template differs from templates already in the same template pool.
_Avoid_: LLM confidence, prompt creativity

**Practice Value**:
The usefulness of a blanked region for helping a learner understand or remember meaningful code.
_Avoid_: randomness, visual variety

**Diversity Check**:
The deterministic acceptance check that decides whether a candidate template is different enough to enter a template pool.
_Avoid_: LLM self-review, manual review

**Practice Value Check**:
The acceptance check that rejects candidate templates whose blanks target low-value code.
_Avoid_: difficulty check, syntax validity

**Language-Aware Check**:
An acceptance check that uses source-language structure to judge blanks.
_Avoid_: string-only check, visual check

**Text Fallback Check**:
An acceptance check that uses generic text and token rules when source-language structure is unavailable.
_Avoid_: no validation, blind acceptance

**Template Acceptance**:
The final system decision that admits a candidate template into a template pool.
_Avoid_: LLM approval, automatic save

**Template Audit Trail**:
The stored evidence explaining how a candidate template was generated, scored, accepted, or rejected.
_Avoid_: debug log, UI history

**Template Difficulty**:
The coarse difficulty level of a fill-in template.
_Avoid_: score, learner grade

**Template Status**:
The lifecycle state of a fill-in template.
_Avoid_: visibility flag, generation result

**Retired Template**:
A fill-in template that is no longer selected for new practice but remains available for existing sessions.
_Avoid_: deleted template, invalid template

**Blank**:
A hidden continuous source-code span that the learner must restore in fill-in practice.
_Avoid_: missing character, token shard

**Semantic Span**:
A continuous source-code span that represents a meaningful code unit.
_Avoid_: arbitrary substring, random token group

**Exact Source Match**:
The answer rule where a learner's fill-in answer must match the original source span exactly.
_Avoid_: semantic equivalence, fuzzy answer

**Blank Answer**:
The original source span that correctly fills one blank.
_Avoid_: hint, placeholder

**Blank Input**:
The answer field where the learner restores one blank in fill-in practice.
_Avoid_: global cursor, typing cursor

**Blank-Level Validation**:
The validation of one blank input against that blank's original source span.
_Avoid_: whole-file cursor validation, LLM answer judging

**Server-Side Answer Validation**:
The server-owned validation of a blank input against its blank answer.
_Avoid_: browser answer check, client-side validation

**Revealed Blank**:
A blank whose answer has been shown to the learner during fill-in practice.
_Avoid_: correct answer, skipped blank

**Independent Completion**:
A fill-in practice completion where every blank was answered without revealing the answer.
_Avoid_: completed with help

**Assisted Completion**:
A fill-in practice completion where at least one blank was revealed.
_Avoid_: failed attempt, independent completion

**Practice Session**:
A learner's resumable progress record for one practice mode on one source file or fill-in template.
_Avoid_: template, file progress

**Fill-in Session**:
A learner's resumable progress record for one fill-in template.
_Avoid_: attempt history, tracing session

**Enter Fill-in Practice**:
The business action that prepares the next fill-in template and resumable session for a learner.
_Avoid_: fetch template, load page data

**Template Selection**:
The product behavior that chooses which fill-in template a learner should practice next.
_Avoid_: template management, manual template browsing

**Template Switch**:
The learner action of moving to another active fill-in template from the same template pool.
_Avoid_: regenerate template, create new template

**Fallback Template**:
A fill-in template generated by local rules when model-based generation cannot provide a usable template.
_Avoid_: broken template, temporary prompt substitute

**Model-Assisted Template Generation**:
The server-orchestrated process that asks a model to propose candidate fill-in templates.
_Avoid_: frontend prompt call, client-side generation

**Model Provider**:
An LLM vendor or compatible endpoint used for model-assisted template generation.
_Avoid_: hardcoded model, prompt engine

**Model Configuration**:
The user-facing configuration that selects the model provider and model used for model-assisted template generation.
_Avoid_: hidden environment setting, one-off API key

**User Model Configuration**:
A learner's default model configuration for model-assisted template generation.
_Avoid_: per-file model setup, template-specific provider

**Model Settings**:
The learner-facing settings area for managing user model configuration.
_Avoid_: practice control, per-template setup

**Development Model Key**:
A non-committed model provider key used only by developers or deployments that explicitly configure it.
_Avoid_: built-in default key, repository-managed key

**User Model Key**:
A learner-provided model provider key used for that learner's model-assisted template generation.
_Avoid_: browser key, shared system key

**Masked Model Key**:
A non-secret display form of a stored model key.
_Avoid_: API key, decrypted key

**Model Source Access**:
The controlled sending of a source file version's content to a model for template generation.
_Avoid_: repository access, client upload to model

**Template History Summary**:
A compact description of existing templates used to guide model-assisted generation away from repetition.
_Avoid_: full template replay, complete answer history

**Persisted Template Pool**:
A server-side template pool that can be reused across visits and devices.
_Avoid_: browser cache, temporary exercise list

**Persisted Fill-in Session**:
A server-side fill-in session that can be resumed across visits and devices.
_Avoid_: local draft, page state

## Relationships

- A **Source File** can have multiple **Source File Versions** over time.
- A **Source File Version** can be practiced through multiple **Practice Modes**.
- **Tracing Practice** and **Fill-in Practice** are separate **Practice Modes**.
- A **Fill-in Template** belongs to exactly one **Source File Version**.
- A **Template Pool** belongs to exactly one **Source File Version**.
- A **Template Pool** grows lazily when a learner enters **Fill-in Practice**, not when a source file is uploaded.
- When a **Template Pool** already has templates, **Fill-in Practice** starts from an existing template while **Template Replenishment** may continue in the background.
- When a **Template Pool** is empty, entering **Fill-in Practice** waits for the first **Fill-in Template**.
- A single **Template Replenishment** run makes at most 3 **Candidate Attempts**.
- The default **Template Limit** is 8 fill-in templates per **Source File Version**.
- **Template Replenishment** stops once the **Template Pool** reaches its **Template Limit**.
- A **Candidate Template** must pass the **Diversity Check** before it becomes a **Fill-in Template**.
- A **Candidate Template** from model-assisted generation must be a **Structured Candidate Template**.
- A **Structured Candidate Template** must identify blanks by source offsets that exactly match the source file version.
- LLM output may help create a **Candidate Template**, but **Template Diversity** is accepted or rejected by the system.
- A **Candidate Template** must pass the **Practice Value Check** before it becomes a **Fill-in Template**.
- High **Practice Value** blanks target meaningful names, expressions, calls, fields, literals, and domain logic.
- Low **Practice Value** blanks target standalone keywords, punctuation, whitespace, or mechanical boilerplate.
- **Practice Value Check** and **Diversity Check** should use **Language-Aware Checks** when available.
- **Text Fallback Checks** are used when language-aware structure is unavailable.
- The first-version **Language-Aware Checks** cover Go, TypeScript, JavaScript, Python, and Rust.
- **Template Acceptance** uses hard gates before weighted scoring.
- LLM scores may contribute to **Template Acceptance**, but cannot override failed hard gates.
- **Template Acceptance** should produce a **Template Audit Trail**.
- **Template Audit Trail** is server-side quality evidence and does not need first-version learner UI.
- **Template Difficulty** uses three levels: easy, medium, and hard.
- **Template Difficulty** is suggested by model-assisted generation and corrected by system checks.
- **Template Status** uses candidate, active, and retired.
- **Template Selection** uses active templates.
- A **Retired Template** is not selected for new practice but remains available to existing sessions.
- A **Blank** should usually cover one **Semantic Span**, not an arbitrary substring.
- A **Fill-in Template** may contain multiple **Blanks**, but each **Blank** should remain meaningful on its own.
- **Fill-in Practice** uses **Exact Source Match** as its default answer rule.
- A **Blank Answer** may be stored server-side but is not returned to the learner by default.
- Revealing a **Blank Answer** is a controlled fill-in practice action that creates a **Revealed Blank**.
- **Fill-in Practice** uses **Blank Inputs** and **Blank-Level Validation**, not a global source-file cursor.
- **Blank-Level Validation** is **Server-Side Answer Validation**.
- **Fill-in Practice** may show the answer for a blank, making it a **Revealed Blank**.
- A **Revealed Blank** does not count as independently answered.
- **Independent Completion** and **Assisted Completion** are distinct outcomes.
- **Tracing Practice** and **Fill-in Practice** can be selected from the same practice page.
- Switching between **Practice Modes** changes the workflow and session model, not just the display.
- **Enter Fill-in Practice** handles template pool lookup, possible replenishment, template selection, and fill-in session recovery.
- A **Fill-in Practice** session uses exactly one **Fill-in Template**.
- A learner has at most one active **Fill-in Session** per **Fill-in Template** in the first version.
- A **Fill-in Session** records blank-level answers and status, not a source-file cursor.
- **Template Selection** is automatic in the first version.
- **Template Selection** should prefer unfinished templates, then templates with higher learner need, then least-recently practiced templates.
- A learner may use **Template Switch** to move among active templates.
- **Template Switch** does not create new templates or bypass the template limit.
- If model-based generation cannot provide the first template for a **Template Pool**, the system should create a **Fallback Template**.
- A **Fallback Template** must still pass the **Practice Value Check** and **Diversity Check** before use.
- **Model-Assisted Template Generation** produces **Candidate Templates**, not accepted **Fill-in Templates**.
- **Model-Assisted Template Generation** is orchestrated by the server so model prompts, credentials, cost controls, and acceptance checks stay out of the client.
- **Model-Assisted Template Generation** uses a configured **Model Provider** rather than a hardcoded model.
- DeepSeek is the default **Model Provider** for the first version.
- OpenAI-compatible GPT models and Anthropic Claude models are supported **Model Provider** targets.
- **Model Configuration** is exposed through user-facing settings.
- The first version uses **User Model Configuration** rather than per-file or per-template model setup.
- Learners manage **User Model Configuration** through **Model Settings**, not through the fill-in practice workflow.
- A **Template Audit Trail** records the model provider and model actually used to generate or score a candidate template.
- A **Development Model Key** may provide a development or deployment default, but it must come from external configuration and must not be committed to the repository.
- A **User Model Key** can override the development or deployment default for that learner.
- **Development Model Keys** and **User Model Keys** are used only by the server.
- A **User Model Key** is stored server-side in encrypted form.
- Model settings shown to the learner use a **Masked Model Key**, not the decrypted key.
- If neither a **User Model Key** nor a **Development Model Key** is available, **Model-Assisted Template Generation** is unavailable and **Fallback Templates** remain available.
- **Model Source Access** is limited to the current **Source File Version** being used for **Fill-in Practice**.
- **Model Source Access** must be configurable off so **Fallback Templates** remain the only generation path.
- **Model Source Access** must not include user credentials, unrelated files, or repository-wide context.
- Model-assisted generation may receive a **Template History Summary**, not full historical template answers.
- A **Template Pool** is a **Persisted Template Pool**, not browser-local state.
- A **Fill-in Session** is a **Persisted Fill-in Session**, not browser-local state.
- A **Practice Session** records learner progress and does not define the exercise content.

## Example Dialogue

> **Dev:** "Should fill-in practice reuse the same cursor as tracing practice?"
> **Domain expert:** "No. **Fill-in Practice** is a separate **Practice Mode**, so its **Practice Session** should track progress against a **Fill-in Template**, not against the whole source file cursor."
>
> **Dev:** "If a learner edits a source file, should existing fill-in templates still apply?"
> **Domain expert:** "Only if they match the same **Source File Version**. A **Fill-in Template** is tied to content, not just a path."
>
> **Dev:** "Should we generate fill-in templates for every uploaded file immediately?"
> **Domain expert:** "No. The **Template Pool** grows only when a learner enters **Fill-in Practice** for that source file version."
>
> **Dev:** "Should a returning learner wait for a new template every time?"
> **Domain expert:** "No. If the **Template Pool** has templates, start **Fill-in Practice** immediately and use **Template Replenishment** to add variety in the background."
>
> **Dev:** "How many generation attempts should one replenishment run make?"
> **Domain expert:** "At most 3 **Candidate Attempts**. If none are accepted, use existing templates or a **Fallback Template** when the pool is empty."
>
> **Dev:** "How many templates should one source file version keep by default?"
> **Domain expert:** "The **Template Limit** is 8. After that, **Template Replenishment** stops and learners reuse the existing pool."
>
> **Dev:** "Can the model decide whether a new template is sufficiently different?"
> **Domain expert:** "The model can propose a **Candidate Template**, but the system's **Diversity Check** decides whether it joins the **Template Pool**."
>
> **Dev:** "Can the model describe the blanking plan in prose?"
> **Domain expert:** "No. Model-assisted generation must return a **Structured Candidate Template** that the system can validate against the source."
>
> **Dev:** "Is a template useful if it blanks keywords like `if`, `for`, or `int`?"
> **Domain expert:** "No. Those blanks have low **Practice Value**. A **Candidate Template** should blank meaningful code, such as conditions, API calls, names, fields, literals, or domain logic."
>
> **Dev:** "Can practice value be judged by string length alone?"
> **Domain expert:** "No. Prefer **Language-Aware Checks** so conditions, calls, fields, and literals can be distinguished from standalone keywords and punctuation."
>
> **Dev:** "Which languages get language-aware checks first?"
> **Domain expert:** "Go, TypeScript, JavaScript, Python, and Rust. Other languages use **Text Fallback Checks** until they are supported."
>
> **Dev:** "Can a high model score save a candidate with invalid ranges or low-value blanks?"
> **Domain expert:** "No. **Template Acceptance** applies hard gates first; model scores only contribute after those gates pass."
>
> **Dev:** "Do we need to keep why a candidate was accepted or rejected?"
> **Domain expert:** "Yes. **Template Acceptance** creates a **Template Audit Trail** so quality, prompt behavior, and rejection reasons can be reviewed later."
>
> **Dev:** "How detailed should template difficulty be?"
> **Domain expert:** "**Template Difficulty** is coarse: easy, medium, or hard. The model may suggest it, but the system can correct it."
>
> **Dev:** "Should low-quality templates be deleted?"
> **Domain expert:** "No. Use **Template Status**. A **Retired Template** stops new selection without breaking existing sessions."
>
> **Dev:** "Can we split one expression into scattered missing characters?"
> **Domain expert:** "No. A **Blank** should cover a continuous **Semantic Span** so the learner recalls meaningful code, not token fragments."
>
> **Dev:** "Can the learner submit semantically equivalent code?"
> **Domain expert:** "No. **Fill-in Practice** uses **Exact Source Match** so the learner restores the original source span."
>
> **Dev:** "Should the browser receive all blank answers when fill-in practice starts?"
> **Domain expert:** "No. A **Blank Answer** can be stored server-side, but it is only returned through a controlled reveal or review action."
>
> **Dev:** "Can the browser decide whether a fill-in answer is correct?"
> **Domain expert:** "No. **Blank-Level Validation** is **Server-Side Answer Validation** so answers and progress remain controlled by the server."
>
> **Dev:** "Should fill-in practice reuse tracing practice's character cursor?"
> **Domain expert:** "No. The learner answers through **Blank Inputs**, and each blank is checked with **Blank-Level Validation**."
>
> **Dev:** "Is the fill-in UI a separate product area?"
> **Domain expert:** "No. Learners choose **Tracing Practice** or **Fill-in Practice** from the same practice page, but each **Practice Mode** has its own workflow and session model."
>
> **Dev:** "Should the client assemble fill-in practice from separate template and session calls?"
> **Domain expert:** "No. **Enter Fill-in Practice** is one business action that returns the next usable template and resumable session."
>
> **Dev:** "If the learner reveals an answer, is that the same as getting it right?"
> **Domain expert:** "No. A **Revealed Blank** can unblock the learner, but it turns the outcome into **Assisted Completion** rather than **Independent Completion**."
>
> **Dev:** "Can one learner have multiple parallel attempts for the same fill-in template?"
> **Domain expert:** "Not in the first version. The learner has one resumable **Fill-in Session** per **Fill-in Template** and can reset it when they want to retry."
>
> **Dev:** "Should learners choose from a template list before practicing?"
> **Domain expert:** "Not in the first version. **Template Selection** is automatic so entering **Fill-in Practice** starts the next useful template."
>
> **Dev:** "Can learners regenerate templates whenever they dislike one?"
> **Domain expert:** "No. They may use **Template Switch** among active templates, but they do not manually regenerate the **Template Pool** in the first version."
>
> **Dev:** "Should fill-in practice be unavailable if model generation fails?"
> **Domain expert:** "No. If the first model-generated template is unavailable, use a **Fallback Template** that still passes the same acceptance checks."
>
> **Dev:** "Can templates and fill-in answers live only in browser storage?"
> **Domain expert:** "No. The **Template Pool** and **Fill-in Session** are persisted server-side so they can be reused and resumed."
>
> **Dev:** "Can the browser call the model directly to create fill-in templates?"
> **Domain expert:** "No. **Model-Assisted Template Generation** is server-orchestrated and only produces **Candidate Templates**."
>
> **Dev:** "Is fill-in generation tied to one hardcoded model?"
> **Domain expert:** "No. **Model-Assisted Template Generation** uses **Model Configuration**. DeepSeek is the default **Model Provider**, with GPT and Claude-style providers supported."
>
> **Dev:** "Does a learner configure a model per file?"
> **Domain expert:** "Not in the first version. The learner has one **User Model Configuration**, while each **Template Audit Trail** records what was actually used."
>
> **Dev:** "Where should model configuration live?"
> **Domain expert:** "In **Model Settings**. The practice page may show generation status, but it should not make provider setup part of the exercise workflow."
>
> **Dev:** "Can the default model key live in the repository?"
> **Domain expert:** "No. A **Development Model Key** is only an externally configured secret for development or deployment, and a **User Model Key** stays server-side."
>
> **Dev:** "Can user model keys live only in the browser?"
> **Domain expert:** "No. A **User Model Key** is encrypted server-side, displayed only as a **Masked Model Key**, and used only by the server."
>
> **Dev:** "Can the model see the source file?"
> **Domain expert:** "Yes, but **Model Source Access** is limited to the current **Source File Version** and can be disabled."
>
> **Dev:** "Should the model see all previous template answers?"
> **Domain expert:** "No. Give it a **Template History Summary** so it can avoid repetition without replaying complete historical templates."

## Flagged Ambiguities

- "mode" could mean a visual switch or a product-level workflow — resolved: **Practice Mode** is a product-level workflow with its own validation and progress model.
- "template" could mean a one-off LLM response or reusable exercise content — resolved: **Fill-in Template** is reusable exercise content for a source file.
- "file" could mean a path or a specific content version — resolved: **Fill-in Template** limits and reuse are based on **Source File Version**.
- "generate templates" could mean upload-time precomputation or practice-time demand generation — resolved: a **Template Pool** grows lazily when **Fill-in Practice** is entered.
- "generate on entry" could mean blocking every time or only when necessary — resolved: block only for the first **Fill-in Template**, then use existing templates while **Template Replenishment** runs in the background.
- "retry generation" could mean unlimited model calls or bounded candidate attempts — resolved: one **Template Replenishment** run allows at most 3 **Candidate Attempts**.
- "template count" could mean a global LLM budget or per-file cap — resolved: **Template Limit** is per **Source File Version**, defaulting to 8.
- "different enough" could mean LLM self-assessment or system-enforced variety — resolved: **Diversity Check** is deterministic and system-owned.
- "candidate template" could mean prose advice or data — resolved: model output must be a **Structured Candidate Template**.
- "valuable blank" could mean any syntactically valid blank or a meaningful exercise target — resolved: **Practice Value Check** rejects blanks over standalone keywords, punctuation, whitespace, and mechanical boilerplate.
- "practice value check" could mean string-only heuristics or language-aware analysis — resolved: use **Language-Aware Checks** when available and **Text Fallback Checks** otherwise.
- "first supported languages" means Go, TypeScript, JavaScript, Python, and Rust for first-version **Language-Aware Checks**.
- "LLM score" could mean final authority or one input into acceptance — resolved: **Template Acceptance** uses hard gates first, then weighted scoring.
- "template audit" could mean temporary logs or durable quality evidence — resolved: **Template Audit Trail** records generation, scoring, and acceptance or rejection evidence.
- "difficulty" could mean a precise numeric learner score or a coarse template attribute — resolved: **Template Difficulty** is easy, medium, or hard.
- "remove a template" could mean delete it or stop selecting it — resolved: use **Retired Template** for no-longer-selected templates.
- "blank" could mean a missing character, token, or full code fragment — resolved: a **Blank** is a continuous **Semantic Span**.
- "correct answer" could mean semantic equivalence or original-source reproduction — resolved: **Fill-in Practice** defaults to **Exact Source Match**.
- "blank answer" could mean public exercise data or controlled server-side answer material — resolved: **Blank Answer** is not returned to learners by default.
- "fill-in input" could mean whole-file typing or one field per blank — resolved: **Fill-in Practice** uses **Blank Inputs** with **Blank-Level Validation**.
- "answer validation" could mean browser-side comparison or server-owned progress mutation — resolved: use **Server-Side Answer Validation**.
- "show answer" could mean success, skip, or assisted progress — resolved: a **Revealed Blank** does not count as independent correctness.
- "enter fill-in" could mean raw data fetching or a business action — resolved: **Enter Fill-in Practice** prepares template selection and session recovery.
- "mode switch" could mean a display toggle or switching workflows — resolved: **Practice Mode** switching changes the workflow and session model.
- "fill-in progress" could mean attempt history or one resumable state — resolved: first version uses one active **Fill-in Session** per learner and **Fill-in Template**.
- "choose a template" could mean manual browsing or product-guided practice — resolved: first version uses automatic **Template Selection**.
- "switch template" could mean select an existing active template or generate a new one — resolved: **Template Switch** does not create templates.
- "LLM-generated" could mean required for fill-in practice to work — resolved: **Fallback Template** keeps **Fill-in Practice** available when model generation fails.
- "template storage" could mean browser-local cache or durable product state — resolved: **Persisted Template Pool** and **Persisted Fill-in Session** are server-side state.
- "LLM integration" could mean browser-side prompting or server-side generation — resolved: **Model-Assisted Template Generation** is server-orchestrated and only yields **Candidate Templates**.
- "model choice" could mean a hardcoded backend model or user-facing provider choice — resolved: use **Model Configuration**, defaulting to DeepSeek while supporting GPT and Claude-style providers.
- "model configuration scope" could mean per-file settings or user-wide defaults — resolved: first version uses **User Model Configuration**.
- "model settings placement" could mean practice controls or account settings — resolved: use **Model Settings** outside the exercise workflow.
- "system default key" could mean a built-in product secret or externally configured development secret — resolved: **Development Model Key** is external secret configuration and never repository-managed.
- "user model key storage" could mean browser-local storage or server-side secret storage — resolved: **User Model Key** is encrypted server-side and displayed only as a **Masked Model Key**.
- "send source to the model" could mean repository-wide access or current-file generation context — resolved: **Model Source Access** is limited to the current **Source File Version** and can be disabled.
- "history for the model" could mean full prior templates or compact guidance — resolved: use a **Template History Summary**.
