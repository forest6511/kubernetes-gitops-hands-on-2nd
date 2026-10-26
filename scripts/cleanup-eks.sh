#!/usr/bin/env bash
# 第6・7章: EKS のクラスターを、課金が止まったことを確かめながら
# 順に消す。ArgoCD が動いたまま Gateway や PVC を消すと作り直される
# ので、この順序を守る。AWS_PROFILE と AWS_REGION を設定し、EKS の
# コンテキストで、リポジトリのルートから実行する
set -euo pipefail
CLUSTER=${CLUSTER:-todo-eks}
wait_until_empty() {
  # $1 は表示する名前。残りは一覧を出すコマンド
  # （空になるまで 10 秒おきに待つ）
  local name=$1 start=$SECONDS
  shift
  until [ -z "$("$@")" ]; do sleep 10; done
  echo "  $name: 0 件（$((SECONDS - start)) 秒）"
}
# #region steps
echo "1. 作り直す仕組みを止める（自動同期・第7章の ops/）"
kubectl patch application todo-app -n argocd --type merge \
  -p '{"spec":{"syncPolicy":null}}'
kubectl delete -f ops/ --ignore-not-found

echo "2. Gateway を消し、ALB が消えるのを待つ"
kubectl delete gateway -n todo-app --all
ALB_QUERY="LoadBalancers[?starts_with(LoadBalancerName,'k8s-todoapp')]"
wait_until_empty "ALB" aws elbv2 describe-load-balancers --output text \
  --query "$ALB_QUERY.LoadBalancerName"

echo "3. PVC を使っている Pod を止める（DB・監視スタック）"
kubectl scale deployment todo-db -n todo-app --replicas=0
helm uninstall kps -n monitoring

echo "4. PVC を消し、EBS が消えるのを待つ"
kubectl delete pvc -n todo-app --all
wait_until_empty "PVC の EBS" aws ec2 describe-volumes --output text \
  --filters "Name=tag:ebs.csi.aws.com/cluster-name,Values=$CLUSTER" \
  --query 'Volumes[].VolumeId'

echo "5. controller を消す（Application・ArgoCD・LBC）"
kubectl delete application -n argocd --all
ARGOCD=https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.6
kubectl delete -n argocd -f "$ARGOCD/manifests/install.yaml"
helm uninstall aws-load-balancer-controller -n kube-system

echo "6. クラスターを消す"
eksctl delete cluster --name "$CLUSTER" --wait
bash scripts/check-aws-empty.sh
# #endregion steps
