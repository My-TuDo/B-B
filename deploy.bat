@echo off
chcp 65001 >nul
title B-B Deploy
echo ==================== B-B Deploy ====================
echo.

:: ---- 1. Check prerequisites ----
echo [1/6] Checking environment...

where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERR] docker not found. Please install Docker Desktop.
    pause & exit /b 1
)

docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    docker-compose version >nul 2>&1
    if %errorlevel% neq 0 (
        echo [ERR] docker compose not found.
        pause & exit /b 1
    )
)
echo [OK]

:: ---- 2. Create .env if missing ----
if not exist ".env" (
    if exist ".env.example" (
        echo [2/6] Creating .env from .env.example...
        copy .env.example .env >nul
        echo.
        echo [TIP] Edit .env to customize secrets, then re-run deploy.bat
        echo        You can edit with: notepad .env
        echo.
        pause & exit /b 0
    )
)

:: ---- 3. Generate SSL certificate ----
if not exist "nginx\ssl\server.crt" (
    echo [3/6] Generating self-signed SSL certificate...
    if not exist "nginx\ssl" mkdir nginx\ssl

    docker run --rm -v "%CD%\nginx\ssl:/ssl" alpine:3.19 sh -c "sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk add --no-cache openssl && openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /ssl/server.key -out /ssl/server.crt -subj '/C=CN/ST=Shanghai/L=Shanghai/O=B-B/OU=Dev/CN=localhost'" >nul 2>&1
    if %errorlevel% neq 0 (
        echo [ERR] Failed to generate SSL certificate. Is Docker running?
        pause & exit /b 1
    )
    echo [3/6] Certificate generated
) else (
    echo [3/6] SSL certificate already exists, skipping
)

:: ---- 4. Check Docker Hub ----
echo [4/6] Checking Docker Hub connectivity...
docker pull alpine:3.19 --quiet >nul 2>&1
if %errorlevel% neq 0 (
    echo [NET] Cannot reach Docker Hub.
    echo.
    echo If you are in China, configure a Docker mirror:
    echo   1. Open Docker Desktop Settings ^> Docker Engine
    echo   2. Add to the JSON:
    echo      "registry-mirrors": ["https://docker.m.daocloud.io"]
    echo   3. Click "Apply ^& Restart"
    echo   4. Re-run deploy.bat
    echo.
    pause & exit /b 1
)
echo [OK]

:: ---- 5. Build & start ----
echo.
echo [5/6] Building images...
docker compose build
if %errorlevel% neq 0 (
    echo [ERR] Build failed. Check the logs above.
    pause & exit /b 1
)

echo [5/6] Starting services...
docker compose up -d
if %errorlevel% neq 0 (
    echo [ERR] Failed to start services.
    pause & exit /b 1
)

:: ---- 6. Health check ----
echo [6/6] Waiting for backend...
for /l %%i in (1,1,30) do (
    curl -sk https://localhost/api/v1/categories/ >nul 2>&1
    if !errorlevel! equ 0 (
        echo [6/6] Backend is ready!
        goto :done
    )
    timeout /t 2 /nobreak >nul
)
:done

echo.
echo ==================== Deploy Complete ====================
echo   Frontend: https://localhost
echo   API:      https://localhost/api/v1
echo   MinIO:    http://localhost:9001
echo.
pause
