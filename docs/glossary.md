# 用語集

| 用語 | 意味 | コード上の対応 |
| --- | --- | --- |
| イメージ (Image) | ユーザーが言葉で表す曲の印象。「雨の日曜」「出発の朝」など | LLM への入力。`MusicalIntent` の元 |
| 音楽的意図 (MusicalIntent) | イメージを LLM が解釈した構造化パラメータ。調・モード・テンポ・雰囲気タグ・進行テンプレート候補 | `intent.Intent`, `intent.Candidate` |
| 進行 (Progression) | 調に依存しないコード進行テンプレート。度数と拍数の列 | `music.Progression`, `music.FromProgression` |
| パターン (Pattern) | 再利用可能なコード進行テンプレート、またはリズムテンプレート | `pattern.Accompaniment`, `pattern.Drums`, `pattern.Apply` |
| タイムライン (Timeline) | 再生・表示・書き出しの対象となる中心データ。テンポ、拍子、調、トラック、ハーモニー注釈 | `music.Timeline` |
| 保存した例 (Saved example) | 名前をつけてブラウザに残した例。あとで開くと同じ例が鳴る。画面はこの呼び名で統一する（コードでは project） | `features/projects`, `shared/lib/storage/projects` |
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
| 度数 (Degree) | 音階上の位置。1 始まり | `Key.Degree`, `Key.Diatonic` |
| ダイアトニックコード | 調の音だけで作られるコード | `Key.Diatonic` |
| ローマ数字 (Roman numeral) | 度数とコードの質をあらわす記法。I, ii, V7, vii° | `music.RomanNumeral` |
| 和声機能 (Harmonic function) | T（トニック）/ S（サブドミナント）/ D（ドミナント） | `music.Function` |
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
| フィル (Fill) | フレーズの終わりで、ドラムがいつもの型を崩して次へつなぐ短い連打。middy は 4 小節ごとに入れる | `pattern.fillHits` |
| 押し (Push, anticipation) | 和音を拍の頭より 8 分か 16 分早く弾き、そのまま次の小節へつなぐこと。和音は元の拍に属したまま、音だけが前に出る | `music.Push`, `PushNotes`, `HarmonyEvent.Push` |
| シンコペーション (Syncopation) | 拍の裏で鳴った音のあと、次の拍の頭が鳴らないこと。音が拍をまたいで前に押し出される（Longuet-Higgins & Lee） | `pattern.classify`, `RhythmPart.Syncopated` |
| スタイル (Style) | 曲の種類（ポップ、バラード、ロック、ジャズ、ボサノバ、ケルト）。合う伴奏・ドラム・ノリを束ねる（ADR-0012） | `pattern.Style`, `intent.Candidate.Style` |
| 終止 (Cadence) | 隣接する 2 コードの進行の分類。正格・変格・半・偽・モーダル | `music.Cadence` |
| 解説 (Explanation) | タイムラインの成り立ちを説明するデータ。度数・機能・終止・文章 | `music.Analysis`（表示は `features/explanation`） |
| スライス (Slice) | 1 PR で完結する、UI〜API〜ドメインを縦に貫く小さな機能単位 | `docs/product/backlog.md` |
