---
type: reference
created: 2026-07-10
updated: 2026-07-12
---

# Skills Catalog

Danh sách toàn bộ skills có sẵn trong `.agents/skills/`. Dùng để tra cứu nhanh khi routing.

## 🏗️ Architecture & Design

| Skill | Dùng khi |
|-------|----------|
| `architecture` | Quyết định kiến trúc hệ thống, ADR, trade-off evaluation |
| `api-patterns` | Thiết kế API: REST vs GraphQL vs tRPC, response format, versioning, pagination |
| `database-design` | Schema design, indexing, ORM selection |
| `design-spec` | Viết DESIGN.md, design token (colors, typography, spacing) trước khi build UI |
| `frontend-architecture` | Cấu trúc frontend: UI/logic/data/type separation, file responsibility, state tiers |

## 🎨 Frontend & UI

| Skill | Dùng khi |
|-------|----------|
| `frontend-design` | Landing page, marketing/product sites, redesigns — anti-slop design |
| `tailwind-patterns` | Tailwind CSS v4, CSS-first config, container queries, design tokens |
| `nextjs-react-expert` | React/Next.js performance: SSR, bundles, waterfalls, optimization |
| `mobile-design` | iOS/Android mobile UI, touch interaction, React Native / Flutter |
| `web-design-guidelines` | Review UI vs. Web Interface Guidelines, accessibility audit |

## 🔧 Backend & Systems

| Skill | Dùng khi |
|-------|----------|
| `nodejs-best-practices` | Node.js: framework selection, async, security, architecture |
| `python-patterns` | Python: framework, async, type hints, project structure |
| `rust-pro` | Rust 2024 edition, Tokio, axum, async patterns, systems programming |
| `bash-linux` | Bash/Linux: commands, piping, error handling, scripting |
| `powershell-windows` | PowerShell: pitfalls, operator syntax, error handling |
| `server-management` | Process management, monitoring, scaling decisions |
| `mcp-builder` | MCP server building: tool design, resource patterns |

## 🧪 Testing & Quality

| Skill | Dùng khi |
|-------|----------|
| `testing-patterns` | Unit, integration, mocking strategies |
| `tdd-workflow` | RED-GREEN-REFACTOR TDD cycle |
| `webapp-testing` | E2E, Playwright, deep audit strategies |
| `code-review-checklist` | Code review: quality, security, best practices |
| `code-review-graph` | Token-efficient review dùng AST graph + SQLite |
| `verify-changes` | Prove code works by running it |
| `lint-and-validate` | Auto quality control, linting, static analysis — sau mỗi code modification |

## 🚀 DevOps & Deployment

| Skill | Dùng khi |
|-------|----------|
| `deployment-procedures` | Production deployment, rollback, verification |

## 🤖 AI & Agent Orchestration

| Skill | Dùng khi |
|-------|----------|
| `app-builder` | Build full-stack app từ đầu, tech stack selection |
| `intelligent-routing` | Auto-select best agent cho mỗi request |
| `parallel-agents` | Multi-agent orchestration, independent tasks chạy song song |
| `coordinator-mode` | Advanced multi-agent: parallel workers + synthesis |
| `behavioral-modes` | AI modes: brainstorm/implement/debug/review/teach/ship |
| `brainstorming` | Socratic questioning, complex requests, unclear requirements |
| `plan-writing` | Structured task planning, dependencies, verification criteria |
| `batch-operations` | Bulk modifications, search-and-replace across codebase |
| `skillify` | Auto-create new skills từ repetitive workflows |
| `memory-system` | Persistent cross-session memory management |
| `context-compression` | Compress context trong long sessions, archive old findings |

## 🔒 Security

| Skill | Dùng khi |
|-------|----------|
| `vulnerability-scanner` | OWASP 2025, supply chain security, attack surface mapping |
| `red-team-tactics` | Red team, MITRE ATT&CK, attack phases, detection evasion |

## 📈 Performance & SEO

| Skill | Dùng khi |
|-------|----------|
| `performance-profiling` | Profiling, measurement, optimization |
| `seo-fundamentals` | SEO, E-E-A-T, Core Web Vitals, Google algorithm |
| `geo-fundamentals` | GEO for AI search engines (ChatGPT, Claude, Perplexity) |

## 🌐 Other

| Skill | Dùng khi |
|-------|----------|
| `clean-code` | Concise, direct code — no over-engineering, no unnecessary comments |
| `simplify-code` | Reduce complexity, remove dead code, flatten nesting |
| `systematic-debugging` | 4-phase debugging, root cause analysis |
| `documentation-templates` | README, API docs, code comments |
| `i18n-localization` | Hardcoded strings detection, translations, RTL support |
| `game-development` | Game dev orchestrator, routes to platform-specific skills |
