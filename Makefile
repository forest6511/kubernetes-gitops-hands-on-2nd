# サンプルの検査をまとめて実行する: make check
# Go（整形・静的検査・テスト・脆弱性）、YAML、マニフェストの
# スキーマ、GitHub Actions のワークフロー、シェルスクリプト
OVERLAYS := $(wildcard manifests/overlays/*)
SCHEMAS := -schema-location default \
  -schema-location 'https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json'

.PHONY: check go yaml manifests actions shell
check: go yaml manifests actions shell

go:
	cd app/api && test -z "$$(gofmt -l .)"
	cd app/api && go vet ./...
	cd app/api && golangci-lint run --config ../../.golangci.yml ./...
	cd app/api && go test ./...
	cd app/api && govulncheck ./...

yaml:
	yamllint -s .

manifests:
	@for o in $(OVERLAYS); do \
	  echo "== $$o"; \
	  kubectl kustomize $$o | kubeconform -strict -summary $(SCHEMAS) || exit 1; \
	done
	@if [ -d argocd ]; then echo "== argocd"; \
	  kubeconform -strict -summary $(SCHEMAS) argocd/*.yaml; fi

actions:
	@if [ -d .github/workflows ]; then actionlint; fi

shell:
	shellcheck scripts/*.sh
