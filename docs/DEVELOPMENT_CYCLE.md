# Development Cycle & Best Practices

## 1. Discovery & Planning
- Define user stories with acceptance criteria.
- Capture non-functional requirements (latency, availability, privacy, cost caps).
- Maintain product scope in lightweight docs or issue templates.

## 2. Design
- **Backend:** API contract, data model, and validation rules.
- **Frontend:** UX flow, component inventory, and state diagrams.
- **AI:** Retrieval strategy, safety constraints, and evaluation metrics.
- Establish success metrics and rollout plan.

## 3. Implementation Standards
- Keep services modular with clear boundaries.
- Prefer composition over inheritance.
- Enforce linting and formatting in CI.
- Use environment-based configuration (no secrets in code).
- Log structured events with request IDs.

## 4. Testing Strategy
- Unit tests for core business logic.
- Integration tests for API + DB queries.
- Contract tests between frontend and backend.
- AI evaluation tests for retrieval quality and hallucination safety.

## 5. Code Review
- Require peer review for all changes.
- Use PR templates to capture rationale and verification steps.
- Block merges without tests or explicit waivers.

## 6. Release & Observability
- Use staged deployments (dev → staging → prod).
- Define SLOs for API latency and error rates.
- Track error budgets and regression alerts.

## 7. Documentation Hygiene
- Keep product and schema docs in `/docs`.
- Update docs as part of each feature PR.
- Use versioned migration files for schema changes.

