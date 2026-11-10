# Architecture Decision Records

設計上の判断とその理由を記録する。**新しい判断は [`0000-template.md`](0000-template.md) をコピーして次の番号で追加する。** 過去の ADR は書き換えず、覆す場合は新しい ADR で「Supersedes」を書く。

| 番号 | タイトル | 状態 |
| --- | --- | --- |
| [0001](0001-record-architecture-decisions.md) | 設計判断を ADR として記録する | Accepted |
| [0002](0002-tech-stack.md) | Go バックエンド + React/TypeScript フロントエンドのモノレポ | Accepted |
| [0003](0003-llm-interprets-engine-generates.md) | LLM は解釈、理論エンジンが生成 | Accepted |
| [0004](0004-api-contract-openapi.md) | API 契約は OpenAPI を唯一の情報源とする | Accepted |
| [0005](0005-vertical-slices-and-backlog-discipline.md) | 縦スライスとバックログ規律による開発 | Accepted |
| [0006](0006-documentation-freshness.md) | ドキュメントは現在の状態だけを書き、機械的検査と定期点検で鮮度を保つ | Accepted |
| [0007](0007-decay-prevention.md) | 開発サイクルの劣化を、矯正・検知・削減の三層で防ぐ | Accepted |
| [0008](0008-local-verification-and-containers.md) | 本番相当の配信形態を「単一 Go バイナリ」に決め、コンテナで検証可能にする | 一部 Superseded（0011、0027） |
| [0009](0009-generated-api-types.md) | TypeScript の API 型を OpenAPI から生成する | Accepted |
| [0010](0010-ux-decisions-follow-product-intent.md) | UI の判断は、プロダクトの意思を根拠に共通の原則と手順で行う | Accepted |
| [0011](0011-settings-only-from-launch-environment.md) | ローカル実行の設定と秘密情報は、起動時の環境変数だけで渡す | Accepted |
| [0012](0012-harmony-groove-style.md) | 例を「和声 × グルーヴ」に分け、「スタイル」で結ぶ（付録: [和声](0012/harmony.md)、[リズム](0012/rhythm.md)、[結びつき](0012/combination.md)、[類似アプリ](0012/apps.md)） | Accepted |
| [0013](0013-breadth-of-proposals.md) | 提案の幅は「表せる空間」と「選ぶ振る舞い」に分けて、測ってから広げる | Accepted |
| [0014](0014-issues-in-the-loop.md) | 観察と企画を GitHub Issue に置き、backlog と ADR はそのまま正典に保つ | Accepted |
| [0015](0015-modes.md) | モードを調の属性として足し、特徴音で解説する | Accepted |
| [0016](0016-ci-speed.md) | CI は、アプリと同じ版をキャッシュで速く取り、検査を 1 回ずつ走らせ、差分でジョブを選ぶ | Accepted |
| [0017](0017-meter.md) | 拍子を例全体の枠にし、3/4 から足す | Accepted |
| [0018](0018-eval-in-ci.md) | 実 API の評価を、PR のラベルで CI に走らせる | Accepted |
| [0019](0019-tempo-counts-the-beat.md) | テンポは拍子の拍で数える | Accepted |
| [0020](0020-slash-bass.md) | 分数コードのベースは和声に置き、度数で書く | Accepted |
| [0021](0021-style-groove.md) | スタイルを選び直したときのグルーヴはエンジンが決める | Accepted |
| [0022](0022-missing-files-are-404.md) | 存在しないファイルは 404 にし、SPA に振り分けるのは画面の URL だけにする | Superseded（0027） |
| [0023](0023-tritone-substitutes.md) | 裏コードは `subV7/x` と書き、セカンダリードミナントの一種として扱う | Accepted |
| [0024](0024-leading-tone-diminished.md) | 導音の減七は `vii°7/x` と書き、セカンダリードミナントの一種として扱う | Accepted |
| [0025](0025-open-chords.md) | sus4・7sus4・パワーコードは「3 度のない和音」として足し、ブルースの属七は分ける | Accepted |
| [0026](0026-blues-dominants.md) | 長調の I7・IV7 は、ブルースの属七として受け付ける | Accepted |
| [0027](0027-nginx-and-go-containers.md) | フロントは nginx のコンテナ、API は Go のコンテナにし、同一オリジンの 1 つのサービスとして配る | Accepted |
| [0028](0028-one-required-check.md) | main の必須チェックは、`make check` に当たるジョブをまとめた 1 つだけにする | Accepted |
| [0029](0029-lint-policy.md) | 推奨される lint は 1 つずつ入れ、道具そのものの陳腐化を定期的に検討する | Accepted |
| [0030](0030-problem-details.md) | エラー応答は RFC 9457 の problem detail にする | Accepted |
| [0031](0031-codeql-advanced-setup.md) | CodeQL はワークフローで動かし、文書だけの変更では走らせない | Accepted |
| [0032](0032-ci-on-main-by-changes.md) | main への push でも、変わったところに関係するジョブだけを走らせ、その実行は取り消さない | Accepted |
| [0033](0033-problem-details-only-for-the-api.md) | problem detail は API の応答にだけ使い、アプリのファイルと画面のエラーは短い平文にする | Accepted |
| [0034](0034-what-variation-is-an-error.md) | 提案のゆらぎは許し、答えの中の理論の誤りと食い違いは許さない | Accepted |
| [0035](0035-pattern-catalogue-as-data.md) | 型のカタログは、版管理され検証されるデータとしてエンジンに含める | Accepted |
| [0036](0036-shuffle-feel.md) | シャッフルを、名前のついたノリとして足す | Accepted |
| [0037](0037-candidate-tensions.md) | テンションは、候補ごとに AI が頼める（足す音はエンジンが決める） | Accepted |
| [0038](0038-narrowed-eval.md) | 評価は、タグで絞って走らせられる（`needs-eval:<tag>`。`needs-eval` は全件） | Accepted |
| [0039](0039-batched-eval.md) | 全件の評価は Message Batches で送れる（`EVAL_BATCH=1`。アプリ本体には使わない） | Accepted |
| [0040](0040-security-headers.md) | web のコンテナがセキュリティヘッダと CSP を付ける（snippet を全 location で。Go も nosniff） | Accepted |
| [0041](0041-wide-screen-bounds.md) | 広い画面の上限: 横幅 2560 px、タイムラインの高さは画面の 45 %（縦横比を保つ） | Accepted |
| [0042](0042-layout-rules.md) | 配置の規則: 骨組みは media、部品は container、しきい値は conventions の表に 1 か所。端末の当て推量は ESLint で禁止 | Accepted |

## いつ ADR を書くか

- 依存ライブラリ・フレームワークを追加／置換するとき
- データモデルや API の互換性を壊す変更をするとき
- 複数の選択肢があり、後から「なぜこうしたか」を聞かれそうなとき
- 既存 ADR の前提が崩れたとき
