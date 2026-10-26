#!/usr/bin/env bash
# 第5〜7章: 本書のクラスター（todo-eks）の課金リソースが残って
# いないかを確かめる。残っていれば一覧を表示して exit 1。
# AWS_PROFILE と AWS_REGION を設定してから実行する
set -u
CLUSTER=${CLUSTER:-todo-eks}
TAG=Name=tag:alpha.eksctl.io/cluster-name,Values=$CLUSTER
found=0
# #region check
check() {
  # $1 は表示する名前。残りは、残っているものの一覧を出す aws のコマンド
  local name=$1 out
  shift
  out=$("$@" --output text 2>&1)
  if [ -n "$out" ] && [ "$out" != "None" ]; then
    echo "残っている: $name"
    printf '  %s\n' "$out"
    found=1
  else
    echo "0 件: $name"
  fi
}
# #endregion check
# #region targets
check "EKS クラスター" aws eks list-clusters \
  --query "clusters[?@=='$CLUSTER']"
check "EC2 インスタンス" aws ec2 describe-instances --filters \
  "Name=tag:eks:cluster-name,Values=$CLUSTER" \
  Name=instance-state-name,Values=pending,running,stopping,stopped \
  --query 'Reservations[].Instances[].InstanceId'
check "NAT Gateway" aws ec2 describe-nat-gateways --filter "$TAG" \
  Name=state,Values=pending,available,deleting \
  --query 'NatGateways[].NatGatewayId'
check "Elastic IP" aws ec2 describe-addresses --filters "$TAG" \
  --query 'Addresses[].PublicIp'
check "PVC から作られた EBS" aws ec2 describe-volumes --filters \
  "Name=tag:ebs.csi.aws.com/cluster-name,Values=$CLUSTER" \
  --query 'Volumes[].VolumeId'
check "どこにも付いていない EBS" aws ec2 describe-volumes \
  --filters Name=status,Values=available --query 'Volumes[].VolumeId'
check "ALB" aws elbv2 describe-load-balancers \
  --query 'LoadBalancers[].LoadBalancerName'
check "ターゲットグループ" aws elbv2 describe-target-groups \
  --query 'TargetGroups[].TargetGroupName'
check "CloudFormation のスタック" aws cloudformation list-stacks \
  --query "StackSummaries[?starts_with(StackName,'eksctl-$CLUSTER-') \
&& StackStatus!='DELETE_COMPLETE'].[StackName,StackStatus]"
# #endregion targets
exit "$found"
