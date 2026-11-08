---
name: propose
description: 観察や相談を検討して「企画」Issue にまとめる。問題の理解、事実と推測、選択肢、推奨、決めてほしいこと、スライス案。実装はしない。
---

# propose

観察（Issue、所有者の報告、評価の結果、自分で見つけた問題）を受けて、検討の結果を `.github/ISSUE_TEMPLATE/proposal.md` の形の Issue にする。使い方と判断の基準は `docs/dev/workflow.md` の「Issue の使い方」にだけ書いてある。

手順:

1. **元を集める。** 関係する `observation` の Issue と、会話での報告を挙げる。`gh api -X GET "repos/<owner>/<repo>/issues?labels=observation&state=open"`。
2. **事実を確かめる。** 推測で書かない。コードを読み、必要なら手元で再現する（fake の LLM、`docker compose config`、スクリーンショット）。事実と推測を分け、確かめ方を書く。
3. **問題を枠にはめる。** 画面に関わるなら `docs/design/ux-principles.md` の手順（目的 → 概念 → 流れ → 画面の約束 → 原則）。設計に関わるなら ADR のテンプレート（文脈 → 決定 → 選択肢 → 結果）。音楽理論なら `docs/architecture/domain-model.md` と ADR-0012 の付録を根拠にする。
4. **選択肢を 2 つ以上書く。** 採らない案にも、採らない理由を書く。
5. **推奨と、決めてほしいことを分ける。** 決めてほしいことは、所有者にしか決められないこと（優先、範囲、流儀の選択）だけ。自分で決められることは推奨に入れる。
6. **スライス案を書く。** 1 PR = 1 スライス。受け入れ条件は検査できる形で、サイズは S / M / L。400 行を超えそうなら分ける。
7. **Issue を作る。** ラベルは `proposal` と `needs-decision`（決めてほしいことがないなら `needs-decision` は付けない）。ラベルがなければ作る: `gh label create proposal --color 1D76DB || true`、`gh label create needs-decision --color D93F0B || true`、`gh label create observation --color FBCA04 || true`。元になった観察 Issue は本文で `refs #N`。
8. **会話では 5 行以内で要約し、Issue へのリンクを示す。** 検討の本文を会話に繰り返さない。
9. **実装はしない。** 所有者が決めてから、スライスを `docs/product/backlog.md` に写し（行に `(#N)` を添える）、通常の開発ループで進める。所有者がすでに決めている事項だけなら、この手順を省いて backlog から着手してよい。

サブエージェントに任せる場合は、手順 2〜6 を渡し、「事実（出典つき）・選択肢・推奨・決めてほしいこと・スライス案」の下書きだけを返させる。Issue の作成と会話での要約は自分で行う。
