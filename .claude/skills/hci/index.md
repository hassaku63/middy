# HCI の知見の索引

項目の形と書かないものは `SKILL.md`。節は「どの判断を助けるか」で分けている。1 つの知見が複数の判断に効くときは、最も効く節に置き、ほかの節から名前で指す。

確認のしかた（2026-11-09）: 開発環境の網の許可の範囲で、 W3C（w3.org）、MDN、NN/g（nngroup.com）、UX Tigers、jnd.org、gihyo.jp、jfly.uni-koeln.de、bottosson.github.io、billbuxton.com、developer.apple.com が開けた。学術出版社（Elsevier、Springer、Wiley、APA、SAGE、Cambridge、OUP、Taylor & Francis）、DOI、ACM DL と interactions.acm.org（Cloudflare のボット遮断）、PMC、HAL、arXiv、著者の大学ページ（Rutgers、NCSU、Missouri、Northwestern、Stanford、York、UCLA、Chris Harrison）、Netflix、BBC、J-STAGE、Web Archive は網の方針で届かなかった。cs.umd.edu と furnas.people.si.umich.edu は TLS の検証に失敗した（証明書の連鎖・名前の不一致。検証は外さない）。「確認日」の欄に何を開いたかを書く。原論文が開けず NN/g や W3C の解説で確かめた項目は、原論文の名は残したうえでそう書いた。

---

## 1. 何を同時に見せるか（群化・近接・全体と部分）

### ゲシュタルトの群化（近接・類同・共通領域・共通運命・連結）

| 欄 | 内容 |
| --- | --- |
| 出典 | Wertheimer (1923) 原典。総説: Wagemans, J. et al. (2012) A century of Gestalt psychology in visual perception: I. *Psychological Bulletin* 138(6). https://doi.org/10.1037/a0029333 。共通領域の原典: Palmer, S. E. (1992) *Cognitive Psychology* 24。解説: NN/g の Proximity https://www.nngroup.com/articles/gestalt-proximity/ 、Similarity https://www.nngroup.com/articles/gestalt-similarity/ 、Common Region https://www.nngroup.com/articles/common-region/ |
| 何を言っているか | 近い・似ている・同じ枠の中にある・同じ向きに動く・線でつながった要素は、ひとまとまりとして知覚される。原理どうしが競合するとき、近接と共通領域は類同（色・形）より強いことが多く、類同は離れた要素を結べる点で粘り強い。余白・枠・色は「どれが何に属するか」を言葉なしで伝える手段で、使いすぎると枠だらけで構造が消える |
| 確からしさ | 100 年分の実験で原理の存在は安定（Wagemans らの総説）。「どの原理が勝つか」は刺激次第で、定量予測は限定的。NN/g は近接・共通領域が類同を上回る、と実務の観察で言う |
| 当たるべき場面 | タイムライン・解説・ツールバーの要素を「どれが何に属するか」目で分からせたいとき。余白・枠・色で群を作るとき。再生位置と連動して動く要素を決めるとき（共通運命） |
| 判断の問い | 見た目の近さ・枠・動きが、意味の所属（どのコードの解説か、どの例に属するか）と一致しているか。意図しない群化（主操作が従属操作と同じ間隔・重さで並ぶ）が起きていないか |
| 適切・不適切な例 | 適切: コードのチップと、そのコードの理由の文を同じ枠に入れ、別のコードの文とは間を空ける。不適切: ツールバーで再生・キー・MIDI・保存が同じ間隔・同じ重さで並び、主操作が群に埋もれる |
| middy との接点 | ツールバー、タイムラインのコード列、解説の列の区切り、度数の地図（機能ごとの色 = 類同） |
| 確認日 | 2026-11-09（NN/g の 3 記事。Wagemans 2012 と Palmer 1992 の原論文は未確認・網の方針外） |

### 分割注意と近接（説明は対象の隣に、同じ時間に）

| 欄 | 内容 |
| --- | --- |
| 出典 | Chandler, P. & Sweller, J. (1992) The split-attention effect as a factor in the design of instruction. *British Journal of Educational Psychology* 62。Ginns, P. (2006) Integrating information: A meta-analysis of the spatial contiguity and temporal contiguity effects. *Learning and Instruction* 16. https://doi.org/10.1016/j.learninstruc.2006.10.001 。Mayer, R. E. (2021) *Multimedia Learning* 3rd ed., Cambridge UP。実務の観察: Nielsen, J. (2012) Tunnel Vision and Selective Attention, NN/g https://www.nngroup.com/articles/tunnel-vision-and-selective-attention/ |
| 何を言っているか | 互いに参照しないと理解できない情報（図と説明、コードと理由）が離れた場所・離れた時刻にあると、探して結びつける負荷がかかり学習が落ちる。近づける（統合する）と、内容が複雑なときほど改善が大きい。NN/g は実験でなく観察から同じ結論に至る: ユーザーは今見ている領域の外をほとんど見ないので、関係するものは隣に置く |
| 確からしさ | メタ分析が複数あり、効果の向きは安定。大きさは材料の複雑さと学習者の知識で変わる。「同じ時間に」は「同じ時間に消える」を含まない点に注意（→ 2 節「流れて消える情報」） |
| 当たるべき場面 | 解説をどこに置くか。PC の横並びと縮小タイムラインの距離。スマホで理由が画面外に出る問題。ux-principles 原則 3 の根拠 |
| 判断の問い | 対象（鳴っているコード）と説明の間に、視線の移動か、スクロールか、時間差があるか。それは内容の理解に必要な往復か |
| 適切・不適切な例 | 適切: コードの真下にそのコードの理由。不適切: スマホで「再生」を押すと、タイムラインは見えるが見出しは画面外 |
| middy との接点 | 結果の画面の配置全般、スマホの再生直後のスクロール位置 |
| 確認日 | 2026-11-09（NN/g の記事。Ginns 2006、Chandler & Sweller 1992、Mayer の本は未確認・網の方針外） |

### 概観+詳細、ズーム、焦点+文脈

| 欄 | 内容 |
| --- | --- |
| 出典 | Cockburn, A., Karlson, A. & Bederson, B. B. (2008) A Review of Overview+Detail, Zooming, and Focus+Context Interfaces. *ACM Computing Surveys* 41(1). DOI 10.1145/1456650.1456652。著者公開 PDF: https://www.microsoft.com/en-us/research/wp-content/uploads/2008/01/cockburn-ComputingSurveys09.pdf 。Shneiderman (1996) The Eyes Have It（概観が先、ズームと絞り込み、詳細は求めに応じて）。Furnas (1986) Generalized Fisheye Views, *CHI '86* |
| 何を言っているか | 全体と部分を同時に扱う手法は 4 つ: 概観+詳細（空間で分ける）、ズーム（時間で分ける）、焦点+文脈（1 つの表示に埋め込む）、手がかり（強調・抑制）。レビューの結論: 概観+詳細は文書理解の課題で最も強く好まれるが、2 つの表示を頭の中で結びつける注意の分割がコストで、動的な作業では不利になりうる。ズームはアニメーションなしだと迷う。歪みを使う焦点+文脈は空間判断とクリック精度を損ねる。詳細の変化は概観に即時に反映し、概観を触っても詳細は動かないのが慣例 |
| 確からしさ | 多数の実験の系統的レビュー。「どれが最良か」は課題次第で、一般的な指針は出せないと著者ら自身が明言 |
| 当たるべき場面 | タイムラインの縮小図+拡大（BL-080）。解説の「全体・流れ」（文脈）と「今のコード」（焦点）の並べ方。PC の 3 列案（概観+詳細の一種）。長い進行の見せ方 |
| 判断の問い | 焦点と文脈を同時に見せる必要がある作業か、切り替えでよいか。2 つの表示を見比べるコスト（視線の往復）は、再生の速さに間に合うか |
| 適切・不適切な例 | 適切: 4 小節までは全体をそのまま、長ければ縮小図+拡大。拡大を動かすと縮小図が追従する。不適切: 再生中に和音ごとに自動でズームする（再定位のコストが毎回かかる） |
| middy との接点 | タイムラインの縮小図と拡大、解説の全体・流れ／今のコード、度数の地図、PC の列構成 |
| 確認日 | 2026-11-09（Cockburn らの全文。Shneiderman 1996 と Furnas 1986 は未確認・IEEE/ACM と著者ページが開けず） |

### 記憶に頼らせない: 作業記憶の容量、外部記憶、認識 > 想起

| 欄 | 内容 |
| --- | --- |
| 出典 | Miller, G. A. (1956) The magical number seven, plus or minus two. *Psychological Review* 63。Cowan, N. (2001) The magical number 4 in short-term memory. *Behavioral and Brain Sciences* 24(1). https://doi.org/10.1017/S0140525X01003922 。Baddeley & Hitch (1974) の作業記憶。Nielsen, J. (2009) Short-Term Memory and Web Usability https://www.nngroup.com/articles/short-term-memory-and-web-usability/ 。Budiu, R. (2018) Working Memory and External Memory https://www.nngroup.com/articles/working-memory-external-memory/ 。Budiu, R. (2024) Memory Recognition and Recall in User Interfaces https://www.nngroup.com/articles/recognition-and-recall/ 。W3C WAI (2021) Making Content Usable for People with Cognitive and Learning Disabilities https://www.w3.org/TR/coga-usable/ |
| 何を言っているか | 見えなくなった情報を頭の中で保持できる量は小さく、数秒〜20 秒ほどで薄れる。容量の数字は資料で違う（NN/g は「約 7 チャンク」、W3C COGA は平均 7・障害があれば 1〜3、Cowan はリハーサルを封じると 3〜5）が、設計への含意は同じ: 保持させず、画面に出しておく（外部記憶）。候補が見えていれば選べるが、手がかりなしに思い出すのは難しい（認識 > 想起）。NN/g は「メニューは 7 個まで」を誤解だと明記している。画面に並ぶものは見えているので記憶の容量は直接は当たらない |
| 確からしさ | 容量の存在と個人差（年齢・教育で大きく違う）は安定。正確な数とその原因（減衰か干渉か）は論争。**よくある誤用:** 7±2 や 4 を「項目数の上限」に使う。当たるのは、見えなくなった情報を保持させる場面だけ |
| 当たるべき場面 | ユーザーに「覚えておいて、あとで照合する」ことを求めていないかの点検（流れていった前のコードの説明と今のを比べる、記法を覚えてキーを入力する）。空の画面の入力例。長い一覧で真ん中が流れる懸念 |
| 判断の問い | この画面で、ユーザーが見えないものを何個保持しなければ目的を果たせないか。それを画面に出しておけば保持は不要になるか。この操作は見えているものを選ぶ形か、記法や前の状態を思い出す形か |
| 適切・不適切な例 | 適切: 直前のコードの名前と機能を薄く残す。度数の地図でモードの 7 和音を常に見せ、例で使う和音を強調する。不適切: 「候補は 4 個以内」と見えている項目数に機械的に適用する。キーの変更が記号の入力を要求する |
| middy との接点 | 解説の文の切り替え、候補カード、度数の地図、キーの選択、言い直し後に畳まれる履歴 |
| 確認日 | 2026-11-09（NN/g の 3 記事と W3C COGA。Miller 1956、Cowan 2001 の原論文は未確認・網の方針外） |

