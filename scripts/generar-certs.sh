#!/usr/bin/env bash
# Genera un certificado TLS autofirmado para DESARROLLO en gateway/certs/.
# Para producción usa certificados reales (Let's Encrypt, mkcert, etc.).
# No sobrescribe uno existente. Los .pem/.key están en .gitignore.
set -euo pipefail
cd "$(dirname "$0")/.."

mkdir -p gateway/certs
if [ -e gateway/certs/cert.pem ]; then
  echo "Ya existe gateway/certs/cert.pem: no se sobrescribe. Bórralo si quieres uno nuevo."
  exit 0
fi

openssl req -x509 -newkey rsa:2048 -nodes -days 365 \
  -keyout gateway/certs/key.pem -out gateway/certs/cert.pem \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" 2>/dev/null

# Clave de DESARROLLO autofirmada: legible por el usuario del contenedor nginx.
# En producción usa certificados reales y permisos estrictos.
chmod 644 gateway/certs/key.pem
echo "Certificado autofirmado creado en gateway/certs/ (válido 365 días, CN=localhost)."
echo "El navegador mostrará un aviso de 'sitio no confiable': es normal con un certificado autofirmado."
