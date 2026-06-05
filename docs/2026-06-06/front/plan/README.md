# 2026-06-06 前端开源栈重构文档

本目录整理 New Haven 前端重构的具体执行文档。核心思路是：不魔改完整开源游戏模板，而是把成熟开源模块接入现有 React + Vite + PixiJS 前端，让项目获得更稳定的 UI、数据、表单、图表、Mock 和测试基础。

## 阅读顺序

1. `01-frontend-open-source-stack.md`
   - 说明每个开源库的用途、边界和暂不引入项。
2. `02-frontend-architecture-boundaries.md`
   - 固化 React、PixiJS、TanStack Query、Zustand、API client 的职责边界。
3. `03-phase-1-implementation-plan.md`
   - 第一阶段具体任务：只建基础设施，不改业务页面行为。
4. `04-codex-frontend-refactor-prompt.md`
   - 可直接交给 Codex 的执行 Prompt，用于后续分阶段实现。

## 总原则

- 不找完整开源游戏来魔改，避免被外部架构绑架。
- 成熟库只负责基础设施，游戏业务仍保留在本项目自己的模块里。
- 每一步都要可构建、可回滚、可独立验收。
- 前端页面组件不直接散写 `fetch`，服务器数据统一通过 API client 和 TanStack Query 管理。
- PixiJS 只负责地图、建筑 sprite、地图交互和轻量动画，不直接调用业务 API。

