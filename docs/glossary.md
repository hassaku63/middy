# 用語集

| 用語 | 意味 | コード上の対応 |
| --- | --- | --- |
| イメージ (Image) | ユーザーが言葉で表す曲の印象。「雨の日曜」「出発の朝」など | LLM への入力。`MusicalIntent` の元 |
| 音楽的意図 (MusicalIntent) | イメージを LLM が解釈した構造化パラメータ。調・モード・テンポ・雰囲気タグ・進行テンプレート候補 | `intent.Intent`, `intent.Candidate` |
| 進行 (Progression) | 調に依存しないコード進行テンプレート。度数と拍数の列 | `music.Progression`, `music.FromProgression` |
| パターン (Pattern) | 再利用可能なコード進行テンプレート、またはリズムテンプレート | `pattern.Accompaniment`, `pattern.Drums`, `pattern.Apply` |
| タイムライン (Timeline) | 再生・表示・書き出しの対象となる中心データ。テンポ、拍子、調、トラック、ハーモニー注釈 | `music.Timeline` |
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
| 終止 (Cadence) | 隣接する 2 コードの進行の分類。正格・変格・半・偽・モーダル | `music.Cadence` |
| 解説 (Explanation) | タイムラインの成り立ちを説明するデータ。度数・機能・終止・文章 | `music.Analysis`（表示は `features/explanation`） |
| スライス (Slice) | 1 PR で完結する、UI〜API〜ドメインを縦に貫く小さな機能単位 | `docs/product/backlog.md` |
