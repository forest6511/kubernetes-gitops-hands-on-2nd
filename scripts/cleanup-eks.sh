#!/usr/bin/env bash
# 第6・7章: EKS のクラスターを、課金が止まったことを確かめながら
# 順に消す。ArgoCD が動いたまま Gateway や PVC を消すと作り直される
# ので、この順序を守る。AWS_PROFILE と AWS_REGION を設定し、EKS の
# コンテキストで、リポジトリのルートから実行する
set -euo pipefail
CLUSTER=${CLUSTER:-todo-eks}
wait_until_empty() {
  # $1 は表示する名前。残りは一覧を出すコマンド（空になるまで
  # 10 秒おきに待つ）。照会に失敗したら、0 件とは見なさずに止まる
  local name=$1 start=$SECONDS out
  shift
  while true; do
    if ! out=$("$@"); then
      echo "  $name: 照会に失敗したので止めます" >&2
      exit 1
    fi
    [ -z "$out" ] && break
    sleep 10
  done
  echo "  $name: 0 件（$((SECONDS - start)) 秒）"
}
# 途中で中断したクラスターでも流せるように、無いものは飛ばす。
# 照会そのものに失敗したら（接続できないなど）、無いとは見なさずに止まる
has() {
  local out
  if ! out=$(kubectl get "$@" --ignore-not-found -o name); then
    echo "kubectl get $* に失敗したので止めます" >&2
    exit 1
  fi
  [ -n "$out" ]
}
ARGOCD=https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.6
ALB_QUERY="LoadBalancers[?starts_with(LoadBalancerName,'k8s-todoapp')]"
# クラスターの状態で分ける。ACTIVE は1段目から、DELETING（削除の途中）
# は6段目から、無ければ残りの確認だけ。それ以外（作成中・更新中など）は
# 止まる。「無い」以外の照会の失敗（認証切れなど）でも止まる
if ! STATUS=$(aws eks describe-cluster --name "$CLUSTER" \
  --query cluster.status --output text 2>&1); then
  if [[ $STATUS != *ResourceNotFoundException* ]]; then
    echo "$STATUS" >&2
    exit 1
  fi
  STATUS=NONE
fi
case $STATUS in
  ACTIVE) ;;
  DELETING | NONE)
    echo "クラスターの状態が $STATUS なので、6段目だけを行う"
    if [ "$STATUS" = DELETING ]; then
      eksctl delete cluster --name "$CLUSTER" --wait
    fi
    bash scripts/check-aws-empty.sh
    exit $? ;;
  *)
    echo "クラスターの状態が $STATUS です。ACTIVE になってから流す" >&2
    exit 1 ;;
esac
# kubectl の接続先が、このクラスターの API の URL と同じかを確かめる
# （minikube や別のクラスターを向いたまま流すと、そちらを消す）。
# eksctl delete cluster は最初に kubeconfig から $CLUSTER を消すので、
# その後に止めた場合は作り直してから流し直す
ENDPOINT=$(aws eks describe-cluster --name "$CLUSTER" \
  --query cluster.endpoint --output text)
SERVER=$(kubectl config view --minify \
  -o jsonpath='{.clusters[0].cluster.server}' 2> /dev/null || true)
if [ "$SERVER" != "$ENDPOINT" ]; then
  echo "kubectl の接続先が $CLUSTER ではありません" >&2
  msg="aws eks update-kubeconfig --name $CLUSTER で接続先を作り直して"
  echo "${msg}から、もう一度流してください" >&2
  exit 1
fi
# #region steps-1
echo "1. 作り直す仕組みを止める（自動同期・第7章の ops/）"
if has crd applications.argoproj.io &&
  has application todo-app -n argocd; then
  kubectl patch application todo-app -n argocd --type merge \
    -p '{"spec":{"syncPolicy":null}}'
fi
kubectl delete -f ops/ --ignore-not-found

echo "2. Gateway を消し、ALB が消えるのを待つ"
if has crd gateways.gateway.networking.k8s.io; then
  kubectl delete gateway -n todo-app --all
fi
wait_until_empty "ALB" aws elbv2 describe-load-balancers --output text \
  --query "$ALB_QUERY.LoadBalancerName"

echo "3. PVC を使っている Pod を止める（DB・監視スタック）"
if has deployment todo-db -n todo-app; then
  kubectl scale deployment todo-db -n todo-app --replicas=0
fi
helm uninstall kps -n monitoring --ignore-not-found
# #endregion steps-1
# #region steps-2
echo "4. PVC を消し、EBS が消えるのを待つ"
kubectl delete pvc -n todo-app --all
wait_until_empty "PVC の EBS" aws ec2 describe-volumes --output text \
  --filters "Name=tag:ebs.csi.aws.com/cluster-name,Values=$CLUSTER" \
  --query 'Volumes[].VolumeId'

echo "5. controller を消す（Application・ArgoCD・LBC）"
if has crd applications.argoproj.io; then
  kubectl delete application -n argocd --all
fi
kubectl delete -n argocd -f "$ARGOCD/manifests/install.yaml" \
  --ignore-not-found
helm uninstall aws-load-balancer-controller -n kube-system \
  --ignore-not-found

echo "6. クラスターを消す"
eksctl delete cluster --name "$CLUSTER" --wait
bash scripts/check-aws-empty.sh
# #endregion steps-2
