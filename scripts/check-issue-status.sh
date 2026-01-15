#!/bin/bash

# InsightFlow 功能实现状态检测脚本
# 检测代码中各功能模块的实现情况

echo "# InsightFlow 实现状态检测报告"
echo ""
echo "检测时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 颜色定义（用于终端输出）
GREEN="✅"
YELLOW="⚠️"
RED="❌"

# 计数器
total_checks=0
passed_checks=0

check_file() {
    local file=$1
    local desc=$2
    total_checks=$((total_checks + 1))
    if [ -f "$file" ]; then
        passed_checks=$((passed_checks + 1))
        echo "$GREEN $desc"
        echo "   文件: $file"
        return 0
    else
        echo "$RED $desc"
        echo "   缺失: $file"
        return 1
    fi
}

check_pattern() {
    local pattern=$1
    local file=$2
    local desc=$3
    total_checks=$((total_checks + 1))
    if grep -q "$pattern" "$file" 2>/dev/null; then
        passed_checks=$((passed_checks + 1))
        echo "$GREEN $desc"
        return 0
    else
        echo "$RED $desc"
        return 1
    fi
}

echo "---"
echo ""
echo "## Issue #179: 内容解析页面 (Insight Canvas)"
echo ""

echo "### 后端检测"
check_file "backend/internal/handlers/insight.go" "Insight Handler"
check_pattern "GetInsightDetail" "backend/internal/handlers/insight.go" "GET /api/v1/insights/:id API"
check_pattern "ProcessInsight\|Process" "backend/internal/handlers/insight.go" "异步处理流程"
check_file "backend/internal/repository/insight.go" "Insight Repository"

echo ""
echo "### 前端检测"
check_file "frontend/components/insights/InsightCanvas.tsx" "InsightCanvas 组件"
check_file "frontend/components/insights/VideoPreview.tsx" "VideoPreview 组件"
check_file "frontend/components/insights/SummarySection.tsx" "SummarySection 组件"
check_file "frontend/components/insights/TranscriptView.tsx" "TranscriptView 组件"
check_pattern "displayMode.*zh.*en.*bilingual\|LanguageToggle\|语言切换" "frontend/components/insights/TranscriptView.tsx" "语言切换功能"
check_pattern "onTimestampClick\|seekTo" "frontend/components/insights/TranscriptView.tsx" "时间戳点击跳转"

echo ""
echo "---"
echo ""
echo "## Issue #178: 时间轴导航栏 (Memory Rail)"
echo ""

check_file "frontend/components/insights/MemoryRail.tsx" "MemoryRail 组件"
check_pattern "Today\|今日\|Yesterday\|昨日" "frontend/components/insights/MemoryRail.tsx" "时间分组显示"
check_pattern "search\|Search\|搜索" "frontend/components/insights/MemoryRail.tsx" "搜索功能"

echo ""
echo "---"
echo ""
echo "## Issue #180: 滑词交互功能"
echo ""

check_pattern "Highlight\|highlight" "backend/internal/handlers/insight.go" "Highlight API (后端)"
check_pattern "CreateHighlight\|AddHighlight" "backend/internal/handlers/insight.go" "创建高亮 API"

echo ""
echo "---"
echo ""
echo "## Issue #181: AI 对话面板 (Chat Console)"
echo ""

echo "### 后端"
check_pattern "GetChatMessages\|chat" "backend/internal/handlers/insight.go" "Chat API 端点"
check_pattern "CreateChatMessage\|SendMessage" "backend/internal/handlers/insight.go" "发送消息 API"

echo ""
echo "### 前端"
if [ -f "frontend/components/insights/ChatPanel.tsx" ] || [ -f "frontend/components/insights/AIChat.tsx" ] || [ -f "frontend/components/insights/ChatConsole.tsx" ]; then
    passed_checks=$((passed_checks + 1))
    total_checks=$((total_checks + 1))
    echo "$GREEN Chat 前端组件"
else
    total_checks=$((total_checks + 1))
    echo "$RED Chat 前端组件"
    echo "   缺失: frontend/components/insights/ChatPanel.tsx 或类似组件"
fi

# 检查 API 客户端
check_pattern "chat\|Chat" "frontend/lib/api/endpoints.ts" "Chat API 客户端定义"

echo ""
echo "---"
echo ""
echo "## Issue #182: 笔记分享功能"
echo ""

echo "### 后端"
check_pattern "GetSharedInsight\|shared" "backend/internal/handlers/insight.go" "分享 API 端点"
check_pattern "ShareToken\|share_token" "backend/internal/models/insight.go" "分享数据模型"

echo ""
echo "### 前端"
if [ -f "frontend/components/insights/ShareDialog.tsx" ] || [ -f "frontend/components/insights/ShareButton.tsx" ]; then
    passed_checks=$((passed_checks + 1))
    total_checks=$((total_checks + 1))
    echo "$GREEN 分享前端组件"
else
    total_checks=$((total_checks + 1))
    echo "$RED 分享前端组件"
    echo "   缺失: frontend/components/insights/ShareDialog.tsx 或类似组件"
fi

echo ""
echo "---"
echo ""
echo "## 📊 总体统计"
echo ""
percent=$((passed_checks * 100 / total_checks))
echo "通过检测: $passed_checks / $total_checks ($percent%)"
echo ""

if [ $percent -ge 90 ]; then
    echo "状态: $GREEN 基本完成"
elif [ $percent -ge 60 ]; then
    echo "状态: $YELLOW 部分完成"
else
    echo "状态: $RED 待开发"
fi

echo ""
echo "---"
echo ""
echo "## 🔍 待完成项建议"
echo ""

# 根据检测结果给出建议
if ! grep -q "ChatPanel\|AIChat\|ChatConsole" frontend/components/insights/*.tsx 2>/dev/null; then
    echo "- [ ] **Issue #181**: 需要创建 Chat 前端组件 (ChatPanel.tsx)"
fi

if ! grep -q "ShareDialog\|ShareButton" frontend/components/insights/*.tsx 2>/dev/null; then
    echo "- [ ] **Issue #182**: 需要创建分享前端组件 (ShareDialog.tsx)"
fi

if ! grep -q "chat" frontend/lib/api/endpoints.ts 2>/dev/null; then
    echo "- [ ] **Issue #181**: 需要在 endpoints.ts 中添加 Chat API 定义"
fi
