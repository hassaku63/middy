# Harmony research for middy's domain model

## How this was checked

Direct page fetches failed in this session. These hosts were not on the network allowlist: `viva.pressbooks.pub`, `en.wikipedia.org`, `human.libretexts.org` and `musictheory.pugetsound.edu`. Web *search* worked, so the citations below rest on search-engine excerpts of the cited pages, plus standard textbook content.

Verification markers:

- **[snippet]**: an excerpt of the cited URL confirmed the claim.
- **[secondhand]**: the claim reached me through another page, for example Wikipedia quoting Kostka & Payne.
- **[unverified]**: standard pedagogy as I know it, with no source confirmed in this session.

---

## 1. Functional classification (T / S / D)

### (a) What the sources say

**Settled**

- I is tonic (T) and V/V7 is dominant (D).
- IV and ii are subdominant or predominant (S/PD).
- vii° (and viiø7, vii°7) is dominant: usually "either a passing harmony or has dominant function" [snippet, Puget Sound *Music Theory for the 21st-Century Classroom*, https://musictheory.pugetsound.edu/mt21c/DiatonicChordsInMinor.html]. It is often taught as an incomplete V7 [unverified, standard in Kostka & Payne].
- Harmonic rhythm and function are separate concerns (see §4).

**Disputed: iii and vi.** The textbooks split three ways.

- **Three-group tonic model.** Open Music Theory 1e says "I, III, and VI as tonic, II and IV as subdominant, V and VII as dominant" [snippet, https://human.libretexts.org/Bookshelves/Music/Open_Music_Theory_1e/03:_Harmony/3.02:_Harmonic_Functions]. middy currently follows this model. Berklee-style jazz harmony also treats iii-7 and vi-7 as tonic substitutes [unverified].
- **Weak-predominant model.** Open Music Theory 2e treats iii and vi as "weak predominants that most often travel through a strong predominant to get to V". In minor, VII, III and VI play the same role [snippet, https://pressbooks.nebraska.edu/openmusictheory/chapter/the-mediant/ and https://pressbooks.nebraska.edu/openmusictheory/?p=358].
- **Riemannian function theory.** vi = Tp (tonic parallel). iii = Dp, which is also Tg (tonic counter-parallel), so iii has a genuine double reading [snippet, https://en.wikipedia.org/wiki/Parallel_and_counter_parallel].

vi as the target of the deceptive cadence (V–vi) behaves as a tonic substitute in every school. vi in I–vi–IV–V behaves as a bridge toward S.

**Disputed: VII in natural minor** (C or C7 in D minor). Four readings appear in the sources:

1. **V/III (applied dominant of the relative major).** Kostka & Payne describe VII as "sounding like the V in the key of the relative major—that is, a V of ♭III". Forte says the same: VII "becomes V in relation to III" [secondhand via https://en.wikipedia.org/wiki/Subtonic]. With a seventh added (C7 → F), every classical textbook would label it V7/III [unverified, but follows directly from the definition].
2. **Passing or predominant** in the descending bass i–VII–VI–V [snippet, search summary of a university text; host not fetched].
3. **Dominant substitute.** This is the jazz "backdoor" ♭VII7 → I, with ♭VII7 standing for V7 [snippet, https://en.wikipedia.org/wiki/Backdoor_progression; https://www.learnjazzstandards.com/blog/learning-jazz/jazz-theory/backdoor-progression/]. The same reading appears in rock and pop as the Aeolian cadence ♭VI–♭VII–i/I [unverified naming]. Spicer treats ♭VII as a V substitute; Everett hears it as an embellishment of IV [secondhand via Temperley, https://mtosmt.org/issues/mto.11.17.4/mto.11.17.4.temperley.php].
4. **Modal mixture / subdominant minor.** In a *major* key, ♭VII is borrowed from the parallel minor (Mixolydian). Berklee's *Harmony 2* has a section on "Subdominant Minor (SDM)" and a separate one on "Dominant Resolution and ♭VII7" [snippet, table of contents at https://docdrop.org/download_annotation_doc/Harmony-2-3rd-Edition--7cfvh.pdf]. A third-party outline lists ♭VII7 among the SDM chords [snippet, https://www.scribd.com/doc/251361724/Modal-Interchange]. Berklee's own label was **[unverified]**.

Open Music Theory 2e classifies the rock ♭VII–IV–I ("double plagal") as a Mixolydian schema [snippet, https://viva.pressbooks.pub/openmusictheory/chapter/modal-schemas/].

**Mainstream conclusion.** No school gives VII in minor a fixed function; the classification depends on what follows.

- VII → III is V/III (classical).
- VII → i is a weak, leading-tone-less dominant (pop, rock, jazz backdoor).
- VII inside a stepwise descent is passing.

### (b) Implications

- **Make function context-dependent.** Keep `Function` T/S/D, but compute it from (degree, quality, *next chord*), not from degree alone. The engine already looks at neighbours for cadences, so this stays a pure function over the progression.
- **Add a strength or role field.** For example `primary | substitute | passing`, so that vi can show "T (substitute)" and VII → i can show "D (weak: no leading tone)".
- **Rule for minor VII.** Next chord III → analyse as V/III. Next chord i → weak D. Otherwise → passing/S. This replaces today's blanket "VII = weak dominant".
- **Explanation text.** For iii and vi, name one convention as middy's own and add a short caveat that other books group these differently. Do not show users three competing labels.
- **Record the choice in an ADR.** "middy uses the three-group model (OMT 1e / Berklee), with context overrides".

---

## 2. Secondary dominants and borrowed chords

### (a) What the sources say

**Notation (settled)**

- Applied chords are written `V/x`, `V7/x`, `vii°/x`, `vii°7/x` and `viiø7/x`, read "five of x".
- Only major or minor triads can be tonicised. In major that is ii, iii, IV, V and vi, so there is no V/vii°. In minor it is III, iv, V, VI and VII (often iv and V) [unverified; universal in Kostka & Payne and Aldwell & Schachter].

**Frequency.** V7/V is "the most frequently encountered" applied chord, and V/ii is close behind, often as I–V/ii–ii–V [snippet, https://en.wikipedia.org/wiki/Secondary_chord; https://www.ethanhein.com/wp/2024/improvising-over-secondary-dominants/]. Wikipedia cites III7 = V7/vi in "The Shoop Shoop Song" [snippet]. In minor, V7/V and V7/iv are the common ones [snippet, Hein].

**Labels for chains (vary by school).** In jazz, chains of applied dominants (III7–VI7–II7–V7) are often written as plain Roman numerals with the quality attached ("VI7", "II7") rather than "V7/ii" [snippet, rhythm-changes search; https://en.wikipedia.org/wiki/Rhythm_changes]. People argue about whether A7 in C is V7/ii or something else [snippet, Reddit thread in the search results]. Classical texts insist on slash notation.

**Borrowed chords (modal mixture).** Open Music Theory says:

- The most common borrowed chords use only ♭6̂: iio6, iiø7 and iv, typically as cadential predominants.
- Borrowing ♭3̂ (♭III) and ♭7̂ (♭VII) is "also common".
- Mixture is more common in major (borrowing from minor) than the reverse [snippet, https://pressbooks.nebraska.edu/openmusictheory/chapter/modal-mixture/].

Pop and jazz add these:

- ♭VII and ♭VI as major triads (Aeolian/Mixolydian colour).
- The backdoor iv7–♭VII7–I [snippet, backdoor sources above].
- Berklee's SDM family: iv-7, iv-6, ♭VImaj7, ♭VII7, iiø7 and ♭IImaj7 [snippet, Scribd outline].

**Frequency data.** Hooktheory's database confirms that diatonic pop progressions dominate (I–V–vi–IV is the most common, and IV→I is about as frequent as V→I) [snippet, https://www.hooktheory.com/blog/i-analyzed-the-chords-of-1300-popular-songs-for-patterns-this-is-what-i-found/]. However, Hooktheory says its borrowed-chord counts are approximate because grouping is incomplete [snippet, https://forum.hooktheory.com/t/computing-chord-statistics-for-hooktheory-trends/112]. I found **no reliable frequency ranking** of ♭VII, ♭VI, iv and ♭III.

**Accidental prefixes (varies by school).** Classical texts (Kostka & Payne, OMT) write minor-key diatonic chords without a flat (VI and VII in minor) and use ♭VI and ♭VII only for chords *borrowed into major*. Jazz and Berklee measure every root against the *major* scale, so they write ♭VI and ♭VII even in minor [unverified as a blanket rule; clearly visible in the backdoor and SDM sources].

**Teaching to beginners.**

- Secondary dominants: "borrow the V of another chord to point at it", with the spotlight or arrow metaphor [unverified; standard at musictheory.net and Hooktheory].
- Borrowed chords: "take the same-degree chord from the parallel minor for a darker, bittersweet colour".

### (b) Implications

A chord's identity needs more than `degree + 7th`:

- `Root`: degree 1–7 plus `alteration` −1/0/+1 relative to the key's own scale.
- `Quality`: derived from the key *unless* overridden.
- `Applied`: optional target degree and the applied chord type (V, V7, vii°, vii°7, viiø7).
- `Source`: for example `diatonic | parallel-minor | harmonic-minor | mixolydian | dorian`.

What the engine needs to analyse each kind:

- **Applied chords:** check that the target is a consonant triad in the key.
  - Expected resolution: the next chord is the target.
  - Deceptive resolution: the next chord is something else.
  - Chained applied chords: the next chord is itself an applied V. These are needed for rhythm changes.
- **Borrowed chords:** each needs a source mode and an inherited function.
  - iv, iiø7 and ♭VI → S.
  - ♭VII → S or weak D, decided by the next chord as in §1.
  - ♭III → T-family colour [unverified as a standard label].
- **Accidental display:** pick one convention and record it in an ADR. The classical convention (no flat for VI and VII in minor) matches middy's "quality comes from the key" model best.
- **UI:**
  - Show the arrow (V/V → V) and the borrowed source ("from C minor").
  - Let users toggle "colour" substitutions, such as IV → iv or V → V7/x before x, as discrete deterministic options rather than free text.

---

## 3. Enharmonic spelling

### (a) What the sources say

**Settled rules for spelling a key**

1. A diatonic scale uses each letter A–G exactly once, so the key signature fixes the spelling.
2. A minor key uses its relative major's signature. The harmonic-minor raised 7̂ is written as an accidental on the same letter: C♯ in D minor, never D♭ [unverified; universal].
3. The enharmonic major pairs are F♯/G♭ (6♯/6♭), C♯/D♭ (7♯/5♭) and B/C♭ (5♯/7♭) [snippet, https://mymusictheory.com/scales-and-keys/major-scales-with-6-sharps-flats/ and https://jadebultitude.com/key-signatures/].

**Convention (not settled):** prefer the spelling with fewer accidentals. So D♭ over C♯ and B over C♭. F♯ versus G♭ is a genuine tie decided by context or habit [snippet, same sources; one says G♭ is "slightly more common", which is **unverified**]. The usual practical list of majors is C G D A E B F♯/G♭ D♭ A♭ E♭ B♭ F. For minors it is A E B F♯ C♯ G♯ D♯/E♭ B♭ F C G D [unverified].

**Chords** are stacked thirds on letters: root, root+2 letters, +4, +6. Chromatic chords keep the *letter of their degree* and change only the accidental. Sources say spelling should follow function and context ("notes … occur in lines … key areas … arpeggios"), and that respelling a note changes which chord and key it implies [snippet, https://tobyrush.com/soundpatterns/chromaticIII/linearchrom.html; https://open.lib.umn.edu/musiccomposition/chapter/the-diminished-seventh-chord/]. Applied to secondary dominants and borrowed chords, this gives:

- V/V in C = D–F♯–A (F♯ is the leading tone of G, never G♭).
- ♭VI in C = A♭–C–E♭, never G♯–B♯–D♯.
- V7/IV in F = F–A–C–E♭.

No source I reached states "spell the applied chord from its target" as a named rule; it follows from rules 1–3 [inference].

### (b) Minimal rule set for a deterministic engine

1. **Represent spelling apart from pitch class.** Store letter (0–6) plus accidental (−2..+2). Pitch class is derived from them, never the other way round.
2. **Fix the tonic spelling per key from a table.** 15 majors and 15 minors, or 12+12 with a fixed tie-break for F♯/G♭ and D♯/E♭ minor. This one change fixes the "A♯ major" defect (B♭ comes from the table).
3. **Spell scale degree n** as letter = tonic letter + (n−1) mod 7, with the accidental chosen to hit the scale's pitch class.
4. **Spell chord tones** as root letter + 2, 4 and 6, with accidentals chosen to hit the quality's intervals.
5. **Applied chords:** root letter = target root letter + 4 (a fifth above). Then apply rule 4.
6. **Borrowed chords:** the same letter as the diatonic degree, with the alteration applied.
7. **Double sharps** appear legitimately in sharp keys, for example V/vi in C♯ major = A♯–C𝄪–E♯. Either support them or restrict the key list to at most six accidentals and avoid the rare cases.

---

## 4. Harmonic rhythm

### (a) What the sources say

**Settled**

- Harmonic rhythm is the rate of chord change, independent of surface rhythm. "Fast" and "slow" are relative [snippet, https://en.wikipedia.org/wiki/Harmonic_rhythm].
- Chords "most often change on strong beats rather than on weak beats, and more often on downbeats than upbeats". The rate often accelerates toward a cadence [snippet, Iowa State *Comprehensive Musicianship*, https://iastate.pressbooks.pub/comprehensivemusicianship/?p=248].
- In common-practice music, harmonic rhythm and metric accent reinforce each other [snippet, Fiveable AP glossary].

**Genre rates.** These are rules of thumb, and I found no quantitative source for them:

| Genre | Typical rate in 4/4 |
| --- | --- |
| Pop / J-pop | 1 chord per bar in verses; 2 per bar (2 beats each) in choruses and at cadences. Changes are often anticipated by an eighth note ("push") [unverified] |
| Jazz, medium and up-tempo standards | 1 or 2 per bar. A ii–V often shares one bar [unverified] |
| Rhythm changes | A sections are 2 beats per chord. The bridge is 2 bars per chord, a deliberate slow-down [snippet, https://en.wikipedia.org/wiki/Rhythm_changes] |
| Jazz ballads | Players reharmonise more densely, up to 1 chord per beat [unverified] |
| Bossa nova | Usually slow: 1 chord per 1–2 bars, as in "The Girl from Ipanema" A section [unverified]. The 2-and-4 accent of the comping pattern is rhythm, not harmonic rhythm [snippet, https://learnjazzstandards.com/blog/bossa-nova-chord-progressions] |

**Teaching.** Students count changes per bar, then see that the cadential tonic lands on a strong bar (bar 4 or 8 of a phrase) [unverified].

### (b) Implications

- Restrict chord length to values that land changes on beat 1 or beat 3 in 4/4: {2, 4, 8} beats, with 1 beat only below a tempo threshold. Validate that a phrase totals whole bars and that the final tonic starts on a downbeat.
- Treat a "push" as an accompaniment-pattern property. Keep it out of the harmony data.
- **UI:** a simple "chords per bar: 1 / 2" control is the beginner-level lever. Show the bar grid with chord boxes so users can see harmonic rhythm.

---

## 5. Common progressions by genre, and the vocabulary each needs

| Progression | Roman numerals | Vocabulary needed | Supported now? |
| --- | --- | --- | --- |
| Royal road (王道進行) | IVmaj7–V7–iii7–vi(7) | Diatonic 7ths | Yes |
| Axis (Hooktheory's most common) | I–V–vi–IV | Diatonic | Yes |
| Canon | I–V–vi–iii–IV–I–IV–V | Diatonic | Yes |
| Komuro progression (小室進行) | vi–IV–V–I | Diatonic | Yes |
| "Just the Two of Us" (丸サ進行) | IVmaj7–III7–vi7–(v7–I7) | V7/vi, V7/IV, borrowed v7 | Planned |
| Major ii–V–I | ii7–V7–Imaj7 | Diatonic 7ths | Yes |
| Minor ii–V–i | iiø7–V7(♭9)–i(m6/m7) | Harmonic-minor V7, ideally ♭9 | Partly (no tensions) |
| Rhythm changes | I–VI7–ii7–V7; bridge III7–VI7–II7–V7 | Applied dominants, chained, without arrival at the target | Planned (chaining is not yet considered) |
| 12-bar blues | I7–IV7–V7 (jazz: ii–V in bars 9–10, I–VI7–ii–V turnaround) | Dominant 7th on I and IV with *tonic and subdominant* function | **No: breaks "quality from key"** |
| Backdoor | iv7–♭VII7–I | Borrowed iv, ♭VII7 | Planned |
| Aeolian / double plagal | ♭VI–♭VII–I, ♭VII–IV–I | Borrowed ♭VI, ♭VII | Planned |
| Andalusian | i–VII–VI–V | Harmonic-minor V, context-dependent VII | Partly |

Sources and verification for the table:

- Royal road: F–G–Em–Am in C; earliest J-pop example 1974 [snippet, https://www.wikipedia.org/wiki/Royal_road_progression]. A claim that 40% of Japan's top-20 singles use it is **[unverified]** (Wikipedia cites a 2023 paper).
- Axis: [snippet, Hooktheory URL above].
- Minor ii–V–i: the ♭9 and ♭13 of V7 come from the tonic's harmonic minor [snippet, https://www.jazzguitar.be/blog/harmonic-minor-scale/; https://www.learnjazzstandards.com/blog/minor-2-5-1/].
- Rhythm changes: [snippet, https://en.wikipedia.org/wiki/Rhythm_changes; https://jazzguitar.be/blog/introduction-to-rhythm-changes].
- Blues: [snippet, https://www.learnjazzstandards.com/blog/learning-jazz/jazz-theory/blues-chord-progressions/; https://www.ethanhein.com/wp/2023/the-twelve-bar-blues/].
- Double plagal: [snippet, OMT modal schemas URL above].
- Komuro, "Just the Two of Us" and Andalusian rows: **[unverified]**, from general knowledge.

**Implications**

- Blues needs a chord *quality override*: I7 in a major key, analysed as T rather than V7/IV. Without it, the engine would label blues I7 as an applied dominant, which is wrong in that style. This is a design decision for an ADR.
- Chained applied dominants need the analysis to accept a target that never arrives.
- A pattern catalog keyed by progression id, with its own explanation template, lets the LLM pick an id without writing chords. This is consistent with ADR-0003.

---

## 6. Sevenths and extensions (9, 11, 13)

### (a) What the sources say

**Jazz and bossa.**

- The 7th is part of the basic chord symbol. Tensions (9, 11, 13 and the altered ♭9, ♯9, ♯11, ♭13) are added by the player as colour.
- In minor, V7 takes ♭9 and ♭13 from harmonic minor [snippet, jazzguitar.be above].
- Bossa specifically favours maj7, m9, 6/9, 13 and dominant ♯11 voicings. Ninths and thirteenths are "popular among bossa nova guitar players" [snippet, https://sites.google.com/view/franksmusicmultiverse/styles/bossa; https://www.pianogroove.com/?p=3033]. In these styles the extensions are part of the style's identity: a bossa on plain triads does not sound like bossa.
- Berklee's chord-scale "available tensions" are a conventional system, not universal theory **[unverified details]**:
  - maj7: 9, ♯11, 13; natural 11 is an avoid note.
  - m7: 9, 11.
  - Diatonic V7: 9, 13.
  - V7 in minor: ♭9, ♭13.
  - m7♭5: 11, ♭13.

**Pop.**

- Triads dominate. 7ths and add9/sus chords are colour, common in city pop and the J-pop royal road (IVmaj7–V7–iii7) but not identity-defining [unverified generalisation].
- Hooktheory-style analysis of pop is written mostly in triads [inference from its published statistics].

### (b) What a beginner-facing app should show

- **Analysis level:** stop the Roman-numeral analysis at the 7th (I, V7, iiø7). Extensions are a *voicing* concern.
- **Voicing level:** let the accompaniment pattern add tensions deterministically, per style. For example, a jazz or bossa feel adds 9 to maj7/m7 and 13 to dominants, and minor-key V7 gets ♭9. The LLM must not choose individual tensions.
- **Display:** show the full chord symbol (e.g. `Cmaj9`) alongside the simpler analysis (`Imaj7`), and explain in one line that "the 9 is colour; the function is unchanged".
- **User control:** offer a "colour: simple / rich" toggle rather than per-tension editing.
- **Data model:** keep `Seventh` in the harmony type. Add tensions as a set of scale-degree intervals on the *voicing* or pattern layer, validated against a per-quality allow-list so that avoid notes (natural 11 over a major or dominant chord) never sound by accident.
