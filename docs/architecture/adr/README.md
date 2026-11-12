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
| [0013](0013-breadth-of-proposals.md) | 提案の幅は「表せる空間」と「選ぶ振る舞い」に分けて、測ってから広げる | Accepted（一部は 0061 で置き換え） |
| [0014](0014-issues-in-the-loop.md) | 観察と企画を GitHub Issue に置き、backlog と ADR はそのまま正典に保つ | Accepted |
| [0015](0015-modes.md) | モードを調の属性として足し、特徴音で解説する | Accepted |
| [0016](0016-ci-speed.md) | CI は、アプリと同じ版をキャッシュで速く取り、検査を 1 回ずつ走らせ、差分でジョブを選ぶ | Accepted |
| [0017](0017-meter.md) | 拍子を例全体の枠にし、3/4 から足す | Accepted |
| [0018](0018-eval-in-ci.md) | 実 API の評価を、PR のラベルで CI に走らせる | Accepted（ラベルの名前は 0063） |
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
| [0038](0038-narrowed-eval.md) | 評価は、タグで絞って走らせられる（ラベルの名前は 0063） | Accepted（名前は 0063 で置き換え） |
| [0039](0039-batched-eval.md) | 全件の評価は Message Batches で送れる（`EVAL_BATCH=1`。アプリ本体には使わない） | Accepted |
| [0040](0040-security-headers.md) | web のコンテナがセキュリティヘッダと CSP を付ける（snippet を全 location で。Go も nosniff） | Accepted |
| [0041](0041-wide-screen-bounds.md) | 広い画面の上限: 横幅 2560 px、タイムラインの高さは画面の 45 %（縦横比を保つ） | Accepted |
| [0042](0042-layout-rules.md) | 配置の規則: 骨組みは media、部品は container、しきい値は conventions の表に 1 か所。端末の当て推量は ESLint で禁止 | Accepted |
| [0043](0043-scrolling-timeline.md) | 長いタイムラインは横スクロールの 1 本と縮小図の帯（拡大とスライダーを置き換え） | Accepted（追随の形は 0066、帯を押したときは 0067） |
| [0044](0044-chord-band-and-function-map.md) | 1 和音の解説と度数の地図を同じ枠に入れ、広い画面ではタイムラインの上の帯に。地図は長調・短調で機能の列、モードで度数の順 | Accepted |
| [0045](0045-no-function-labels-in-modes.md) | モードの例では、画面で和音を機能で呼ばない（札・読み上げ・地図の色）。地図は主和音に印。文は #366 | Accepted |
| [0046](0046-modal-explanations-without-functions.md) | モードの解説の文は機能を名指ししない（主和音は「モードの主和音」、ほかは「N 度の和音」）。長調・短調は今まで通り | Accepted |
| [0047](0047-modal-roles.md) | モードの和音に役割（主和音・特徴和音・三全音の和音・そのほか）を音階の規則で付ける。解説の文に役割の一文 | Accepted |
| [0048](0048-modal-roles-on-screen.md) | モードの地図を役割の列（主和音 → 特徴和音 → そのほか → 三全音）に分け、札と読み上げを役割の言葉にする | Accepted |
| [0049](0049-sounding-chord-roles.md) | モードの解説の役割は鳴っている和音で決める（七の和音が三全音を足せば三全音の和音）。地図は三和音のまま | Accepted |
| [0050](0050-borrowed-chord-functions.md) | 長調の借用和音の機能は同主短調での働きで決める（♭VI は S、♭VII は次が IV なら S、それ以外は弱い D） | Accepted |
| [0051](0051-cadences-end-phrases.md) | 終止はフレーズ（4 小節）の終わりでだけ判定する。終止の四六は D で終止なし、転回形の V → I は弱い正格終止 | Accepted |
| [0052](0052-loop-seam-without-cadence-names.md) | ループの継ぎ目の文で終止の名前を言わない（どう戻るかだけを言う） | Accepted |
| [0053](0053-bass-line-under-slash-examples.md) | 分数コードのある例では、すべての和音のベースを、前のベースに近い高さで 1 本の線として鳴らす | Accepted |
| [0054](0054-borrowed-function-in-the-major-key.md) | 借用和音の機能の根拠は「長調の中での働き」（♭VI は iv の代わり）。短調の VI はトニックの代理のまま | Accepted |
| [0055](0055-sounding-chord-by-lightness.md) | 鳴っている和音は明るさの反転と琥珀の縁で示し、機能の色（緑・青・橙）は地図と札だけが使う | Accepted |
| [0056](0056-degree-map-current-chord-glow.md) | 度数の地図の今の和音は、セルの外側の枠と動かない光で示す（追従は琥珀、固定は青。動く案は却下） | Accepted |
| [0057](0057-timeline-fills-the-stage.md) | タイムラインは結果の幅いっぱいに描き、高さの上限（45 %）に当たるときは小節を横に伸ばす。文字と行の高さは変えない（ADR-0041 の縦横比を置き換える） | Accepted |
| [0058](0058-layout-e2e-in-required.md) | 配置の E2E（全サイズで横スクロールなし・axe・操作の集合・固定が残る）を、フロントエンドを変える PR の CI `layout` ジョブで走らせ、必須の検査に入れる。イメージに頼らない | Accepted |
| [0059](0059-timeline-scrolls-on-narrow-screens.md) | 狭い画面のタイムラインは縮めずに横へスクロールし、行名の列を左に留める。行名は 12 px | Accepted |
| [0060](0060-sixteenth-swing-feel.md) | 16 分のスウィング `swing-16` を名前のついたノリとして足す。比 1.3 : 1（約 57 %）、テンポによらない。pop と lofi で選べる | Accepted |
| [0061](0061-show-what-could-not-be-expressed.md) | 表せなかったもの（`unconstrained`）を本番の答えでも求め、要約の下に「AI の見立て」として小さく出す | Accepted |
| [0062](0062-sus2.md) | sus2 を 3 度のない和音に足す（テンションは足さない）。3 度のない和音の分数コードのベースは、根音からの音程で読む | Accepted |
| [0063](0063-eval-label-names.md) | 評価のラベルは `needs-eval-` のあとが走らせるもの、`:` のあとが絞り込み（`needs-eval-cases`、`needs-eval-cases:<tag>`、`needs-eval-converse`） | Accepted |
| [0064](0064-conversation-reviewer.md) | 会話の査読は会話 1 本を通して読み、依頼に従ったかを 4 つの区分（反映されていない・頼んでいないのに変わった・指した例の取り違え・返事が変更点を誤って言う）で数える。報告だけ。鍵を持つのは `converse judge` だけ | Accepted |
| [0065](0065-tolerance-in-taking-up-requests.md) | 頼まれたものを採り入れる程度には幅を認める（程度・一部だけ・初心者向けの単純化）。計測は答え単位に ✓／△／✗ と k/n で数え、幅の内は見せるが数えない | Accepted |
| [0066](0066-centre-follow.md) | 再生中の追随は、再生位置を（行名の列を除いた）見えている絵の中央に保つ。先頭と末尾はスクロールの端で止め、動きを減らす設定ではページめくり | Accepted |
| [0067](0067-click-to-move-the-playhead.md) | タイムラインの和音以外を押すと、拍の頭へ再生位置が動く。停止中は次の再生の開始位置、停止で戻り、鳴り終えたら先頭。和音は今どおり固定 | Accepted |
| [0068](0068-pause-following-on-manual-scroll.md) | 再生中に手でスクロールしたら追随を一時停止し、「再生位置へ戻る」・帯のクリック・次の再生で戻る（DAW の Follow に倣う） | Accepted |
| [0069](0069-overview-strip-picks-the-view.md) | 縮小図の帯は見る場所を選ぶ道具。枠のドラッグと帯のクリックは本体をスクロールするだけで、再生位置は動かさない。再生中は追随を一時停止する | Accepted |
| [0070](0070-play-pause-and-stop.md) | 再生は ▶/⏸ の切り替えと ■ の 2 つ。▶ は再生位置から、⏸ はその場で、■ は先頭へ戻して止める | Accepted |
| [0071](0071-drag-the-playhead.md) | マウスとペンでは、押した瞬間に再生位置が動き、押したまま動かすとついてくる。再生中はドラッグのあいだ音を止め、離すと続きから鳴る | Accepted |

## いつ ADR を書くか

- 依存ライブラリ・フレームワークを追加／置換するとき
- データモデルや API の互換性を壊す変更をするとき
- 複数の選択肢があり、後から「なぜこうしたか」を聞かれそうなとき
- 既存 ADR の前提が崩れたとき
