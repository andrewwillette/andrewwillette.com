(function () {
  const LETTERS = ["C", "D", "E", "F", "G", "A", "B"];
  const NATURAL_SEMITONES = { C: 0, D: 2, E: 4, F: 5, G: 7, A: 9, B: 11 };
  const MAJOR_SCALE_STEPS = [0, 2, 4, 5, 7, 9, 11];
  const KEYS = ["C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"];

  function mod(n, m) {
    return ((n % m) + m) % m;
  }

  // root like "D", "Bb", or "Bbb". A key change can land on a
  // double-flat or double-sharp root, e.g. bIII of Gb is Bbb.
  function parseRoot(root) {
    const letter = root[0];
    let accidental = 0;
    for (const c of root.slice(1)) accidental += c === "b" ? -1 : 1;
    return { letter, accidental };
  }

  // Returns the 7 {letter, accidental} degrees of root's major scale,
  // built by walking the natural-letter cycle rather than a lookup table
  // so e.g. key F's 4th degree spells as Bb, not A#.
  function majorScale(root) {
    const { letter, accidental } = parseRoot(root);
    const rootLetterIdx = LETTERS.indexOf(letter);
    const rootSemitone = mod(NATURAL_SEMITONES[letter] + accidental, 12);
    const scale = [];
    for (let i = 0; i < 7; i++) {
      const stepLetter = LETTERS[mod(rootLetterIdx + i, 7)];
      const target = mod(rootSemitone + MAJOR_SCALE_STEPS[i], 12);
      let acc = target - NATURAL_SEMITONES[stepLetter];
      if (acc > 6) acc -= 12;
      if (acc < -6) acc += 12;
      scale.push({ letter: stepLetter, accidental: acc });
    }
    return scale;
  }

  function spell(letter, accidental) {
    if (accidental === 0) return letter;
    return accidental > 0 ? letter + "#".repeat(accidental) : letter + "b".repeat(-accidental);
  }

  // degreeToken like "4", "b7", "#4"; quality like "m", "maj7", "".
  function toLetterChord(key, degreeToken, quality) {
    const prefix = degreeToken.length > 1 ? degreeToken[0] : "";
    const degree = parseInt(degreeToken[degreeToken.length - 1], 10);
    const { letter, accidental } = majorScale(key)[degree - 1];
    let acc = accidental;
    if (prefix === "b") acc -= 1;
    if (prefix === "#") acc += 1;
    return spell(letter, acc) + quality;
  }

  function applyKey(container, key) {
    const body = container.querySelector(".chord-chart-body");
    // overflow is how many columns the current line has been pushed right by
    // labels that outgrew their cells (e.g. "I " -> "Bb "). Later cells on the
    // same line give those columns back from their padding, so the bars after
    // them return to where the chart put them.
    let overflow = 0;
    body.childNodes.forEach((node) => {
      if (node.nodeType === Node.TEXT_NODE) {
        if (node.textContent.includes("\n")) overflow = 0;
        return;
      }
      if (!node.classList) return;
      // A key change target in a section label, e.g. the "IV" in
      // "B (-> IV):", shown as the new key's letter.
      if (node.classList.contains("modulation")) {
        node.textContent = key === "numerals" ? node.dataset.numeral : toLetterChord(key, node.dataset.degree, "");
        return;
      }
      if (!node.classList.contains("chord")) return;
      const degree = node.dataset.degree;
      const quality = node.dataset.quality || "";
      const numeral = node.dataset.numeral;
      // Chords in a section with a key change count from that degree of
      // the chart's key, e.g. data-tonic="4" spells the section in IV.
      const tonic = node.dataset.tonic;
      const chordKey = tonic && key !== "numerals" ? toLetterChord(key, tonic, "") : key;
      const label = key === "numerals" ? numeral : toLetterChord(chordKey, degree, quality);
      // Pad to the cell width the chart was written with, less any overflow
      // to absorb; keep at least one space if the cell had trailing spaces.
      const width = parseInt(node.dataset.width, 10) || numeral.length;
      const minWidth = width > numeral.length ? label.length + 1 : label.length;
      const padded = Math.max(minWidth, width - overflow);
      overflow += padded - width;
      node.textContent = label.padEnd(padded);
    });
  }

  document.addEventListener("DOMContentLoaded", () => {
    const container = document.getElementById("single-chord-chart-page");
    const select = document.getElementById("key-picker");
    if (!container || !select) return;
    const defaultKey = container.dataset.defaultKey;
    if (defaultKey && KEYS.includes(defaultKey)) select.value = defaultKey;
    applyKey(container, select.value);
    select.addEventListener("change", () => applyKey(container, select.value));
  });
})();
