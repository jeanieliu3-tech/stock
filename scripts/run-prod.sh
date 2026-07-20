#!/bin/bash
# 波段趋势共振助手 - 生产启动脚本 (bash版)
# 单文件运行，不需要 Node.js / pnpm
# 默认端口 8080，可通过 PORT 环境变量修改
# 示例: PORT=80 ./run-prod.sh

set -Eeuo pipefail

cd "$(dirname "$0")/../backend"

PORT="${PORT:-8080}"

echo "==========================================="
echo " 波段趋势共振助手 - 生产模式"
echo " 端口: $PORT"
echo " 前端页面: http://localhost:$PORT"
echo " API:      http://localhost:$PORT/api/health"
echo "==========================================="
echo ""

BACKEND_PORT="$PORT" ./server.exe
