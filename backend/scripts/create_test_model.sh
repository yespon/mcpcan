#!/bin/bash

# 创建测试模型配置脚本
# 使用方法: ./create_test_model.sh [TOKEN]

set -e

# 配置
API_BASE="http://localhost:8080/market"
TOKEN="${1:-YOUR_TOKEN_HERE}"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== 创建 AI 模型测试配置 ===${NC}"
echo ""

# 检查服务是否运行
echo "检查服务状态..."
if ! curl -s -f "${API_BASE}/../health" > /dev/null 2>&1; then
    echo -e "${RED}错误: 服务未运行,请先启动后端服务${NC}"
    echo "提示: cd backend && go run cmd/market/main.go"
    exit 1
fi
echo -e "${GREEN}✓ 服务运行正常${NC}"
echo ""

# 创建模型配置
echo "创建 Qwen3 Coder Plus 配置..."
RESPONSE=$(curl -s -X POST "${API_BASE}/ai/models" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "name": "Qwen3 Coder Plus (Test)",
    "provider": "openai",
    "baseUrl": "https://coding.dashscope.aliyuncs.com/v1",
    "apiKey": "sk-sp-de748c01d0924e6b9b2c818c5edbaae0",
    "modelName": "qwen3-coder-plus"
  }')

# 检查响应
if echo "$RESPONSE" | grep -q '"code":0'; then
    MODEL_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo -e "${GREEN}✓ 模型配置创建成功${NC}"
    echo "Model ID: $MODEL_ID"
    echo ""
    echo "完整响应:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    echo ""
    echo -e "${YELLOW}下一步:${NC}"
    echo "  测试连接: ./scripts/test_model_connection.sh $MODEL_ID $TOKEN"
else
    echo -e "${RED}✗ 创建失败${NC}"
    echo "响应:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    exit 1
fi
