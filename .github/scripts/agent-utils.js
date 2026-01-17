/**
 * Agent 共享工具函数
 *
 * 提供给 agent-simple.yml, agent-medium.yml 等共享使用的函数
 */

/**
 * 更新 Issue 进度评论
 */
async function updateProgressComment(github, context, {
  issueNumber,
  progressCommentId,
  stage,
  status,
  percentage,
  details = '',
  prInfo = '',
  actionUrl = ''
}) {
  if (!progressCommentId) return;

  const progressBar = '█'.repeat(Math.floor(percentage / 5)) + '░'.repeat(20 - Math.floor(percentage / 5));

  const body = [
    `## ${stage}`,
    "",
    `**当前阶段**: ${status}`,
    "",
    "```",
    `${progressBar} ${percentage}%`,
    "```",
    "",
    details,
    prInfo,
    actionUrl ? `\n**Actions 日志**: [查看详情](${actionUrl})` : ''
  ].filter(Boolean).join("\n");

  try {
    await github.rest.issues.updateComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      comment_id: progressCommentId,
      body: body
    });
  } catch (e) {
    console.log("无法更新进度评论:", e.message);
  }
}

/**
 * 更新 Issue 标签
 */
async function updateIssueLabels(github, context, {
  issueNumber,
  removeLabels = [],
  addLabels = []
}) {
  // 移除标签
  for (const label of removeLabels) {
    try {
      await github.rest.issues.removeLabel({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: issueNumber,
        name: label
      });
    } catch (e) {
      // 标签可能不存在，忽略
    }
  }

  // 添加标签
  if (addLabels.length > 0) {
    await github.rest.issues.addLabels({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      labels: addLabels
    });
  }
}

/**
 * 查找关联的 PR
 */
async function findRelatedPR(github, context, issueNumber) {
  try {
    const { data: prs } = await github.rest.pulls.list({
      owner: context.repo.owner,
      repo: context.repo.repo,
      state: 'all',
      sort: 'created',
      direction: 'desc',
      per_page: 20
    });

    const relatedPR = prs.find(pr =>
      pr.title.includes(`#${issueNumber}`) ||
      pr.title.toLowerCase().includes(`issue-${issueNumber}`) ||
      pr.head.ref.includes(`issue-${issueNumber}`) ||
      pr.head.ref.includes(`${issueNumber}`) ||
      pr.body?.includes(`#${issueNumber}`) ||
      pr.body?.includes(`Closes #${issueNumber}`) ||
      pr.body?.includes(`Fixes #${issueNumber}`)
    );

    if (relatedPR) {
      console.log(`✅ 找到关联 PR: #${relatedPR.number}`);
      return relatedPR;
    }

    console.log(`⚠️ 未找到关联 PR`);
    return null;
  } catch (e) {
    console.log("获取 PR 信息失败:", e.message);
    return null;
  }
}

/**
 * 检查重复运行
 */
async function checkDuplicateRun(github, context, {
  workflowId,
  issueNumber,
  currentRunId,
  timeWindowMs = 15 * 60 * 1000
}) {
  // 检查是否有同一个 workflow 正在运行
  try {
    const { data: runs } = await github.rest.actions.listWorkflowRuns({
      owner: context.repo.owner,
      repo: context.repo.repo,
      workflow_id: workflowId,
      status: 'in_progress',
      per_page: 20
    });

    for (const run of runs.workflow_runs) {
      if (run.id === currentRunId) continue;

      const runTime = new Date(run.created_at);
      const now = new Date();
      const timeDiff = Math.abs(now - runTime);

      if (timeDiff < timeWindowMs) {
        console.log(`⚠️ 发现可能的重复运行 (run_id: ${run.id})`);
        return true;
      }
    }
  } catch (e) {
    console.log(`检查重复运行失败: ${e.message}`);
  }

  // 检查 issue 是否已经在处理中
  try {
    const { data: issue } = await github.rest.issues.get({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber
    });

    const labels = issue.labels.map(l => l.name);
    if (labels.includes('ai:processing')) {
      console.log(`⚠️ Issue #${issueNumber} 已在处理中`);
      return true;
    }
  } catch (e) {
    console.log(`检查 issue 状态失败: ${e.message}`);
  }

  return false;
}

/**
 * 分析错误类型并给出建议
 */
function analyzeError(errorMessage) {
  if (errorMessage.includes("Edit") || errorMessage.includes("old_string") || errorMessage.includes("not found")) {
    return {
      type: "代码编辑失败 - 目标代码块未找到（可能已被修改）",
      suggestion: "检查目标文件是否已更新，或手动完成此任务"
    };
  }

  if (errorMessage.includes("timeout") || errorMessage.includes("Timeout")) {
    return {
      type: "执行超时 - 任务过于复杂",
      suggestion: "使用 `/agent-complex` 将任务拆分为多个子任务"
    };
  }

  if (errorMessage.includes("rate limit") || errorMessage.includes("429") || errorMessage.includes("Too Many Requests")) {
    return {
      type: "API 请求限流",
      suggestion: "等待几分钟后重试"
    };
  }

  if (errorMessage.includes("permission") || errorMessage.includes("Permission")) {
    return {
      type: "权限不足",
      suggestion: "检查 GitHub Token 权限配置"
    };
  }

  if (errorMessage.includes("build") || errorMessage.includes("compile")) {
    return {
      type: "构建失败 - 代码存在语法或编译错误",
      suggestion: "查看 Actions 日志定位具体错误"
    };
  }

  return {
    type: "执行过程中出现错误",
    suggestion: "查看 Actions 日志获取详细信息"
  };
}

/**
 * 获取 Issue 信息
 */
async function getIssueInfo(github, context, inputIssueNumber) {
  let issueNumber;
  let issue;

  if (context.eventName === 'workflow_dispatch' && inputIssueNumber) {
    issueNumber = parseInt(inputIssueNumber);
    const { data } = await github.rest.issues.get({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber
    });
    issue = data;
  } else {
    issue = context.payload.issue;
    issueNumber = issue.number;
  }

  return {
    issueNumber,
    title: issue.title,
    body: issue.body || ''
  };
}

/**
 * 创建或查找进度追踪评论
 */
async function findOrCreateProgressComment(github, context, {
  issueNumber,
  initialBody,
  searchPattern = '🔄 自动迭代'
}) {
  // 查找已有的进度评论
  const { data: comments } = await github.rest.issues.listComments({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: issueNumber,
    per_page: 100
  });

  for (let i = comments.length - 1; i >= 0; i--) {
    const comment = comments[i];
    if (comment.body && comment.body.includes(searchPattern)) {
      console.log(`找到进度追踪评论: #${comment.id}`);
      return { id: comment.id, body: comment.body, isNew: false };
    }
  }

  // 创建新评论
  const newComment = await github.rest.issues.createComment({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: issueNumber,
    body: initialBody
  });

  console.log(`创建新的进度追踪评论: #${newComment.data.id}`);
  return { id: newComment.data.id, body: initialBody, isNew: true };
}

module.exports = {
  updateProgressComment,
  updateIssueLabels,
  findRelatedPR,
  checkDuplicateRun,
  analyzeError,
  getIssueInfo,
  findOrCreateProgressComment
};
