FROM node:22-alpine

WORKDIR /app

COPY web-ui/package*.json ./
RUN npm install

COPY web-ui/ .

EXPOSE 5173

CMD ["sh", "-c", "npm install && npm run dev -- --host 0.0.0.0"]
