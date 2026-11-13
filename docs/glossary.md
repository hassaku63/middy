# 用語集

| 用語 | 意味 | コード上の対応 |
| --- | --- | --- |
| イメージ (Image) | ユーザーが言葉で表す曲の印象。「雨の日曜」「出発の朝」など | LLM への入力。`MusicalIntent` の元 |
| 音楽的意図 (MusicalIntent) | イメージを LLM が解釈した構造化パラメータ。調・モード・テンポ・雰囲気タグ・進行テンプレート候補 | `intent.Intent`, `intent.Candidate` |
| 進行 (Progression) | 調に依存しないコード進行テンプレート。度数と拍数の列 | `music.Progression`, `music.FromProgression` |
| パターン (Pattern) | 再利用可能なコード進行テンプレート、またはリズムテンプレート | `pattern.Accompaniment`, `pattern.Drums`, `pattern.Apply` |
| タイムライン (Timeline) | 再生・表示・書き出しの対象となる中心データ。テンポ、拍子、調、トラック、ハーモニー注釈。画面の部品としての意味は「画面の部品」の節 | `music.Timeline` |
| とっておいた例 (Kept example) | 題名の横の星（☆／★）でブラウザに残した例。名前は候補の題名になり、一覧で変えられる（#440）。あとで開くと同じ例が鳴る。キーやスタイルを変えた例は別の例。画面はこの呼び名で統一する（コードでは project） | `features/projects`, `shared/lib/storage/projects` |
| トラック (Track) | 楽器ごとのノート列 | `music.Track` |
| ノート (NoteEvent) | ピッチ・ベロシティ・開始 tick・長さ | `music.NoteEvent` |
| ハーモニー注釈 (HarmonyEvent) | ある区間を支配するコードとローマ数字。解説レイヤーの素材 | `music.HarmonyEvent` |
| SMF (Standard MIDI File) | DAW が読める MIDI ファイル形式。middy は format 0 を書き出す | `smf.Write` |
| tick | 時間の最小単位。`DefaultPPQ`=480 tick が四分音符 | `music.Tick` |
| ピッチクラス (PitchClass) | オクターブを無視した音名。C=0 … B=11 | `music.PitchClass` |
| ピッチ (Pitch) | MIDI ノート番号 0..127。C4=60 | `music.Pitch` |
| 調 (Key) | 主音 + モード | `music.Key` |
| モード（旋法, Mode） | 音階をどの音を中心に聞かせるかで区別したもの。長調はイオニアン、自然短音階はエオリアンで、どちらもモードの 1 つ。ドリアン（短調に似て 6 度が長い）とミクソリディアン（長調に似て 7 度が短い）がある（ADR-0015） | `music.Mode` |
| スケール（音階, Scale） | 音を高さの順に並べた集合。長音階、自然短音階、和声的短音階、ペンタトニックなど。モードより広い言葉 | `Mode.Intervals` |
| 度数 (Degree) | 音階上の位置。1 始まり。解説の文では、言い方で意味を分ける。音階の音は「音階の第 N 音」、その度数の上の和音は「N 度の和音」、和音の中の音は「和音の N 度（の音）」、主音からの音程は「主音から長 6 度」のように音程の名前で言う（#390） | `Key.Degree`, `Key.Diatonic` |
| ダイアトニックコード | 調の音だけで作られるコード | `Key.Diatonic` |
| ローマ数字 (Roman numeral) | 度数とコードの質をあらわす記法。I, ii, V7, vii° | `music.RomanNumeral` |
| 和声機能 (Harmonic function) | T（トニック）/ S（サブドミナント）/ D（ドミナント） | `music.Function` |
| 主音・主和音・トニック | 解説の文では役割で分ける。主音は調の中心の音（C major の C）、主和音はその上の和音（I・i）、トニック（T）は機能の名前（I と、その仲間の iii・vi）。一時的な中心や和音の動きの行き先は「主和音」と言う（#390） | `Key.Tonic`, `music.Tonic` |
| 語の注 | 入門者が最初に出会う語には、出てくる文の中で短い注を付ける。導音（主音の半音下の音）、三全音（全音 3 つ分の音程。トライトーン）、属七（V7 と同じ形の和音）、終止の四六（V の前に置く、I の第 2 転回形）。和音の転回形は「第 1 転回形」のように言い、通奏低音の数字（iv6）は使わない（#390） | `music.Analyze` |
| 分数コード (Slash chord) | 和音の下に、根音以外の音をベースとして置いたもの。「C/E」。middy のラベルでは `/` のあとに調の度数を書く（`I/3`）（ADR-0020） | `music.Bass`, `HarmonyEvent.Bass` |
| 転回形 (Inversion) | 和音の音（3 度・5 度・7 度）がベースにある形。第 1 転回形は 3 度、第 2 は 5 度、第 3 は 7 度がベース | `bassText` |
| ペダル (Pedal) | 和音が変わってもベースが同じ音に留まること。和音にない音でも続く | `bassText` |
| ライン・クリシェ (Line cliché) | 同じ和音を保ったまま、1 つの声部（middy ではベース）が半音ずつ動く書き方。Am → Am/G# → Am/G → Am/F# | `lineText` |
| バックビート (Backbeat) | 4/4 の 2・4 拍目に強いスネアを置くこと。ロックやポップの土台 | `pattern.backbeatText` |
| ハーフタイム (Half time) | スネアを 3 拍目だけに置き、テンポの半分の速さで揺れて聞こえるようにすること | `pattern.backbeatText` |
| スウィング比 (Swing ratio) | スウィングで組にした 2 音の長さの比。2 : 1 は 3 連符、3 : 1 は付点と短い音。遅いテンポほど深く、速いほど浅くなる傾向がある | `pattern.Swing`, `DefaultSwingRatio` |
| ボイスリーディング (Voice leading) | 和音が替わるとき、各声部をなるべく近い音へ動かすこと。middy は前の和音から動きが最小の転回形を選ぶ | `music.LeadVoicings` |
| ウォーキング・ベース (Walking bass) | ベースが 4 分音符で歩き、和音の最後の拍で次の和音の根音へ半音で入る弾き方。ジャズの 4 ビート | `pattern.VoiceWalk`, `walkLine` |
| テンション (Tension) | 七の和音の上に重ねる 9 度・11 度・13 度の音。middy ではスタイルが足し、調の音階にある音だけを使う。ローマ数字には書かない | `pattern.Tensions` |
| 裏コード (Tritone substitution) | V7 の代わりに置く、根音が三全音離れた属七。3 度と 7 度が同じなので同じように解決し、根音が半音上から降りる。middy では `subV7/x` と書く（C major の `subV7/I` は D♭7） | `music.Step.Substitute` |
| パッシング・ディミニッシュ (Passing diminished) | 半音ずつ上がるベースの途中に置く減七。次の和音の導音の上に積み、そこへ半音で解決する。middy では `vii°7/x` と書く（C major の I – `vii°7/ii` – ii は C – C#°7 – Dm） | `music.Step.LeadingTone` |
| 掛留（sus4） (Suspension) | 3 度の代わりに 4 度を鳴らす和音。長調とも短調とも決まらない響きで、4 度が 3 度へ下がって解決することが多い。middy では `Vsus4`、`V7sus4` と書く | `music.Sus4` |
| sus2 | 3 度の代わりに 2 度を鳴らす和音。長調とも短調とも決まらない、澄んで開いた響き。sus4 と違い、解決を求めない。middy では `Isus2` と書き、テンションは足さない | `music.Sus2` |
| パワーコード (Power chord) | 3 度を抜き、根音と 5 度だけを鳴らす和音。ロックのギターに多い。middy では `I5`、`bVII5` と書く | `music.Power` |
| フィル (Fill) | フレーズの終わりで、ドラムがいつもの型を崩して次へつなぐ短い連打。middy は 4 小節ごとに入れる | `pattern.fillHits` |
| 押し (Push, anticipation) | 和音を拍の頭より 8 分か 16 分早く弾き、そのまま次の小節へつなぐこと。和音は元の拍に属したまま、音だけが前に出る | `music.Push`, `PushNotes`, `HarmonyEvent.Push` |
| シンコペーション (Syncopation) | 拍の裏で鳴った音のあと、次の拍の頭が鳴らないこと。音が拍をまたいで前に押し出される（Longuet-Higgins & Lee） | `pattern.classify`, `RhythmPart.Syncopated` |
| スタイル (Style) | 曲の種類（ポップ、バラード、ロック、ジャズ、ボサノバ、ケルト）。合う伴奏・ドラム・ノリを束ねる（ADR-0012） | `pattern.Style`, `intent.Candidate.Style` |
| 終止 (Cadence) | 隣接する 2 コードの進行の分類。正格・変格・半・偽・モーダル | `music.Cadence` |
| 解説 (Explanation) | タイムラインの成り立ちを説明するデータ。度数・機能・終止・文章 | `music.Analysis`（表示は `features/explanation`） |
| スライス (Slice) | 1 PR で完結する、UI〜API〜ドメインを縦に貫く小さな機能単位 | `docs/product/backlog.md` |

