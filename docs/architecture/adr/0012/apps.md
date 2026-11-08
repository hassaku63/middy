# How similar apps model and present harmony and rhythm

Survey for middy (beginner composition / theory PWA; LLM picks parameters, deterministic engine writes notes).

**Method and confidence.** Direct page fetches were blocked in this session's network sandbox (hooktheory.com, irealpro.com, support.apple.com did not resolve), so everything below comes from search-engine results that quote official manuals, vendor pages and reviews. Each claim links the page it was quoted from. Claims marked **[unverified]** come from background knowledge or from a single third-party source, and need a check against the product before anyone relies on them.

## 1. Comparison table

| App | Harmony input / display | Rhythm input / display | Harmony and rhythm separate? How bound | Teaching |
| --- | --- | --- | --- | --- |
| **Hookpad / TheoryTab** | Scale-degree chord palette in the chosen key/mode; chords coloured by function; Roman numerals always shown with chord symbols; borrowed and secondary chords; "Magic Chord" next-chord suggestions ranked by corpus frequency | Chord duration set on the timeline; band templates and drum presets **[unverified detail]** | Partly. Harmony and melody are separate layers stored as scale degrees + beat + duration; accompaniment comes from instrument presets | Strongest of the set: function colours, corpus statistics ("what usually comes next"), the book |
| **iReal Pro** | Typed chord chart (lead-sheet symbols); transposition; chord diagrams and scale suggestions | About 50 named styles (e.g. Medium Swing 2-feel / 4-feel, Bossa Nova); per-style mixer; tempo | **Yes, fully.** Chart = harmony only; style = rhythm/orchestration; the same chart plays in any style | Chord-scale suggestions; practice automation (tempo ramps, key cycling) |
| **Band-in-a-Box** | Chord symbols typed into bar cells; part markers a/b (up to 24 substyles) | Style (.STY) chosen in StylePicker; RealTracks/RealDrums audio replacements | **Yes.** Song = chords + a style reference; you can force one style across songs or change style at given bars | Little theory teaching; it is a practice and arranging tool |
| **Scaler 2/3** | Key/scale detection; 1,000+ chord sets tagged by mood/genre/artist; voicing (voice grouping, drop voicings) | "Motions" (formerly Phrases/Performances), tagged by mood; humanise timing/velocity; arpeggio/strum | Yes. Chord set and the Motion applied to it are independent | Shows chords in scale, but teaches little "why" |
| **Captain Plugins** | Key + scale, then preset progressions written as Roman numerals (transposable) | Rhythm presets by genre, with segments splittable down to 8ths; Captain Beat has swing, quantise, half/double time | Yes. One shared rhythm-pattern library is reused across Chords, Deep (bass) and Melody | Numeral notation, explained in its wiki |
| **Logic Pro Drummer & Session Players** | Chord Track / region chords; the Keyboard Player has voicing styles (block, broken, extended) | Drummer: genre + drummer persona, XY pad (simple-complex × soft-loud), Fills, Swing, per-kit-piece choices; Session Players: Complexity and Intensity sliders | **Yes.** Chords live on the timeline; players follow them; Drummer can "Follow" another track's rhythm | None for theory; the "persona" metaphor does the explaining |
| **EZkeys 2 / EZdrummer 3** | EZkeys Song Track holds your chords | Grooves library by genre; Bandmate matches grooves to your audio or MIDI; Tap2Find (tap a rhythm to search); grid editor with humanise | Yes. Song Track chords + a separately chosen play style | Chord suggestions only |
| **Chordbot** | Chord sequence, 60 chord types, slash chords; random generator | 50+ comping styles built from mixable instrument tracks; length per chord | Yes. Progression + style | None |
| **Ableton Live 12** | Stacks generator (up to 4 chords); scale awareness | Rhythm and Euclidean generators; Groove Pool (.agr) with Base, Quantize, Timing, Random, Velocity, Global Amount | Yes, and rhythm is a transform you can apply to any clip | None |
| **Magenta (GrooVAE, Drumify, Coconet)** | Coconet harmonises a melody in Bach chorale style | Drumify turns any rhythm into a drum groove; Groove adds human timing and velocity | Yes, and separated by design: score ↔ groove | Bach Doodle presented as a demo, not a lesson |
| **Groove Pizza** | (separate "Chord pizzas" experiments) | Circular step grid; polygons; swing drawn on the grid itself | n/a (rhythm only) | Strong: geometry makes rhythm visible |
| **musictheory.net / Musicca** | Lessons and drills on Roman numerals and chord identification; chord player | Rhythm lessons | Taught separately | Drill-based |
| **Suno / Udio** | Free text; key and BPM in the prompt are not reliably followed | Genre words in the prompt | **No.** One opaque audio output | None |

