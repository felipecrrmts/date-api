FROM golang:1.22.3 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download  && go mod verify

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server ./cmd/date-api

FROM alpine:3
RUN apk add --no-cache ca-certificates

ENV MONGODB_URI="mongodb+srv://melon:LOygT4ZPKq60nV2J@meloncluster.xzumd1y.mongodb.net/?retryWrites=true&w=majority"
ENV MONGODB_DATABASE="melon"

COPY --from=builder /app/server /server

CMD ["/server"]