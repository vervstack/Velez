# Velez UI Redesign — Roadmap

Reference design: [`docs/paas/ui/examples/Velez Dashboard.html`](examples/Velez%20Dashboard.html)

The goal is to rebuild the UI to match the dark-themed dashboard design from the example. Work is split into 6
milestones, bottom-up per feature-sliced design: tokens → base components → complex components → widgets → layout →
pages.

Each task is self-contained and executable by a model without prior context. Tasks within a milestone can be done in any
order. Complete each milestone before starting the next.

---

## Milestone 1 — Design Tokens

| #   | Task                                                         | File                   |
|-----|--------------------------------------------------------------|------------------------|
| T01 | [Design tokens CSS variables](tasks/M1-T01-design-tokens.md) | `src/index.module.css` |

---

## Milestone 2 — Base Components (`src/components/base/`)

| #   | Task                                                              | Files                                                                                           |
|-----|-------------------------------------------------------------------|-------------------------------------------------------------------------------------------------|
| T02 | [StatusDot](tasks/M2-T02-status-dot.md)                           | `StatusDot.tsx` + `.module.css`                                                                 |
| T03 | [Badge](tasks/M2-T03-badge.md)                                    | `Badge.tsx` + `.module.css`                                                                     |
| T04 | [MiniBar (progress bar)](tasks/M2-T04-mini-bar.md)                | `MiniBar.tsx` + `.module.css`                                                                   |
| T05 | [Chip (EnvChip, IncidentChip, FreezeChip)](tasks/M2-T05-chips.md) | `chips/EnvChip.tsx`, `chips/IncidentChip.tsx`, `chips/FreezeChip.tsx`, `chips/chips.module.css` |
| T06 | [IconButton](tasks/M2-T06-icon-button.md)                         | `IconButton.tsx` + `.module.css`                                                                |
| T07 | [Button (rebuild)](tasks/M2-T07-button.md)                        | `Button.tsx` + `Button.module.css`                                                              |
| T08 | [SectionLabel](tasks/M2-T08-section-label.md)                     | `SectionLabel.tsx` + `.module.css`                                                              |
| T09 | [StatCard](tasks/M2-T09-stat-card.md)                             | `StatCard.tsx` + `.module.css`                                                                  |

---

## Milestone 3 — Complex Components (`src/components/complex/`)

| #   | Task                                                              | Files                                                            |
|-----|-------------------------------------------------------------------|------------------------------------------------------------------|
| T10 | [ThreeDotMenu (context dropdown)](tasks/M3-T10-three-dot-menu.md) | `ThreeDotMenu/ThreeDotMenu.tsx` + `.module.css`                  |
| T11 | [ServiceCard (kanban card)](tasks/M3-T11-service-card.md)         | rebuild `src/components/service/ServiceCard.tsx` + `.module.css` |
| T12 | [ServiceListRow (table row)](tasks/M3-T12-service-list-row.md)    | `src/components/service/ServiceListRow.tsx` + `.module.css`      |
| T13 | [NodeCard](tasks/M3-T13-node-card.md)                             | `src/components/node/NodeCard.tsx` + `.module.css`               |
| T14 | [VCNPeerRow](tasks/M3-T14-vcn-peer-row.md)                        | `src/components/vcn/VCNPeerRow.tsx` + `.module.css`              |
| T15 | [CodeBlock](tasks/M3-T15-code-block.md)                           | `src/components/complex/CodeBlock/CodeBlock.tsx` + `.module.css` |

---

## Milestone 4 — Widgets (`src/widgets/`)

| #   | Task                                                              | Files                                               |
|-----|-------------------------------------------------------------------|-----------------------------------------------------|
| T16 | [Sidebar](tasks/M4-T16-sidebar.md)                                | `sidebar/Sidebar.tsx` + `.module.css`               |
| T17 | [TopBar](tasks/M4-T17-top-bar.md)                                 | `topbar/TopBar.tsx` + `.module.css`                 |
| T18 | [DeploymentFilters (toolbar)](tasks/M4-T18-deployment-filters.md) | `deployments/DeploymentFilters.tsx` + `.module.css` |
| T19 | [KanbanBoard](tasks/M4-T19-kanban-board.md)                       | `deployments/KanbanBoard.tsx` + `.module.css`       |
| T20 | [ServiceListView](tasks/M4-T20-service-list-view.md)              | `deployments/ServiceListView.tsx` + `.module.css`   |
| T21 | [NodeHealthList](tasks/M4-T21-node-health-list.md)                | `controlplane/NodeHealthList.tsx` + `.module.css`   |
| T22 | [PluginMatrix](tasks/M4-T22-plugin-matrix.md)                     | `controlplane/PluginMatrix.tsx` + `.module.css`     |
| T23 | [NetworkTopologyMap](tasks/M4-T23-network-topology.md)            | `vcn/NetworkTopologyMap.tsx` + `.module.css`        |
| T24 | [VCNPeerTable](tasks/M4-T24-vcn-peer-table.md)                    | `vcn/VCNPeerTable.tsx` + `.module.css`              |

---

## Milestone 5 — Layout / Segments

| #   | Task                                                             | Files                           |
|-----|------------------------------------------------------------------|---------------------------------|
| T25 | [MainLayout (sidebar+topbar shell)](tasks/M5-T25-main-layout.md) | `src/app/router/MainLayout.tsx` |

---

## Milestone 6 — Pages

| #   | Task                                                             | Files                                                         |
|-----|------------------------------------------------------------------|---------------------------------------------------------------|
| T26 | [ControlPlanePage (rebuild)](tasks/M6-T26-control-plane-page.md) | `src/pages/controlplane/ControlPlanePage.tsx` + `.module.css` |
| T27 | [DeploymentsPage (rebuild)](tasks/M6-T27-deployments-page.md)    | `src/pages/deployments/DeploymentsPage.tsx` + `.module.css`   |
| T28 | [VCNPage (rebuild)](tasks/M6-T28-vcn-page.md)                    | `src/pages/vcn/VervClosedNetworkPage.tsx` + `.module.css`     |
| T29 | [SearchPage](tasks/M6-T29-search-page.md)                        | `src/pages/search/SearchPage.tsx` + `.module.css`             |

---

## Coding Rules (must be followed in every task)

- **Function declarations only** — no `const Foo = () => {}`. Use `function Foo() {}`.
- **Named handlers** — all event handlers inside components must also be named function declarations.
- **One file = one component** (except chips grouped in T05).
- **CSS Modules** — all styles in `.module.css`. No inline `style={{}}` for new code.
- **Root class suffix** is `Container`; wrapper classes use `Wrapper`.
- **`@/`** resolves to `src/` — use it for all imports.
- **`rem` units** for font sizes and spacing. Never hardcode `px` for spacing/font.
- **No `!important`**, no `z-index` hardcoding.
- **CSS nesting** for child selectors inside a module.
- **Animations**: CSS `transition`/`@keyframes` first; use `framer-motion` only if CSS cannot achieve the effect.
- **No comments** unless the WHY is non-obvious.