### 行長と読みやすさ

| 欄 | 内容 |
| --- | --- |
| 出典 | W3C WCAG 2.2 SC 1.4.8 Visual Presentation (AAA) と Understanding https://www.w3.org/WAI/WCAG22/Understanding/visual-presentation.html 。Dyson, M. C. & Haselgrove, M. (2001) The influence of reading speed and line length on the effectiveness of reading from screen. *IJHCS* 54(4). https://doi.org/10.1006/ijhc.2001.0458 。Dyson (2004) *Behaviour & IT* 23(6)（総説） |
| 何を言っているか | WCAG 1.4.8 は、1 行を 80 字以内（CJK は 40 字以内）にできる仕組み、両端揃えにしない、行間 1.5 倍以上、200% 拡大で横スクロールなし、を挙げる。Understanding は、行が長いと次の行の頭に戻りにくい読者がいること、CJK の 1 字は欧文の約 2 字幅に当たることを根拠に書く。Dyson らの画面実験では 55 字（英数）前後で理解の成績が最良で、55〜100 字で読む速さに差がなかったとされる（原論文は未確認） |
| 確からしさ | WCAG は規範（AAA。値は上限であり最適値ではない）。画面の行長の実験は 2000 年前後のものが中心で数は多くなく、日本語（分かち書きなし）での直接の実験は見つけていない。「100 字が最速」という通俗的な要約は支持が弱い |
| 当たるべき場面 | 解説の列の幅（PC で 248 px、1 行 15 字程度）を決めるとき。文が縦に細長くなっていないかの点検 |
| 判断の問い | 1 行の文字数は、文を「列」ではなく「文」として読める長さか（日本語で上限 40 字。15 字は短すぎて行替えが増える）。200% に拡大しても横スクロールが出ないか |
| 適切・不適切な例 | 適切: PC の空いている横幅を解説に回し、理由を 25〜35 字/行にする。不適切: 248 px の列に 110 字の文を 7〜8 行で積む |
| middy との接点 | 解説の列の幅、PC の列構成 |
| 確認日 | 2026-11-09（WCAG 2.2 Understanding 1.4.8。Dyson らの原論文は未確認・網の方針外） |

### 強調と手がかり: 前注意的特徴、シグナリング、同期ハイライトの限界

| 欄 | 内容 |
| --- | --- |
| 出典 | Treisman, A. & Gelade, G. (1980) A feature-integration theory of attention. *Cognitive Psychology* 12。Healey, C. G. & Enns, J. T. (2012) Attention and visual memory in visualization and computer graphics. *IEEE TVCG* 18(7)。解説: Laubheimer, P. (2017) Dashboards: Making Charts and Graphs Easier to Understand, NN/g https://www.nngroup.com/articles/dashboards-preattentive/ 。Richter, J., Scheiter, K. & Eitel, A. (2016) Signaling text-picture relations in multimedia learning: A comprehensive meta-analysis. *Educational Research Review* 17. https://doi.org/10.1016/j.edurev.2015.12.003 。Gerbier, E., Bailly, G. & Bosse, M.-L. (2018) Audio–visual synchronization in reading while listening to texts. *Computer Speech & Language* 47 |
| 何を言っているか | 色・明度・向き・大きさ・長さ・2 次元の位置・動きなど少数の特徴は、周囲と 1 つだけ違えば一瞬で浮き上がる（前注意的）。組み合わせ（赤くて丸い）はそうならず探索になる。NN/g は、長さと位置は量に、色と形はカテゴリに向くとし、色覚多様性のため色は「2 番目の手がかり」にせよと言う。図と文の対応を色・矢印・同時の強調で示すと（シグナリング）、自力で対応づけできない初心者ほど理解が上がる（メタ分析、小〜中の効果）。ただし読み上げに同期して語を光らせる実験では、視線は誘導されたが学習は増えなかった |
| 確からしさ | 前注意的特徴は実験の積み重ねがある（理論の細部は修正されたが「単一特徴なら速い」は安定）。シグナリングは複数のメタ分析で向きが一致。「同期ハイライトで読める・分かる」は根拠がなく、視線の誘導にとどまる |
| 当たるべき場面 | 「今鳴っているコード」を文を読まずに一瞬で見つけさせたいとき。強調に使う特徴を 1 つに絞るか決めるとき。タイムライン・度数の地図・解説の見出しの 3 か所で同じ和音を結ぶ方法。ピアノロールや再生位置の作り込みの期待値 |
| 判断の問い | 現在位置の強調は、1 つの特徴（明度か枠か動き）で、周囲のどれとも違うか。機能色と競合していないか。同じ和音は 3 か所で同じ色・同じ形で結ばれているか。同期ハイライトに期待しているのは「どこが鳴っているか」（根拠あり）か「解説が頭に入る」（根拠なし）か |
| 適切・不適切な例 | 適切: 鳴っているコードだけ明度を変え、ほかは同じ彩度に抑える。解説の見出しのコード名をタイムラインの該当チップと同じ色・形で出す。不適切: 機能色の上に現在位置も別の色で塗る。文を 1 語ずつカラオケ式に光らせて読めている気にさせる |
| middy との接点 | タイムラインの `aria-current` の見た目、度数の地図の「使う和音」と「解説中の和音」の 2 段、解説の h2、理由の文の用語 |
| 確認日 | 2026-11-09（NN/g の記事。Treisman & Gelade、Healey & Enns、Richter ら、Gerbier らの原論文は未確認・網の方針外） |

### 画面幅のクラスと配置の切り替え

| 欄 | 内容 |
| --- | --- |
| 出典 | Apple *Human Interface Guidelines* "Layout"（compact / regular の size class）https://developer.apple.com/design/human-interface-guidelines/layout 。Material Design 3 "Window size classes"（compact / medium / expanded / large / extra-large） |
| 何を言っているか | 配置の分岐は「スマホか PC か」ではなく「使える幅（と高さ）がどれだけか」で決める。幅が変わっても、できることは変えない（配置だけ変える）。安全領域と余白を尊重し、文字サイズの拡大でも崩れないようにする |
| 確からしさ | プラットフォームの実務指針 |
| 当たるべき場面 | PC の列構成とスマホの縦積みを切り替える幅を決めるとき。1280×800 の E2E と 1440×900 の測定、390 幅のスマホ |
| 判断の問い | 分岐の条件は幅（と高さ）か。幅が狭いときに、できること（固定、キー変更）が失われていないか。文字を大きくしたら崩れないか |
| 適切・不適切な例 | 適切: 幅が広ければ会話／舞台／解説を横に、狭ければ再生とコード列を上に固定して縦に。不適切: `userAgent` で分岐する。middy では縦の高さも効く（1280×800 では舞台が 227 px しか見えない） |
| middy との接点 | 結果の画面の配置、会話と結果の上下関係、入力欄の位置。規則としきい値の表は `docs/dev/conventions.md` の「配置」（ADR-0042） |
| 確認日 | 2026-11-09（Apple HIG。Material 3 は JS 描画で本文が取れず未確認）。2026-11-10 に Android Developers の window size classes と "Support different display sizes" を確認。どちらも Material 3 の境目に対応し、物理的な端末の値で配置を決めない、部品は与えられた幅を使う、と明記している（https://developer.android.com/develop/ui/compose/layouts/adaptive/support-different-display-sizes ） |

## 2. 時間で変わる情報（再生に同期して出す・消える）

### 流れて消える情報: 一過性情報効果と WCAG 2.2.1 / 2.2.2

| 欄 | 内容 |
| --- | --- |
| 出典 | Leahy, W. & Sweller, J. (2011) Cognitive load theory, modality of presentation and the transient information effect. *Applied Cognitive Psychology* 25(6)。Wong, A., Leahy, W., Marcus, N. & Sweller, J. (2012) *Learning and Instruction* 22。W3C WCAG 2.2 SC 2.2.1 Timing Adjustable (A) https://www.w3.org/WAI/WCAG22/Understanding/timing-adjustable.html 、SC 2.2.2 Pause, Stop, Hide (A) https://www.w3.org/WAI/WCAG22/Understanding/pause-stop-hide.html 。Nielsen, J. (2013) Auto-Forwarding Carousels and Accordions Annoy Users and Reduce Visibility, NN/g https://www.nngroup.com/articles/auto-forwarding/ |
| 何を言っているか | 出て消える情報（音声、アニメーション、自動で切り替わる文）は作業記憶で保持し続けねばならず、留まる情報より負荷が高い。短ければ差は出ないが、長くなると留まる文字の方が良くなる。分節して短くすると害は減る。WCAG はこれを規範にしている: 読む速さを超えて進む・更新される内容は「時間制限」であり（2.2.1）、自動で始まり他の内容と並行して 5 秒を超えて動くもの、自動更新されるものには、一時停止・停止・非表示か更新頻度の制御が要る（2.2.2。活動に必須なら例外）。同期字幕やカルーセルは、操作で戻れる限り 2.2.1 の対象外。5 秒で消えるトーストは、同じ情報が別の場所で得られるなら可。NN/g の観察: 5 秒ごとに切り替わる箱は「読む時間がなかった」と言われ、大きく出ていても見つけてもらえない |
| 確からしさ | 一過性情報効果は 2010 年代に複数の実験で反復（原論文は未確認）。「どこから長いか」の境界は材料次第で、一般化された秒数はない。WCAG は規範（A） |
| 当たるべき場面 | **再生中に解説の文を自動で切り替える設計そのもの。** middy の文は文字だが、1.8〜3.3 秒で消えるので行動としては一過性情報になっている。文を短くするか、留めるか（固定・一覧）、分節するかの判断。トーストや通知の表示時間 |
| 判断の問い | この情報はユーザーが読み終える前に消えるか。消えるなら、あとで戻れる場所に同じ情報が留まっているか。短いもの（見出し）と長いもの（理由）で、流すものと留めるものを分けているか。自動で切り替わる領域に、止める・頻度を変える手段があるか |
| 適切・不適切な例 | 適切: 短い見出し（コード名・機能・終止）だけを再生と同期して流し、長い理由は留まる一覧か、止めたときに読む。不適切: 110 字の半終止の説明を 1.8 秒のコードに同期して差し替える |
| middy との接点 | 結果の画面の解説の文、再生・停止、固定、更新の通知 |
| 確認日 | 2026-11-09（WCAG 2.2 Understanding 2.2.1 と 2.2.2、NN/g の記事。Leahy & Sweller 2011、Wong ら 2012 は未確認・網の方針外） |

### 分節化と学習者ペース（止められるだけでは足りない）

