# Kubernetes実践ハンズオン（改訂版）— サンプルコード

書籍『Kubernetes実践ハンズオン ― GitHub Actions・ArgoCD・Amazon EKSで本番運用』（2026年10月 改訂版）のサンプルコードです。
1つの Todo アプリ（画面・API・DB）を題材に、CI・GitOps・監視を手元の minikube で組み、最後に Amazon EKS へ移して本番運用の基礎となる設定を当てます。

> **本書および本リポジトリは、Amazon Web Services, Inc.、Cloud Native Computing Foundation、Argo Project とは無関係の非公式な出版物です。**

Kindle 版: [Kubernetes実践ハンズオン](https://www.amazon.co.jp/dp/B0D33JTNFF)

## 料金について

> [!CAUTION]
> **このリポジトリの手順を実行すると、AWS の利用料金が発生します。**
>
> - 第1〜4章（minikube）で課金の対象になるのは、Amazon ECR のイメージの保存量と、イメージを取得するときのデータ転送量です
> - 第5〜7章（Amazon EKS）は、クラスター・EC2 のノード・NAT ゲートウェイ・ALB などが、起動している間ずっと時間で課金されます
> - 各章の冒頭に費用の目安を記載しています。作業しない日は、章末の「中断する場合」の手順でクラスターを削除してください
> - 実行前に AWS Budgets で予算アラートを設定してください（書籍 第1章）
> - 作業を終えたら、`scripts/check-aws-empty.sh` と請求とコスト管理（Cost Explorer）で、課金が止まったことを確認してください

詳細は [DISCLAIMER.md](DISCLAIMER.md) を参照してください。

## 動作環境

本書の手順は次の版で確認しています（macOS・Apple シリコン）。

- minikube v1.39.0（Kubernetes v1.37.0）、kubectl 1.37
- Envoy Gateway v1.9.2（Gateway API v1.6.1）
- ArgoCD v3.5.6、Sealed Secrets（チャート 2.20.0）
- kube-prometheus-stack（チャート 92.13.1）
- Amazon EKS（Kubernetes 1.35）、eksctl 0.231.0、AWS CLI v2
- AWS Load Balancer Controller v3.6.2（Gateway API v1.6.0）
- リージョン: `ap-northeast-1`（東京）

## 使い方

第2章の GitHub Actions と第3章の ArgoCD は、読者自身のリポジトリを使います。最初にフォークしてください。

```bash
# 1. このリポジトリを自分の GitHub アカウントにフォークし、クローンする
git clone https://github.com/<自分のアカウント>/kubernetes-gitops-hands-on-2nd.git
cd kubernetes-gitops-hands-on-2nd

# 2. AWS の認証情報はプロファイルで渡す（アクセスキーをコードに書かない）
export AWS_PROFILE=k8s-gitops-book AWS_REGION=ap-northeast-1
```

続いて、書籍の指示に従って次の2か所を自分の値に書き換えます。

- `manifests/base/kustomization.yaml` の `123456789012` → 自分の AWS アカウント ID（第1章）
- `argocd/*.yaml` の `YOUR_GITHUB_USER` → 自分の GitHub のユーザー名（第3章）

## ディレクトリ構成

章ごとにディレクトリを分けるのではなく、全章で同じファイルを使います。章の進み具合は、`manifests/overlays/` の「入口」を切り替えて表します。

- `app/api/` — Todo API（Go）。`/metrics` で Prometheus の値を公開します
- `app/frontend/` — 画面（非 root で動く nginx）
- `manifests/base/` — どの環境でも使うマニフェスト（Deployment・Service・ConfigMap・PVC・HTTPRoute）
- `manifests/components/` — 章ごとに足す差分（Kustomize の components）
- `manifests/overlays/` — 章の状態ごとの入口（`kubectl apply -k` や ArgoCD の `path` に渡すディレクトリ）
- `.github/workflows/ci.yml` — 第2章の CI（テスト・ビルド・Amazon ECR への push）
- `aws/` — 第1章の予算アラートと、第2章の IAM の設定
- `argocd/` — 第3章の ArgoCD の Application と ApplicationSet、第7章の HPA 用の Application
- `monitoring/` — 第4章の kube-prometheus-stack の values
- `eks/` — 第5章の eksctl の ClusterConfig と、第6章の AWS Load Balancer Controller の values
- `ops/` — 第7章の HPA・RBAC・NetworkPolicy・ResourceQuota
- `scripts/` — 認証 Secret の作成・SealedSecret の作成・通信の送信・EKS の削除と残りの確認
- `docs/logs/` — 書籍に載せた実行結果

章ごとの入口は次のとおりです。

- 第1・2章 — `manifests/overlays/ch01-minikube`
- 第3章 — `manifests/overlays/ch03-minikube-sealed`（Secret を Sealed Secrets で暗号化）と `env-dev`・`env-staging`・`env-prod`（ApplicationSet）
- 第4章 — `manifests/overlays/ch04-minikube-monitored`（ServiceMonitor・アラートのルール・ダッシュボードを追加）
- 第6・7章 — `manifests/overlays/ch06-eks`（ALB の Gateway・EBS の StorageClass に差し替え）

## 検査

```bash
make check   # Go の整形・静的検査・テスト・脆弱性、YAML、マニフェストのスキーマ、ワークフロー、シェル
```

## 利用上の注意

> [!WARNING]
> **このコードは学習用です。本番環境にそのまま使わないでください。**
> 本書の構成は、手順を追いやすくするために単純化しています。実運用では次の検討が必要です。
>
> - 最小権限の IAM ポリシー（本書の IAM ユーザーは学習用に AdministratorAccess を使います）
> - NAT ゲートウェイの冗長化（本書は1つだけ）
> - Secret の扱い（第3章の平文の見本は学習用です）
> - バックアップ・マルチクラスター・コスト管理の方針

## 初版のリポジトリ

初版（2026年4月）の読者は、旧リポジトリ [kubernetes-gitops-hands-on](https://github.com/forest6511/kubernetes-gitops-hands-on) を使ってください。

## ライセンス

[MIT License](LICENSE)

ライセンスはコードの利用許諾です。損害の免責については [DISCLAIMER.md](DISCLAIMER.md) を参照してください。
