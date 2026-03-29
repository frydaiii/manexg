## 项目概述

BanExg 是一个用Go语言开发的多交易所统一SDK类库，旨在为加密货币交易提供统一的API接口。目前已完整支持Binance交易所，部分支持Bybit交易所，并为后续接入更多交易所（如OKX）奠定了坚实的架构基础。

## 一、整体文件组织架构

### 1.1 项目根目录结构
```
banexg/
├── 核心接口层
│   ├── intf.go              # 核心接口定义 (BanExchange)
│   ├── types.go             # 核心数据结构 (Exchange, Market, Order等)
│   ├── base.go              # 基础功能实现
│   ├── biz.go               # 通用业务逻辑
│   ├── common.go            # 通用工具函数
│   ├── data.go              # 常量和配置数据
│   ├── exts.go              # 扩展工具函数
│   └── websocket.go         # WebSocket客户端实现
├── 交易所实现层
│   ├── binance/             # Binance交易所完整实现
│   ├── bybit/               # Bybit交易所部分实现
│   └── china/               # 中国区域交易所（模拟）
├── 基础设施层
│   ├── utils/               # 工具函数库
│   ├── log/                 # 日志系统
│   ├── errs/                # 错误处理
│   └── bex/                 # 交易所工厂注册
├── 测试和文档
│   ├── */testdata/          # 测试数据
│   ├── */readme.md          # 各模块文档
│   └── contribute.md        # 贡献指南
└── 配置文件
    ├── go.mod               # Go模块配置
    └── */local.json         # 本地配置文件
```

### 1.2 架构分层设计

**四层架构模式：**
1. **接口抽象层** - 定义统一的交易所操作接口
2. **业务逻辑层** - 实现通用的交易业务逻辑
3. **适配器层** - 各交易所的具体实现适配
4. **基础设施层** - 工具、日志、错误处理等支撑服务

## 二、核心结构体和接口

### 2.1 最重要的核心接口

#### BanExchange接口 (`intf.go`)
**位置：** `intf.go`

**作用：** 定义所有交易所必须实现的统一接口，包含100+个方法，涵盖：
- **市场数据获取**：`LoadMarkets`, `FetchTicker`, `FetchOHLCV`, `FetchOrderBook`
- **交易操作**：`CreateOrder`, `CancelOrder`, `EditOrder`, `FetchOrder`
- **账户管理**：`FetchBalance`, `FetchPositions`, `SetLeverage`
- **实时数据**：`WatchOrderBooks`, `WatchTrades`, `WatchBalance`
- **工具方法**：`CalculateFee`, `CalcMaintMargin`, `PrecAmount`

### 2.2 最重要的核心结构体

#### Exchange结构体 (`types.go:59`)
**位置：** `types.go:59`

**作用：** 交易所的基础实现结构，所有具体交易所都嵌入此结构
**关键字段：**
- `ExgInfo`: 交易所基本信息（ID、名称、版本）
- `Apis`: API端点映射表
- `Accounts`: 多账户管理
- `Markets`: 市场信息缓存
- `WSClients`: WebSocket客户端池
- `Options`: 配置选项管理


## 三、项目整体架构风格

### 3.1 设计模式运用

- 核心`BanExchange`接口定义统一标准
- 通过`bex`模块支持动态交易所注册
- 可配置的重试策略和错误处理
- 基础`Exchange`提供通用业务流程
- 统一的初始化和配置流程
- 使用标准的状态和返回格式，各交易所按需转换

## 四、重要开发流程

### 4.1 添加新接口
1. 在`entry.go`下`Apis`添加对应的定义
2. 在`data.go`下添加`MethodXXX`方法名常量
3. 在合适位置如`biz_order_create.go`下添加对应的业务逻辑实现
4. 在`types.go`下添加返回数据类型定义


## AI准则
- 严格遵守DRY准则，Dont Repeat Yourself，添加新代码前检查是否已有相似代码，有则提取为子函数
- 始终用最少的代码完成任务，只生成当前需要的核心代码，不要过度设计，不要提前生成以后可能需要的代码。

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **manexg** (3757 symbols, 12639 relationships, 300 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## When Debugging

1. `gitnexus_query({query: "<error or symptom>"})` — find execution flows related to the issue
2. `gitnexus_context({name: "<suspect function>"})` — see all callers, callees, and process participation
3. `READ gitnexus://repo/manexg/process/{processName}` — trace the full execution flow step by step
4. For regressions: `gitnexus_detect_changes({scope: "compare", base_ref: "main"})` — see what your branch changed

## When Refactoring

- **Renaming**: MUST use `gitnexus_rename({symbol_name: "old", new_name: "new", dry_run: true})` first. Review the preview — graph edits are safe, text_search edits need manual review. Then run with `dry_run: false`.
- **Extracting/Splitting**: MUST run `gitnexus_context({name: "target"})` to see all incoming/outgoing refs, then `gitnexus_impact({target: "target", direction: "upstream"})` to find all external callers before moving code.
- After any refactor: run `gitnexus_detect_changes({scope: "all"})` to verify only expected files changed.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Tools Quick Reference

| Tool | When to use | Command |
|------|-------------|---------|
| `query` | Find code by concept | `gitnexus_query({query: "auth validation"})` |
| `context` | 360-degree view of one symbol | `gitnexus_context({name: "validateUser"})` |
| `impact` | Blast radius before editing | `gitnexus_impact({target: "X", direction: "upstream"})` |
| `detect_changes` | Pre-commit scope check | `gitnexus_detect_changes({scope: "staged"})` |
| `rename` | Safe multi-file rename | `gitnexus_rename({symbol_name: "old", new_name: "new", dry_run: true})` |
| `cypher` | Custom graph queries | `gitnexus_cypher({query: "MATCH ..."})` |

## Impact Risk Levels

| Depth | Meaning | Action |
|-------|---------|--------|
| d=1 | WILL BREAK — direct callers/importers | MUST update these |
| d=2 | LIKELY AFFECTED — indirect deps | Should test |
| d=3 | MAY NEED TESTING — transitive | Test if critical path |

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/manexg/context` | Codebase overview, check index freshness |
| `gitnexus://repo/manexg/clusters` | All functional areas |
| `gitnexus://repo/manexg/processes` | All execution flows |
| `gitnexus://repo/manexg/process/{name}` | Step-by-step execution trace |

## Self-Check Before Finishing

Before completing any code modification task, verify:
1. `gitnexus_impact` was run for all modified symbols
2. No HIGH/CRITICAL risk warnings were ignored
3. `gitnexus_detect_changes()` confirms changes match expected scope
4. All d=1 (WILL BREAK) dependents were updated

## Keeping the Index Fresh

After committing code changes, the GitNexus index becomes stale. Re-run analyze to update it:

```bash
npx gitnexus analyze
```

If the index previously included embeddings, preserve them by adding `--embeddings`:

```bash
npx gitnexus analyze --embeddings
```

To check whether embeddings exist, inspect `.gitnexus/meta.json` — the `stats.embeddings` field shows the count (0 means no embeddings). **Running analyze without `--embeddings` will delete any previously generated embeddings.**

> Claude Code users: A PostToolUse hook handles this automatically after `git commit` and `git merge`.

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
