FROM node:22-alpine AS builder

WORKDIR /app

COPY admin/package.json admin/
RUN cd admin && npm install

COPY admin/ admin/
RUN cd admin && npm run build

FROM nginx:alpine
COPY --from=builder /app/admin/dist /usr/share/nginx/html
EXPOSE 80