| 欄 | 内容 |
| --- | --- |
| 出典 | Mayer, R. E. & Chandler, P. (2001) When learning is just a click away. *Journal of Educational Psychology* 93(2)。Rey, G. D. et al. (2019) A meta-analysis of the segmenting effect. *Educational Psychology Review* 31. https://doi.org/10.1007/s10648-018-9456-4 。Tabbers, H. K., Martens, R. L. & van Merriënboer, J. J. G. (2004) *British Journal of Educational Psychology* 74。関連: W3C Media Accessibility User Requirements（拡張字幕: ユーザーが読む速さを決め、追いつくために再生を止められること）https://www.w3.org/TR/media-accessibility-reqs/ 。Schade, A. (2014) Video Usability, NN/g https://www.nngroup.com/articles/video-usability/ |
| 何を言っているか | 連続して流れる説明を、学習者が「次へ」で進める分節にすると、保持と転移が上がる（メタ分析で小〜中の効果）。ただしシステムが決めた間隔で止める分節と、学習者が止める分節は別物で、初心者は「いつ止めるべきか」を知らないため、止められるだけでは利益が小さいという報告がある。学習者が自分のペースで進める条件では「音声の方が文字より良い」という様相効果が消える・逆転する。W3C の MAUR は、拡張字幕の要件として、読む速さをユーザーが決め、追いつくために再生を一時停止できることを挙げる。NN/g も映像は「制御できること」「逐次にしか読めない欠点を補う代替」を条件にする |
| 確からしさ | 分節化の効果はメタ分析で支持（原論文は未確認）。学習者ペースの境界条件（初心者は止めどころを知らない）は 2020 年前後の研究で議論中。MAUR は W3C のノート（規範ではない） |
| 当たるべき場面 | 再生と解説の同期の設計。音楽の再生は本質的にシステムペースで、音を止めずに文だけ止めることはできない。この制約の中で、何を学習者のペースにするか |
| 判断の問い | ユーザーは読みたいときに「読むためのペース」を自分で取れるか（止める、戻る、1 和音ずつ進める）。止める操作は聴きながらでも届く大きさ・位置か。初めての人に「ここで止めて読むと良い」が伝わるか |
| 適切・不適切な例 | 適切: 止めると位置が残ってその和音の理由が出る。タイムラインのコードを押すとそこへ飛ぶ（1 和音 = 1 分節）。不適切: 再生中は読めず、止めると「鳴っている」強調が消えて、どの和音の解説か分からなくなる |
| middy との接点 | 再生・停止、固定、縮小図のコードを押して移動（BL-080）、experience.md の「止まっているあいだは強調しない」の約束（止めた位置を残すなら読み替えが要る） |
| 確認日 | 2026-11-09（W3C MAUR、NN/g Video Usability。Mayer & Chandler、Rey ら、Tabbers らの原論文は未確認・網の方針外） |

### 読みの速さと字幕の基準

| 欄 | 内容 |
| --- | --- |
| 出典 | Nielsen, J. (2008) How Little Do Users Read?, NN/g https://www.nngroup.com/articles/how-little-do-users-read/ （英語の読む速さの前提 200 wpm、高リテラシーで 250 wpm）。豊倉省子・山田優 (2017) 字幕翻訳における文字数制限の妥当性についての検証〜1秒4文字は妥当か？〜. *通訳翻訳研究への招待* 18。Netflix "Japanese Timed Text Style Guide"。BBC Subtitle Guidelines https://www.bbc.co.uk/accessibility/forproducts/guides/subtitles/ 。Trauzettel-Klosinski & Dietz (2012) IReST, *IOVS* 53(9) |
| 何を言っているか | 英語の読みは実務で 200〜250 語/分が前提に置かれる（NN/g）。日本語の黙読は平均 600 字/分前後（約 10 字/秒）で個人差が 3 倍以上あると紹介されるが、原資料は未取得。映像の字幕は絵も見ながら読むので、ずっと遅い基準が使われる: 日本の字幕の慣習は 1 秒 4 文字で、豊倉・山田の視線計測では 4 字と 6 字に差がなく 8 字/秒で負荷が上がったとされる。Netflix の日本語ガイドは 4 字/秒（日本語音声の字幕は 7 字/秒）、英語の放送字幕は 160〜180 語/分（いずれも未確認） |
| 確からしさ | 字幕の数値は業界の規範で、科学的な裏づけは限定的。黙読速度は個人差が非常に大きく、平均で設計すると半分の人に速すぎる。**この項目の日本語の数値は、この session では一次情報を開けていない。数値を根拠に使う前に開くこと。** 向きだけは確か: 音を聴きながら絵も見る状況では、黙読の速さではなく字幕の基準に近い |
| 当たるべき場面 | 再生中に出す文の文字数の上限、文が留まる最短時間、1 行の文字数を決めるとき |
| 判断の問い | この文は、そのコードが鳴っている秒数 × 字幕の速さ（4〜6 字/秒。要確認）に収まるか。収まらない文は、流さずに留める設計になっているか |
| 適切・不適切な例 | 適切: 1.8 秒のコードには「V7 半終止」程度だけ流す。不適切: 1.8 秒に 110 字（読むには 10 秒以上かかる） |
| middy との接点 | 解説の見出しと理由の分離、再生中に流す文の上限 |
| 確認日 | 2026-11-09（NN/g の記事のみ。豊倉・山田、Netflix、BBC、IReST、日本語黙読の原資料は未確認・網の方針外） |

### 注意の捕捉と見落とし（出現・動き始め、変化の見落とし、選択的注意）

| 欄 | 内容 |
| --- | --- |
| 出典 | Yantis, S. & Jonides, J. (1984) Abrupt visual onsets and selective attention. *JEP: HPP* 10(5)。Abrams & Christ (2003) Motion onset captures attention. *Psychological Science* 14。Rensink, O'Regan & Clark (1997) To see or not to see. *Psychological Science* 8。Simons & Chabris (1999) Gorillas in our midst. *Perception* 28。解説: Harley, A. (2014) Animation for Attention and Comprehension, NN/g https://www.nngroup.com/articles/animation-usability/ 。Budiu, R. (2018) Change Blindness in UX https://www.nngroup.com/articles/change-blindness-definition/ 。Nielsen, J. (2012) Tunnel Vision and Selective Attention https://www.nngroup.com/articles/tunnel-vision-and-selective-attention/ |
| 何を言っているか | 周辺視野の動き、とくに新しく現れるものと位置が変わるものは、意図と関係なく注意を引く（スライドインはフェードインより強く引く）。逆に、視線が外れた瞬間（瞬き、サッカード、再描画）に起きた変化や、注意を向けていない領域の変化は、大きくても気づかれない。操作に応じて 1 か所が変わると、ユーザーはそこを見るので、同時に別の場所で起きた変化は見落とされる。「見ている = 気づいている」ではない。因果を感じさせる動きは操作から 0.1 秒以内に始める |
| 確からしさ | 出現・動き始めの捕捉、変化の見落とし、非注意性盲目はいずれも実験の積み重ねが厚い（再現性の議論がある心理学の中では安定した部類。原論文は未確認）。NN/g の観察とも一致 |
| 当たるべき場面 | 再生中に何かを動かす・切り替えるとき。解説の文が切り替わる瞬間、通知やトーストを出すとき。ユーザーが音に注意を向けている間に画面側で何かを変えるとき。「出しておけば読むだろう」と考えたくなったとき |
| 判断の問い | 再生中に、注意を向けてほしい場所以外で、出現や動き始めが起きていないか。聴くことに集中しているユーザーがこの変化に気づく根拠はあるか。気づかなくても困らない（あとで戻れる）か。操作に応じて現れるものは、操作した場所の近くに出るか |
| 適切・不適切な例 | 適切: 再生中に動くのは再生位置の印だけにし、解説は固定の一覧の中でハイライトだけが移る。再生後も鳴った和音の一覧と理由が残る。不適切: 再生中に解説列の高さが毎コード変わって下の要素が跳ねる。再生中にだけ出て終わると消える解説 |
| middy との接点 | 再生位置の印、解説の文の切り替え、固定時に全体の文が 1 行増えて理由が押し下がる挙動、更新通知、言い直し後に「最初の例が選ばれた状態に戻る」挙動 |
| 確認日 | 2026-11-09（NN/g の 3 記事。Yantis & Jonides、Abrams & Christ、Rensink ら、Simons & Chabris の原論文は未確認・網の方針外） |

### 動きと「動きを減らす」設定

| 欄 | 内容 |
| --- | --- |
| 出典 | WCAG 2.2 SC 2.3.3 Animation from Interactions (AAA) https://www.w3.org/WAI/WCAG22/Understanding/animation-from-interactions.html 、SC 2.2.2 Pause, Stop, Hide (A)。*Media Queries Level 5* `prefers-reduced-motion` https://www.w3.org/TR/mediaqueries-5/#prefers-reduced-motion 。MDN https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-reduced-motion 。Apple HIG "Motion" |
| 何を言っているか | 操作をきっかけに起きる必須でない動き（視差、ズーム演出）は切れるようにする。自動で 5 秒以上動き続けるものは止められるようにする。前庭障害の人には吐き気・頭痛を起こしうる。OS の「動きを減らす」設定は `prefers-reduced-motion: reduce` で読める。再生位置の移動のような、情報そのものである動きは「必須」に当たる |
| 確からしさ | 規範（WCAG）と仕様（CSS。主要ブラウザで実装済み） |
| 当たるべき場面 | 再生中の強調の移動、縮小図から拡大への移動のアニメーション、候補の切り替え演出を足すとき。「ズームの遷移はアニメーションで負荷が減る」（概観+詳細の項）と両立させる |
| 判断の問い | この動きは情報か（再生位置）、演出か。演出なら `prefers-reduced-motion` で止められるか。自動で動き続けるものは止められるか |
| 適切・不適切な例 | 適切: 拡大の 4 小節が移るときの滑らかな移動は、reduce 設定では即時切り替えにする。不適切: 再生中に背景が脈打つ装飾 |
| middy との接点 | タイムラインの再生位置、拡大の移動、候補の切り替え、待機中の表示。フロントに `prefers-reduced-motion` の参照はまだない（`prefers-color-scheme` はある） |
| 確認日 | 2026-11-09（WCAG 2.2、MQ5、MDN、Apple HIG） |

## 3. 状態とモード

### モードとモードエラー、準モード（quasimode）

