# go-png-meta-web-strip

[![Go Reference](https://pkg.go.dev/badge/github.com/ideamans/go-png-meta-web-strip.svg)](https://pkg.go.dev/github.com/ideamans/go-png-meta-web-strip)
[![Go Report Card](https://goreportcard.com/badge/github.com/ideamans/go-png-meta-web-strip)](https://goreportcard.com/report/github.com/ideamans/go-png-meta-web-strip)

[English README](README.md)

Web表示に必要な情報を保持しながら、不要なメタデータを削除してPNG画像を最適化するGoライブラリです。

## 機能

- **選択的なチャンク削除**: ブラックリスト方式で不要なPNGチャンクを削除
- **重要データの保持**: 透明度、ガンマ補正、カラープロファイル、DPI設定などの重要なチャンクを保持
- **画像の完全性**: ピクセルデータは変更されずに保持されます
- **高性能**: 最小限のメモリオーバーヘッドで効率的なチャンクベースの処理
- **CRC検証**: チャンクの整合性を検証し、CRC値を再計算

### 削除されるチャンク

- tEXt/zTXt/iTXt: テキストメタデータとコメント
- tIME: 最終更新時刻
- bKGD: 背景色
- sPLT: 推奨パレット
- hIST: ヒストグラム
- eXIf: EXIFメタデータ
- その他すべての補助チャンク

### 保持されるチャンク

- IHDR: 画像ヘッダー（必須）
- PLTE: パレット（インデックスカラー画像で必須）
- IDAT: 画像データ（必須）
- IEND: 画像トレーラー（必須）
- tRNS: 透明度情報
- gAMA: ガンマ補正
- cHRM: 色度
- sRGB: sRGB色空間
- iCCP: ICCカラープロファイル
- sBIT: 有効ビット数（色精度）
- pHYs: 物理的なピクセル寸法（DPI）

## インストール

```bash
go get github.com/ideamans/go-png-meta-web-strip
```

## 使い方

```go
package main

import (
    "fmt"
    "os"
    pngmetawebstrip "github.com/ideamans/go-png-meta-web-strip"
)

func main() {
    // PNGファイルを読み込む
    pngData, err := os.ReadFile("input.png")
    if err != nil {
        panic(err)
    }

    // 不要なメタデータを削除
    cleanedData, result, err := pngmetawebstrip.PngMetaWebStrip(pngData)
    if err != nil {
        panic(err)
    }

    // クリーンなPNGを書き込む
    err = os.WriteFile("output.png", cleanedData, 0644)
    if err != nil {
        panic(err)
    }

    // 結果を表示
    fmt.Printf("削除されたチャンク:\n")
    fmt.Printf("  テキストチャンク: %d バイト\n", result.Removed.TextChunks)
    fmt.Printf("  タイムチャンク: %d バイト\n", result.Removed.TimeChunk)
    fmt.Printf("  背景: %d バイト\n", result.Removed.Background)
    fmt.Printf("  EXIFデータ: %d バイト\n", result.Removed.ExifData)
    fmt.Printf("  その他のチャンク: %d バイト\n", result.Removed.OtherChunks)
    fmt.Printf("合計削除: %d バイト\n", result.Total)
    
    // サイズ削減率を計算
    reduction := float64(result.Total) / float64(len(pngData)) * 100
    fmt.Printf("サイズ削減率: %.1f%%\n", reduction)
}
```

## APIリファレンス

### 主要関数

#### PngMetaWebStrip
```go
func PngMetaWebStrip(data []byte) ([]byte, *Result, error)
```
PNGデータを処理し、不要なメタデータチャンクを削除します。

#### PngMetaWebStripReader
```go
func PngMetaWebStripReader(r io.Reader) ([]byte, *Result, error)
```
io.ReaderからPNGデータを処理します。

#### PngMetaWebStripWriter
```go
func PngMetaWebStripWriter(data []byte, w io.Writer) (*Result, error)
```
PNGデータを処理し、結果をio.Writerに書き込みます。

### Result構造体
```go
type Result struct {
    Removed struct {
        TextChunks  int // tEXt, zTXt, iTXt
        TimeChunk   int // tIME
        Background  int // bKGD
        ExifData    int // eXIf
        OtherChunks int // その他の削除されたチャンク
    }
    Total int // 削除された合計バイト数
}
```

## テストデータジェネレーター

パッケージには、特定のチャンクの組み合わせを持つPNGファイルを作成するテストデータジェネレーターが含まれています。

### 使い方

```bash
# カスタムジェネレーターを使用してテストデータを生成
go run testgen/main.go

# またはImageMagickベースのジェネレーターを使用
go run datacreator/cmd/main.go
```

### 生成されるテスト画像

テストジェネレーターは`testdata`ディレクトリに様々なPNGファイルを作成します：

| ファイル名                     | 説明                           | チャンク/メタデータ                      |
| ------------------------------ | ------------------------------ | ---------------------------------------- |
| `basic_copy.png`               | オリジナルの基本コピー         | 最小限のチャンク                         |
| `with_text_chunks.png`         | tEXt/zTXt/iTXtチャンク付きPNG | コメント、キーワード、メタデータ         |
| `with_time.png`                | tIMEチャンク付きPNG           | 最終更新時刻                             |
| `with_background.png`          | bKGDチャンク付きPNG           | 背景色                                   |
| `with_exif.png`                | eXIfチャンク付きPNG           | EXIFメタデータ                           |
| `with_gamma.png`               | gAMAチャンク付きPNG           | ガンマ2.2（保持）                        |
| `with_chromaticity.png`        | cHRMチャンク付きPNG           | 色度（保持）                             |
| `with_srgb.png`                | sRGBチャンク付きPNG           | sRGBインジケーター（保持）               |
| `with_physical_dims.png`       | pHYsチャンク付きPNG           | 300 DPI（保持）                          |
| `with_transparency.png`        | tRNSチャンク付きPNG           | 透明度（保持）                           |
| `with_significant_bits.png`    | sBITチャンク付きPNG           | 有効ビット数情報（保持）                 |

### テストデータ生成の要件

- Go 1.22以上（カスタムジェネレーター用）
- ImageMagick（`magick`コマンド）- ImageMagickベースのジェネレーター用（オプション）
- ExifTool（`exiftool`コマンド）- eXIfチャンク操作用（オプション）

## テスト

```bash
# すべてのテストを実行
go test ./...

# 詳細出力付きで実行
go test -v ./...

# 特定のテストを実行
go test -v -run TestPngMetaWebStrip

# カバレッジレポートを生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## テストケース

パッケージには包括的なテストが含まれています：

1. **チャンク削除テスト**: 特定のチャンクタイプが削除されることを確認
2. **チャンク保持テスト**: 重要なチャンクが保持されることを確認
3. **無効なデータの処理**: 無効な入力に対するエラー処理をテスト
4. **画像整合性テスト**: ピクセルデータが変更されないことを確認
5. **包括的テスト**: 混合チャンクシナリオ
6. **透明度の保持**: 透明度を持つ画像でtRNSチャンクが保持されることを確認
7. **パフォーマンスベンチマーク**: 処理速度を測定

## パフォーマンス

ライブラリは最小限のメモリ割り当てで高性能になるよう設計されています：

- メモリに画像全体をロードせずにチャンクを順次処理
- データ整合性のためCRCチェックサムを検証
- 典型的な処理速度: チャンク構成に応じて約100-500 MB/s

ベンチマーク結果の例：
```
BenchmarkPngMetaWebStrip-8    10000    112337 ns/op    24576 B/op    12 allocs/op
```

## ユースケース

- **Web最適化**: より高速なWebページ読み込みのためにPNGファイルサイズを削減
- **プライバシー保護**: 機密情報を含む可能性のあるメタデータを削除
- **ストレージ最適化**: 不要なチャンクを削除してストレージスペースを節約
- **CDN最適化**: より小さな画像を配信して帯域幅コストを削減

## 要件

- Go 1.22以上
- コア機能に外部依存関係なし

## コントリビューション

コントリビューションを歓迎します！お気軽にプルリクエストを送信してください。

## ライセンス

MIT License

Copyright (c) 2024 IdeaMans Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.