package diffwriter

import (
	"os"
	"testing"
)

func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "testdiff*.txt")
	if err != nil {
		t.Fatalf("一時ファイル作成エラー: %v", err)
	}

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("ファイル書き込みエラー: %v", err)
	}
	tmpFile.Close()

	return tmpFile.Name()
}

// ... (existing test cases)

// エラーケースとエッジケースのテスト
func TestErrorCases(t *testing.T) {
	// ... (test cases)
}

// パス制限のテスト
func TestPathRestrictions(t *testing.T) {
	// ... (test cases)
}

// オプション機能のテスト
func TestAppendTextOptions(t *testing.T) {
	// ... (test cases)
}

// 挿入位置のエッジケース
func TestInsertTextEdgeCases(t *testing.T) {
	// ... (test cases)
}
