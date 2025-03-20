# アクセス制御システム PoC

本プロジェクトは、Open Policy Agent (OPA) を使用したプロキシベースのアーキテクチャによるロールベースアクセス制御（RBAC）システムを実装したPoCです。

システムは、Policy Enforcement Point (PEP) がリバースプロキシとして機能し、すべてのリクエストをインターセプトしてアプリケーションに到達する前にアクセス制御を実施します。

## 主要な検証項目と発見事項

本PoCでは、OPAを使用したアクセス制御の実装に関する以下の重要な側面を検証・実証しています：

### 1. 認可アーキテクチャ

1. 関心の分離
   - 認可ロジックをビジネスロジックから効果的に分離する方法
   - 集中管理されたポリシー管理の利点
   - サービスの保守性と開発速度への影響

2. プロキシベースの実施
   - リバースプロキシによるアクセス制御の効果
   - プロキシレイヤーのパフォーマンスへの影響
   - 実装の複雑さと利点のトレードオフ

### 2. OPA統合

1. ポリシー実装
   - Regoポリシーの作成と管理
   - ポリシーのテストと検証アプローチ
   - ポリシーのバージョン管理とデプロイ戦略

2. パフォーマンス特性
   - ポリシー評価のレイテンシー
   - レスポンスタイムへの影響
   - スケーラビリティに関する考慮点

3. 開発体験
   - Rego言語の学習曲線
   - ポリシーのデバッグとテストツール
   - 開発者ワークフローの改善点

### 3. アクセス制御機能

1. フィールドレベルのアクセス制御
   - きめ細かなデータフィルタリングの実装
   - フィールドフィルタリングのパフォーマンス影響
   - フィールドレベルポリシーの保守性

2. ロールベースの権限
   - 柔軟なロール定義
   - 権限の継承と階層
   - ロール割り当てと管理

### 4. 実践的な適用性

1. マイクロサービス統合
   - サービスの独立性と疎結合
   - サービス間でのポリシー一貫性
   - デプロイメントと運用上の考慮点

2. 実装上の課題
   - ポリシーの配布と更新
   - 監視とトラブルシューティング
   - エラー処理とフォールバック戦略

3. 本番環境対応
   - 必要なインフラコンポーネント
   - 運用上の考慮点
   - パフォーマンス最適化の必要性

## はじめ方

### 前提条件

- Docker
- Docker Compose
- Make

### インストールとセットアップ

1. リポジトリのクローン
```bash
git clone git@github.com:bmf-san/poc-opa-access-control-system.git
```

2. `/etc/hosts`の更新:
```sh
127.0.0.1 employee.local
127.0.0.1 pdp.local
127.0.0.1 pep.local
127.0.0.1 pip.local
```

3. Docker Composeによる全サービスの起動
```bash
make up
```

追加のコマンド:
```bash
# 全サービスのログ表示
make logs

# 特定サービスのログ表示
make log SERVICE=pep

# 全サービスの停止
make down

# 全サービスの再起動
make restart

# データベースCLIへのアクセス
make employee-db  # 従業員データベース用
make prp-db      # PRPデータベース用

# テストの実行
make test
```

## アクセス制御モデルのデモンストレーション

以下の例は、クライアントがPEPプロキシを介して従業員サービスとやり取りする方法を示しています。すべてのリクエストはpep.localを通過し、employee.local:8083への転送前にアクセス制御が実施されます。

各リクエストは以下の詳細なログを生成します：
- リクエストの受信と解析
- リソースとアクションの識別
- ポリシー評価
- アクセス判断とリクエスト転送

### RBACの例
```bash
# マネージャーロール：全従業員フィールドの参照可能
# John Manager（エンジニアリングマネージャー）
curl -X GET http://employee.local/employees \
  -H "X-User-ID: 11111111-1111-1111-1111-111111111111"
# レスポンス：200 OK とフィルタリング済みデータ
{
  "employees": [{
    "id": "11111111-1111-1111-1111-111111111111",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "employment_type": "Full-time",
    "employment_type_id": "11111111-1111-1111-1111-111111111111",
    "department_id": "dep1",
    "department_name": "Engineering",
    "position": "Engineer",
    "joined_at": "2023-01-01T00:00:00Z"
  }]
}

# 従業員ロール：idとnameフィールドのみ参照可能
# Bob Engineer（一般従業員）
curl -X GET http://employee.local/employees \
  -H "X-User-ID: 44444444-4444-4444-4444-444444444444"
# レスポンス：200 OK とフィルタリング済みデータ
{
  "employees": [{
    "id": "11111111-1111-1111-1111-111111111111",
    "name": "John Doe",
    "employment_type": "Full-time",
  }]
}

# 制御対象外リソースへのアクセスは拒否
# マネージャーを含むすべてのユーザーが departments などで403を受信
curl -X GET http://employee.local/departments \
  -H "X-User-ID: 11111111-1111-1111-1111-111111111111"
# レスポンス：403 Forbidden - アクセス拒否

curl -X GET http://employee.local/invalid_resource \
  -H "X-User-ID: 11111111-1111-1111-1111-111111111111"
# レスポンス：403 Forbidden - アクセス拒否

# User ID未指定：不正なリクエスト
curl -X GET http://employee.local/employees
# レスポンス：400 Bad Request - X-User-IDヘッダー不足
```

## ドキュメント

本プロジェクトのドキュメントは以下のセクションで構成されています：

### アーキテクチャと設計

以下を含む技術ドキュメント：
- アクセス制御アーキテクチャとモデル
- コンポーネントの責務
- アクセス制御フロー図
- API仕様
- データモデル
- OPA統合分析
- 運用設計
- 今後の検討事項

詳細は[設計ドキュメント](docs/design/DESIGN.ja.md)を参照してください。

### データベースドキュメント

データベーススキーマの詳細と関係：
- RBACテーブル
- エンティティ関係

以下のデータベースドキュメントを参照してください：
- PRPデータベース：[docs/db/prp](docs/db/prp/README.md)
- 従業員データベース：[docs/db/employee](docs/db/employee/README.md)

## 開発

### テストの実行
```bash
# 全テストの実行
make test

# 変更を反映してサービスをビルド・起動
make up
```

### データベースドキュメント
```bash
# データベースドキュメントの生成
make gen-dbdocs
```

## 貢献

Issue と Pull Request は常に歓迎します。

皆様のご貢献をお待ちしています。

貢献する前に以下のドキュメントを確認してください：

- [CODE_OF_CONDUCT](https://github.com/bmf-san/poc-opa-access-control-system/blob/master/.github/CODE_OF_CONDUCT.md)
- [CONTRIBUTING](https://github.com/bmf-san/poc-opa-access-control-system/blob/master/.github/CONTRIBUTING.md)

## 参考文献

- [www.openpolicyagent.org - Open Policy Agent](https://www.openpolicyagent.org/)
- [zenn.dev - OPA/Rego入門](https://zenn.dev/mizutani/books/d2f1440cfbba94)
- [kenfdev.hateblo.jp - アプリケーションにおける権限設計の課題](https://kenfdev.hateblo.jp/entry/2020/01/13/115032)

## ライセンス

MIT Licenseに基づいています。

[LICENSE](https://github.com/bmf-san/poc-opa-access-control-system/blob/master/LICENSE)

## 作者

[bmf-san](https://github.com/bmf-san)

- Email: bmf.infomation@gmail.com
- Blog: [bmf-tech.com](http://bmf-tech.com)
- Twitter: [@bmf-san](https://twitter.com/bmf-san)
