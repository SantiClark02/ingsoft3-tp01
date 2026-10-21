# syntax=docker/dockerfile:1

# ---------- Etapa 1: build ----------
FROM node:20-alpine AS build

WORKDIR /app

# npm ci instala exactamente lo que dice package-lock.json.
# Es reproducible, a diferencia de npm install.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# ---------- Etapa 2: tests (parte del build: ya tiene Node y las devDependencies) ----------
# Va EN EL MEDIO: la última etapa del archivo tiene que seguir siendo nginx.
FROM build AS test
ENTRYPOINT ["npm", "run", "test:ci"]

# ---------- Etapa 3: runtime ----------
FROM nginx:1.27-alpine

# TP6: se copia una PLANTILLA (templates/), no una config fija (conf.d/).
# Al arrancar, nginx reemplaza ${BACKEND_URL} y ${DNS_RESOLVER} por las variables
# del entorno. Los valores de abajo son los de docker compose (por defecto);
# cada entorno los pisa con los suyos.
COPY default.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
ENV BACKEND_URL=http://backend:8080 \
    DNS_RESOLVER=127.0.0.11

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]