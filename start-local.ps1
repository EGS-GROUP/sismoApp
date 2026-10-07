# SismoMonitor / NextCollege Sismos - Lanzador Local (Windows 11)
$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " Iniciando NextCollege Sismos en Modo Desarrollo Local    " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Asegurar PATH de Go en la sesión actual
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User") + ";$env:Path"

# 2. Iniciar Backend Go en una nueva ventana de PowerShell
Write-Host "`n[1/2] Iniciando Backend en Go (Puerto 8082)..." -ForegroundColor Green
Start-Process powershell.exe -ArgumentList "-NoExit", "-Command", "cd backend; Write-Host '=== BACKEND SISMONITOR (GO) ===' -ForegroundColor Green; go run ."

# 3. Iniciar Frontend Vite en una nueva ventana de PowerShell
Write-Host "[2/2] Iniciando Frontend en Vite (Puerto 5173)..." -ForegroundColor Green
Start-Process powershell.exe -ArgumentList "-NoExit", "-Command", "cd frontend; Write-Host '=== FRONTEND SISMONITOR (VITE) ===' -ForegroundColor Green; npm run dev"

Write-Host "`nServicios iniciados en ventanas separadas." -ForegroundColor Yellow
Write-Host "Abre tu navegador en: http://localhost:5173" -ForegroundColor Cyan
Write-Host "Las llamadas a /api/ se redirigen automáticamente a http://localhost:8082" -ForegroundColor DarkGray
Write-Host "==========================================================" -ForegroundColor Cyan