| 欄 | 内容 |
| --- | --- |
| 出典 | Sellen, A., Kurtenbach, G. & Buxton, W. (1992) The prevention of mode errors through sensory feedback. *Human-Computer Interaction* 7(2). 著者公開版: https://billbuxton.com/ModeErrors.html 。Raskin, J. (2000) *The Humane Interface*（モードの定義、quasimode）。Laubheimer, P. (2019) Modes in User Interfaces, NN/g https://www.nngroup.com/articles/modes/ |
| 何を言っているか | モード = 同じ入力が状態によって違う結果になること。モードエラーは、状態を取り違えて、別の状態なら正しい操作をしてしまうスリップ。Sellen らの実験では、画面全体の色を変える視覚フィードバックも効いたが、ペダルを踏み続けて保つ「ユーザー自身が維持するモード」の方がエラーも負荷も減った。Raskin は「注意の焦点に状態があるか」でモードを定義し、押している間だけ有効な準モードを勧める。NN/g は、モードを使うなら表示を 2 つ以上重ねて示し、取り返しのつかない場面では使うなと言う。「同じデータの別の見え方」はモードではない |
| 確からしさ | モードエラーの存在とフィードバックでの低減は実験で支持。「すべてのモードをなくす」は理念で、選択状態や文脈もモードの一種 |
| 当たるべき場面 | 「再生に追従して解説が流れる」状態と「固定したコードを見る」状態のように、同じ操作（コードを押す、再生を押す）の意味が状態で変わるとき。トグル（固定／解除、再生／停止）を足すとき。ux-principles 原則 5 の根拠 |
| 判断の問い | 今どちらの状態かは、注意を解説に向けているユーザーにも見えるか。取り違えたとき、結果は取り返せるか。押している間だけ有効（準モード）にできないか |
| 適切・不適切な例 | 適切: 固定中は解説の見出しに印と解除ボタンを隣に置き、タイムライン側でも固定したコードを別の強調にする。不適切: 固定中かどうかが解説の文章の中にしか書かれていない。middy の「固定 = 追従を止める」は実質モードで、再生を押し直したとき追従に戻るのか固定を保つのかを一意に決めて見せる必要がある |
| middy との接点 | 追従／固定、再生／停止、例の切り替え中の状態 |
| 確認日 | 2026-11-09（Sellen らの全文、NN/g。Raskin は未確認・書籍） |

### 実行の淵と評価の淵、7 段階の行為

| 欄 | 内容 |
| --- | --- |
| 出典 | Hutchins, E. L., Hollan, J. D. & Norman, D. A. (1985) Direct Manipulation Interfaces. *Human-Computer Interaction* 1(4). DOI 10.1207/s15327051hci0104_2。Norman, D. A. (2013) *The Design of Everyday Things, Revised* 第 2 章。解説: Whitenton, K. (2018) The Two UX Gulfs, NN/g https://www.nngroup.com/articles/two-ux-gulfs-evaluation-execution/ |
| 何を言っているか | 人は「目標 → 意図 → 計画 → 実行 → 知覚 → 解釈 → 評価」を繰り返す。目標から操作へ渡る側の隔たりが実行の淵、画面の状態から「目標に対してどうなったか」を読み取る側が評価の淵。評価の淵は「見えている」だけでは埋まらず「意味が解釈できる」必要がある。評価が外れると、その後の実行も外れる |
| 確からしさ | 広く使われる枠組み。実験で検証された法則ではなく、失敗の原因を分解する道具 |
| 当たるべき場面 | 操作後に何が起きたかが伝わっているかの点検。状態（再生中、固定中、キー変更後）の表示。新しい操作の前に「ユーザーはどの状態を見て判断するか」を書くとき |
| 判断の問い | この操作の結果を、ユーザーはどこを見て、どう解釈するか。今の状態（再生中か、固定中か、どの例か）は見えて、かつ意味がわかるか |
| 適切・不適切な例 | 適切: キーを変えると、コード名の表記が変わると同時に「ローマ数字は同じ」ことがその場でわかる。不適切: トグルのラベルが「今の状態」か「押すと起きること」か曖昧 |
| middy との接点 | 再生、固定／解除、キー・スタイルの変更、例の切り替え、待機中の表示 |
| 確認日 | 2026-11-09（NN/g の解説と DOET の目次。Hutchins ら 1985 の本文は未確認・網の方針外） |

### 概念モデル、システムイメージ、自然な対応づけ、制約

| 欄 | 内容 |
| --- | --- |
| 出典 | Norman, D. A. (2013) *The Design of Everyday Things, Revised and Expanded*, Basic Books。第 1 章（概念モデル／システムイメージ）、第 3 章（自然な対応づけ）、第 4 章（4 種の制約、強制機能）。目次: https://jnd.org/the-design-of-everyday-things-revised-and-expanded/ 。Nielsen のヒューリスティクス 2「現実世界との一致」 |
| 何を言っているか | 設計者の頭の中のモデルは直接ユーザーに届かない。ユーザーは「システムイメージ」（画面・文書・振る舞いから見えるもの全部）から自分のモデルを組み立てる。操作子と結果の空間的な対応（自然な対応づけ）と、できることを絞る制約は、説明なしでモデルを伝える手段。対応づけの一部（左右と上下、色の意味）は文化で変わる |
| 確からしさ | 枠組み。実務での合意は厚いが、個々の対応づけの効果は文脈で測る |
| 当たるべき場面 | 新しい概念（experience.md の表）を足すとき、画面でどう見えるかを決めるとき。「例」「成り立ち」「固定」の関係が画面から読めるかの点検 |
| 判断の問い | 画面と振る舞いだけから、ユーザーは「例・成り立ち・固定」の関係を正しく組み立てられるか。操作子の並びは、それが効く対象の並びと対応しているか |
| 適切・不適切な例 | 適切: タイムラインの上に縮小図、下に拡大（位置の対応が時間の対応になっている）。不適切: 解説の文が、どのコードの説明かを文の中でしか示さない |
| middy との接点 | experience.md の概念表、度数の地図、タイムラインの縮小図と拡大、キーの選択 |
| 確認日 | 2026-11-09（目次のみ。本文は未確認・書籍） |

### 直接操作と委任（会話型 AI の UI、明示の壁、狭いスコープ）

| 欄 | 内容 |
| --- | --- |
| 出典 | Shneiderman, B. (1983) Direct Manipulation: A Step Beyond Programming Languages. *IEEE Computer* 16(8)。Shneiderman, B. & Maes, P. (1997) Direct manipulation vs. interface agents. *ACM Interactions* 4(6)（著者公開 PDF は cs.umd.edu）。Nielsen, J. (2023, 2025 更新) The Articulation Barrier, UX Tigers https://www.uxtigers.com/post/ai-articulation-barrier 。Moran, K. (2025) Scope in Generative AI Features, NN/g https://www.nngroup.com/articles/scope-ai-features/ 。Mugunthan, T. (2024) Response Outlining with Generative-AI Chatbots, NN/g |
| 何を言っているか | 直接操作は、対象を常に表示し、対象への操作で動かし、小刻みで取り消し可能にし、結果をすぐ見せる型。予測可能性と制御感が利点で、対象が見えない・多すぎる作業では委任（エージェント）が向く。散文で意図を書かせる UI は書く力を要求し多くの人に重い（Nielsen の推論）。狭いスコープの AI 機能は目的が見えやすく、短い入力で済み、出力の形が予測できるので構造化した UI にでき、採用がよかった（NN/g の定性調査）。推奨は GUI で選べる要素を足したハイブリッド |
| 確からしさ | 直接操作は概念の枠組みで、多数の事例と実験で支持されるが「常に優れる」わけではない。会話型 AI の知見は意見記事と少人数の定性調査で、方向は一致、実験の積み重ねは薄い |
| 当たるべき場面 | AI に言葉で頼む操作と、画面で直接いじる操作の切り分け。入力例・言い直しの支援。「入口はイメージだけ」の約束の運用 |
| 判断の問い | この操作の結果はユーザーが予見できるか（直接操作向き）、何が出てくるか自体を相談したいのか（会話向き）。ユーザーが「どう言えばいいか」を知らないとき、画面は手がかり（入力例、押せる選択肢）を出しているか |
| 適切・不適切な例 | 適切: キー・スタイルの変更を AI に聞き直さず画面で行う。空の画面に入力例を出す。不適切: 「もう少し明るく」という解釈の要る変更をスライダーで数値指定させる。middy のスコープ「イメージ → 和音とリズムの例」は狭く、出力の形が予測できるので候補カードやタイムラインとして構造化できる、という強みを活かす |
| middy との接点 | 会話の入力欄と入力例、候補カード、キー・スタイル・固定・例の切り替え、保存・書き出し |
| 確認日 | 2026-11-09（UX Tigers、NN/g の 2 記事。Shneiderman 1983 と 1997 は未確認・IEEE と cs.umd.edu の TLS 検証失敗） |

### ヒューリスティクス（Nielsen の 10 と、Shneiderman の黄金律の独自 2 点）

| 欄 | 内容 |
| --- | --- |
| 出典 | Nielsen, J. (1994, 2020 更新・2024 再確認) 10 Usability Heuristics for User Interface Design, NN/g https://www.nngroup.com/articles/ten-usability-heuristics/ 。元は Nielsen & Molich (1990) と 249 件の問題の因子分析（Nielsen 1994）。Shneiderman ほか *Designing the User Interface* 第 6 版（2016）の 8 つの黄金律（著者の一覧 https://www.cs.umd.edu/~ben/goldenrules.html ） |
| 何を言っているか | 状態の可視化、現実との一致、制御と自由（取り消し・脱出）、一貫性と標準、エラーの予防（スリップとミステイクを分ける）、記憶より認識、柔軟性と効率、美的で最小限、エラーの認識と回復、ヘルプ。ヒューリスティクスは見つけた問題を分類する言葉で、満たせば使いやすいという十分条件ではない。黄金律で独自なのは「作業のまとまりに区切りと完了感を与える」「ユーザーが始めて、システムが応じる」 |
| 確からしさ | 1994 年版は因子分析に基づく。黄金律は実務の経験則。**よくある誤用:** チェックリストとして使い、満たしたら終わりにする |
| 当たるべき場面 | ux-principles にない観点の点検: 制御と自由（固定の解除、言い直しの取り消し、保存の削除）、一貫性と標準、エラーの予防（スリップとミステイクで対策が違う）、エラーメッセージ（何が起きたか・次にすること）、流れ（送る → 待つ → 届く → 聴く）の区切りと完了感 |
| 判断の問い | 間違えたとき 1 手で戻れるか。この間違いはスリップ（注意 → 制約と既定値）か、ミステイク（モデル → 説明と確認）か。この流れのどこで「終わった」とわかるか。システムが勝手に状態を変えていないか |
| 適切・不適切な例 | 適切: 例が届いたら最初の例が選ばれ、再生できる状態で止まる（届いた = 区切り）。固定の解除が解説の隣にある。不適切: 待つあいだに開いていた別の例が、届いた例に勝手に切り替わる。エラーに符号だけ出す |
| middy との接点 | 固定の解除、保存の一覧と削除、失敗時の表示、送って待つ → 届く の流れ、オフラインの通知 |
| 確認日 | 2026-11-09（NN/g。黄金律は cs.umd.edu の TLS 検証失敗で未確認・項目名は複数の二次資料で一致） |

## 4. 操作の対象と手がかり

### アフォーダンス（Gibson）とシグニファイア（Norman）

