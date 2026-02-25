#!/bin/bash

# 测试模型连接脚本
# 使用方法: ./test_model_connection.sh MODEL_ID [TOKEN]

set -e

# 参数
MODEL_ID="$1"
TOKEN="${2:-YOUR_TOKEN_HERE}"
API_BASE="http://localhost:8080/market"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查参数
if [ -z "$MODEL_ID" ]; then
  echo -e "${RED}错误: 缺少 MODEL_ID 参数${NC}"
  echo "使用方法: $0 <model_id> [token]"
  exit 1
fi

echo -e "${YELLOW}=== 测试 AI 模型连接 ===${NC}"
echo "Model ID: $MODEL_ID"
echo ""

# 测试连接
echo "正在测试连接..."
START_TIME=$(date +%s)

RESPONSE=$(curl -s -X POST "${API_BASE}/ai/models/${MODEL_ID}/test" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}")

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# 解析响应
SUCCESS=$(echo "$RESPONSE" | grep -o '"success":[^,}]*' | cut -d':' -f2 | tr -d ' ')
MESSAGE=$(echo "$RESPONSE" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)
LATENCY=$(echo "$RESPONSE" | grep -o '"latencyMs":[0-9]*' | cut -d':' -f2)

echo ""
if [ "$SUCCESS" = "true" ]; then
    echo -e "${GREEN}✓ 连接测试成功${NC}"
    echo -e "${BLUE}延迟: ${LATENCY}ms${NC}"
    echo -e "${BLUE}总耗时: ${DURATION}s${NC}"
    echo "消息: $MESSAGE"
    echo ""
    echo "完整响应:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    echo ""
    echo -e "${YELLOW}下一步:${NC}"
    echo "  1. 创建会话: curl -X POST ${API_BASE}/ai/sessions -H 'Authorization: Bearer ${TOKEN}' -d '{\"name\":\"Test\",\"modelAccessId\":${MODEL_ID}}'"
    echo "  2. 发送消息: curl -X POST ${API_BASE}/ai/sessions/{SESSION_ID}/chat -H 'Authorization: Bearer ${TOKEN}' -d '{\"content\":\"Hello\"}'"
else
    echo -e "${RED}✗ 连接测试失败${NC}"
    echo "消息: $MESSAGE"
    echo ""
    echo "完整响应:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    exit 1
fi