## 2. Per-app notes

### Hooktheory Hookpad, TheoryTab, the books
- **Harmony.** You pick chords from a palette of scale degrees in the current key. Hookpad "shows only the notes and chords in the song's key, coloring each note to denote its musical function", and its chord palettes are "persistent UIs that give you new ways to find and enter chords" ([Hooktheory blog](https://www.hooktheory.com/blog/hookpad-chord-progressions/), [Produce Like A Pro review](https://producelikeapro.com/blog/hooktheory-hookpad-review/)).
  - Magic Chord ranks the next chord by how often it appears in popular music.
  - Bass Sets list common chords over a given bass note.
- **Reference scale.** Staff say the major scale is the reference for numbering borrowed chords, and that Hookpad "always displays modern chord symbols along with Roman numerals". The colouring stands in for the bass note ([forum: Minor keys, Roman numerals](https://forum.hooktheory.com/t/minor-keys-roman-numerals/107?page=2), [forum: add9 display](https://forum.hooktheory.com/t/about-the-new-add-9-roman-numeral-display/725)). Secondary chords of borrowed chords are supported ([forum](https://forum.hooktheory.com/t/secondary-subdominants/2097)).
- **Rhythm.** Hookpad 2 shipped "42 guitar presets, 108 drums, 72 leads, and 61 basses"; a later listing says "500+ instruments across 11 band templates" ([Hooktheory blog](https://www.hooktheory.com/blog/introducing-hookpad-2-a-musical-sketchpad-with-music-theory-built-in/), [Hyperbits](https://hyperbits.com/blog/hooktheory/)). The chord rhythm itself is the chord's position and length on the timeline. I could not confirm how drum and comping patterns are chosen **[unverified]**.
- **Data.** The Hookpad clipboard JSON has `notes` and `chords` arrays with fields such as `sd` (scale degree), `octave`, `beat`, `duration` and `isRest`. Harmony and melody are separate, key-relative layers ([HF discussion](https://huggingface.co/datasets/m-a-p/HookTheory/discussions/2), [CMU pyquist](https://www.cs.cmu.edu/~15322/_modules/pyquist/web/theorytab.html)). The public Trends API exposes next-chord probabilities ([API docs](https://hooktheory.com/api/trends/docs)).
- **Teaching.** Book I teaches function with corpus evidence such as "which chords are most commonly used after ii?" ([Musical U review](https://www.musical-u.com/learn/review-hooktheory-music-theory-book)), built on 40,000+ TheoryTab analyses ([free resources](https://www.hooktheory.com/blog/free-hooktheory-resources/)).

### iReal Pro
- **Harmony and rhythm are separate.** A chart is a chord grid; the style is picked separately. Each style has its own instrument set in the mixer: piano, Rhodes, guitars, bass, drums, vibes, organ ([forum: instruments in styles](https://forums.irealpro.com/goto/post?id=39726), [App Store](https://apps.apple.com/us/app/298206806)).
- **Style catalogue.** Named feels sit side by side as separate styles, e.g. "Medium Swing (2 feel)", "Medium Swing (4 feel)" and "Bossa Nova".
- **Practice.** "automatic tempo increase, automatic key transposition" and "scale suggestions for improvisation" ([App Store](https://apps.apple.com/uz/app/id298206806)).
- **Pitfall.** Users report that changing style also changes tempo, and ask to be able to lock both ([forum](https://forums.irealpro.com/goto/post?id=57039)). Rhythm parameters leaking into the chart is a known pain point.
- **Format.** The `irealbook://` scheme has 6 fields: title, composer, style, key, `n`, progression. Newer `irealb://` links are scrambled and reverse-engineered rather than published. The progression string looks like `T44*A{C^7 |A-7 |D-9 |G7#5 }` ([protocol-launcher](https://zhensherlock.github.io/protocol-launcher/apps/ireal-pro.md), [pyRealParser](https://pypi.org/project/pyRealParser/)). The style is just a name in the chart; the chart carries no rhythm.

### Band-in-a-Box
- **Song.** Chord symbols go into bar cells, part markers switch substyles, and Bar Settings can change key, style or tempo at a given bar ([2026 manual ch. 8](https://www.pgmusic.com/manuals/bbw2026full/chapter8.htm)).
- **Changing style.** "Load Style from StylePicker" swaps the style. A forced-styles option makes "all subsequent songs ... not change the style", so many charts can play in one style.
- **Style anatomy (StyleMaker).** Patterns are recorded per instrument: drums, bass, piano, guitar, strings. A minimal style needs only bass, drums and piano ([2024 manual ch. 13](https://www.pgmusic.com/manuals/bbm2024full/chapter13.htm)).
  - Bass and piano/guitar/strings patterns are filed by **how long the chord lasts (1, 2, 4 or 8 beats)** and recorded in 2-bar chunks.
  - Drum patterns are always one bar.
  - "a" and "b" are the two substyles; columns hold variations.
  - Masks restrict a pattern to particular bars, beats, chord types or Roman numerals (e.g. only on II).
- **RealTracks.** These are audio performances that replace a MIDI instrument inside a style ([2025 manual](https://d.pgmusic.com/manuals/bbm2025full/chapter13.htm)). This is the clearest documented case of a rhythm object built to fit any chord sequence: patterns are picked by chord duration and chord type, then transposed to the actual chord.

### Scaler 2/3
- **Chord sets.** 1,000+ sets of chords, keys and modes "categorized by mood, artist, and genre".
- **Motions.** Scaler 3 regrouped Phrases, Performances, Melodies and Basslines as **Motions** with mood tags, plus an Arrange page with lanes for chords, melody, bass and phrases.
- **Detection.** Key and scale detection from MIDI or audio ([Plugin Boutique](https://www.pluginboutique.com/products/14479-Scaler-3-Upgrade)).
- **Voicing and feel.** Voice grouping and drop voicings; humanise varies velocity and timing, and a review adds swing ([Scaler 2 product page](https://scalermusic.com/products/scal), [Producer Sphere](https://producersphere.com/plugin-boutique-scaler-2-review)).
- **Reception.** Reviewers call it good for beginners but note it is "not the graphical piano roll". A chord set and its Motion are chosen independently.

### Captain Plugins (Chords / Deep / Beat / Melody)
- **Harmony.** Pick key and scale, then a progression from presets written in numerals, e.g. "68 of the best chord progressions ... all in numeral form, so they can be transposed" ([how-to guide](https://mixedinkey.com/captain-plugins/how-to-guide/captain-chords/)).
- **Shared rhythm library.** Deep (bass) and Melody "draw on the same library of Rhythm patterns as Captain Chords" ([MusicRadar](https://musicradar.com/reviews/mixed-in-key-captain-plugins)). Rhythm segments can be split down to 8th notes ([Mixed In Key](https://mixedinkey.com/?p=550)).
- **Captain Beat.** Adds swing, quantise, and half-time / double-time ([Mixed In Key](https://mixedinkey.com/?p=577)).
- **Criticism.** MusicTech found the rhythm selection "a little restrictive, as you have to cycle through to find a pattern" ([MusicTech](https://www.musictech.com/reviews/mixed-in-key-captain-plugins-2-0-review-the-composing-assistant-you-need/)). This is the finite-menu problem the owner wants to avoid.

### Logic Pro Drummer and Session Players
- **Drummer.** Choose a drummer type (Acoustic / Electronic / Percussion), then a drummer persona by genre ([iPad guide](https://support.apple.com/en-my/guide/logicpro-ipad/choose-a-drummer-type-and-drummer-lpip3f568c7f/1.1/ipados/17.0)).
- **XY pad.** One puck sets simple ↔ complex on X and soft ↔ loud on Y ([MusicRadar](https://musicradar.com/tuition/tech/how-to-make-beats-with-logic-pro-xs-drummer-and-drum-kit-designer-584981), [Apple help 10.4](https://help.apple.com/logicpro/mac/10.4/en.lproj/lgcp898c7532.html)).
- **Other Drummer controls.** Fills and Swing knobs, a choice of which kit pieces play, and **Follow**: the kick and snare can track another track's rhythm ([Peachpit](https://peachpit.com/articles/article.aspx?p=2414580)).
- **Session Players (Logic 11).** Bass and Keyboard players follow the Chord Track or region chords ("Follow Region Chords"). Complexity and Intensity sliders work like the XY pad axes, and the Keyboard Player ranges from "simple block chords to ... extended harmony" ([Apple newsroom](https://apple.com/newsroom/2024/05/logic-pro-takes-music-making-to-the-next-level-with-new-ai-features), [iPad Session Player editor](https://support.apple.com/en-my/guide/logicpro-ipad/the-session-player-editor-lpipdc3bba32/2.0/ipados/17.0), [MusicRadar](https://musicradar.com/how-to/logic-session-keyboard-player)).
- **Why it matters for middy.** Rhythm here is a **continuous space**, not a menu. A few named characters set the starting point, and two axes move within it.

### Toontrack EZkeys 2 / EZdrummer 3
- **EZkeys Song Track.** Applies "playing styles from EZkeys MIDI to your own chord sequences in real time" ([Toontrack](https://www.toontrack.com/product/ezkeys-2/)).
- **Bandmate.** Matches grooves to the user's audio or MIDI ([KVR](https://www.kvraudio.com/news/toontrack-releases-ezkeys-2-57839)).
- **Tap2Find (EZdrummer 3).** Searches the groove library by a tapped rhythm; grooves can be mixed per part ([Sweetwater](https://www.sweetwater.com/insync/?p=233119)).
- **Takeaway.** Retrieval by example, not by menu.

### Chordbot
- Progression of 60 chord types with a length per chord, played in 50+ comping styles built from mixable instrument tracks.
- Random generator; MIDI/WAV export ([App Store](https://apps.apple.com/au/app/chordbot/id394754767), [No Treble](https://www.notreble.com/?p=22424)).
- Closest in shape to middy's current model. It has no teaching.

### Ableton Live 12
- **Generators.** Rhythm (drum patterns), Seed, Shape, Stacks (up to 4 chords) and Euclidean ([Sound On Sound](https://www.soundonsound.com/techniques/ableton-live-12-midi-generators)).
- **Groove Pool.** A groove (.agr) is applied non-destructively to any clip ([Live 12 manual: Using Grooves](https://www.ableton.com/en/live-manual/12/using-grooves/)). Its parameters:
  - Base: the grid the groove is measured against.
  - Quantize: straight quantisation applied first.
  - Timing.
  - Random: humanisation.
  - Velocity: −100 to +100.
  - Global Amount: up to 130%.
- **Takeaway.** **Groove (micro-timing and accent) is a third object, separate from both chords and the pattern of onsets.**

### Magenta
- **GrooVAE.** Splits a drum performance into a quantised **score** and a **groove** (timing offsets plus velocities). It was trained on 13.6 h of drummers.
  - Groove (humanise) adds feel to a score.
  - Drumify writes drums for any input rhythm.
  - Tap2Drum lets a non-drummer tap a rhythm ([GrooVAE](https://magenta.tensorflow.org/groovae), [paper](https://arxiv.org/pdf/1905.06118v2.pdf), [Magenta Studio](https://magenta.tensorflow.org/studio/standalone)).
- **Coconet.** Harmonises a melody, trained on 306 Bach chorales ([Coconet](https://magenta.tensorflow.org/coconet), [Google blog](https://blog.google/technology/ai/honoring-js-bach-our-first-ai-powered-doodle/)).
- **Contrast with middy.** These are ML note generators, which middy's ADR-0003 forbids. Their **representation** (score ⟂ groove) is still useful.

### Pedagogy tools
- **Groove Pizza (NYU MusEDLab).** Draws a rhythm as rings ("rhythm necklaces"); joining the active steps makes polygons. It shows swing "graphically on the background grid", supports 8th or 16th swing, and aims at "novice drummers" ([evolution of the Groove Pizza](https://wp.nyu.edu/musedlab/?p=41)).
- **musictheory.net and Musicca.** Teach Roman numerals and chord recognition as separate drills, with a chord player and ear training ([classic.musictheory.net](https://classic.musictheory.net/), [Musicca songwriting](https://www.musicca.com/music-theory-for-songwriters)).
- **What they show.** Pedagogy tools teach harmony and rhythm separately and lean on visuals plus sound.

### Suno / Udio (contrast)
- **Inputs.** Only free-text style prompts. BPM and key in the prompt "is not a guaranteed lock", and chord direction is "flexible unless reinforced" ([layer3labs guide](https://www.layer3labs.io/guides/how-to-write-suno-prompts), [Blake Crosley](https://blakecrosley.com/blog/suno-style-field-style-influence)). These are community sources; I found no official statement **[unverified]**.
- **MIDI and stems.** Availability depends on the plan and I could not verify it **[unverified]**.
- **Explanation.** None. Harmony and rhythm are not separable objects.

## 3. Patterns worth adopting / avoiding for middy

### Adopt

1. **Keep harmony and rhythm as separate objects, as iReal Pro, Band-in-a-Box, Logic and Chordbot all do.** Every tool that lets a user "hear the same chords another way" keeps the progression free of rhythm and the style free of chords, and binds them at render time.
   - For middy that means a `Progression` (key, Roman numerals, chord lengths) and a `Feel` (accompaniment pattern, drum pattern, swing, density), joined by the engine.
   - This supports "few concepts": the learner gets two knobs, "what chords" and "how they're played", and can change one while hearing the other stay fixed. That is ideal for listening-first learning.
   - iReal Pro's complaint that changing style also changed tempo is a warning to keep tempo with the song, not inside the rhythm preset.

2. **Choose patterns by chord duration (Band-in-a-Box StyleMaker).** BiaB files bass and comp patterns by chord length (1, 2, 4, 8 beats) and uses masks for chord type or Roman numeral. That lets a finite pattern set fit any progression deterministically, a near-exact precedent for middy's engine. It also gives an explainable rule: "this bar uses the 2-beat pattern because the chord changes every 2 beats."

3. **Replace the finite rhythm menu with a few named starting points plus continuous axes (Logic Drummer and Session Players).**
   - The axes are simple ↔ complex, soft ↔ loud and swing amount.
   - The LLM can map mood words onto points in that space ("restless" → complex, loud) without writing notes. The engine turns axis values into onsets by deterministic rules, e.g. density thinning or adding off-beats.
   - This meets "rhythm should not be a finite menu" while keeping ADR-0003.
   - Captain Plugins' "cycle through to find a pattern" criticism shows what happens with a menu alone.

4. **Treat groove (micro-timing and accent) as its own small layer (Ableton Groove Pool, Magenta score ⟂ groove).**
   - Swing, humanise and accent can be one `groove` value applied after the pattern.
   - It stays separate from which steps sound, so the UI can say "same rhythm, more swing" and the learner hears exactly that change.
   - Keep it to a single swing control; the Groove Pool's five parameters are too many for beginners.

5. **Colour by function and pair Roman numerals with chord symbols (Hookpad).** Hookpad always shows numerals alongside chord symbols and colours chords by scale degree or function. Beginners get the abstract and concrete names together, and colour gives a quick sense of the harmony without reading. Hookpad's key-relative storage (`sd`, `beat`, `duration`) also matches middy's Roman-numeral `Progression` and makes key change trivial.

6. **Explain with evidence where possible, not only with rules (Hooktheory's "what usually comes next").** middy cannot ship corpus statistics without new data and an ADR, but the explanation layer can use the same framing deterministically: "V → I is a cadence; this is the most common way to close a phrase." Use domain-model.md definitions.

7. **Show rhythm as geometry (Groove Pizza).** A circular or step view with swing drawn on the grid itself, not only on the notes, makes "rhythm placement" visible at a glance. It also suits phone width better than a wide piano roll.

### Avoid

- **Huge preset catalogues as the main interface** (Scaler's 1,000+ chord sets, Captain's preset cycling, iReal Pro's 50 styles). They work for producers but conflict with "few concepts" and "understanding over output volume." Use a small number of curated pattern families plus axes instead.
- **Mixer-level orchestration** (iReal Pro per-style mixer, BiaB five-instrument styles, Hookpad's 500+ instruments). This is a lot of surface that teaches no theory.
- **Rhythm fields inside the progression** (iReal Pro's style name and tempo live in the chart). This makes "change the feel, keep the chords" leak, so keep the binding outside both objects.
- **Opaque end-to-end generation** (Suno/Udio, Magenta note models). There is nothing to explain and no stable object to vary. This is the contrast case that confirms ADR-0003.
- **Black-box "follow" binding** (Logic Session Players follow chords, but give no explanation of the bass choices). If middy binds rhythm to chords, it should be able to say why: "the bass plays the root on beat 1 and the fifth on beat 3."

### Suggested model shape (for discussion, not a decision)

`Song = { key, tempo, Progression[romanNumeral, beats], Feel{ patternFamily, complexity 0..1, intensity 0..1, swing 0..1 } }`

- The engine renders `Timeline = render(Progression, Feel)`.
- Patterns are chosen by chord length (BiaB-style) and thinned or filled by complexity (Logic-style).
- Swing is applied last as groove (Ableton/Magenta-style).
- The LLM only fills these fields.
- The UI shows two cards, "Chords" and "Feel", each with its own explanation lane.
- This would need an ADR before adoption.
