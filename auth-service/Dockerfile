FROM node:24-alpine AS builder

WORKDIR /usr/app

COPY package.json package-lock.json tsconfig.json src/ ./

RUN npm install

RUN npm run build

EXPOSE 8000
ENV OTEL_EXPORTER_OTLP_ENDPOINT "http://collector:4317"
ENV DB_HOST "postgres"

CMD npm start