| 欄 | 内容 |
| --- | --- |
| 出典 | Gibson, J. J. (1979) *The Ecological Approach to Visual Perception*。Norman, D. A. (2008) Signifiers, not affordances. *ACM Interactions* 15(6). 著者公開版: https://jnd.org/signifiers-not-affordances/ 。Norman (2013) DOET 第 1 章 |
| 何を言っているか | Gibson のアフォーダンスは環境と行為者の関係で、知覚されるかどうかと無関係に存在する。Norman は 1988 年に「知覚されたアフォーダンス」の意味で設計に持ち込んだが誤用が広がり、2008 年に「設計者が与えるべきは、何ができるかを示す手がかり = シグニファイア」と言い直した |
| 確からしさ | 概念の枠組みで、実験による検証ではない。**よくある誤用:** 「ボタンにアフォーダンスを付ける」。問題になるのはアフォーダンスの有無ではなく、見えるか（シグニファイア） |
| 当たるべき場面 | 要素が操作できることをどう示すか。押せるもの／押せないものの見分け。新しい操作（固定、縮小図の押下）を足すとき |
| 判断の問い | 初見の人が、これを押せる・動かせると気づく手がかりは何か。押せないものが押せそうに見えていないか |
| 適切・不適切な例 | 適切: リンクは色と下線、ボタンは枠と形で、押せることが見てわかる。不適切: 見出しに背景色を付けてボタンに見えるが押せない。middy: タイムラインのコードは押すと固定できるが、その手がかりがコードの見た目にない |
| middy との接点 | タイムラインのコード（`role="option"`）、縮小図、度数の地図、候補カード、キーの選択 |
| 確認日 | 2026-11-09（Norman 2008 の著者公開版と DOET の目次。Gibson 1979 は未確認・書籍） |

### フラット UI とクリック可能性の手がかり

| 欄 | 内容 |
| --- | --- |
| 出典 | Moran, K. (2017) Flat UI Elements Attract Less Attention and Cause Uncertainty, NN/g https://www.nngroup.com/articles/flat-ui-less-attention-cause-uncertainty/ 。Loranger, H. (2015) Beyond Blue Links: Making Clickable Elements Recognizable, NN/g https://www.nngroup.com/articles/clickable-elements/ |
| 何を言っているか | 9 サイト 18 ページを強い手がかり版と弱い手がかり版（ゴーストボタン、本文と同じ見た目のリンク）にして 71 人の視線を計った実験。弱い版では課題時間が平均 22%、注視回数が 25% 増え、視線が画面全体に散った。低い情報密度・慣習的な配置・目立つ位置の 3 条件がそろえばフラットでも害は小さい |
| 確からしさ | 1 本の視線計測実験。課題は「どこを押すか見つける」だけで、実作業全体の遅れはこれより小さいと著者自身が注記 |
| 当たるべき場面 | 押せる要素の見た目を決めるとき。装飾を減らすとき。1 画面の要素が多いとき。ux-principles 原則 1 の「主操作」を見た目で区別する根拠 |
| 判断の問い | 押せる要素と押せない要素は、形・色・位置のどれで区別されるか。情報密度が高い画面で手がかりを弱めていないか。主操作はほかと違う形・大きさか |
| 適切・不適切な例 | 適切: ツールバーの操作は枠のあるボタンにそろえ、本文のリンクは色+下線。主操作（再生）だけ形と大きさを変える。不適切: 再生・キー・スタイル・MIDI・保存・候補名が同じ重さで並ぶ |
| middy との接点 | ツールバー、候補カード、タイムラインのコード、度数の地図 |
| 確認日 | 2026-11-09 |

### Fitts の法則とタッチターゲットの大きさ・間隔

| 欄 | 内容 |
| --- | --- |
| 出典 | Fitts, P. M. (1954) The information capacity of the human motor system in controlling the amplitude of movement. *Journal of Experimental Psychology* 47。MacKenzie, I. S. (1992) Fitts' law as a research and design tool in HCI. *Human-Computer Interaction* 7。解説: Budiu, R. (2022) Fitts's Law and Its Applications in UX, NN/g https://www.nngroup.com/articles/fitts-law/ 。Nielsen, J. (2026 閲覧) Fitts's Law: Big and Near Beats Small and Far, UX Tigers https://www.uxtigers.com/post/fitts-law 。WCAG 2.2 SC 2.5.8 Target Size (Minimum) (AA) 24×24 CSS px https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html 、SC 2.5.5 (AAA) 44×44。Apple HIG "Accessibility"（44×44 pt、最小 28×28 pt）。Material 3（48×48 dp、間隔 8 dp。未確認） |
| 何を言っているか | 標的に到達する時間は、距離が長いほど、標的が小さいほど、対数的に増える。小さい標的ほど減速の段階が長く、誤りも増える。画面の縁は行き過ぎがないので「無限の標的」になる（マウスのみ。タッチでは効かず、むしろ縁は遅い）。指は接触面が広く遮蔽もするので、画面上の px 計算より誤りが増え（FFitts）、44 pt / 48 dp / 1 cm といった下限が別に置かれる。隣接する標的との間隔は大きさと同じくらい効く。ひとつの画面で大きくするのは主操作だけにし、破壊的な操作は小さく・遠く・隣接させない。WCAG の 24 px は全ポインタ入力の最低線で、Working Group 自身がタッチには大きい値を勧めると注記。ズームしても CSS px は変わらないので「拡大すれば足りる」は通らない |
| 確からしさ | HCI で最も再現性の高いモデルの 1 つ（原論文は未確認）。タッチの下限値は運動制御の研究と実務の折衷。WCAG は規範（AA） |
| 当たるべき場面 | タイムラインのチップの大きさ（スマホで 7.4 px の文字、0.53 倍の当たり判定）、固定の操作、ツールバーのボタン、縮小図のコード、度数の地図の和音 |
| 判断の問い | 再生中に押してほしい標的は、視線を向けずに押せる大きさ・位置か。24×24 CSS px の正方形を内側に描けるか。描けないなら隣との中心距離は 24 px 以上か。主操作（再生、固定）は 44 px に届くか。隣に押し間違えたら何が起きるか |
| 適切・不適切な例 | 適切: 主操作（再生／停止）を大きく、親指の届く位置に固定。再生中に押せる「今の和音を固定する」大きなボタンを置く（2.5.8 の「同等」の例外にも沿う）。不適切: 聴きながら 0.53 倍に縮んだチップを狙って押して固定する |
| middy との接点 | スマホのタイムライン、固定、ツールバー、縮小図、度数の地図 |
| 確認日 | 2026-11-09（NN/g、UX Tigers、WCAG 2.2 Understanding 2.5.8、Apple HIG。Fitts 1954 と MacKenzie 1992 の原論文は未確認・網の方針外。Material 3 は未確認） |

### オブジェクト指向 UI（OOUI / OOUX）

| 欄 | 内容 |
| --- | --- |
| 出典 | ソシオメディア、上野学、藤井幸多 (2020) 『オブジェクト指向UIデザイン』技術評論社。改訂新版 2026-07。書誌: https://gihyo.jp/book/2020/978-4-297-11351-3 、https://www.sociomedia.co.jp/ 。Collins, D. (1995) *Designing Object-Oriented User Interfaces*。IBM *CUA Guide to User Interface Design* (1991)。Prater, S. V. の OOUX（ORCA）https://www.ooux.com/ |
| 何を言っているか | タスク（動詞）ではなくオブジェクト（名詞）を起点に UI を組み立てる。オブジェクトが知覚でき直接働きかけられる、オブジェクトが自分の性質と状態を体現する、「対象を選んでから操作を選ぶ」順序、オブジェクト同士が協調して UI を成す。手順は、名詞の抽出 → メインオブジェクトとそのコレクション／シングルのビュー → レイアウトパターン |
| 確からしさ | 実務の経験則の体系化（コンサルティングの事例が根拠）。対照実験はない。**よくある誤用:** 「オブジェクト = DB のテーブル」。CUA も Collins も、ユーザーの関心対象とプログラムのオブジェクトを分けている |
| 当たるべき場面 | ux-principles 原則 6「画面より先に名詞を決める」を実際に適用するとき（3 ステップが手順になる）。新しい操作を「どの名詞の、コレクションとシングルのどちらのビューに属するか」で位置づけるとき |
| 判断の問い | この操作はどのオブジェクト（例・成り立ち・保存）に属し、一覧（コレクション）と 1 つ（シングル）のどちらの画面にあるべきか。先に対象を選ばせてから操作を出しているか |
| 適切・不適切な例 | 適切: 例を選ぶ（対象）→ 再生・キー・保存（操作）の順で、操作は選んだ例に従属する。不適切: 「書き出す」を先に選ばせて、次にどの例かを聞く。middy の「成り立ち」はシングルビューで、全和音の一覧というコレクションビューを持つかどうかが論点 |
| middy との接点 | experience.md の概念表、候補カード（例のコレクション）、舞台（例のシングル）、保存した例の一覧 |
| 確認日 | 2026-11-09（gihyo.jp の書誌と目次、sociomedia.co.jp。本文、Collins、CUA、ooux.com は未確認） |

### 段階的開示

| 欄 | 内容 |
| --- | --- |
| 出典 | Nielsen, J. (2006) Progressive Disclosure, NN/g https://www.nngroup.com/articles/progressive-disclosure/ 。関連: WCAG 2.2 SC 1.4.13 Content on Hover or Focus |
| 何を言っているか | 最初は重要な少数だけ見せ、残りは求められたら出す。初期画面に出ていること自体が「重要だ」という合図になる。うまくいく条件は、初期と二次の切り分けが正しいこと（頻繁に使うものは初期に）と、二次へ進む入口が見えて中身が予想できること。3 段以上は迷子になりやすい |
| 確からしさ | 実務の経験則（NN/g の多数の評価研究の要約）。「2 段まで」は目安 |
| 当たるべき場面 | ux-principles 原則 2「状態が生まれるまで出さない」の実装。解説の量（見出しだけ／理由まで／リズムまで）をいつ出すか。「もっと見る」を足すとき。熟達の逆転（7 節）と組み合わせて「慣れたら薄くできる」を作るとき |
| 判断の問い | 初期に出すものは、聴く段階で頻繁に要るものだけか。二次の情報（理由の全文、リズムの解説）への入口は見えて、押すと何が出るか予想できるか。段が 3 つ以上になっていないか |
| 適切・不適切な例 | 適切: 再生中は見出しだけを出し、止めると理由が出る。不適切: 初期画面に全解説を同じ重さで並べ、スクロールで探させる。middy: リズムの解説（PC で 644 px）が理由のすぐ下に同じ重さで出ており、二次に回す候補 |
| middy との接点 | 解説（全体・流れ／今のコード／リズム）、これまでのやり取り（畳む）、保存の一覧 |
| 確認日 | 2026-11-09 |

## 5. 選択と既定値

### Hick–Hyman の法則と、その誤用

