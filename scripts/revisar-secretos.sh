#!/usr/bin/env bash
# Revisión previa a subir a GitHub: busca secretos en los archivos que Git va a subir.
# Uso:  ./scripts/revisar-secretos.sh        (sale con código 1 si encuentra algo)
# Recomendado además: instalar gitleaks (https://github.com/gitleaks/gitleaks) como hook de pre-commit.
set -uo pipefail
cd "$(dirname "$0")/.."
problemas=0
aviso() { echo "  ✘ $1"; problemas=$((problemas + 1)); }

# Lista de archivos a revisar: los que Git rastrea o va a rastrear (respeta .gitignore).
if git rev-parse --git-dir >/dev/null 2>&1; then
  mapfile -t ARCHIVOS < <(git ls-files --cached --others --exclude-standard)
else
  mapfile -t ARCHIVOS < <(find . -type f -not -path './.git/*' -not -path '*/node_modules/*' -not -path './frontend/dist/*' -not -name '.env' | sed 's#^\./##')
fi

echo "Revisando ${#ARCHIVOS[@]} archivos..."

echo "1. Archivos sensibles que no deben subirse"
for f in "${ARCHIVOS[@]}"; do
  case "$f" in
    .env|.env.*) [ "$f" = ".env.ejemplo" ] || aviso "$f (archivo de entorno con secretos)";;
    *.pem|*.key|*.p12|*.pfx|*id_rsa*) aviso "$f (llave o certificado)";;
  esac
done

echo "2. Contenido que parece un secreto"
buscar() { # buscar <descripción> <regex>
  local resultado
  resultado=$(grep -InE "$2" "${ARCHIVOS[@]}" 2>/dev/null | grep -v 'scripts/revisar-secretos.sh' | head -5)
  [ -z "$resultado" ] || { aviso "$1"; echo "$resultado" | sed 's/^/      /'; }
}
buscar "llave privada"                 'BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY'
buscar "token JWT escrito en un archivo" 'eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.'
buscar "clave de acceso de AWS"        'AKIA[0-9A-Z]{16}'
buscar "token de GitHub"               'gh[pousr]_[A-Za-z0-9]{30,}'
buscar "URL con contraseña escrita"    '(postgres|postgresql|amqp|mysql|mongodb)://[^:/ ]+:[^@${ ]{3,}@'
buscar "contraseña asignada con valor literal" '^\s*[A-Z_]*(CLAVE|PASSWORD|SECRETO|SECRET|TOKEN)[A-Z_]*\s*[:=]\s*[A-Za-z0-9+/_-]{8,}\s*$'

echo
if [ "$problemas" -eq 0 ]; then
  echo "Resultado: no se encontraron secretos ✔"
else
  echo "Resultado: $problemas hallazgo(s) ✘. Corrígelos antes de subir. Si ya se subió un secreto, cámbialo: borrarlo del historial no basta."
  exit 1
fi
