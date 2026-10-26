#!/usr/bin/env bash
# 第4章: Todo API に、追加・完了・削除・一覧の4つのリクエストを1秒ごとに送る。
# Grafana のグラフとアラートの練習で、値が変化するようにするために使う。
# 使い方: bash scripts/send-requests.sh [回数]（既定 300 回。約5分で終わる）
set -u
# #region send
URL=${URL:-http://127.0.0.1:8080/api/todos}
COUNT=${1:-300}
for i in $(seq "$COUNT"); do
  # 追加した Todo の id を応答から取り出す（DB が止まっていると空になる）
  id=$(curl -s -X POST -H 'Content-Type: application/json' \
    -d "{\"title\":\"request $i\"}" "$URL" |
    sed -n 's/.*"id":\([0-9]*\).*/\1/p')
  if [ -n "$id" ]; then
    curl -s -o /dev/null -X PUT -H 'Content-Type: application/json' \
      -d '{"completed":true}' "$URL/$id"
    curl -s -o /dev/null -X DELETE "$URL/$id"
  fi
  curl -s -o /dev/null "$URL"
  sleep 1
done
# #endregion send
