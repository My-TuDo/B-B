@echo off
chcp 65001 >nul
title B-B Deploy
echo === B-B Deploy (Windows) ===
echo.

:: ---- Check Docker ----
echo [Check] docker...
where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: docker not found. Please install Docker Desktop first.
    pause
    exit /b 1
)
echo [OK]

:: ---- Check docker compose ----
echo [Check] docker compose...
docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    docker-compose version >nul 2>&1
    if %errorlevel% neq 0 (
        echo Error: docker compose not found.
        pause
        exit /b 1
    )
)
echo [OK]

:: ---- Check .env ----
if not exist ".env" (
    if exist ".env.example" (
        echo [Env] Creating .env from .env.example...
        copy .env.example .env >nul
        echo [Env] Please edit .env to customize secrets, then re-run deploy.bat
        echo [Env] You can edit .env with: notepad .env
        pause
        exit /b 0
    ) else (
        echo [Env] Warning: no .env or .env.example found
    )
)

:: ---- Generate SSL certificate (if not exists) ----
if not exist "nginx\ssl\server.crt" (
    echo [SSL] Generating self-signed certificate...
    if not exist "nginx\ssl" mkdir nginx\ssl

    :: Use Docker to generate cert (avoids needing OpenSSL on host)
    docker run --rm -v "%CD%\nginx\ssl:/ssl" alpine:3.19 sh -c "\
        apk add --no-cache openssl >/dev/null 2>&1 && \
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout /ssl/server.key \
            -out /ssl/server.crt \
            -subj '/C=CN/ST=Shanghai/L=Shanghai/O=B-B/OU=Dev/CN=localhost' \
            >/dev/null 2>&1" >nul 2>&1
    if %errorlevel% neq 0 (
        echo [SSL] Failed to generate certificate via Docker.
        echo [SSL] Make sure Docker is running and try again.
        pause
        exit /b 1
    )
    echo [SSL] Certificate generated
)

:: ---- Build and start ----
echo [Docker] Building images...
docker compose build
if %errorlevel% neq 0 (
    echo [Error] Build failed. Check the logs above.
    pause
    exit /b 1
)

echo [Docker] Starting services...
docker compose up -d
if %errorlevel% neq 0 (
    echo [Error] Failed to start services.
    pause
    exit /b 1
)

:: ---- Wait for backend health ----
echo [Health] Waiting for backend...
setlocal enabledelayedexpansion
for /l %%i in (1,1,30) do (
    curl -s -k https://localhost/api/v1/categories/ >nul 2>&1
    if !errorlevel! equ 0 (
        echo [Health] Backend is ready
        goto :done
    )
    timeout /t 2 /nobreak >nul
)
echo [Health] Backend may still be starting up...
:done
endlocal

echo.
echo === Deploy Complete ===
echo Frontend: https://localhost
echo API:      https://localhost/api/v1
echo MinIO:    http://localhost:9001
echo.
pause
