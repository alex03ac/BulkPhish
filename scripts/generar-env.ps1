# Crea el archivo ".env" con claves aleatorias y únicas (Windows / PowerShell).
# Ejecútalo UNA vez, antes de "docker compose up". No sobrescribe un .env existente.
$ErrorActionPreference = 'Stop'
Set-Location (Join-Path $PSScriptRoot '..')

if (Test-Path .env) {
    Write-Host 'Ya existe un archivo .env: no se modifica. Bórralo si quieres generar uno nuevo.'
    exit 1
}

$generador = [System.Security.Cryptography.RandomNumberGenerator]::Create()
function Aleatorio([int]$bytes) {
    $b = New-Object byte[] $bytes
    $generador.GetBytes($b)
    ($b | ForEach-Object { $_.ToString('x2') }) -join ''
}
function ClaveHumana([int]$largo) {
    $letras = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'.ToCharArray()
    $b = New-Object byte[] $largo
    $generador.GetBytes($b)
    -join ($b | ForEach-Object { $letras[$_ % $letras.Length] })
}

$claveAdmin = ClaveHumana 20
$contenido = @"
# Generado por scripts/generar-env.ps1. NO subir a Git.
JWT_SECRETO=$(Aleatorio 32)
ADMIN_CORREO=admin@bulkphish.test
ADMIN_NOMBRE=Administrador
ADMIN_CLAVE=$claveAdmin
CLAVE_BD_ADMIN=$(Aleatorio 24)
CLAVE_BD_AUTENTICACION=$(Aleatorio 24)
CLAVE_BD_CAMPANAS=$(Aleatorio 24)
CLAVE_BD_RESULTADOS=$(Aleatorio 24)
RABBIT_USUARIO=bulkphish
RABBIT_CLAVE=$(Aleatorio 24)
BIND_GATEWAY=127.0.0.1
URL_PUBLICA=http://localhost:8080
"@
# UTF-8 sin BOM y con saltos de línea LF, como espera Docker Compose
[System.IO.File]::WriteAllText((Join-Path (Get-Location) '.env'), ($contenido -replace "`r`n", "`n"), (New-Object System.Text.UTF8Encoding($false)))

# Las contraseñas de PostgreSQL solo se aplican cuando el volumen de datos se crea por primera vez.
if (Get-Command docker -ErrorAction SilentlyContinue) {
    $volumenes = docker volume ls -q 2>$null
    if ($volumenes -contains 'bulkphish_datos_postgres') {
        Write-Host 'ADVERTENCIA: ya existe un volumen de datos de una ejecución anterior (bulkphish_datos_postgres).'
        Write-Host '  Conserva las contraseñas viejas y los servicios no podrán conectarse.'
        Write-Host '  Antes de levantar, ejecuta:  docker compose down -v   (borra los datos de prueba)'
        Write-Host ''
    }
}
Write-Host 'Archivo .env creado con claves aleatorias.'
Write-Host ''
Write-Host '  Usuario administrador : admin@bulkphish.test'
Write-Host "  Contraseña            : $claveAdmin"
Write-Host ''
Write-Host 'Guarda esa contraseña en un lugar seguro (también está dentro de .env).'
Write-Host 'Ahora ejecuta:  docker compose up --build'
