#!/bin/bash

# テストスクリプト
cd /home/deepzen/workspaces/mcp/mcp-filesystem-server

# 依存関係の更新
go mod tidy

# テストの実行（詳細モード）
go test -v ./pkg/diffwriter

# カバレッジレポートの生成
go test -cover ./pkg/diffwriter

# カバレッジレポートをHTML形式で出力
go test -coverprofile=coverage.out ./pkg/diffwriter
go tool cover -html=coverage.out -o coverage.html

echo "テストと カバレッジレポートの生成が完了しました。"
