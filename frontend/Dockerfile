# syntax=docker/dockerfile:1

# build stage
FROM node:18 as build

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

# producción con nginx
FROM nginx:stable-alpine

COPY --from=build /app/dist /usr/share/nginx/html

# copiamos configuración personalizada de nginx si quisieras
# COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
