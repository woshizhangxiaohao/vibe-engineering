/**
 * Vibe Router - AI 需求复杂度分析
 *
 * 分析 Issue 复杂度并返回分类结果
 */

const fs = require('fs');

module.exports = async ({ github, context, core, promptContent }) => {
  const issue = context.payload.issue;
  const title = issue.title || '';
  const body = issue.body || '';

  console.log("=".repeat(60));
  console.log("🔍 VIBE ROUTER - 分析需求复杂度");
  console.log("=".repeat(60));
  console.log(`Issue #${issue.number}: ${title}`);

  // 从配置文件读取配置（如果存在）
  let config;
  try {
    const configPath = '.github/config/workflow-config.json';
    const configContent = fs.readFileSync(configPath, 'utf8');
    config = JSON.parse(configContent);
  } catch (error) {
    console.warn(`⚠️ 无法读取配置文件，使用默认配置: ${error.message}`);
    config = {
      agents: { router_model: "google/gemini-2.0-flash-001" }
    };
  }

  // 使用从模板加载的 prompt
  const prompt = promptContent;
  console.log("📄 使用模板: router/complexity-analyzer.md");

  // API 调用辅助函数（带重试）
  const MAX_RETRIES = 3;
  const RETRY_DELAY_MS = 1000;

  const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));
  const getBackoffDelay = (attempt) => {
    const baseDelay = RETRY_DELAY_MS * Math.pow(2, attempt);
    const jitter = Math.random() * 500;
    return Math.min(baseDelay + jitter, 30000);
  };

  const isRetryableError = (status) => {
    return status === 429 || (status >= 500 && status < 600);
  };

  async function callAPIWithRetry() {
    let lastError = null;

    for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
      try {
        const response = await fetch("https://openrouter.ai/api/v1/chat/completions", {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${process.env.OPENROUTER_API_KEY}`,
            "Content-Type": "application/json",
            "HTTP-Referer": "https://github.com/lessthanno/vibe-engineering-playbook",
            "X-Title": "Vibe Router"
          },
          body: JSON.stringify({
            model: config.agents?.router_model || "google/gemini-2.0-flash-001",
            messages: [{ role: "user", content: prompt }],
            temperature: 0.1,
            response_format: { type: "json_object" }
          })
        });

        if (response.status === 429) {
          const retryAfter = response.headers.get('Retry-After');
          const waitTime = retryAfter ? parseInt(retryAfter) * 1000 : getBackoffDelay(attempt);
          console.log(`⚠️ Rate limited. Waiting ${waitTime}ms before retry...`);
          await sleep(waitTime);
          continue;
        }

        if (!response.ok) {
          if (isRetryableError(response.status) && attempt < MAX_RETRIES) {
            const waitTime = getBackoffDelay(attempt);
            console.log(`⚠️ Attempt ${attempt + 1} failed (${response.status}). Retrying in ${waitTime}ms...`);
            await sleep(waitTime);
            continue;
          }
          throw new Error(`API 请求失败: ${response.status}`);
        }

        return await response.json();

      } catch (error) {
        lastError = error;
        if (attempt < MAX_RETRIES) {
          const waitTime = getBackoffDelay(attempt);
          console.log(`⚠️ Attempt ${attempt + 1} failed: ${error.message}. Retrying in ${waitTime}ms...`);
          await sleep(waitTime);
        }
      }
    }

    throw lastError || new Error('API call failed after all retries');
  }

  try {
    const data = await callAPIWithRetry();
    let result = data.choices?.[0]?.message?.content || '';

    // 提取 JSON
    const jsonMatch = result.match(/\{[\s\S]*\}/);
    if (jsonMatch) {
      result = JSON.parse(jsonMatch[0]);
    } else {
      throw new Error("无法解析 AI 响应");
    }

    console.log("✅ AI 分析结果:", JSON.stringify(result, null, 2));

    const complexity = result.complexity || 'M';
    const reasoning = result.reasoning || '默认中等复杂度';
    const areas = result.affected_areas || [];

    core.setOutput('complexity', complexity);
    core.setOutput('reasoning', reasoning);

    // 添加复杂度标签（使用配置文件中的标签名）
    const labelMap = config.labels?.complexity || {
      'S': 'complexity:simple',
      'M': 'complexity:medium',
      'L': 'complexity:complex',
      'skip': 'needs-triage'
    };

    // 映射复杂度到标签
    const complexityLabelMap = {
      'S': labelMap.simple || 'complexity:simple',
      'M': labelMap.medium || 'complexity:medium',
      'L': labelMap.complex || 'complexity:complex',
      'skip': 'needs-triage'
    };

    const labels = [complexityLabelMap[complexity] || 'complexity:medium'];

    // 添加影响区域标签
    if (areas.includes('frontend')) labels.push(config.labels?.scope?.frontend || 'frontend');
    if (areas.includes('backend')) labels.push(config.labels?.scope?.backend || 'backend');
    if (areas.includes('database')) labels.push(config.labels?.scope?.database || 'database');

    // 移除 needs-route 标签（已完成路由）
    try {
      await github.rest.issues.removeLabel({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: issue.number,
        name: 'needs-route'
      });
    } catch (e) {
      // 标签可能不存在，忽略
    }

    await github.rest.issues.addLabels({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issue.number,
      labels: labels
    });

    // 发布分析评论
    const complexityEmoji = {
      'S': '🟢',
      'M': '🟡',
      'L': '🔴',
      'skip': '⏭️'
    };

    const complexityName = {
      'S': '简单任务',
      'M': '中等任务',
      'L': '复杂任务',
      'skip': '跳过'
    };

    await github.rest.issues.createComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issue.number,
      body: [
        `${complexityEmoji[complexity]} **Vibe Router 分析结果**`,
        "",
        `**复杂度**: ${complexityName[complexity]} (${complexity})`,
        `**原因**: ${reasoning}`,
        areas.length > 0 ? `**影响范围**: ${areas.join(', ')}` : '',
        result.estimated_hours ? `**预估工时**: ${result.estimated_hours} 小时` : '',
        "",
        "---",
        "",
        complexity === 'S' ? "🚀 **简单任务 Agent 已自动触发**，请稍候查看 PR。" :
        complexity === 'M' ? "🔧 **中等任务处理中**，AI 将先分析再开发。" :
        complexity === 'L' ? "📋 **复杂任务需要拆分**，AI 将生成子 Issue 列表。" :
        "⏸️ 此 Issue 需要人工处理。",
        "",
        "> 💡 如需手动触发，可使用 `/agent-simple`、`/agent-medium` 或 `/agent-complex` 命令。"
      ].filter(Boolean).join("\n")
    });

    return { complexity, reasoning, areas };

  } catch (error) {
    console.error("❌ 分析失败:", error.message);

    // 失败时默认为中等复杂度
    core.setOutput('complexity', 'M');
    core.setOutput('reasoning', '自动分析失败，默认中等复杂度');

    await github.rest.issues.addLabels({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issue.number,
      labels: ['complexity:medium', 'needs-review']
    });

    return { complexity: 'M', reasoning: '自动分析失败，默认中等复杂度' };
  }
};
