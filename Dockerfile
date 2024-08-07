FROM golang:latest AS build_go

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /artifacts

FROM node:alpine AS build_react

WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH

COPY frontend/package.json ./
COPY frontend/package-lock.json ./

RUN npm install --silent

COPY frontend/. ./
RUN npm run build

FROM alpine:latest

WORKDIR /
COPY --from=build_go /artifacts /artifacts
COPY --from=build_react /app/build /frontend/build

ENTRYPOINT ["/artifacts"]
