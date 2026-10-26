# #region build
# 1段目: Go でビルドする（このイメージは配らない）。ビルドする
# 側の CPU で動かし、目的の CPU 向けに Go でクロスコンパイルする
FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
# 静的にリンクして、2段目の最小のイメージで動くようにする
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /todo-api .

# 2段目: 実行に要るものだけを入れる。nonroot は UID 65532 で動く
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /todo-api /todo-api
EXPOSE 8080
ENTRYPOINT ["/todo-api"]
# #endregion build
