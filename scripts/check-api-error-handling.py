#!/usr/bin/env python3
"""
检查后端 API 错误处理是否符合规范

检查项：
1. Handler 是否使用了标准化的 ErrorResponse 格式（models.ErrorResponse）
2. 是否正确处理了 gorm.ErrRecordNotFound 错误
3. 错误日志是否包含了必要的字段（error_code, request_id）
4. 404 错误是否返回了正确的错误码
"""

import os
import re
import sys
from pathlib import Path
from typing import List, Dict, Tuple

# 检查结果
issues: List[Dict[str, str]] = []

def check_file(file_path: Path) -> List[Dict[str, str]]:
    """检查单个 Go 文件"""
    file_issues = []
    
    try:
        content = file_path.read_text(encoding='utf-8')
    except Exception as e:
        return [{
            'file': str(file_path),
            'line': 0,
            'severity': 'error',
            'message': f'无法读取文件: {e}'
        }]
    
    lines = content.split('\n')
    
    # 检查是否导入了 models 包
    has_models_import = 'models.ErrorResponse' in content or '"vibe-backend/internal/models"' in content
    
    # 检查是否导入了 gorm 包
    has_gorm_import = 'gorm.io/gorm' in content or 'gorm.ErrRecordNotFound' in content
    
    # 检查是否导入了 errors 包
    has_errors_import = 'errors' in content or '"errors"' in content
    
    # 1. 检查是否使用了 gin.H 而不是 models.ErrorResponse（仅针对错误状态码）
    # 需要检查多行，因为 gin.H 可能跨行
    for i, line in enumerate(lines, 1):
        # 检查错误状态码
        if re.search(r'http\.Status(?:BadRequest|NotFound|InternalServerError|Unauthorized|Forbidden)', line):
            # 检查接下来的几行是否使用了 gin.H
            next_lines = '\n'.join(lines[i-1:min(i+5, len(lines))])
            if 'gin.H' in next_lines and 'models.ErrorResponse' not in next_lines:
                # 检查是否包含错误相关的字段
                if 'error' in next_lines.lower() or '"error"' in next_lines or "'error'" in next_lines:
                    file_issues.append({
                        'file': str(file_path),
                        'line': i,
                        'severity': 'error',
                        'message': '使用了 gin.H 而不是 models.ErrorResponse，请使用标准化的错误响应格式'
                    })
    
    # 2. 检查错误响应是否包含 Code 字段
    error_response_pattern = r'c\.JSON\(http\.Status(?:BadRequest|NotFound|InternalServerError|Unauthorized|Forbidden),?\s*models\.ErrorResponse\s*\{'
    has_error_response = False
    for i, line in enumerate(lines, 1):
        if re.search(error_response_pattern, line):
            has_error_response = True
            # 检查接下来的几行是否包含 Code 字段
            next_lines = '\n'.join(lines[i-1:min(i+10, len(lines))])
            if 'Code:' not in next_lines and 'Code' not in next_lines:
                file_issues.append({
                    'file': str(file_path),
                    'line': i,
                    'severity': 'error',
                    'message': 'ErrorResponse 缺少 Code 字段'
                })
    
    # 3. 检查是否正确处理了 gorm.ErrRecordNotFound
    # 查找 First/Find 等数据库查询操作
    db_query_patterns = [
        r'\.First\([^)]+\)',
        r'\.Find\([^)]+\)',
        r'\.GetByID\([^)]+\)',
        r'\.GetBy[^(]+\([^)]+\)',
    ]
    
    has_record_not_found_check = False
    for pattern in db_query_patterns:
        for i, line in enumerate(lines, 1):
            if re.search(pattern, line):
                # 检查后续是否有错误处理
                # 查找 if err != nil 块
                err_check_start = None
                for j in range(i, min(i+20, len(lines))):
                    if 'if err != nil' in lines[j] or 'if err :=' in lines[j]:
                        err_check_start = j
                        break
                
                if err_check_start:
                    # 检查错误处理块中是否有 ErrRecordNotFound 检查
                    error_block = '\n'.join(lines[err_check_start:min(err_check_start+15, len(lines))])
                    if 'ErrRecordNotFound' in error_block or 'errors.Is' in error_block:
                        has_record_not_found_check = True
                        # 检查是否返回了正确的错误码
                        if 'StatusNotFound' in error_block:
                            if 'ANALYSIS_NOT_FOUND' not in error_block and 'INSIGHT_NOT_FOUND' not in error_block and 'NOT_FOUND' not in error_block:
                                file_issues.append({
                                    'file': str(file_path),
                                    'line': err_check_start + 1,
                                    'severity': 'warning',
                                    'message': '404 错误应该使用具体的错误码（如 ANALYSIS_NOT_FOUND, INSIGHT_NOT_FOUND）'
                                })
                    elif has_gorm_import and 'First' in line or 'Find' in line:
                        # 如果有 gorm 导入但没检查 ErrRecordNotFound
                        file_issues.append({
                            'file': str(file_path),
                            'line': i,
                            'severity': 'warning',
                            'message': '数据库查询后应该检查 gorm.ErrRecordNotFound 并返回 404 错误'
                        })
    
    # 4. 检查错误日志是否包含 error_code 字段
    log_error_pattern = r'h\.log\.Error\(|s\.log\.Error\(|log\.Error\('
    for i, line in enumerate(lines, 1):
        if re.search(log_error_pattern, line):
            # 检查是否包含 error_code
            next_lines = '\n'.join(lines[i-1:min(i+5, len(lines))])
            if 'error_code' not in next_lines and 'ErrorCode' not in next_lines:
                # 但如果是简单的日志，可能不需要 error_code
                if 'StatusInternalServerError' in next_lines or 'StatusNotFound' in next_lines:
                    file_issues.append({
                        'file': str(file_path),
                        'line': i,
                        'severity': 'warning',
                        'message': '错误日志应该包含 error_code 字段（zap.String("error_code", "...")）'
                    })
    
    # 5. 检查错误日志是否包含 request_id
    for i, line in enumerate(lines, 1):
        if re.search(log_error_pattern, line):
            next_lines = '\n'.join(lines[i-1:min(i+5, len(lines))])
            if 'request_id' not in next_lines and 'RequestID' not in next_lines:
                if 'StatusInternalServerError' in next_lines or 'StatusNotFound' in next_lines or 'StatusBadRequest' in next_lines:
                    file_issues.append({
                        'file': str(file_path),
                        'line': i,
                        'severity': 'warning',
                        'message': '错误日志应该包含 request_id 字段'
                    })
    
    return file_issues


