# Velez PaaS — Roadmap

Goal: evolve Velez into a self-contained PaaS deployment manager accessible from a browser.

## Milestones

| #  | Name                                                        | Status      | Details                                                     |
|----|-------------------------------------------------------------|-------------|-------------------------------------------------------------|
| M1 | [Core Platform](milestones/m1-core-platform/overview.md)    | In Progress | Services, deployments, deploy flow, settings, Verv services |
| M2 | [Cluster & Networking](milestones/m2-cluster-networking.md) | Backlog     | Control plane, VCN, multi-node, node scheduling tags        |
| M3 | [Observability](milestones/m3-observability.md)             | Backlog     | Logs, container metrics, deployment history                 |
| M4 | [Access & Multi-tenancy](milestones/m4-access-control.md)   | Backlog     | Auth, RBAC, namespaces                                      |
| M5 | [PaaS Automation](milestones/m5-paas-automation.md)         | Backlog     | Auto-rollback, health-gating, scaling policies              |

## UI Redesign (complete)

```mermaid
flowchart LR
    classDef done fill: #1a3a1a, stroke: #4caf50, color: #c8e6c9

    subgraph M1["M1 — Foundation"]
        T01["T01 Design Tokens"]
    end

    subgraph M2["M2 — Base Components"]
        T02["T02 StatusDot"]
        T03["T03 Badge"]
        T04["T04 MiniBar"]
        T05["T05 Chips"]
        T06["T06 IconButton"]
        T07["T07 Button"]
        T08["T08 SectionLabel"]
        T09["T09 StatCard"]
    end

    subgraph M3["M3 — Complex Components"]
        T10["T10 ThreeDotMenu"]
        T11["T11 ServiceCard"]
        T12["T12 ServiceListRow"]
        T13["T13 NodeCard"]
        T14["T14 VCNPeerRow"]
        T15["T15 CodeBlock"]
    end

    subgraph M4["M4 — Widgets"]
        T16["T16 Sidebar"]
        T17["T17 TopBar"]
        T18["T18 DeploymentFilters"]
        T19["T19 KanbanBoard"]
        T20["T20 ServiceListView"]
        T21["T21 NodeHealthList"]
        T22["T22 PluginMatrix"]
        T23["T23 NetworkTopologyMap"]
        T24["T24 VCNPeerTable"]
    end

    subgraph M5["M5 — Layout"]
        T25["T25 MainLayout"]
    end

    subgraph M6["M6 — Pages"]
        T26["T26 ControlPlanePage"]
        T27["T27 DeploymentsPage"]
        T28["T28 VCNPage"]
        T29["T29 SearchPage"]
    end

    M1 --> M2 --> M3 --> M4 --> M5 --> M6
class T01, T02, T03, T04, T05, T06, T07, T08, T09 done
class T10,T11, T12, T13, T14, T15 done
class T16,T17, T18, T19, T20, T21, T22, T23, T24 done
class T25,T26, T27, T28, T29 done
```