package diffwriter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// DiffWriterConfig セキュリティと設定のための構造体
type DiffWriterConfig struct {
	MaxFileSize    int64
	MaxDiffSize    int64
	MaxInsertSize  int64
	MaxReplaceSize int64
	AllowedPaths   []string
}

// AppendTextOptions テキスト追記のオプション
type AppendTextOptions struct {
	Position int64  // 追記位置 (-1 は末尾)
	NewLine  bool   // 改行を追加するか
}

// InsertTextOptions テキスト挿入のオプション
type InsertTextOptions struct {
	Position int64  // 挿入位置
}

// ReplaceTextOptions テキスト置換のオプション
type ReplaceTextOptions struct {
	StartPos int64  // 置換開始位置
	EndPos   int64  // 置換終了位置
}

// DefaultConfig デフォルトのセキュリティ設定
var DefaultConfig = DiffWriterConfig{
	MaxFileSize:    10 * 1024 * 1024, // 10MB
	MaxDiffSize:    1 * 1024 * 1024,  // 1MB
	MaxInsertSize:  64 * 1024,        // 64KB
	MaxReplaceSize: 256 * 1024,       // 256KB
	AllowedPaths:   []string{},
}

// validatePath セキュリティチェック: パスの妥当性検証
func validatePath(path string, config DiffWriterConfig) error {
	// 絶対パスに変換
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// 許可されたパスチェック
	if len(config.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range config.AllowedPaths {
			if strings.HasPrefix(absPath, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return errors.New("path is not in allowed directories")
		}
	}

	// ファイルサイズチェック
	if config.MaxFileSize > 0 {
		info, err := os.Stat(absPath)
		if err == nil && info.Size() > config.MaxFileSize {
			return fmt.Errorf("file size exceeds maximum allowed size: %d", config.MaxFileSize)
		}
	}

	return nil
}

// AppendText ファイルにテキストを追記
func AppendText(filePath string, content string, opts ...AppendTextOptions) error {
	// デフォルトオプション
	option := AppendTextOptions{
		Position: -1,
		NewLine:  true,
	}
	if len(opts) > 0 {
		option = opts[0]
	}

	// セキュリティチェック
	if err := validatePath(filePath, DefaultConfig); err != nil {
		return err
	}

	// サイズ制限チェック
	if int64(len(content)) > DefaultConfig.MaxInsertSize {
		return fmt.Errorf("追記サイズが最大許容サイズ(%d)を超えています", DefaultConfig.MaxInsertSize)
	}

	// ファイルサイズと追記サイズチェック
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if fileInfo.Size()+int64(len(content)) > DefaultConfig.MaxFileSize {
		return errors.New("追記後のファイルサイズが最大許容サイズを超えています")
	}

	// 追記処理
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 追記位置の調整
	if option.Position == -1 {
		// 末尾に追記
		if _, err := file.Seek(0, io.SeekEnd); err != nil {
			return err
		}
	} else {
		// 指定位置に追記
		if _, err := file.Seek(option.Position, io.SeekStart); err != nil {
			return err
		}
	}

	// 改行の追加
	finalContent := content
	if option.NewLine {
		finalContent += "\n"
	}

	// 書き込み
	_, err = file.WriteString(finalContent)
	return err
}

// InsertText 指定位置にテキストを挿入
func InsertText(filePath string, content string, opts ...InsertTextOptions) error {
	// デフォルトオプション
	option := InsertTextOptions{
		Position: 0,
	}
	if len(opts) > 0 {
		option = opts[0]
	}

	// セキュリティチェック
	if err := validatePath(filePath, DefaultConfig); err != nil {
		return err
	}

	// サイズ制限チェック
	if int64(len(content)) > DefaultConfig.MaxInsertSize {
		return fmt.Errorf("挿入サイズが最大許容サイズ(%d)を超えています", DefaultConfig.MaxInsertSize)
	}

	// ファイル読み込み
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 挿入位置の調整
	if option.Position > int64(len(data)) {
		option.Position = int64(len(data))
	}

	// 指定位置にテキストを挿入
	result := make([]byte, 0, len(data)+len(content))
	result = append(result, data[:option.Position]...)
	result = append(result, []byte(content)...)
	result = append(result, data[option.Position:]...)

	// ファイルに書き戻し
	return os.WriteFile(filePath, result, 0644)
}

// ReplaceText 指定範囲のテキストを置換
func ReplaceText(filePath string, content string, opts ...ReplaceTextOptions) error {
	// デフォルトオプション
	option := ReplaceTextOptions{
		StartPos: 0,
		EndPos:   -1,
	}
	if len(opts) > 0 {
		option = opts[0]
	}

	// セキュリティチェック
	if err := validatePath(filePath, DefaultConfig); err != nil {
		return err
	}

	// サイズ制限チェック
	if int64(len(content)) > DefaultConfig.MaxReplaceSize {
		return fmt.Errorf("置換サイズが最大許容サイズ(%d)を超えています", DefaultConfig.MaxReplaceSize)
	}

	// ファイル読み込み
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 置換範囲の調整
	if option.EndPos == -1 {
		option.EndPos = int64(len(data))
	}

	// 置換処理
	result := make([]byte, 0, len(data)-int(option.EndPos-option.StartPos)+len(content))
	result = append(result, data[:option.StartPos]...)
	result = append(result, []byte(content)...)
	result = append(result, data[option.EndPos:]...)

	// ファイルに書き戻し
	return os.WriteFile(filePath, result, 0644)
}

// TruncateFile ファイルを指定サイズに切り詰め
func TruncateFile(filePath string, size int64) error {
	// セキュリティチェック
	if err := validatePath(filePath, DefaultConfig); err != nil {
		return err
	}

	return os.Truncate(filePath, size)
}

// PatchText JSON/テキストパッチの適用
func PatchText(filePath string, patchContent string) error {
	// セキュリティチェック
	if err := validatePath(filePath, DefaultConfig); err != nil {
		return err
	}

	// ファイル読み込み
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// パッチの適用（簡易実装）
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(patchContent)
	if err != nil {
		return err
	}

	// パッチの適用
	result, successes := dmp.PatchApply(patches, string(data))

	// すべてのパッチが適用されたかチェック
	allSucceeded := true
	for _, success := range successes {
		if !success {
			allSucceeded = false
			break
		}
	}

	if !allSucceeded {
		return errors.New("一部のパッチが適用できませんでした")
	}

	// 結果をファイルに書き戻し
	return os.WriteFile(filePath, []byte(result), 0644)
}
