# Harmony × Rhythm: how they combine in real composition

**Method and limits.** Sources were found by web search. Full-page fetches were blocked in this session, so claims rest on search excerpts of the cited pages. **[u]** = unverified (background knowledge or a single weak source).

---

## 1. How separable are harmony and rhythm?

### (a) What sources say

**Notation draws the line at "what chord, for how long". The groove is left open.**

- A lead sheet specifies melody, lyrics and harmony. It "does not describe the chord voicings, voice leading, bass line or other aspects of the accompaniment", and the performer is expected to improvise them ([Wikipedia: Lead sheet](https://en.wikipedia.org/wiki/Lead_sheet); [Berklee Today, "Why Lead Sheets?"](https://www.berklee.edu/berklee-today/summer-2018/lead-sheet)). The groove is a one-word mark at the top ("Med. Swing", "Bossa"); a tune marked only "Med." is played as swing ([Q&A](https://ftyjkyo.blogspot.com/2019/11/what-does-med-as-style-mean-in-jazz.html), weak). Settled: chart and groove are separate layers.
- Nashville number charts go further than lead sheets, but still only into *harmonic rhythm and band-wide accents*. Underlined numbers share a bar, a diamond means "hold the whole bar", `<`/`>` marks a push (attack an eighth early), and a fermata or marcato means a stop or choke. Groove is still not written ([Soundfly](https://flypaper.soundfly.com/write/what-is-the-nashville-number-system/); [JotChord reference](https://www.jotchord.com/quick-reference-guide)). The push symbol conventions **vary between sources**, so a chart needs a legend.
- Arranger keyboards are the clearest engineering precedent for the split. A *style* is a set of rhythm patterns recorded in one source chord (Yamaha's default is CM7). At playback, note-transposition rules (NTR) and tables (NTT) map those patterns onto whatever chord the player holds. A "Guitar" NTR handles guitar fingering ([Yamaha Tyros3 manual](https://www.manualslib.com/manual/477787/Yamaha-Tyros3.html?page=46); [CVP-709 reference](https://www.manualowl.com/m/Yamaha/CVP-709/Manual/702498?page=58)). Korg also maps chord types (M7, m7…) to different pattern *variations*, so the pattern itself can depend on chord quality ([Korg Pa4x](https://www.manualslib.com/manual/1354785/Korg-Pa4x.html?page=245)).
- Practice tools keep one chart and swap styles. In iReal Pro, switching style also resets tempo to the style's default ([iReal Pro help](https://technimo.helpshift.com/hc/en/3-ireal-pro/faq/151-two-different-playback-styles-in-one-song); [changelog](https://mac.softpedia.com/progChangelog/iReal-b-Changelog-92150.html)). Band-in-a-Box stores pushes, rests, shots and holds on the *chord sheet*, while styles carry their own pushes; each can be switched off ([Sound on Sound review](https://www.soundonsound.com/reviews/pg-music-band-box-2017-0); [PG Music manual](https://d.pgmusic.com/manuals/bbm2018full/chapter15.htm)).

**Cognition: they are not independent in perception.** Dawe, Platt & Racine found that where the chords change was the main determinant of where listeners heard the bar begin, and of how they grouped rhythms. Musicians relied on harmonic accents more than non-musicians ([McMaster record, 1993](https://experts.mcmaster.ca/display/publication235905); [thesis](https://macsphere.mcmaster.ca/handle/11375/15895)). Harmonic rhythm is therefore itself a metrical cue.

**Verdict on the hypothesis.** It is confirmed: harmony and groove are separable as *data* (what is played versus how it is articulated in time). They are not separable in the *result*. The listed seams are right. I would extend the list:

The five listed seams hold: harmonic rhythm, style binding, quality-dependent comping (Korg's tables), pushes, and harmony-derived bass. Bass needs the *next* chord for approach notes. Seven more:

6. Tempo ↔ harmonic rhythm (fast tempo, slow changes).
7. Meter (3/4, 6/8, 12/8) sets the unit harmonic rhythm is counted in.
8. Chord-change placement *is* a metric accent (cognition).
9. Band-wide articulations: stops, shots, holds (Nashville, BiaB).
10. Form: harmonic rhythm differs by section; fills and crashes mark phrase boundaries.
11. Style *colorization*: the same numeral becomes 6/9 in bossa, 9 in funk, a triad in rock.
12. Voice leading: comping depends on the *previous* voicing (state).

### (b) Implications

- **Domain model.** Use three entities, not two. `Harmony` holds key, chord events (degree, quality, duration, optional push) and phrase structure. `Groove` holds meter, feel, and drum, comping-rhythm and bass-rhythm patterns. `Style` is a named binding: preferred vocabulary, colorization rule, default tempo range, feel, compatible grooves, bass approach and voicing approach. The renderer is `Realize(Harmony, Groove, Style) → Timeline`.
- The swing-jazz failure is a missing `Style` constraint. Today `Candidate` picks `Accompaniment`, `Drums` and `Feel` independently from the progression, so nothing stops "jazz harmony + bossa drums". If the LLM chose a `style` id first, and `Validate` checked that drums, accompaniment and feel belong to that style's compatible set, this would become a deterministic check. It would not rely on prompting. BL-077 (rhythm as a description) should keep that style-compatibility check either way.
- **UI.** Show harmony and rhythm on one shared timeline, aligned in time and visually together. Give them independent controls: a "chords" axis, a "groove" axis, and a "style" preset that sets both. A user can lock one axis and vary the other.

---

## 2. Harmonic rhythm and metric placement

### (a) What sources say

- **Settled:** in common-practice tonal music, chord changes tend to fall on strong beats, so harmonic rhythm reinforces meter ([Fiveable AP glossary](https://fiveable.me/ap-music-theory/key-terms/harmonic-rhythm.md); [Wikipedia: Harmonic rhythm](https://en.wikipedia.org/wiki/Harmonic_rhythm)).
- **Hypermeter.** Bars group into hypermeasures, most commonly of four. In cadential music, hyperbeat 3 leads to a cadence on hyperbeat 4 ([Open Music Theory: Hypermeter](https://viva.pressbooks.pub/openmusictheory/chapter/hypermeter_v2)). **Practice varies:** scholars dispute whether four-bar groups always start strong. Cone warned of the "tyranny of the four-measure phrase", and Rothstein treats four-bar segments as metrical units rather than phrases (same source). For pop and beginner material, 4- and 8-bar units are a safe default.
- **Turnarounds** occupy the last two bars of a form, typically I–vi–ii–V at two beats per chord ([Guitar Player](https://www.guitarplayer.com/lessons/blues-turnarounds-part-1); [Yamaha Hub](https://hub.yamaha.com/to-be-reviewed/a-bassists-guide-to-turnarounds/); [LearnJazzStandards](https://www.learnjazzstandards.com/blog/jazz-turnarounds/)). Harmonic rhythm doubles at the end of the phrase.
- **Anticipations/pushes.** The chord is attacked about an eighth note early, usually on the "and of 4", and held over the barline. Sixteenth and triplet pushes are also common ([Darling, Encyclopedia of Music](https://daviddarling.info/encyclopedia_of_music/A/anticipation.html); [ledutu](https://ledutu.com/en/learn/jazz/syncopation-and-anticipation/)). Band-in-a-Box writes `^C` for an eighth push and `^^` for a sixteenth push ([Sound on Sound](https://www.soundonsound.com/reviews/pg-music-band-box-2017-0)). A push still *belongs* to the next bar harmonically; only its attack moves.
- **Pop songwriting practice** (one author's guidelines): keep one harmonic rhythm for about 80% of a section, let verse and chorus differ, pair faster tempos with slower changes ([Secrets of Songwriting](https://www.secretsofsongwriting.com/?p=5732)). Holding a chord longer than the established rate highlights a lyric ([Berklee](https://www.berklee.edu/node/2474896)).
- **Interaction with groove.** A push is a property of the chord event, but every layer has to play it: comping and bass move their attack, and drums usually reinforce it with kick and crash. Under swing, an eighth push lands on the third triplet. In funk, harmonic rhythm can drop to near zero (a one-chord vamp) and the groove carries the piece ([MusicRadar](https://musicradar.com/how-to/develop-funk-rhythm-guitar-grooves-with-these-classic-chord-vamps); [Wikipedia: Funk](https://www.wikipedia.org/wiki/Funk)).

### (b) Implications

- **Model.** Give `ChordEvent` a `Push` (`none | eighth | sixteenth`). It shifts the attack of every pitched layer but keeps `HarmonyEvent.Start` (bar ownership) for analysis, so "the V arrives in bar 4" stays true. `Feel` must be applied after the push so that a swung push lands on the triplet.
- Add a `Phrase`/`Section` entity (4 or 8 bars, cadence position). A deterministic check can then explain "the cadence lands on the downbeat of the phrase's last bar". Right now cadence detection does not look at metric position.
- `Candidate.BeatsPerChord` is uniform, but `music.Step.Beats` is per step. Turnarounds need per-step beats at the intent boundary (a gap, not a build recommendation).
- **UI.** Show bar lines and group bars into 4-bar hyperbars. Draw a push as a small arrow just before the barline. Put the cadence label on the downbeat it resolves to. When the groove changes, harmony markers must stay where they are.

---

## 3. Style as the binding

### (a) Compact table

Tempo figures are rules of thumb. Sources are blogs and Q&A unless noted, and **[u]** marks unverified entries.

| Style | Harmony vocabulary | Harmonic rhythm | Tempo (BPM) | Groove / feel | Bass | Comping |
|---|---|---|---|---|---|---|
| Pop / J-pop | Diatonic triads and 7ths. Royal road IV–V–iii–vi ([Wikipedia](https://en.wikipedia.org/wiki/Royal_road_progression), with the claim that it appears in 40% of Japan's top-20 singles). Komuro vi–IV–V–I | 1–2 chords per bar, steady within a section | 90–130 [u] | Straight 8ths or 16ths, backbeat on 2 and 4 | Root 8ths, octaves | Pad or arpeggio, 8th or 16th pulse |
| Rock | I, IV, V, ♭VII, VI most frequent. IV is the most common pre-tonic chord, ♭VII the third ([de Clercq & Temperley 2011](https://rockcorpus.midside.com/2011_paper/declercq_temperley_2011.pdf); [Temperley, "Cadential IV"](https://mtosmt.org/issues/mto.11.17.1/mto.11.17.1.temperley.html)). Triads and power chords [u] | 1 per bar, often 2 | 100–140 [u] | Straight 8ths, eight-beat | Root 8th pedal [u] | Strummed or power-chord 8ths |
| Pop ballad | Diatonic with 7ths, sus, slash-bass descents [u] | 1 per bar or slower | 60–80 [u] | Straight, sparse, half-time feel | Long roots, stepwise | Sustained pad or arpeggio |
| Jazz swing | ii–V–I, sevenths everywhere, I–vi–ii–V turnarounds | 1–2 per bar, 2 per bar in turnarounds | "Med." ≈100–150 ([Q&A](https://ftyjkyo.blogspot.com/2019/11/what-does-med-as-style-mean-in-jazz.html)) | Swing 8ths: ride "spang-a-lang", hi-hat foot on 2 and 4, feathered kick ([Soundbrenner](https://www.soundbrenner.com/blogs/articles/rhythm-library-jazz-swing)). Swing ratio varies with tempo | **Walking:** root on 1, chord or scale tones on 2–3, chromatic approach to the next root on 4 ([Guitar World](https://www.guitarworld.com/lessons/bring-a-new-level-of-rhythm-and-groove-to-your-guitar-playing-with-jazz-walking-basslines); [FreeJazzLessons](https://freejazzlessons.com/?p=53401)) | Shell voicings (3rd and 7th), sparse syncopated hits such as the "Charleston" [u for Charleston] |
| Bossa nova | maj7, 6/9, m9, 9 (6 and 9 chords often preferred over maj7) ([jazzguitar.be](https://www.jazzguitar.be/blog/blue-bossa-chords/)) | 1 per bar, often 2-bar chords | 90–130 ([weak](https://beatkey.pages.dev/how-to-make-bossa-nova-music)) | Straight 8ths, syncopated one- or two-bar cycle [u] | Root and fifth on 1 and 3 (**disputed**: some say root only) | Thumb bass plus syncopated finger chords, close voicings |
| Funk | dom9, m7. One- or two-chord vamp, changes only in bridges ([MusicRadar](https://musicradar.com/how-to/develop-funk-rhythm-guitar-grooves-with-these-classic-chord-vamps); [Piano with Jonny](https://pianowithjonny.com/piano-lessons/funk-chords-the-complete-guide/)) | Very slow (vamp) | 90–115 [u] | 16th grid, accent on "the One", muted 16ths | Syncopated 16ths anchored on 1 | Rootless 3–7 plus tensions, short stabs |
| Lo-fi hip-hop | maj7, m7, m9, 4-bar jazz-derived loops ([Native Instruments](https://blog.native-instruments.com/how-to-make-lofi-music/)) | 1 per bar, looped | 70–90 ([Mixed In Key](https://mixedinkey.com/captain-plugins/wiki/how-to-make-lofi-hip-hop/)) | Kick-snare 2 and 4, hats swung about 55–65%, loose quantize | Sparse roots | Sustained Rhodes-type 7th chords |
| Waltz / 6/8 ballad | Diatonic. Doo-wop I–vi–IV–V common in 12/8 [u for progression] | 1 per bar | 6/8 ballad ≈65–100 ([ProducerHive](https://producerhive.com/?p=41187)) | 3/4: oom-pah-pah. 6/8 and 12/8: kick on 1, snare on 4 (of 6), hats on every 8th ([Soundbrenner 12/8](https://www.soundbrenner.com/blogs/articles/rhythm-library-12-8)) | 3/4: root on 1 [u] | 3/4: chords on 2 and 3. 6/8: triplet arpeggios [u] |

**Settled vs varying.** The broad mapping is uncontroversial: swing pairs with 7ths and ii–V, bossa with 6/9 and maj7, rock with triads and ♭VII, funk with vamps. Exact tempos, swing ratios and bass details vary by player and era. The bossa bass disagreement and the dependence of swing ratio on tempo are explicitly noted in the sources.

### (b) Implications

- `Style` is a catalogue entity holding: compatible drum, accompaniment and feel ids; meter; tempo range; preferred vocabulary (a preference, not a ban); a colorization rule ("I → 6/9, V → 9"); bass approach (`root`, `root-fifth`, `walking`, `syncopated`); voicing approach (`close`, `shell`, `rootless`).
- Colorization stays a deterministic engine rule: the LLM returns `I`, the style adds the 9th (ADR-0003 intact). It needs tensions, currently unsupported.
- The engine cannot yet render several rows: no swing ride or walking bass, no 6/9 or 9 chords, 4/4 only (`pattern.Apply` rejects other meters). Keep such styles out of the catalogue until renderable rather than approximating; the swing-jazz failure is an approximation leaking through.
- **UI.** A style chip sits on the example card next to key and tempo, and the explanation says what the style changed ("in bossa the I chord gets a 9th").

---

## 4. Arrangement layers

### (a) What sources say

A minimal combo arrangement is bass, comping/pad, drums, and optionally melody. Each layer has a different derivation:

| Layer | Pitches derive from | Timing derives from | Needs |
|---|---|---|---|
| Bass | Harmony: root, fifth, chord tones, chromatic approach to the *next* root ([FreeJazzLessons](https://freejazzlessons.com/?p=53401)) | Groove pattern (walking quarters, root 8ths, bossa 1 and 3) | Look-ahead to the next chord |
| Comping / pad | Harmony: chord quality → voicing, inversion chosen by voice leading | Groove pattern, plus pushes from harmony | Previous voicing (state) |
| Drums | none | Groove pattern only | Form: fills and crashes at phrase ends and after pushes |
| Melody | Both: chord tones on strong beats | Both | Not in middy's current scope |

**Voice leading (settled in jazz and classical pedagogy).** Guide tones (3rd and 7th) define the chord. If the shell is inverted at each change, a ii–V–I moves only one voice by a half step at a time: in C, F–C → F–B → E–B ([Piano with Jonny](https://pianowithjonny.com/piano-lessons/how-to-use-guide-tones-for-jazz-piano/); [MusicLessons.com](https://www.musiclessons.com/pubs/lesson/116/9/voice-leading-comping-with-3rds-and-7ths)). Root-position chords lack this voice leading ([beginner guide](https://pianowithjonny.com/piano-lessons/the-beginner-jazz-piano-accompaniment-guide/)). Pease & Pullig's Berklee text *Modern Jazz Voicings* is canonical here but was not read.

Arranger keyboards encode the same split. The pattern stores *roles* relative to the chord, and the transposition table resolves those roles to pitches per chord type ([Korg Pa1000](https://www.manualslib.com/manual/1300117/Korg-Pa1000.html?page=257); [Yamaha](https://www.manualslib.com/manual/196986/Yamaha-Tyros.html?page=121)). middy's `Hit.Voice` (`chord`/`bass`/`root`/`third`/`fifth`/`top`) is the same design.

### (b) Implications

- **Pipeline:** `Harmony → VoicingPlan → Layers`. A `VoicingPlan` picks pitches per chord by a deterministic minimal-motion rule (hold common tones, move others by the smallest step). Patterns pick voices from the plan instead of `Chord.Voicing(octave)`, which is root-position close voicing, so I–IV–V jumps in parallel.
- The bass layer needs `(current, next)` chord context, not just `current`.
- Drum patterns need a "phrase end" variant (fill) and a "push" reaction. Both are rhythm data, triggered by harmony and form.
- **UI.** Use one lane per layer with solo and mute ("bass only" shows how the bass spells the harmony). Draw voice-leading lines between consecutive comping voicings. Mark derived pitches as coming "from the chords" and pattern-driven timing as coming "from the groove".

---

## 5. Learning value

### (a) What sources say

- **Separate, then combine.** Gordon's Music Learning Theory teaches tonal patterns and rhythm patterns separately: tonal patterns are sung free of rhythm, rhythm patterns are chanted. Rationale: clarity, less confusion ([MTNA handout](https://www.mtna.org/downloads/conference/handouts/2019/MLT.pdf); [Dansville FAQ](https://sites.google.com/a/dansvillecsd.org/primary-music/frequently-asked-questions)); evidence is teacher-training material.
- **Vary one aspect against a constant background.** Marton's variation theory holds that learners notice an aspect when it varies while the others stay invariant ([learningdiscourses](https://learningdiscourses.com/discourse/variation-theory)). (critique: [Scand. J. Ed. Research](https://informahealthcare.com/doi/full/10.1080/00313831.2012.656277)). This is the basis for "same progression, different groove" and "same groove, different progression".
- **Whole music matters too.** Lucy Green found popular musicians learn from whole real songs by ear, not sequenced exercises ([UCL IOE project](https://www.ucl.ac.uk/ioe/research-projects/2018/oct/adapting-popular-musicians-practices-classroom)). Isolation should therefore happen in the *controls*, while what the learner *hears* stays a complete, musical result.
- **Established practice:** drill ii–V–I in all twelve keys ([Piano with Jonny](https://pianowithjonny.com/?p=28709)); rehearse one chart in many styles (iReal Pro). Ableton's Learning Music teaches beats, chords and basslines as separate chapters that build into songs ([Engadget](https://www.engadget.com/2017/05/08/ableton-interactive-lessons/)).
- **Cognition caveat:** non-musicians use harmonic cues for meter less than musicians do (Dawe et al., above). Beginners may not *hear* harmonic rhythm as rhythm, so it is worth teaching explicitly.

### (b) Implications

- **One-variable controls.** BL-078 (change key only) already does this; extend to "change groove, keep progression", "change progression, keep groove" and "change harmonic rhythm, keep chords". The engine re-renders and the LLM is not called.
- **Deterministic diff explanation.** After a single change, explain only what changed ("the chords are identical; the bass now walks and the hi-hat is on 2 and 4").
- **Mismatches can be lessons.** A user may deliberately pair swing harmony with eight-beat drums; allow it, label it as unusual, and say why. LLM candidates stay held to style compatibility.
- **Default view combined, controls separated** (Green): a coherent style first, axes next.

---

## Unverified or weak points

- I read every source through search excerpts only. No full page could be fetched in this session.
- Tempo ranges for pop, rock, ballad and funk; rock bass and power-chord norms; waltz and 6/8 comping and bass; the Charleston comping rhythm; doo-wop I–vi–IV–V in 12/8: all from background knowledge **[u]**.
- Bossa and medium-swing tempo ranges come from single weak sources.
- Nashville push symbol direction conflicts between sources.
- The Berklee harmonic-rhythm article's author and the content of Pease & Pullig, *Modern Jazz Voicings*, were not verified.
- The strength of evidence for Gordon's MLT and for variation theory in music teaching was not assessed.
