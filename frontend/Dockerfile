# Etapa 1: compilar el panel con Node 20 (Node solo existe aquí, no en la imagen final).
FROM node:20-alpine AS compilador
WORKDIR /app
COPY package.json package-lock.json ./
# npm ci instala EXACTAMENTE lo que fija package-lock.json; --ignore-scripts impide que
# una dependencia ejecute código arbitrario durante la instalación.
RUN npm ci --ignore-scripts --no-audit --no-fund
COPY tsconfig.json webpack.config.js ./
COPY src ./src
RUN npm run build

# Etapa 2: servir solo los archivos estáticos con NGINX sin privilegios (usuario nginx, puerto 8080).
FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=compilador /app/dist /usr/share/nginx/html
