# kubernetes-gitops-hands-on-2nd

書籍『Kubernetes実践ハンズオン』（2026年改訂版）のサンプルリポジトリです。
GitHub Actions・ArgoCD・Prometheus + Grafana・Amazon EKS を、Todo アプリを題材に順に動かします。

> ⚠️ **このリポジトリのコードを実行すると AWS の利用料金が発生します。** 先に [DISCLAIMER.md](DISCLAIMER.md) を読んでください。

## 使い方

1. このリポジトリを自分の GitHub アカウントにフォークし、クローンする（第2章の GitHub Actions と第3章の ArgoCD は、自分のリポジトリを使う）
2. `manifests/base/kustomization.yaml` の `123456789012` を、自分の AWS アカウント ID に書き換える
3. 各章の手順は書籍に従う

## 構成

- `app/api/` — Todo API（Go）
- `app/frontend/` — 画面（nginx）
- `manifests/base/` — どの環境でも使うマニフェスト
- `manifests/components/` — 章ごとに足す差分（Kustomize の components）
- `manifests/overlays/` — 章の状態ごとの入口（`kubectl apply -k` に渡すディレクトリ）
- `.github/workflows/ci.yml` — 第2章の CI
- `aws/` — 第1章の予算アラート・第2章の IAM の設定
- `scripts/` — 補助のスクリプト

## 検査

```bash
make check   # Go の整形・静的検査・テスト・脆弱性、YAML、マニフェストのスキーマ、ワークフロー、シェル
```

初版（2026年4月）の読者は、旧リポジトリ [kubernetes-gitops-hands-on](https://github.com/forest6511/kubernetes-gitops-hands-on) を使ってください。
