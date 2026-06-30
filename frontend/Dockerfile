FROM node:22-bookworm-slim AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm config set fetch-retries 5 \
  && npm config set fetch-retry-maxtimeout 120000 \
  && npm install
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /src/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
