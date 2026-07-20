@echo off
REM 波段趋势共振助手 - 生产启动脚本
REM 单文件运行，不需要 Node.js / pnpm
REM 默认端口 8080，可通过环境变量 PORT 修改
REM 示例: set PORT=80 && run-prod.bat

cd /d "%~dp0..\backend"

if "%PORT%"=="" (
    set PORT=8080
)

echo ===========================================
echo  波段趋势共振助手 - 生产模式
echo  端口: %PORT%
echo  前端页面: http://localhost:%PORT%
echo  API:      http://localhost:%PORT%/api/health
echo ===========================================
echo.

server.exe
