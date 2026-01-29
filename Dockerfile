FROM golang:1.25.5 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/rssd ./cmd/rssd
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/tg-bot ./cmd/tg-bot

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app

COPY --from=build /out/rssd /app/rssd
COPY --from=build /out/tg-bot /app/tg-bot

COPY migrations /app/migrations

USER nonroot:nonroot