| 欄 | 内容 |
| --- | --- |
| 出典 | Hick, W. E. (1952) On the rate of gain of information. *Quarterly Journal of Experimental Psychology* 4。Hyman, R. (1953) *Journal of Experimental Psychology* 45。Seow, S. C. (2005) Information-theoretic models of HCI. *Human-Computer Interaction* 20。Cockburn, A., Gutwin, C. & Greenberg, S. (2007) A predictive model of menu performance. *CHI '07*. DOI 10.1145/1240624.1240723 |
| 何を言っているか | 等確率で、刺激と反応の対応を練習済みの選択では、反応時間は選択肢数の対数で増える。これは「知っている選択肢から 1 つを決める」時間で、初めて見るメニューを目で探す時間ではない。Cockburn らは、初心者のメニュー探索は項目数に比例（線形）、熟達者は Hick–Hyman + Fitts で予測できるとした |
| 確からしさ | 法則自体は安定。**よくある誤用:** 「選択肢を減らせば対数的に速くなる」。初心者には線形の探索コストで、見出し・並び順・群化の方が効く。ux-principles 原則 4 の根拠として Hick を挙げるのは、結論（既定値で決める）は良いが根拠としては弱い |
| 当たるべき場面 | 候補カードの数、キーの選択肢（12 調）、スタイルの選択肢の見せ方。「数を減らす」より「探しやすく並べる」で解けないかの点検 |
| 判断の問い | ユーザーはこの選択肢を知っている（対数）か、初めて見て探す（線形）か。後者なら、並び順と群化は探索を助けているか |
| 適切・不適切な例 | 適切: キーの選択肢を五度圏や近親調の順で並べ、今のキーを強調する。不適切: 「Hick の法則によりスタイルは 3 つまで」と数だけで決める |
| middy との接点 | 候補カード、キーの選択、スタイルの切り替え、原則 4 の根拠 |
| 確認日 | 未確認（ACM DL はボット遮断、Taylor & Francis は網の方針外。NN/g に該当する解説記事はなかった。書誌は検索結果の抄録で照合） |

### 選択過多（再現性に注意）と既定値の力

| 欄 | 内容 |
| --- | --- |
| 出典 | Iyengar, S. S. & Lepper, M. R. (2000) When choice is demotivating. *JPSP* 79(6)。Scheibehenne, B., Greifeneder, R. & Todd, P. M. (2010) Can there ever be too many options? *Journal of Consumer Research* 37(3). https://doi.org/10.1086/651235 。Chernev, A., Böckenholt, U. & Goodman, J. (2015) *Journal of Consumer Psychology* 25(2)。Nielsen, J. (2005) The Power of Defaults, NN/g https://www.nngroup.com/articles/the-power-of-defaults/ 。Loranger, H. (2015) Simplicity Wins over Abundance of Choice, NN/g https://www.nngroup.com/articles/simplicity-vs-choice/ |
| 何を言っているか | 「選択肢が多いと選べない・満足が下がる」という 2000 年の有名な結果は、2010 年のメタ分析（50 実験）では平均効果がほぼゼロで研究間のばらつきが大きく、2015 年のメタ分析は、選択肢の複雑さ・課題の難しさ・好みの不確かさ・目的の 4 条件がそろうと出て、そろわなければ消えるとした（いずれも未確認）。一方、既定値の力は観察が安定している: 検索結果の 1 位と 2 位を密かに入れ替えても、1 位のクリック率は 42% → 34% に下がるだけで、ユーザーは上にあるものを選ぶ（Joachims らの研究、NN/g）。既定値は「その場の見本」と「よくある値の省力」の 2 つの役を持つ。NN/g は多機能化が判断の労力を増やすと経験則として言う |
| 確からしさ | **選択過多は再現性に議論がある。** 「選択肢は少ないほど良い」は一般則として成り立たない。成り立つのは、初心者が好みを言語化できず、違いが分かりにくい選択肢を決めなければならないとき。既定値の力は実験と観察で安定 |
| 当たるべき場面 | 候補の数、キー・スタイルの選択肢、ux-principles 原則 4 の根拠欄 |
| 判断の問い | ユーザーはこの選択肢の違いを判断できる状態か。できないなら、数を減らすより、既定値で鳴らして比べられるようにする方が効くか。既定値は最初の価値（最初の音）を速く届けるか |
| 適切・不適切な例 | 適切: 最初の例を選んだ状態で届け、ほかは切り替え先として並べる（今の約束）。不適切: 「選択過多だから選択肢は 3 つまで」と一般則として書く |
| middy との接点 | 候補カード、キーの選択、原則 4 の根拠 |
| 確認日 | 2026-11-09（NN/g の 2 記事。Iyengar & Lepper、Scheibehenne ら、Chernev らの原論文は未確認・網の方針外） |

### 望ましい困難と「学習と成績の乖離」

| 欄 | 内容 |
| --- | --- |
| 出典 | Bjork, E. L. & Bjork, R. A. (2011) Making things hard on yourself, but in a good way. In *Psychology and the Real World*, Worth。Soderstrom, N. C. & Bjork, R. A. (2015) Learning versus performance: An integrative review. *Perspectives on Psychological Science* 10(2). https://doi.org/10.1177/1745691615569000 。Roediger, H. L. & Karpicke, J. D. (2006) Test-enhanced learning. *Psychological Science* 17(3) |
| 何を言っているか | 練習中の成績を上げる条件（まとめて練習、同じ順、答えをすぐ見せる）は長期の保持や転移を下げ、成績を下げる条件（間隔を空ける、混ぜる、思い出させる、自分で生成させる）が学習を上げることが多い。ただし困難は学習者が乗り越えられるものに限る。テストで思い出す方が再読より 1 週間後の保持が良い |
| 確からしさ | 間隔効果・テスト効果は教育心理で最も再現されている部類。「生成」「混ぜる」は条件依存。UI への適用は慎重に: 困難は内容（和声を自分で予想する）に置き、操作に置かない |
| 当たるべき場面 | 原則 4（既定値）と学習の衝突（ux-principles 4 章）。「次の和音を当てる」「理由を読む前に予想する」のような学習の仕掛けを足すとき |
| 判断の問い | この困難は、和声の理解に向いているか、画面の操作に向いているか。初めての人が乗り越えられるか |
| 適切・不適切な例 | 適切: 理由の文を最初は畳んでおき、「なぜ V7 か」を考えてから開ける。不適切: 固定の操作を難しくして「よく聴かせる」 |
| middy との接点 | 解説の見せ方（見出しだけ → 理由）、将来のバリエーション |
| 確認日 | 未確認（UCLA の著者ラボ、SAGE が網の方針外。書誌は検索結果で照合） |

## 6. 待ちと応答

### 応答時間のしきい値（0.1 / 1 / 10 秒）と進捗表示の形

| 欄 | 内容 |
| --- | --- |
| 出典 | Miller, R. B. (1968) Response time in man-computer conversational transactions. *AFIPS FJCC* 33。Card, Robertson & Mackinlay (1991) The information visualizer. *CHI '91*。Nielsen, J. (1993) *Usability Engineering* 5 章、同 Response Times: The 3 Important Limits https://www.nngroup.com/articles/response-times-3-important-limits/ 、Powers of 10: Time Scales in UX https://www.nngroup.com/articles/powers-of-10-time-scales-in-ux/ 。Sherwin, K. (2014) Progress Indicators Make a Slow System Less Insufferable, NN/g https://www.nngroup.com/articles/progress-indicators/ 。Harrison, C. et al. (2007) Rethinking the progress bar. *UIST '07*。Doherty & Thadhani (1982) IBM 技術報告 |
| 何を言っているか | 0.1 秒以内なら自分が直接操作している感覚が保たれ、1 秒以内なら思考の流れが切れず、10 秒を超えると注意がほかへ移る。1 秒を超える待ちには働いていることの表示、10 秒を超えるなら完了の見込み（割合）と中断の手段を出し、戻ってきたユーザーが再定位する前提で設計する。2〜10 秒はループするアニメーションで足り、1 秒未満にそれを出すと却って気になる。操作の直後に受け付けたことを示す（押された状態）。進捗表示は体感を変え、ある研究では表示ありの方が 3 倍長く待てた。速すぎて追えない変化（表示の慣性）も避ける。Harrison らは、同じ時間でも後半で加速するバーが速く感じられると報告（未確認） |
| 確からしさ | 0.1 / 1 / 10 は知覚の実験と実務の経験則を合わせた目安で、精密な定数ではない。**Doherty しきい値（400 ms）は 1980 年代の端末でのプログラマの生産性の観察で、人間の知覚限界ではない。** 「Laws of UX」などで法則のように引かれるが、独立した再現は見つけていない。使うなら「速いほど操作が増える」という向きだけ |
| 当たるべき場面 | 再生ボタンから音が出るまで、コードを押して固定するまで、キーを変えて再描画するまでの遅延。LLM の応答待ち（数秒〜十数秒、長さが予測できない）。ux-principles 原則 8 |
| 判断の問い | 操作から反応まで 0.1 秒を超える操作はどれか（音の開始は AudioContext の起動で遅れがち）。1 秒を超える待ちに、進んでいることの表示はあるか。10 秒を超えうる LLM 待ちで、注意が離れても戻れる手がかりと中断の手段はあるか。待ちの長さは予測できるか。できないなら、偽の進捗より「状態の変化」（送った言葉が残っている、前の例は聴ける）を見せているか |
| 適切・不適切な例 | 適切: 送信直後に「考え中」を出し、届いたら視線が行くべき場所に結果が現れる。待つあいだも前の例を聴ける。不適切: 再生を押してから音が出るまで無音で、押せたか分からない。止まったり戻ったりする進捗バー |
| middy との接点 | 再生、固定、キー変更、送って待つ段階（experience.md）、原則 8 |
| 確認日 | 2026-11-09（NN/g の 3 記事。Miller 1968、Card ら 1991、Harrison ら 2007、Doherty & Thadhani は未確認・ACM のボット遮断と著者ページの網の方針外） |

## 7. 学習の負荷

### 認知負荷理論（内在・外在）と一貫性（関係のない情報は削る）

