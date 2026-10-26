#!/usr/bin/env bash
# 第3章: 平文の Secret の見本を、このクラスターの Sealed Secrets の鍵で
# 暗号化し、SealedSecret のファイルを上書きする。
# 鍵はクラスターごとに違うので、著者の SealedSecret は読者のクラスターでは
# 復号できない。必ず自分のクラスターでこのスクリプトを実行する
set -euo pipefail
cd "$(dirname "$0")/.."

# #region seal
kubeseal --format yaml \
  < manifests/components/secret-plain/secret.yaml \
  > manifests/components/secret-sealed/sealedsecret.yaml
# #endregion seal
echo "wrote manifests/components/secret-sealed/sealedsecret.yaml"
