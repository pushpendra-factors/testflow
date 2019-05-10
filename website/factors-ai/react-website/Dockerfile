FROM node:8.15.0-alpine

WORKDIR /website
COPY dist dist
COPY serve-dist.js .
RUN npm install express

ENTRYPOINT ["node", "serve-dist.js"]
EXPOSE 3001