| 欄 | 内容 |
| --- | --- |
| 出典 | Sweller, J. (1988) Cognitive load during problem solving. *Cognitive Science* 12。Sweller, van Merriënboer & Paas (2019) Cognitive architecture and instructional design: 20 years later. *Educational Psychology Review* 31. https://doi.org/10.1007/s10648-019-09465-5 。Whitenton, K. (2013) Minimize Cognitive Load to Maximize Usability, NN/g https://www.nngroup.com/articles/minimize-cognitive-load/ 。Sundararajan, N. & Adesope, O. O. (2020) Keep it coherent: A meta-analysis of the seductive details effect. *Educational Psychology Review* 32(3)。Mayer の一貫性原理（*Multimedia Learning* 3rd ed.） |
| 何を言っているか | 学習時の作業記憶の負荷を、内容そのものの複雑さ（内在）と、提示の仕方で生じる無駄（外在）に分け、外在を削って作業記憶を内容に回す。NN/g の言い換え: 内在負荷は消せないし消すべきでもない。外在負荷（意味のない装飾、記憶や判断を強いる手順）を削り、既存のメンタルモデルに乗せ、課題を肩代わりする（見せる、再表示する、既定値）。学習目標に関係ない挿話・装飾を足すと、面白くはなるが学習は下がる（誘惑的詳細。メタ分析で小〜中の負の効果） |
| 確からしさ | 多数の効果が実験で反復。負荷の測定法と「学習関連負荷」の定義には批判がある。誘惑的詳細の効果は小さく、動機を保つ効果との兼ね合いは未解決。ux-principles 3 章の冒頭は理論の要点を正しく踏まえている |
| 当たるべき場面 | 学習画面で何を削り、何を残すかの判断すべて。「内容の難しさは削らず、画面の理解に使う負荷を削る」の根拠。解説に雰囲気の形容や雑学を足したくなったとき。LLM の解釈文とエンジンの分析文の並べ方 |
| 判断の問い | この要素の理解に使う負荷は、内容（和声）の理解に向かうか、画面の操作・配置の解読に消えるか。この文・絵は和声の理解に寄与するか、雰囲気づくりか。雰囲気づくりなら理解の文と同じ場所・同じ重さで出す必要があるか |
| 適切・不適切な例 | 適切: 理由の文を短くするのではなく、文と対象の視線の往復（外在）を減らす。全体・流れの文はモードの性格と進行の区切りに絞る。不適切: 理論の説明を簡単にして内在負荷を削る（ADR-0003 と intent の優先 1 に反する）。理由の隣に曲の情景描写を並べる |
| middy との接点 | 解説の全部、ツールバー、スマホの縮小タイムライン（7 px の文字を解読する負荷は外在）、会話の LLM 応答と結果の解説の分離 |
| 確認日 | 2026-11-09（NN/g の記事。Sweller 1988、Sweller ら 2019、Sundararajan & Adesope 2020 は未確認・Springer が網の方針外） |

### 熟達の逆転効果（初心者に効く支援が、経験者には邪魔になる）

| 欄 | 内容 |
| --- | --- |
| 出典 | Kalyuga, S., Ayres, P., Chandler, P. & Sweller, J. (2003) The expertise reversal effect. *Educational Psychologist* 38(1). https://doi.org/10.1207/S15326985EP3801_4 。Kalyuga, S. (2007) *Educational Psychology Review* 19 |
| 何を言っているか | 統合された説明、作業例、シグナリングなどの支援は、知識の少ない学習者には効くが、知識が増えると効果が消え、逆に成績を下げることがある。すでに持っているスキーマと提示された説明が重複し、処理の無駄になるため |
| 確からしさ | 複数領域で反復され、認知負荷理論の中でも安定した効果。「いつ逆転するか」の境界は課題依存 |
| 当たるべき場面 | 解説の量や常時表示を決めるとき。同じユーザーが使い込むにつれて説明が煩わしくなる問題。段階的開示（4 節）と組み合わせる |
| 判断の問い | この説明は、初めての人のためか、慣れた人も毎回読まされるか。慣れたら薄くできる（畳む・見出しだけにする）か |
| 適切・不適切な例 | 適切: 再生中は見出しだけ、読みたいときに止めて理由を読む。不適切: 全和音の理由を常時全文で並べ、経験者にも初心者にも同じ量を読ませる |
| middy との接点 | 解説の密度、全体・流れの文、リズムの解説の常時表示 |
| 確認日 | 未確認（Taylor & Francis、Springer が網の方針外。書誌は検索結果の抄録で照合） |

## 8. 色とコントラスト

### 色だけに頼らない（WCAG 1.4.1）

| 欄 | 内容 |
| --- | --- |
| 出典 | WCAG 2.2 SC 1.4.1 Use of Color (A) https://www.w3.org/TR/WCAG22/#use-of-color 。Understanding: https://www.w3.org/WAI/WCAG22/Understanding/use-of-color.html |
| 何を言っているか | 情報の伝達、操作の指示、応答の促し、要素の区別を色だけで行わない。色相だけでなく明度も違い、相対輝度のコントラスト比が 3:1 以上なら「別の視覚的な区別」と数えてよい。ただし「緑なら有効」のように特定の色を言い当てる必要がある場合は、比に関係なく別の手がかりが要る。色分け自体を禁じるものではない |
| 確からしさ | 規範（A） |
| 当たるべき場面 | 和声機能 T/S/D を色で分けるとき。今鳴っているコードの強調。固定中の表示。エラーの表示 |
| 判断の問い | 色を白黒にしても、同じ情報が読めるか。T/S/D は文字・形・位置でも区別されているか |
| 適切・不適切な例 | 適切: 機能の色の帯に T/S/D の文字を重ねる。不適切: 度数の地図で「例で使う和音」を色の濃さだけで示す。middy: `ExplanationPanel.css` と `DegreeMap.css` は塗り分けつつ文字も出している。`aria-current` の強調が色だけになっていないかは点検対象 |
| middy との接点 | 度数の地図、解説の機能の表示、タイムラインの強調、固定の表示、エラー |
| 確認日 | 2026-11-09 |

### コントラスト: 文字（1.4.3）と文字以外（1.4.11）、WCAG 3 / APCA の位置づけ

| 欄 | 内容 |
| --- | --- |
| 出典 | WCAG 2.2 SC 1.4.3 Contrast (Minimum) (AA): 文字 4.5:1、大きな文字 3:1。SC 1.4.11 Non-text Contrast (AA): 3:1。Understanding Non-text Contrast: https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html 。W3C *WCAG 3.0* Working Draft https://www.w3.org/TR/wcag-3.0/ 。Somers, A. *APCA* https://github.com/Myndex/SAPC-APCA |
| 何を言っているか | 部品は「存在と操作法がわかるのに必要な部分」が 3:1 あればよく、フォーカス表示は常に必要、非活性の部品は対象外。細い線はアンチエイリアスで見かけの濃さが落ちるので太くするか比を余らせる。状態の印は部品内部の色に対して測る。WCAG 3 は作業草案で、文字コントラストのアルゴリズムは未決定。APCA は候補の 1 つで、明暗の極性とフォントサイズ・太さを考慮する知覚ベースの指標 |
| 確からしさ | 1.4.3 / 1.4.11 は規範（AA）。3:1 の閾値は丸めない。**WCAG 3 は規範ではなく、適合の根拠は今も WCAG 2.2。** **よくある誤用:** 「APCA を使えば WCAG 2 の比は無視してよい」 |
| 当たるべき場面 | ダーク／ライトの両テーマでトークンを決めるとき。タイムラインの音の矩形、再生位置の線、度数の地図の塗り、縮小図の小さな要素。「WCAG 3 対応」をうたいたくなったとき |
| 判断の問い | 文字は 4.5:1、部品の輪郭・状態・図形の必要な部分は 3:1 を、両テーマで満たすか。細い線は太くできるか。APCA を使うなら「補助」と明記しているか |
| 適切・不適切な例 | 適切: 再生位置の線を 2 px 以上にし背景と 3:1 以上。2.2 の比を満たしたうえでダークの小さな文字を APCA でも確かめる。不適切: 固定中のコードの枠を 1 px の薄い色で示す。APCA の Lc 値だけを根拠に 2.2 の比を下回る色を採る |
| middy との接点 | タイムライン（音、コード名、再生位置）、度数の地図、ツールバー、配色トークン |
| 確認日 | 2026-11-09 |

### 色覚多様性とカテゴリ色: Okabe–Ito の Color Universal Design

| 欄 | 内容 |
| --- | --- |
| 出典 | 岡部正隆・伊藤啓 "Color Universal Design (CUD)" https://jfly.uni-koeln.de/color/ （引用時は両名を記すようにと明記）。Ichihara, Okabe, Iga, Tanaka, Musha, Ito (2008) *Proc. SPIE* 6807. DOI 10.1117/12.765420 |
| 何を言っているか | 色覚の型が違っても、印刷しても見分けられる 8 色の組（黒、橙、空色、青緑、黄、青、朱、赤紫）。赤は朱にして 1 型色覚でも黒と区別、黄と緑の間を避け、緑は青みに、紫は赤紫に。細い線と小さな文字には空色・黄より濃い青・橙。暖色と寒色を交互に、同系なら明度か彩度に差を。「色の名前で指さない」、色以外の手がかりを重ねる |
| 確からしさ | 色覚のシミュレーションと当事者の確認に基づく実務の提案。広く使われ（R・matplotlib の既定パレット）実務上の標準に近い。「8 色まで」は見分けの限界で、カテゴリは少ないほどよい |
| 当たるべき場面 | T/S/D など少数のカテゴリに色を割り当てるとき。ダーク／ライトで同じカテゴリ色を保つとき。グラフや地図を足すとき |
| 判断の問い | カテゴリは 3〜5 以内か。色覚シミュレーションで見分けがつくか。小さな要素に黄や空色を使っていないか。色の名前で説明していないか（「赤いコード」ではなく「右上の V」） |
| 適切・不適切な例 | 適切: T/S/D に緑・青・琥珀のように明度も違う色を当て、文字を重ねる。不適切: 赤と緑で T と D を分ける。middy の今の T/S/D は Okabe–Ito の青緑・青・橙に近く、方向は合う |
| middy との接点 | 度数の地図、解説の機能の表示、今後のリズムの可視化 |
| 確認日 | 2026-11-09（jfly のページ。SPIE 論文は未確認） |

### 知覚的に均一な色空間でトークンを作る（CIELAB / Oklab / OKLCH）

| 欄 | 内容 |
| --- | --- |
| 出典 | Ottosson, B. (2020) A perceptual color space for image processing https://bottosson.github.io/posts/oklab/ 。W3C *CSS Color Module Level 4* §9 https://www.w3.org/TR/css-color-4/ 。*CSS Color Module Level 5* `color-mix()` https://www.w3.org/TR/css-color-5/ 、MDN https://developer.mozilla.org/en-US/docs/Web/CSS/color_value/color-mix |
| 何を言っているか | HSL/HSV は同じ「明度」でも黄と青で見た目の明るさが大きく違う。Oklab は明度 L・色度が互いに独立に近く、補間が滑らかになるよう設計された。CSS では `oklch()` / `oklab()` / `lab()` / `lch()` が標準化され、`color-mix()` で混ぜる色空間を選べる |
| 確からしさ | 色彩科学の応用。「知覚的に均一」は近似。トークンの決め方の道具であって、WCAG のコントラスト比の代替ではない |
| 当たるべき場面 | 両テーマで T/S/D などの色を「同じ明るさ・同じ彩度」にそろえるとき。段階色（強調の濃淡）を作るとき。`color-mix` の色空間を選ぶとき |
| 判断の問い | トークンは OKLCH の L を揃えて作っているか。混色は `in oklch` か `in srgb` か。できた色は WCAG 2.2 の比を満たすか（均一性は比を保証しない） |
| 適切・不適切な例 | 適切: T/S/D を OKLCH で同じ L・C にそろえ、色相だけ変える。不適切: sRGB の 16 進を目で合わせ、黄だけ浮く。middy: `DegreeMap.css` は `color-mix(in srgb, ...)` を使っている |
| middy との接点 | 配色トークン、度数の地図、タイムラインの強調の濃淡、ダークテーマ |
| 確認日 | 2026-11-09 |

