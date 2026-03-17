FROM node:24-alpine AS builder

WORKDIR /usr/app

COPY package.json package-lock.json tsconfig.json src/ ./

RUN npm install

RUN npm run build

EXPOSE 8000

CMD npm start
