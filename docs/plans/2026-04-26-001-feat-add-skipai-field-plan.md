---
title: feat: Add SkipAI field to UriSpot for AI API tagging control
type: feat
status: completed
date: 2026-04-26
---

# feat: Add SkipAI field to UriSpot for AI API tagging control

## Overview

为 `UriSpot` 结构体添加 `SkipAI` 字段，允许在特定 API 操作（Create/Update/Delete）的文档注释中添加 `// @Tags skipai` 标记，用于控制 AI 行为。

## Problem Statement / Motivation

现有的 `UriSpot.Ignore` 字段用于忽略生成特定操作的代码，但缺少一种机制来在生成的 API 文档中添加特殊标记（如 `@Tags skipai`）来告诉 AI 跳过某些接口。需要在文档注释生成时（Summary 和 Description 附近）添加 `// @Tags skipai` 注释。

## Proposed Solution

### 1. 在 `UriSpot` 结构体添加 `SkipAI` 字段

参考现有的 `Ignore` 字段模式：

```go
// scripts/codegen/gens/type_webapi.go:55-75
type UriSpot struct {
    // ... existing fields ...
    SkipAI string `yaml:"skipAI,omitempty"`
}
```

### 2. 定义操作字符映射

在文件顶部添加 `SkipAI` 操作字符常量定义（参考现有的 `msmethods` 和 `mslabels`）：

```go
var skipaiActions = map[string]string{
    "List":   "L",
    "Get":    "G",
    "Create": "C",
    "Update": "U",
    "Delete": "D",
}
```

### 3. 在 `Handle` 结构体继承 SkipAI

由于 `Handle` 使用 `yaml:",inline"` 嵌入 `UriSpot`，`SkipAI` 会自动继承。需要在 `genHandle` 方法中将 `SkipAI` 传递到 `Handle` 结构体。

### 4. 修改 `CommentCodes` 方法添加 skipai 标签

在生成文档注释时（@Summary 附近），检查当前操作是否在 `SkipAI` 中，如果是则添加 `// @Tags skipai` 注释。

插入位置：`@Summary` 注释之后（line 488 附近）。

```go
// 如果当前操作需要跳过 AI，则添加 @Tags skipai
if h.shouldSkipAI() {
    st.Comment("@Tags skipai").Line()
}
```

### 5. 添加辅助方法 `shouldSkipAI()`

```go
func (h *Handle) shouldSkipAI() bool {
    if len(h.SkipAI) == 0 {
        return false
    }
    char, ok := skipaiActions[h.act]
    if !ok {
        return false
    }
    return strings.ContainsRune(h.SkipAI, rune(char[0]))
}
```

## Technical Considerations

- **继承关系**：`UriSpot.SkipAI` 通过 `yaml:",inline"` 嵌入到 `Handle`，无需单独定义
- **向后兼容**：`skipAI` 字段是可选的（`omitempty`），不影响现有配置
- **插入位置**：选择在 `@Summary` 之后添加，与 `// @Tags skipai` 保持一致的注释风格
- **字符映射**：使用单字符标识，与 `Ignore` 的 `CU` 模式保持一致

## System-Wide Impact

- **API 文档生成**：影响 `type_webapi.go` 的文档注释生成逻辑
- **无运行时行为变更**：仅影响代码生成阶段

## Acceptance Criteria

- [ ] `UriSpot` 结构体添加 `SkipAI string` 字段，带 `yaml:"skipAI,omitempty"` tag
- [ ] 添加 `skipaiActions` 映射表：`L=List, G=Get, C=Create, U=Update, D=Delete`
- [ ] `Handle` 继承 `SkipAI` 字段（通过 `UriSpot` inline 嵌入自动获得）
- [ ] 添加 `shouldSkipAI()` 辅助方法检查当前操作是否需要跳过 AI
- [ ] 修改 `CommentCodes` 方法，在 `@Summary` 之后添加 `@Tags skipai` 注释
- [ ] 当配置 `skipAI: "CUD"` 时，Create/Update/Delete 接口文档会生成 `// @Tags skipai`
- [ ] 配置 `skipAI: "L"` 时，List 接口会生成 `// @Tags skipai`
- [ ] 配置 `skipAI: "G"` 时，Get 接口会生成 `// @Tags skipai`
- [ ] 不配置 `skipAI` 或为空时，不生成 `@Tags skipai` 注释

## Implementation Phases

### Phase 1: 添加 SkipAI 字段和映射表

- 在 `UriSpot` 结构体添加 `SkipAI string` 字段
- 在文件顶部添加 `skipaiActions` 映射表

### Phase 2: 添加辅助方法和修改 CommentCodes

- 添加 `shouldSkipAI()` 方法
- 修改 `CommentCodes` 方法在 `@Summary` 之后添加条件判断

### Phase 3: 测试验证

- 编写测试用例验证各种组合（L, G, C, U, D, CU, CUD 等）

## Context

- **参考实现**：`UriSpot.Ignore` 字段（line 60, 209-214）提供了类似的模式
- **相关代码**：`CommentCodes` 方法（line 453-560）负责生成 API 文档注释
- **操作方法**：`cutMethod` 函数从 `Method` 中提取 action（line 368-371）
