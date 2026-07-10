FROM golang:1.25-alpine AS builder

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

FROM alpine:3.19

# Use Alpine mirror (uncomment or change if apk.dl-cdn.alpinelinux.org is unreachable)
ARG APK_MIRROR=mirrors.ustc.edu.cn
RUN if [ -n "${APK_MIRROR}" ]; then \
      sed -i "s/dl-cdn.alpinelinux.org/${APK_MIRROR}/g" /etc/apk/repositories; \
    fi

RUN apk add --no-cache ca-certificates tzdata ffmpeg

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