## 画面の部品

画面の部品の呼び名（#493）。この節の呼び名だけは、決まりではなく、指示のあいまいさを少し和らげる道具として使う。上の表の音楽とデータの用語は、これまでどおり厳密に使い、呼び替えない。ここにない言い方（別の言い回し、日本語と英語、略し方）も通じるものとして扱う。AI は文脈から部品を推し量り、迷ったら聞き返す。文書はこの呼び名を目安にそろえるが、そろっていないことで検査は落とさない。呼び名は DAW の語彙を優先する。

「枠」は単独では使わず、何の枠かを付ける（帯の枠、解説の枠）。単独の「枠」が出たら文脈で判断する。

「タイムライン」は 2 つの意味で使う。広義は、帯を含めたタイムラインの部品全体（`frontend/src/features/timeline/TimelineOverview.tsx`）。狭義は、ノートエリアとトラックヘッダー。本体だけを指すときは「メインのタイムライン」と言う。

| 呼び名 | 別名 | 画面の文言・読み上げ名 | コード |
| --- | --- | --- | --- |
| 会話の列 | 左の列、会話、サイドバー | 中の会話は「イメージから」 | `frontend/src/app/App.tsx` の `.app__side`、`frontend/src/features/chat/ChatPanel.tsx` |
| 結果 | 右の区分、開いている例 | 「結果」 | `frontend/src/app/App.tsx` の `.app__timeline` |
| ステージ | — | — | `frontend/src/app/Stage.tsx`（結果のうち、タイムラインと解説の部分） |
| 会話の引き出し | ドロワー | 「会話」「会話を閉じる」 | `frontend/src/app/ChatDrawer.tsx`（横向きで高さの低い画面だけ。ADR-0073） |
| 入力欄 | — | 「イメージ」 | `frontend/src/features/chat/ChatPanel.tsx` |
| 例のカード | 候補カード、カード | 例の題名。並びは「候補」 | `frontend/src/features/chat/CandidateCard.tsx` |
| これまでの例 | 前の例の行 | 「これまでの例」 | `frontend/src/features/chat/ExampleList.tsx` |
| とっておいた例の一覧 | — | 「とっておいた例の一覧」 | `frontend/src/features/projects/ProjectList.tsx`（上の「とっておいた例」） |
| ツールバー | 再生のツールバー | — | `frontend/src/app/App.tsx` の `.app__toolbar` |
| 再生ボタン | ▶、一時停止ボタン、⏸ | 「再生」「一時停止」 | `frontend/src/features/player/PlayerControls.tsx` |
| 停止ボタン | ■ | 「停止」 | `frontend/src/features/player/PlayerControls.tsx` |
| 位置の表示 | 再生位置の表示、3.1 | 「再生位置」 | `frontend/src/features/player/PlayerControls.tsx` |
| キー、スタイル、MIDI 書き出し | — | 「キー」「スタイル」「MIDI 書き出し」 | `frontend/src/features/key-change/KeySelect.tsx`、`frontend/src/features/regroove/StyleSelect.tsx`、`frontend/src/features/export/DownloadMidiButton.tsx` |
| メインのタイムライン | 本体 | 「タイムライン」 | `frontend/src/features/timeline/TimelineView.tsx`、横にスクロールする箱は `.timeline-overview__scroll` |
| 帯 | 縮小図、縮小タイムライン、オーバービュー | 支援技術から隠す | `frontend/src/features/timeline/TimelineOverview.tsx` の `.timeline--overview` |
| 帯の枠 | 表示範囲、ビューの枠 | — | `.timeline__highlight` |
| ルーラー | 小節番号 | — | `frontend/src/features/timeline/TimelineRuler.tsx` |
| コードトラック | コード列、コードレーン、上部のコード表示 | 「コード」 | `frontend/src/features/timeline/ChordLane.tsx` |
| 和音のブロック | — | 「V7 B7」の形 | `frontend/src/features/timeline/ChordLane.tsx` の `.timeline__chord` |
| トラックヘッダー | 行名の列、音名の列、鍵盤 | — | `frontend/src/features/timeline/TimelineLabels.tsx` |
| ノートエリア | 譜面、音符の画面 | — | ピアノロールとドラムトラック |
| ピアノロール | — | 支援技術から隠す | `frontend/src/features/timeline/PianoRoll.tsx` |
| ドラムトラック | 打楽器の段、ドラムレーン | 「ドラム」 | `frontend/src/features/timeline/DrumLane.tsx` |
| プレイヘッド | 再生位置の線、ヘッド | — | `frontend/src/features/timeline/TimelineView.tsx` の `.timeline__playhead`。指でつかむ帯は `frontend/src/features/timeline/useHeadGrip.ts` |
| 再生位置 | — | — | 時刻そのもの（`positionTick`）。線はプレイヘッド、「3.1」は位置の表示 |
| 「再生位置へ戻る」 | — | 「再生位置へ戻る」 | `frontend/src/features/timeline/TimelineOverview.tsx`（ADR-0068） |
| 解説の枠 | コードの解説、1 和音の解説 | 「コードの解説」 | `frontend/src/features/explanation/ChordFrame.tsx` |
| 度数の地図 | 地図 | 「度数の地図」 | `frontend/src/shared/ui/DegreeMap.tsx` |
| 解説のパネル | 全体の解説 | 「解説」 | `frontend/src/features/explanation/ExplanationPanel.tsx` |
