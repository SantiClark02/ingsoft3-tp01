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

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]