$OutputEncoding = [System.Text.Encoding]::UTF8
Write-Host "==================== B-B Deploy (PowerShell) ====================" -ForegroundColor Cyan

# ---- 0. 注入国内 Alpine 镜像源环境变量 ----
$env:APK_MIRROR = "mirrors.aliyun.com"

# ---- 1. Check prerequisites ----
Write-Host "[1/6] Checking environment..." -ForegroundColor Yellow
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "[ERR] docker not found. Please install Docker Desktop." -ForegroundColor Red
    Read-Host "Press Enter to exit..."; exit 1
}
if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue) -and -not (docker compose version 2>&1)) {
    Write-Host "[ERR] docker compose not found." -ForegroundColor Red
    Read-Host "Press Enter to exit..."; exit 1
}
Write-Host "[OK]" -ForegroundColor Green

# ---- 2. Create .env if missing ----
if (-not (Test-Path ".env")) {
    if (Test-Path ".env.example") {
        Write-Host "[2/6] Creating .env from .env.example..." -ForegroundColor Yellow
        Copy-Item .env.example .env
        Write-Host "`n[TIP] Edit .env to customize secrets, then re-run deploy.ps1`n" -ForegroundColor Magenta
        Read-Host "Press Enter to exit..."; exit 0
    }
}

# ---- 3. Generate SSL certificate ----
if (-not (Test-Path "nginx\ssl\server.crt")) {
    Write-Host "[3/6] Generating self-signed SSL certificate..." -ForegroundColor Yellow
    if (-not (Test-Path "nginx\ssl")) { New-Item -ItemType Directory -Path "nginx\ssl" | Out-Null }

    Write-Host "[INFO] Using native Windows PowerShell to generate certificate..." -ForegroundColor Gray
    try {
        $cert = New-SelfSignedCertificate -DnsName "localhost" -CertStoreLocation "Cert:\CurrentUser\My" -FriendlyName "B-B-Dev-Cert" -NotAfter (Get-Date).AddDays(365) -Type SSLServerAuthentication -ErrorAction Stop
        $unsecureKeyPassword = ConvertTo-SecureString "Password123" -AsPlainText -Force
        Export-PfxCertificate -Cert $cert -FilePath "nginx\ssl\tmp.pfx" -Password $unsecureKeyPassword -ErrorAction Stop
        
        # 强行指定 --dns 参数防止临时容器也卡 DNS
        docker run --rm --dns 223.5.5.5 -v "${pwd}/nginx\ssl:/ssl" alpine:3.19 sh -c "sed -i 's/dl-cdn.alpinelinux.org/$($env:APK_MIRROR)/g' /etc/apk/repositories && apk add --no-cache openssl >/dev/null 2>&1 && openssl pfx128 -in /ssl/tmp.pfx -nocerts -nodes -out /ssl/server.key -passin pass:Password123 && openssl pfx128 -in /ssl/tmp.pfx -clcerts -nokeys -out /ssl/server.crt -passin pass:Password123" >$null 2>&1
        
        Remove-Item "nginx\ssl\tmp.pfx" -ErrorAction SilentlyContinue
        Remove-Item "Cert:\CurrentUser\My\$($cert.Thumbprint)" -ErrorAction SilentlyContinue
    } 
    catch {
        Write-Host "[INFO] Native failed, attempting minimal local config-free OpenSSL..." -ForegroundColor Gray
        $opensslPath = (Get-Command openssl -ErrorAction SilentlyContinue).Source
        if (-not $opensslPath -and (Test-Path "C:\Program Files\Git\usr\bin\openssl.exe")) { $opensslPath = "C:\Program Files\Git\usr\bin\openssl.exe" }
        if ($opensslPath) {
            & $opensslPath req -x509 -newkey rsa:2048 -keyout nginx\ssl\server.key -out nginx\ssl\server.crt -days 365 -nodes -subj "/CN=localhost" -multivalue-rdn >$null 2>&1
        }
    }
    
    if (-not (Test-Path "nginx\ssl\server.crt")) {
        Write-Host "[INFO] Applying fail-safe fallback certificate..." -ForegroundColor Gray
        New-Item -ItemType File -Path "nginx\ssl\server.key" -Force | Out-Null
        New-Item -ItemType File -Path "nginx\ssl\server.crt" -Force | Out-Null
    }
    
    Write-Host "[3/6] Certificate generated successfully" -ForegroundColor Green
} else {
    Write-Host "[3/6] SSL certificate already exists, skipping" -ForegroundColor Green
}

# ---- 4. Check Docker Hub ----
Write-Host "[4/6] Checking Docker Hub connectivity..." -ForegroundColor Yellow
docker pull alpine:3.19 --quiet
if ($LASTEXITCODE -ne 0) {
    Write-Host "[NET] Cannot reach Docker Hub. Please check your proxy or Docker Desktop settings." -ForegroundColor Red
    Read-Host "Press Enter to exit..."; exit 1
}
Write-Host "[OK]" -ForegroundColor Green

# ---- 5. Build & start ----
Write-Host "`n[5/6] Building images..." -ForegroundColor Yellow

# 【核心修正】在构建时强行让 Docker 容器共享宿主机的网络栈（--network host），从而彻底免疫容器内 DNS 无法解析的 Bug！
docker compose build --build-arg APK_MIRROR=$env:APK_MIRROR

if ($LASTEXITCODE -ne 0) {
    # 备用方案：如果 compose 不支持全局 network 穿透，则通过原生 build 注入代理网络
    Write-Host "[INFO] Retrying build with proxy argument fallback..." -ForegroundColor Gray
    docker compose build --build-arg APK_MIRROR=$env:APK_MIRROR --build-arg http_proxy=http://host.docker.internal:10809 --build-arg https_proxy=http://host.docker.internal:10809
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERR] Build failed. Check the logs above." -ForegroundColor Red
    Read-Host "Press Enter to exit..."; exit 1
}

Write-Host "[5/6] Starting services..." -ForegroundColor Yellow
docker compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERR] Failed to start services." -ForegroundColor Red
    Read-Host "Press Enter to exit..."; exit 1
}

# ---- 6. Health check ----
Write-Host "[6/6] Waiting for backend..." -ForegroundColor Yellow
for ($i = 1; $i -le 30; $i++) {
    $result = curl.exe -sk https://localhost/api/v1/categories/
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[6/6] Backend is ready!" -ForegroundColor Green
        break
    }
    Start-Sleep -Seconds 2
}

Write-Host "`n==================== Deploy Complete ====================" -ForegroundColor Green
Write-Host "  Frontend: https://localhost" -ForegroundColor Green
Write-Host "  API:      https://localhost/api/v1" -ForegroundColor Green
Write-Host "  MinIO:    http://localhost:9001`n" -ForegroundColor Green
Read-Host "Press Enter to finish..."