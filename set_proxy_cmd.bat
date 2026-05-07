@echo off
chcp 65001
set HTTP_PROXY=http://127.0.0.1:7897
set HTTPS_PROXY=http://127.0.0.1:7897
set ALL_PROXY=http://127.0.0.1:7897
set NO_PROXY=localhost,127.0.0.1

echo 代理环境变量已设置：
echo   HTTP_PROXY  = %HTTP_PROXY%
echo   HTTPS_PROXY = %HTTPS_PROXY%
echo   ALL_PROXY   = %ALL_PROXY%
echo   NO_PROXY    = %NO_PROXY%
echo.

cmd /k