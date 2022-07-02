FROM node:16.14-alpine
WORKDIR /app
COPY --chown=node:node . .
EXPOSE 80
RUN npm install && npm run build
CMD ["npm","run","start-prod"]
