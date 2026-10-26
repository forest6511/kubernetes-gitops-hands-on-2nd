#!/usr/bin/env bash
# ECR の認証トークン（12時間で切れる）で、minikube の Pod が
# イメージを取得するための Secret（ecr-login）を作り直す。
# 使い方: scripts/ecr-login-minikube.sh [Namespace]
set -euo pipefail

# #region login
ns="${1:-todo-app}"
region="${AWS_REGION:-ap-northeast-1}"
account="$(aws sts get-caller-identity --query Account --output text)"
registry="${account}.dkr.ecr.${region}.amazonaws.com"

# Namespace が無ければ作る（apply なので何度実行してもよい）
kubectl create namespace "$ns" --dry-run=client -o yaml |
  kubectl apply -f -
# 同じ名前の Secret を上書きする
kubectl -n "$ns" create secret docker-registry ecr-login \
  --docker-server="$registry" \
  --docker-username=AWS \
  --docker-password="$(aws ecr get-login-password --region "$region")" \
  --dry-run=client -o yaml | kubectl apply -f -
# #endregion login
