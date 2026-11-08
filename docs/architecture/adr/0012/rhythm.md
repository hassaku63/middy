# Rhythm and groove: research for middy's rhythm model

Scope: what the literature says, and what follows for a rhythm description an LLM writes and a deterministic engine (4/4, 480 PPQ) checks, renders and explains. Markers: **[settled]** means textbooks and research agree; **[varies]** means practice or findings differ; **[unverified]** means I could not confirm the claim against a primary source in this session.

---

## 1. Meter, subdivision and grids

### (a) What the sources say

- **[settled]** Meter is a hierarchy of pulses (bar > beat > subdivision); each position has a metrical weight, highest on the downbeat (Lerdahl & Jackendoff's metrical grid, 1983 **[unverified: not re-read]**). Every syncopation measure in §3 starts from it.
- **[settled]** Simple meters split the beat into two (4/4: 8ths, then 16ths). Compound meters split it into three (6/8 has two dotted-quarter beats; 12/8 has four). A 4/4 shuffle or swing written with triplets is, in tick terms, the same grid as 12/8. The difference is notation and teaching convention.
- **[settled]** Perception sets lower limits. London's *Hearing in Time* (2004) puts the shortest interval that still works as a metrical element at about 100 ms. Beats are heard roughly between 200 ms and 1.8 s, and listeners prefer about 600 ms (≈100 BPM) ([summary](https://teach.alimomeni.net/2013fall2/?cat=51); [London et al. 2019](https://jyx.jyu.fi/bitstream/123456789/66341/1/london2019tappingdoesnthelpsynchronizeds.pdf)). These are second-hand figures; page numbers in the book are **[unverified]**.
- Grid resolution needed per style **[varies, practice-based]**:
  - rock/pop 8-beat: 8ths (8 steps per bar).
  - 16-beat pop, funk, disco, hip-hop, lo-fi: 16ths (16 steps per bar).
  - bossa nova and samba: 16ths in 4/4 notation (they are often written in 2/4).
  - swing, shuffle, blues, 6/8 and 12/8 ballads: 8th-note triplets (12 steps per bar in 4/4).
  - jazz fills, gospel/neo-soul: 16th triplets (24 steps), rare for beginners.
- No single 16-step grid covers both duple and triple subdivision. The smallest grid that holds both 16ths and 8th triplets has 48 steps per bar (12 per beat, 40 ticks at 480 PPQ).

### (b) Implications

- A rhythm description should state its **subdivision per beat** explicitly (2, 3, 4 or 6) and not assume 16 steps. At 480 PPQ every one of these divides the beat evenly: 240, 160, 120 and 80 ticks.
- The validator can do a deterministic, explainable **tempo-aware check**: a step shorter than about 100 ms is suspect ("too fast to hear as separate notes"), e.g. 16th triplets at 180 BPM are 83 ms.
- Keep a `meter` field even while it can only be 4/4, so "12/8 feel" is explained as "4/4 with three pulses per beat" and true 6/8 can come later via the backlog.
- UI: show a step grid with beat lines and labels such as *1 e & a* or *1 trip let*. Let users switch the subdivision of a lane.

## 2. Swing and shuffle

### (a) What the sources say

- **[settled]** Swing means unequal eighth-note pairs, long then short. It is measured as the swing ratio or **beat-upbeat ratio (BUR)**: the duration of the on-beat eighth divided by the off-beat eighth ([Benadon 2006, *Ethnomusicology* 50(1)](https://olympias2.lib.uoi.gr/items/eaff0c54-8887-4221-b172-093e060aeeb2/full)). 1:1 is straight, 2:1 is the triplet, 3:1 is dotted-eighth plus sixteenth.
- **[settled] The ratio depends on tempo, and 2:1 is only one point on the range.** Friberg and Sundström measured ride-cymbal swing ratios on jazz recordings. They found ratios up to about 3.5:1 at slow tempi, falling to about 1:1 at fast tempi. The "triplet feel" appeared only around medium tempi. Soloists swung less than the cymbal and played their downbeats late while meeting the cymbal on off-beats ([Friberg & Sundström 2002, *Music Perception* 19(3)](https://cir.nii.ac.jp/crid/1362262943521999744); [ASA summary](https://acoustics.org/pressroom/httpdocs/137th/friberg.html)). Honing and de Haas confirmed with professional drummers that swing ratio varies systematically with tempo, and that expressive timing does not simply scale with tempo ([Honing & de Haas 2008, *Music Perception* 25(5)](https://mcg.uva.nl/mcg-2023/papers/honing-haas-2008.pdf)). Exact tempo breakpoints are **[unverified]**.
- **[varies]** Instruments swing differently: drummers' triplet-like eighths mark the beat, soloists' straighter eighths keep momentum ([Butterfield 2011, *Music Theory Spectrum* 33(1)](https://docdrop.org/download_annotation_doc/Butterfield---2011---Why-do-jazz-musicians-swing-their-eight-tbhj3.pdf)). Benadon reports soloist BURs well below 2:1 (secondary report: up to about 1.7).
- **Swing vs shuffle [varies, pedagogy]**: a shuffle (blues, rock, R&B) uses the same long-short division, but the whole band states it on every beat and it stays near 2:1; jazz swing is looser and tempo-dependent **[practice; no primary source checked]**.
- **Sixteenth-note swing** underlies hip-hop, R&B and lo-fi feel, expressed as a drum-machine "swing %" (first note's share of the pair: 50% straight, ~66% triplet), dating from Roger Linn's LM-1 and MPC ([Attack Magazine](https://www.attackmagazine.com/technique/passing-notes/daw-drum-machine-swing/)). "Sweet spot" values such as 54–62% for 16ths are **[unverified]** (Linn 2013 interview, [listing](https://rogerlinndesign.com/about/about-press), not fetched). Conversion: percentage = r / (1 + r), so 2:1 is 66.7% and 3:1 is 75%.
- **Microtiming (push / lay-back) [varies, contested].** Keil's "participatory discrepancies" theory holds that small asynchronies between players create groove **[unverified: Keil 1987 not retrieved]**. Danielsen's group studies laid-back and early placement in groove music ([*Musical Rhythm in the Age of Digital Reproduction*, 2010](https://www.routledge.com/Musical-Rhythm-in-the-Age-of-Digital-Reproduction/Danielsen/p/book/9781138246843)). Controlled experiments do not support the theory as a general rule:
  - Perfectly quantized rock patterns were rated grooviest. Displacements of 15–25 ms lowered ratings, and early displacements lowered them more ([Frühauf, Kopiez & Platz 2013](https://www.frontiersin.org/journals/psychology/articles/10.3389/fpsyg.2015.01232/full), via Senn et al.).
  - Systematic microtiming in jazz, funk and samba patterns lowered groove ratings, except for the simple short-long jazz shuffle ([Davies et al. 2013](https://repositorio.inesctec.pt/handle/123456789/3901?mode=full)).
  - Quantized and original expert microtiming were rated equally ([Senn et al. 2016](https://pmc.ncbi.nlm.nih.gov/articles/PMC5050221)).
  - "Dilla time" (conflicting placements between layers; Charnas, *Dilla Time*, 2022; [Linn's remarks](https://brettworks.com/2013/07/23/roger-linn-on-drum-machine-groove-and-j-dillas-off-beat-sound/)) is real practice but **[unverified]** as measured data.

### (b) Implications

- Model **feel** as parameters, not as an enum:
  - `swing.ratio`: a continuous value, clamped to about 1.0–3.0.
  - `swing.level`: which subdivision is swung, `8th` or `16th`.
  - `offsetMs` per voice (optional): small and bounded (for example ±30 ms).
- Keep the current `straight` and `swing` values as named presets (1:1 and 2:1).
- The engine can map tempo to a *default* ratio (wider when slow, straighter when fast) and explain it as a tendency, not a rule.
- Do not add random humanisation: the evidence does not show it helps groove, and it breaks determinism. If ever added: seeded, off by default, explained.
- UI:
  - a swing control showing the ratio three ways: as "long-short" bar lengths, as a percentage, and as "≈ triplet".
  - an A/B toggle between straight and swing.
  - a note that soloists usually swing less than the drums.

## 3. Syncopation and groove

### (a) What the sources say

- **[settled] Syncopation can be measured deterministically against the metrical grid.**
  - Longuet-Higgins and Lee (LHL): a syncopation occurs when a note on a weaker position is followed by a rest or tie on a stronger one. Its strength is the difference in metrical weight. Listeners prefer the reading of a rhythm that has the least syncopation ([LHL 1984, *Music Perception* 1(4)](https://www.scilit.net/publications/374f7f188f2252ed833871ed371f7ffc)).
  - With LHL as a graded scale, listeners "re-heard" highly syncopated patterns as less syncopated by shifting the pulse, and remembered simpler ones better ([Fitch & Rosenfeld 2007](https://research-portal.st-andrews.ac.uk/en/publications/perception-and-production-of-syncopated-rhythms/)).
  - Alternative measures include Weighted Note-to-Beat Distance ([Gómez et al. 2005, Bridges](https://archive.bridgesmathart.org/2005/bridges2005-73.html); [reference implementation](https://code.soundsoftware.ac.uk/hg/syncopation-dataset/annotate/76ce27beba95/Syncopation%20models/WNBD.py)) and Toussaint's off-beatness.
- **[settled]** Groove is the pleasurable urge to move; high-groove music increases spontaneous movement and tapping accuracy ([Janata, Tomic & Haberman 2012, *JEP: General* 141(1)](https://escholarship.org/content/qt06k2j4h8/qt06k2j4h8.pdf?t=plo4qb)). Pressing grounds groove in predictable pulse plus patterned deviation from it ([Pressing 2002](https://jazztheory.ssac.carleton.ca/view.php?id=449)).
- **[varies] The "inverted U".** In these studies, medium syncopation produced more wanting-to-move and pleasure than low or high syncopation:
  - funk drum breaks, using a polyphonic LHL-style index ([Witek et al. 2014, PLOS ONE](https://doaj.org/article/743d9013fbde436999b78a244c718d37));
  - synthesized melodies ([Sioros et al. 2014](https://pmc.ncbi.nlm.nih.gov/articles/PMC4165312));
  - rhythm combined with harmony, where complex chords dampened the effect ([Matthews et al. 2019](https://pmc.ncbi.nlm.nih.gov/articles/PMC6328141)).

  Counter-evidence: across 248 real drum patterns, syncopation and event density had only small positive effects (R² ≈ 0.01 each). Whether the listener liked the style mattered far more (R² ≈ 0.15) ([Senn et al. 2018](https://pmc.ncbi.nlm.nih.gov/articles/PMC6025871)). A 2024 replication did not confirm the inverted U ([DOAJ](https://doaj.org/article/80b7e08cac2d4344916b360d7696a983)).

### (b) Implications

- **Current gap:** the engine marks a part `Syncopated` when off-beat hits are ≥ 1/4 of the total. That is *off-beatness*. Under LHL an off-beat hit counts only if the next stronger position is *silent* (kick on "& of 2" with 3 empty = anticipation; 8th hats with every beat sounded = not syncopated). The engine should compute LHL per voice and for the combined kick and snare (as Witek did) and report:
  - which steps syncopate,
  - against which beat,
  - a total score with fixed bands (none / light / strong).
- Explain the *mechanism* ("the kick arrives before beat 3 and beat 3 is silent, so it feels pushed forward"), not a contested promise like "medium syncopation = more groove".
- Also report density (onsets per bar per voice), the other robust correlate.
- UI: highlight syncopated cells on the grid and draw an arrow to the silent strong beat. Show the score as a three-level meter.

## 4. Anatomy of drum grooves

### (a) What the sources say (drum pedagogy; details **[varies]**)

**Roles [settled]:**
- kick: anchors the downbeats and the bass rhythm.
- snare: the backbeat, on 2 and 4 in rock, pop and R&B.
- hi-hat or ride: the subdivision "clock".
- cross-stick: a soft snare voice used in ballads and Latin styles.
- hi-hat foot: the 2-and-4 "chick" in jazz.
- toms: mostly fills.

Defining features by style. Positions are 16ths in a 4/4 bar, 0–15, unless noted.

| Style | Recognisable by |
| --- | --- |
| Rock/pop 8-beat | 8th hats; snare 4 and 12; kick 0 and 8 (variants add the "& of 3", step 10) |
| 16-beat | 16th hats (or alternating accents); kick on 16th off-beats |
| Funk | 16th hats; snare ghost notes at low velocity; syncopated kick; strong "one" |
| Half-time | snare only on beat 3 (step 8); feels half the tempo |
| Shuffle | triplet grid; ride or hat on triplet partials 1 and 3; backbeat; Purdie variant adds ghost notes on partial 2 |
| Jazz swing | ride "spang-a-lang": quarters plus the swung "& of 2" and "& of 4"; hi-hat foot on 2 and 4; kick feathered on quarters; snare comps freely ([Vic Firth](https://ae.vicfirth.com/wp-content/uploads/FADrumset-swingin-easy.pdf), [LJS](https://www.learnjazzstandards.com/blog/learning-jazz/drums/jazz-drumming-mastering-ride-cymbal-part-2/)) |
| Two-beat | kick and bass on 1 and 3; hats on 2 and 4 (Dixieland, country, early swing) **[unverified source]** |
| Bossa nova | 8th hats; kick plays a surdo-like long-short figure; cross-stick plays a 2-bar clave-like timeline ([Brit+Co guide](https://guides.brit.co/guides/play-a-bossa-nova-groove-on-the-drum-set); [Rock Drumming System](https://www.rockdrummingsystem.com/underground/drum-lessons/bossa-nova.php)) |
| Samba | continuous 16ths; surdo accent on beat 2 of 2/4 **[unverified]** |
| Clave | son 3-2 onsets at pulses 0, 3, 6, 10, 12 of a 16-pulse cycle (Toussaint; in practice one 4/4 bar of 16ths or two bars of cut time); rumba moves the third to 7 ([Toussaint](https://archive.bridgesmathart.org/2002/bridges2002-157.html)) |
| Ballad | sparse; cross-stick or snare on 2 and 4; often a 12/8 triplet feel |
| Boom-bap / lo-fi | about 80–95 BPM; heavy snare on 2 and 4; kick on 1 plus 16th off-beats around beat 2–3; 8th or 16th hats with 16th swing; lo-fi adds late placement and low velocity **[practice; tempos unverified]** |

### (b) Implications

- A drum description should use a **fixed voice vocabulary**: kick, snare, cross-stick, closed hat, open hat, hat-foot, ride, ride-bell, crash, toms. These map to GM notes, and each voice carries a role label the explanation can quote.
- Explainable pattern-level checks: "backbeat" = snare/cross-stick on 2 and 4 at top velocity; "half-time" = main snare on 3 only; "clock voice" = densest regular voice.
- Clave, bossa, samba and fills need a **two-bar cycle**: allow 1, 2 or 4 bars.
- Adding a swing ride pattern and a shuffle to the catalogue fixes the "swing jazz" failure independently of the step-grid change.
- UI: let users solo or mute each voice ("hear the backbeat alone"). Label each lane with its role.

## 5. Accompaniment and bass figures

### (a) What the sources say (**[varies]**, jazz, pop and classical pedagogy)

**Comping:**
- Charleston rhythm: beat 1 plus the "& of 2". The reverse form is the "& of 1" plus beat 3 ([jazzguitar.be](https://jazzguitar.be/blog/jazz-guitar-chord-rhythms)).
- Freddie Green four-to-the-bar guitar: short, even quarters with time ahead of voicings ([freddiegreen.org](https://www.freddiegreen.org/technique/berens.html)).

**Other accompaniment figures:**
- Block chords.
- Arpeggios and Alberti bass (low–high–middle–high in 8ths).
- Strumming: down-strokes on beats and up-strokes on off-beats, with "ghost" strums that miss the strings keeping the hand moving.
- Anticipation: the next chord arrives an 8th early, on the "& of 4". This is a rhythmic device that moves *harmony*.

**Bass:**
- Root only: on beat 1, or as 8th "pump" roots, sometimes alternating octaves (disco and rock).
- Root–fifth: two-beat feel, country, Latin.
- Walking bass: quarters, with a chord tone (usually the root) on beat 1 and an approach tone on beat 4. The approach tone is a scale or chromatic step to the next root.
- Bossa bass: root on 1 and fifth on 3, each with a long-short (dotted quarter + 8th) figure that locks with the kick ([PianoGroove](https://pianogroove.com/bossa-nova-lessons/basslines-rhythm-grooves/)). Exact variants differ.

**Rhythm vs harmony [settled in principle]:** rhythm fixes onsets, durations, accents and articulation. Harmony fixes which pitch fills each onset, given a role: root, fifth, chord tone, or approach. The figures that cross the boundary are walking bass and anticipation, because both depend on the *next* chord.

### (b) Implications

- Keep the existing role-based `Voice` field (chord/bass/root/third/fifth/top/octave). It already separates rhythm from harmony. Add two roles:
  - `approach`: a step toward the next chord's root, resolved by a fixed deterministic rule such as "chromatic from below".
  - `next-chord`: for anticipations. It needs a defined lookahead to the following `Harmony` segment.
- Comping should share the drums' step grid so analysis is shared ("guitar and kick land together on the & of 2").
- UI: bass and comping lanes under the drums on one grid, pitch roles coloured (root / fifth / approach) so rhythm and harmony are visibly separate.

## 6. Parameters vs templates

### (a) What the sources say

- Generative and analytic work describes rhythm with **scalar features**: event density, syncopation (LHL or WNBD), complexity, beat salience, and microtiming ([Senn et al. 2018](https://pmc.ncbi.nlm.nih.gov/articles/PMC6025871)). Sioros et al. generated controlled syncopation levels with a deterministic transformation algorithm ([2014](https://pmc.ncbi.nlm.nih.gov/articles/PMC4165312)).
- Scalars *measure* a pattern but do not determine one: many patterns share density and syncopation scores. Style identity comes from *specific positions* (backbeat, clave, spang-a-lang); Toussaint shows timelines differ in exact onset geometry ([Toussaint, *The Geometry of Musical Rhythm*](https://www.routledge.com/The-Geometry-of-Musical-Rhythm-What-Makes-a-Good-Rhythm-Good-Second-Edition/Toussaint/p/book/9780815370970)).
- Pedagogy teaches named patterns plus variations (ghost notes, accents, a fill every 4 or 8 bars), not parameter settings.

### (b) Implications (recommended shape)

**What the LLM writes** is an explicit **step-grid description**:

```
meter, bars (1|2|4), subdivision per beat (2|3|4|6),
voices: [{voice, role, steps: [{index, level: accent|normal|ghost}]}],
feel: {swingRatio, swingLevel, offsetsMs?},
basedOn?: catalogue id
```

**What the engine computes** from that grid, deterministically:
- validation:
  - indices in range;
  - subdivision divides PPQ;
  - known voices;
  - velocity mapped from 3 fixed levels;
  - IOI ≥ ~100 ms at the given tempo;
  - a playability limit (at most 2 hand voices plus 2 feet at one step);
  - density bounds.
- explanation: LHL syncopations, density, backbeat and half-time detection, swing in words, and **distance to the nearest catalogue pattern**. For example, Hamming distance on the grid gives "this is the jazz ride pattern with the kick moved to the & of 4".

**How templates fit in:** they stay as reference examples and named starting points, not as the only options. Fills are a separate one-bar variation placed by a fixed phrase rule.

**What to avoid:** scalar-only requests ("density 0.6, syncopation medium"); placement then becomes unexplainable ("why is the kick there?") unless a documented deterministic algorithm decides it.

**UI:**
- an editable grid with toggle and accent/ghost cycling;
- the derived labels updating live: backbeat ✓, syncopations, density, swing;
- a "closest known groove" badge.

---

### Verification notes

- Only abstracts and secondary summaries were read for Friberg & Sundström, Honing & de Haas, Benadon, Butterfield, LHL, Witek, Frühauf and Davies. attackmagazine.com was blocked by the session network (Linn percentages unverified). Genre tempos, two-beat/samba details, Keil, Lerdahl & Jackendoff and Charnas are from general knowledge.