def main():
    """主函数"""
    backend_handlers_dir = Path('backend/internal/handlers')
    
    if not backend_handlers_dir.exists():
        print(f"❌ 目录不存在: {backend_handlers_dir}")
        sys.exit(1)
    
    # 获取所有 Go 文件
    go_files = list(backend_handlers_dir.glob('*.go'))
    
    if not go_files:
        print("⚠️ 未找到任何 Go 文件")
        sys.exit(0)
    
    print(f"🔍 检查 {len(go_files)} 个文件...\n")
    
    # 检查每个文件
    for go_file in go_files:
        file_issues = check_file(go_file)
        issues.extend(file_issues)
    
    # 输出结果
    if issues:
        print("❌ 发现以下问题：\n")
        
        # 按严重程度分组
        errors = [i for i in issues if i['severity'] == 'error']
        warnings = [i for i in issues if i['severity'] == 'warning']
        
        if errors:
            print("## 🔴 错误（必须修复）\n")
            for issue in errors:
                print(f"- **{Path(issue['file']).name}:{issue['line']}** - {issue['message']}")
        
        if warnings:
            print("\n## 🟡 警告（建议修复）\n")
            for issue in warnings:
                print(f"- **{Path(issue['file']).name}:{issue['line']}** - {issue['message']}")
        
        print(f"\n\n总计: {len(errors)} 个错误, {len(warnings)} 个警告")
        
        # 如果有错误，返回非零退出码
        if errors:
            sys.exit(1)
        else:
            sys.exit(0)
    else:
        print("✅ 所有检查通过！")
        sys.exit(0)


if __name__ == '__main__':
    main()