### ダークモードとコントラスト極性

| 欄 | 内容 |
| --- | --- |
| 出典 | Budiu, R. (2020) Dark Mode vs. Light Mode: Which Is Better?, NN/g（文献レビュー）https://www.nngroup.com/articles/dark-mode/ 。一次研究（NN/g の記述による）: Piepenbrock ほか (2013) *Ergonomics*、同 (2013) *Human Factors*、Piepenbrock, Mayr & Buchner (2014) *Ergonomics*、Dobres ほか (2017) *Applied Ergonomics*。Apple HIG "Dark Mode" |
| 何を言っているか | 通常視力の人では、明るい背景に暗い文字（正極性）の方が読みの成績がよく、文字が小さいほど差が開く。機序は瞳孔が縮んで被写界深度が増すこと。暗い環境光では差がさらに出る。本人の主観は差を感じない。白内障などではダークが有利な場合がある。NN/g の結論: 既定はライト、切り替えは必ず提供し、OS の設定に従う。Apple は両方の外観で試し、小さな文字は 7:1 を目標にせよと言う |
| 確からしさ | 実験室研究が複数あり方向は一致。効果は環境光・文字サイズ・年齢で変わる。NN/g 自身は独自調査をしていないと明記 |
| 当たるべき場面 | ダークテーマの文字サイズとコントラストを決めるとき。スマホで小さい文字をダークで出すとき |
| 判断の問い | ダークテーマで、小さな文字は大きくするか比を上げているか。OS の設定に従い、両方で試しているか |
| 適切・不適切な例 | 適切: ダークでは小さな文字の比を 7:1 近くまで上げる。不適切: ライトで決めた薄い灰色の文字をそのまま反転する。middy: スマホの 4.8 px の音名はダークでいっそう読めない |
| middy との接点 | ダークテーマ、タイムラインの小さな文字、解説の本文 |
| 確認日 | 2026-11-09（NN/g と Apple HIG。Piepenbrock ら・Dobres らの原著は未確認） |

## 9. 支援技術

### WAI-ARIA Authoring Practices: ツールバー、リストボックス／グリッド、スライダー

| 欄 | 内容 |
| --- | --- |
| 出典 | W3C WAI *ARIA Authoring Practices Guide*: Toolbar https://www.w3.org/WAI/ARIA/apg/patterns/toolbar/ 、Listbox https://www.w3.org/WAI/ARIA/apg/patterns/listbox/ 、Grid https://www.w3.org/WAI/ARIA/apg/patterns/grid/ 、Slider https://www.w3.org/WAI/ARIA/apg/patterns/slider/ 。*WAI-ARIA 1.2* https://www.w3.org/TR/wai-aria-1.2/ |
| 何を言っているか | ツールバー: 3 つ以上の操作を 1 つの Tab 停止にまとめ、矢印キーで中を移動。リストボックス: 選択肢の名前は短く、先頭語の繰り返しを避ける。選択肢の中に見出しやボタンなどの意味を持つ要素は置けない。中に操作できる要素を並べたいならグリッド。スライダー: タッチの支援技術はジェスチャをキー操作に変換するので実機で試す |
| 確からしさ | W3C の実装指針（規範ではない）。支援技術との相性は組み合わせで変わる |
| 当たるべき場面 | ツールバーの Tab 回数を減らすとき。タイムラインのコード（`role="listbox"` / `option`）の中にコード名以外の情報や操作を足したくなったとき。再生位置のスライダーを足すとき |
| 判断の問い | 主操作まで Tab 何回か（ux-principles 6 章の指標）。option の名前は短いか。option の中にボタンや見出しを入れようとしていないか（→ グリッドか、外に出す） |
| 適切・不適切な例 | 適切: 和音の一覧を listbox にするなら各行の名前は「V7 B7 D」程度にし、理由の文は外に出す。不適切: option の中に解説文と解除ボタンを入れる。middy: `listbox` と `aria-current` は E2E が頼っているので、構造を変えるなら E2E も同じ PR で |
| middy との接点 | ツールバー、タイムラインのコード、縮小図、再生位置 |
| 確認日 | 2026-11-09 |

### ライブリージョンと状態メッセージ（うるさい読み上げの危険）

| 欄 | 内容 |
| --- | --- |
| 出典 | *WAI-ARIA 1.2* Live Regions（aria-live, aria-atomic, aria-relevant, aria-busy）https://www.w3.org/TR/wai-aria-1.2/ 。WCAG 2.2 SC 4.1.3 Status Messages (AA) https://www.w3.org/WAI/WCAG22/Understanding/status-messages.html 。MDN "ARIA live regions" https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Guides/Live_regions |
| 何を言っているか | ライブリージョンは、フォーカスが別の場所にあるときに変わる領域を支援技術に知らせる。`polite` は手が空いたときに読む、`assertive` は割り込む（時間に厳しい重要事項だけ）。MDN は「重要だが、うるさいほど速くは変わらない更新」に polite を使えと言う。4.1.3 は、フォーカスを奪わずに結果・待機・進捗・エラーを知らせるメッセージを支援技術が読める形にせよという要件 |
| 確からしさ | 規範（WCAG）と実装仕様（ARIA）。支援技術の実装差が大きい |
| 当たるべき場面 | 「今鳴っているコード」を支援技術に伝えるとき。解説が再生に追従して変わる領域に `aria-live` を付けるかどうか |
| 判断の問い | この領域は 1.8〜3.3 秒ごとに変わるが、それを全部読み上げさせたいか。読み上げるなら短い状態（コード名・機能）だけにして理由の文は別領域にできないか。役割は status（結果の通知）か、読者が自分で辿る内容か |
| 適切・不適切な例 | 適切: 「送信しました」「接続できません」を `role="status"` で通知。不適切: 解説パネル全体（全体・流れ、度数の地図、見出しと理由、リズム）を 1 つの polite なライブリージョンにし、再生中に和音ごとに全文が読み上げ候補になる。middy: `ExplanationPanel.tsx` の `<section aria-live="polite">` がこれに当たる |
| middy との接点 | 解説パネル、今鳴っているコード、通知（送信中、エラー、オフライン、更新） |
| 確認日 | 2026-11-09 |

---

## 外したもの

検討して索引に入れなかった知見と理由。再び提案されたとき、同じ検討を繰り返さないために残す。

- **自我消耗（Ego depletion）・「決定疲れ」**: Hagger et al. 2016（23 ラボ）と Vohs et al. 2021（36 ラボ）の多施設事前登録研究で効果が出なかった。「決定疲れ」を根拠に選択肢を減らすのは根拠が崩れている。
- **Doherty しきい値（400 ms）を人間の知覚限界として引くこと**: 1982 年の IBM 社内報告で、端末の応答とプログラマの作業量の関係。応答時間の項の注記にとどめた。
- **Miller の 7±2 を項目数の上限として引くこと**: 作業記憶の項に誤用として書いた。NN/g 自身が「メニューは 7 個まで」を誤解だと明記している。
- **Hick の法則で「選択肢を減らせ」**: 初心者の視覚探索には当たらない。Hick–Hyman の項に誤用として書いた。
- **学習スタイル（視覚型・聴覚型に合わせる）**: Pashler, McDaniel, Rohrer & Bjork (2008) *PSPI* 9(3) の総説で、適合させると学習が上がる証拠は見つからなかった。「聴く人」「読む人」に出し分ける設計の根拠にはならない。
- **言葉の冗長性原理と様相原理（図の説明は音声で、音声に同じ文字を重ねない）**: 「音声」が言葉である場合の研究で、音楽を聴きながら文字を読む middy の状況には直接当たらない。学習者ペースで消え、長いと逆転することも含め、分節化と一過性情報の項に注記で残した。将来、音声ナレーションを足す案が出たら再検討する。
- **ミラーニューロン、Dunbar 数、ドーパミンと変動報酬**（Weinschenk の本が扱う主張）: 人間のミラーニューロンの役割は論争、Dunbar 数は 2021 年の再分析で信頼区間が極端に広く、変動報酬で「通知で釣る」設計は middy の意思に反する。関係が薄く根拠も弱い。
- **目標勾配、満足化、ピーク・エンドの法則**（同上）: 消費者行動や行動経済学の知見で、中程度の反復はあるが、middy の判断に直接効く場面がない。満足化は「既定値で決める」に集約される。
- **「注意持続時間は 8 秒（金魚以下）」**: 一次情報が存在しない俗説。
- **カラオケ式の語単位ハイライトで読解が上がる、という期待**: Gerbier et al. 2018 で学習は増えなかった。強調と手がかりの項に書いた。
- **ステアリングの法則の詳細**: middy にドラッグでなぞる操作がないので、Fitts の項の参照にとどめた。
- **ゲシュタルトの「プレグナンツ」「図と地」「良い連続」**: 群化の項で足りる。
- **Shneiderman の視覚情報探索マントラと Furnas の魚眼を単独項目にすること**: 概観+詳細の項に含めた。歪みを使う焦点+文脈は、拍の位置が情報であるタイムラインには向かない。
- **色の意味の文化差（赤 = 危険、など）**: 結果が割れ、設計判断に落ちる形になっていない。middy の色は和声機能のカテゴリで、文化的な意味は負わせない。1.4.1 と「色は文字で補う」で足りる。
- **IBM CUA、Collins (1995)、Raskin『ヒューメイン・インタフェース』全体を独立項目にすること**: 絶版・一次情報が開けず、OOUI とモードの項に歴史として含めた。ZUI や LEAP は middy と関係が薄い。
- **会話 UI の「チャット + 右側の成果物パネル」レイアウトの指針**: その配置そのものの評価研究は見つからなかった。「狭いスコープ」「ハイブリッド UI」「出力が予測できるなら構造化する」の方向だけを直接操作と委任の項に書いた。配置の是非は概観+詳細（分割注意のコスト）と原則 3 で判断する。
- **Shneiderman の 8 つの黄金律のうち Nielsen と重なる項目の個別解説**: 重複。独自の 2 点だけ残した。
- **WCAG 2.2 SC 1.4.13（ホバーで出す内容）を単独項目にすること**: middy はホバーに頼らない方針（タッチが前提）。段階的開示の項に参照だけ残した。
- **Material 3 の数値の断定**: 本文が取得できなかったので、Fitts と画面幅の項に「未確認」の注付きで触れるにとどめた。
- **日本語の黙読速度と字幕の数値を根拠に使うこと（この session では）**: 読みの速さの項に残したが、原資料は 1 つも開けていない。数値を根拠に使う前に開くこと。
- **Weinschenk『インタフェースデザインの心理学』そのものを出典にすること**: 本文を開けておらず、本は二次資料。本が引く研究のうち middy に関係するもの（短期記憶の訂正、選択的注意と変化の見落とし、行長、見出しと先行オーガナイザ、段階的開示、進捗の見せ方）は、それぞれの一次情報の項に振り分けた。
