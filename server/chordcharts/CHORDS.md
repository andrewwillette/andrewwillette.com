# Chord Chart Format (`.chord`)

Chord charts live in `server/chordcharts/charts/` as `.chord` files. They are
written in Roman numerals, so the site can show any chart in any key. The key
picker on the chart page converts the numerals to letter chords in the
browser.

Run `andrewwillettedotcom sync-chord-charts` to validate every file and upload
the charts to S3. It reads from `--dir`, else the `CHORD_CHARTS_DIR` config
value, else `server/chordcharts/charts` relative to the current directory. If any file fails validation, the command uploads nothing.

Run `andrewwillettedotcom format-chord-charts` to line up the bars. It reads
the same directory as `sync-chord-charts`. With `--check`, it changes nothing,
lists the files that need formatting, and exits with status 1 if there are any.
See [Formatting](#formatting).

## File layout

A `.chord` file has a header, a `---` separator line, and then the chart body:

```
Title: Soldier's Joy
Artist: Traditional (Benton Flippen)
Key: D
---
A part:
I     |  I     |  I     |  V     |
I     |  I     |  V     |  I     |

B part:
I     |  IV    |  I     |  V     |
I     |  IV    |  I  V  |  I     |
```

### File name

- Must end in `.chord`.
- The name without the extension becomes the URL slug, for example
  `soldiers_joy.chord` → `/chord-charts/soldiers_joy`. It may contain only
  lowercase letters, digits, and underscores.

### Header

Each header line is `Field: value`. Field names are case-insensitive, and
blank lines in the header are ignored.

| Field    | Required | Notes                                                                 |
|----------|----------|-----------------------------------------------------------------------|
| `Title`  | yes      | The title shown on the list page and the chart page.                  |
| `Artist` | no       | Credit or source, such as `Traditional (Benton Flippen)`.             |
| `Key`    | yes      | The default key. It must match `[A-G]`, plus an optional `#` or `b`, plus an optional `m`, such as `D`, `Bb`, `F#`, or `Am`. |
| `Time`   | no       | The time signature for the whole chart, such as `4/4`, `3/4`, or `6/8`. Leaving it out means 4/4, and no time is shown. Single bars can override it. See [Time signatures](#time-signatures). |
| `Notes`  | no       | Free text shown under the title. It can span several lines, and it must be the last header field. See [Notes](#notes). |
| `Draft`  | no       | `true` or `false`. A draft is not validated or uploaded by `sync-chord-charts`. See [Drafts](#drafts). |

An unknown header field or a line with no `:` is an error. The first line
whose only content is `---` ends the header.

### Drafts

Put `Draft: true` in the header to keep a chart off the site while you work
on it:

```
Draft: true
Title: Lady Be Good
Key: G
---
```

`sync-chord-charts` still formats a draft, but it doesn't validate or upload
it, and lists it under "Skipped drafts". A draft that fails validation
doesn't stop the other charts from syncing. Because the sync deletes S3
charts that it doesn't upload, marking a published chart as a draft takes it
off the site. Remove the line or change it to `false` to publish the chart.

Put the `Draft` line first. The header is read top to bottom, so if a header
line above `Draft` is malformed, the file is reported as an error instead of
skipped. A file with no `---` line is always an error.

### Notes

`Notes:` starts a block of free text that appears on the chart page, under
the title and above the key picker. The notes run from `Notes:` to the `---`
line, so `Notes` must be the last field in the header:

```
Title: Salt River
Key: A
Notes: Play it AABB.

  Kick off with the last four bars of the B part.
---
```

- Text on the `Notes:` line is optional; the notes can start on the next line.
- Line breaks, blank lines, and indentation are kept. Blank lines at the
  start and end, and trailing spaces, are removed.
- Every line in the block is part of the notes, even one that looks like a
  header field. In the example, a `Key: A` line placed after `Notes:` would
  show up as notes, and the chart would have no key.
- The notes are plain text. The key picker doesn't convert numerals in them.

### Body

Every line in the body must be one of the following:

1. **A blank line.**
2. **A section label.** This is any line ending in `:`, such as `A part:`,
   `B part:`, or `Verse:`. The label appears as plain text. A label can also
   change the key for its section, such as `B (-> IV):`. See
   [Key changes](#key-changes).
3. **A chord line.** Every whitespace-separated token on the line must be
   a bar (`|`, `||`, `|||`, …), a chord token (see below), or a
   [time signature](#time-signatures).

Any other line fails validation. This includes lyrics, comments, and a chord
line containing even one unrecognized token.

The body is displayed in a `<pre>` block, so spacing is kept exactly. Line up
the bars with spaces so the measures form columns. To put two chords in one
measure, write both between the bars (`|  I  V7  |`).

### Formatting

`format-chord-charts` rewrites every chord line in a file so that:

- The bars line up in columns across all the chord lines in the file. Each
  column is as wide as its widest measure.
- There is one space between each bar and the chord after it, and one space
  between chords in the same measure.
- The first measure starts at the left edge, and trailing spaces are removed.

Each chord gets room for its longest letter spelling, a two-character root
plus its quality (`I` → `Bb`, `ii7` → `Bbm7`), so the bars stay lined up
after the key picker converts numerals to letters. The header, section labels,
blank lines, and any line that isn't a valid chord line are left exactly as
they are. Formatting a formatted file changes nothing.

```
Before:                               After:
I     |  I IV  |  I     |  I      |   I  | I IV  | I  | I      |
V |   V |  IV    |  2/4 IV | I |      V  | V     | IV | 2/4 IV | I  |
```

If one bar run in a column is longer than another (`|` and `||`), the shorter
one is padded, which leaves an extra space before the chord after it.

## Time signatures

A chart can set a time signature for the whole chart and override it for any
single bar. Both are optional. A chart with neither is treated as 4/4.

### Syntax

A time signature is a number from 1 to 99, a `/`, and a denominator of `1`,
`2`, `4`, `8`, or `16`, with no spaces:

| Valid                       | Invalid                                        |
|-----------------------------|------------------------------------------------|
| `4/4`, `3/4`, `2/4`, `5/4`  | `3/5` (the denominator must be a power of two, up to 16) |
| `6/8`, `9/8`, `12/8`        | `0/4`, `100/4` (the top number must be 1–99)   |
| `2/2`, `7/16`               | `3 / 4`, `(3/4)`, `C`, `waltz`                 |

### The whole chart: the `Time` header

```
Title: Angel's Waltz
Artist: Bobby Hicks
Key: G
Time: 3/4
---
```

The page shows it as "Time: 3/4" under the artist. If you leave the field out,
the chart is treated as 4/4 and no time is shown. Write `Time: 4/4` if you
want 4/4 shown.

### A single bar: an inline time signature

Write the time signature as the first token of a measure to change the meter
for that one bar. This is for crooked tunes, which mostly stay in one meter
but have the odd short or long bar:

```
Title: Rebecca
Artist: Herschel Sizemore
Key: B
Time: 4/4
---
A Part:
I     |  I IV  |  I     |  I      |
V     |  V     |  IV    |  2/4 IV |  I     |  I     |
```

The inline time signature applies **only to the bar it starts**. The next bar
goes back to the chart's `Time`, or 4/4 if there is none. In the example,
only the `IV` bar is in 2/4, and the two `I` bars after it are in 4/4 again.
To change the meter for several bars in a row, mark each bar:

```
I     |  3/4 IV  |  3/4 IV  |  I     |
```

#### Where an inline time signature may go

It must be the first token in its measure, so it must come either at the
start of the line or right after a bar (`|`, `||`, …):

| Line                     | Valid? | Why                                           |
|--------------------------|--------|-----------------------------------------------|
| `I  |  2/4 V  |  I  |`   | yes    | It starts the second measure.                 |
| `3/4 I  |  I  |`         | yes    | It starts the line, so it starts the first measure. |
| `I  2/4 V  |`            | no     | It comes after a chord in the same measure.   |
| `|  2/4 3/4 V  |`        | no     | A measure can have only one time signature.   |
| `|  2/4  |`              | yes    | A measure with only a time signature is allowed, though it has no chord to show. |

`sync-chord-charts` rejects a misplaced time signature and names it in the
error, for example `time signature "2/4" must be the first token in its
measure`.

### How it looks on the page

- Inline time signatures are shown in a separate color from the chords.
- The key picker never changes them. Only the chords are converted to
  letters.
- A time signature takes up columns like any other token. Pad the other rows
  so the bars still line up. In the Rebecca example, the first row has
  `I      |` (seven columns) to match `2/4 IV |` below it.

### Deploying

Time signatures are parsed by the server when it renders a page, not by
`sync-chord-charts` alone. If a chart that uses them is synced to S3 while
production is running an older build, that build doesn't recognize `2/4` as a
token. It shows every line containing an inline time signature as plain text:
the numerals on those lines aren't converted by the key picker and aren't
colored. It also doesn't show the `Time` header. Redeploy the site
(`./scripts/deploy-prod-podman.sh`) to fix this.

## Key changes

A section can move to a key related to the chart's key, such as a polka whose
B part goes to the IV. Put `(-> NUMERAL)` in the section label:

```
Title: Example Polka
Key: A
---
A:
IV    |  I     |  V     |  I     |

B (-> IV):
I     |  IV    |  V7    |  I     |
```

Inside that section, numerals count from the new key, so the B part is
written as if it were its own tune in I. In A, the IV is D, so that
section's `I` is D, `IV` is G, and `V7` is A7. With the picker set to G, the B
part is in C instead.

### Scope

The key change lasts **until the next section label**. A label without a
marker goes back to the chart's `Key`. To stay in the new key for several
sections, mark each one:

```
B (-> IV):
...
B ending (-> IV):
...
A:
...            <- back in the chart's key
```

The target is always counted from the chart's `Key`, not from the previous
section. `C (-> V):` after `B (-> IV):` goes to the V of the chart's key.

### Syntax

| Label                  | Valid? | Why                                                      |
|------------------------|--------|----------------------------------------------------------|
| `B (-> IV):`           | yes    |                                                          |
| `B part (->bVII):`     | yes    | Spaces around `->` are optional. `b` and `#` are allowed. |
| `B (-> iv):`           | no     | The new key is always major, so the numeral must be uppercase. |
| `B (-> IV7):`          | no     | The target is a key, not a chord, so it takes no suffix. |
| `B (-> D):`            | no     | Use a numeral, so the change follows the picker.         |
| `B -> IV:`             | no     | The parentheses are required.                            |

`sync-chord-charts` rejects any label containing `->` that isn't written as
`(-> NUMERAL)`.

A key change always lands on a major key. To change to a minor key, mark the
relative major and chart the section from it, with `vi` as home.

### How it looks on the page

In `Numerals` mode the label shows the numeral as written (`B (-> IV):`). In a
letter key, the picker replaces it with the new key's letter (`B (-> D):` in
A).

### Deploying

Like time signatures, key changes are handled when a page renders. A chart
synced to S3 while production runs an older build shows the label as plain
text and spells that section's chords **in the chart's key**, which is wrong.
Redeploy the site before syncing charts that use key changes.

## Chord tokens

A chord token has the form `[accidental]NUMERAL[quality]`, such as `I`, `vi`,
`V7`, `ii7`, `bVII`, or `IVmaj7`.

### Numeral

The numeral is a scale degree, `I` to `VII`, counted from the major scale of
the current key. **Case sets the third:**

- **Uppercase** (`I II III IV V VI VII`) is a **major** chord.
- **Lowercase** (`i ii iii iv v vi vii`) is a **minor** chord.

Mixed case, such as `Iv`, is invalid. Arabic numbers (`1`, `5`) are no longer
accepted.

### Accidental (optional)

| Prefix | Meaning                      | Example in G   |
|--------|------------------------------|----------------|
| (none) | the diatonic degree          | `VI` → `E`     |
| `b`    | lowered one semitone         | `bVII` → `F`   |
| `#`    | raised one semitone          | `#IV` → `C#`   |

Only one accidental is allowed. `bbVII` is not valid.

The accidental and the case are independent. For example, in G the diatonic
`VII` is F#, so an F major chord ("flat seven major") is `bVII`: `b` lowers F#
to F, and the uppercase numeral makes it major (`bvii` would be Fm).

### Quality (optional)

A token can have at most one suffix. It is added after the numeral, and the
numeral's case decides how it is spelled:

| Suffix | Uppercase (major)            | Lowercase (minor)              |
|--------|------------------------------|--------------------------------|
| (none) | `I` → `G`                    | `vi` → `Em`                    |
| `7`    | `V7` → `D7` (dominant 7th)   | `ii7` → `Am7` (minor 7th)      |
| `6`    | `IV6` → `C6`                 | `iv6` → `Cm6`                  |
| `9`    | `V9` → `D9`                  | `ii9` → `Am9`                  |
| `maj7` | `IVmaj7` → `Cmaj7`           | `imaj7` → `Gmmaj7`             |
| `add9` | `Iadd9` → `Gadd9`            | `iadd9` → `Gmadd9`             |
| `dim`  | `#IVdim` → `C#dim`           | `viidim` → `F#dim`             |
| `dim7` | `#IVdim7` → `C#dim7`         | `viidim7` → `F#dim7`           |
| `m7b5` | `VIIm7b5` → `F#m7b5`         | `viim7b5` → `F#m7b5`           |
| `aug`  | `Vaug` → `Daug`              | not allowed                    |
| `sus2` | `Isus2` → `Gsus2`            | not allowed                    |
| `sus4` | `Vsus4` → `Dsus4`            | not allowed                    |

(Examples are in the key of G.)

- `dim`, `dim7`, and `m7b5` already have a minor third, so they are spelled the
  same in either case. Lowercase (`viidim`) is the conventional choice.
- `aug`, `sus2`, and `sus4` don't have a minor third, so they are rejected on a
  lowercase numeral.
- There is no `m` suffix. Use a lowercase numeral instead (`vi`, not `VIm`).

Combined suffixes are not supported, such as `7sus4`, `maj9`, `13`, and slash
chords like `I/III`. To add a suffix, edit `chordTokenRe` in `render.go`, and
list longer alternatives before their prefixes (for example `m7b5` before
`maj7`). If the new suffix is a lowercase case, also update the minor-handling
`switch` in `parseChord`.

### Letter spelling

When a key is selected, `server/static/chordcharts.js` spells the chord from
that key's major scale. It walks the letter names in order, so for example `IV`
in F is `Bb`, not `A#`. The `b` or `#` prefix then lowers or raises that
spelled note. The quality from the table above is added after it. The
`Numerals` option shows each token exactly as written in the file.

## Key picker caveats

For a major-key chart, the picker lists `Numerals` and the 12 major keys `C Db
D Eb E F Gb G Ab A Bb B`. For a minor-key chart (a `Key` ending in `m`, such as
`Gm`), it lists `Numerals` and the 12 minor keys `Cm C#m Dm Ebm Em Fm F#m Gm
G#m Am Bbm Bm`. The chart opens in its `Key` only when that key is in the
list. Otherwise, such as with `F#` or `Dbm`, it opens in `Numerals`.

Numerals are always counted from the major scale of the key's tonic, so chart
a minor-key tune with `i` as home and the other degrees flattened as needed
(`i`, `bIII`, `iv`, `v`, `bVI`, `bVII`, …). In `Gm`, `i` is `Gm`, `bVII` is
`F`, and `v` is `Dm`.
