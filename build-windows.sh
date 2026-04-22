#!/bin/sh
# Кросс-компил mscaner.exe под Windows через Docker.
# Go на хосте не нужен. Результат — dist/mscaner.exe + config.env.example.
#
# modernc.org/sqlite — чистый Go, поэтому CGO_ENABLED=0 и никаких mingw.

set -eu

mkdir -p dist

docker run --rm \
    -v "$PWD":/app \
    -w /app \
    -e GOOS=windows \
    -e GOARCH=amd64 \
    -e CGO_ENABLED=0 \
    golang:alpine \
    go build -trimpath -ldflags "-s -w" -o dist/mscaner.exe ./cmd/mscaner

cp config.env.example dist/config.env.example

echo ""
echo "Собрано: dist/mscaner.exe"
echo "Шаблон конфига: dist/config.env.example"
echo ""
echo "Для деплоя:"
echo "  1. Скопируйте dist/ на Windows-хост."
echo "  2. Переименуйте config.env.example → config.env и впишите UNC-путь."
echo "  3. Запустите mscaner.exe (или зарегистрируйте как службу через NSSM)."
