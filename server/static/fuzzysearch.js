// Fuzzy list filter. Usage: an <input data-fuzzy-list="LIST_ID" data-fuzzy-empty="EMPTY_ID">
// where each child of the list has a data-search attribute (falls back to textContent).
(function () {
    // Fuzzy subsequence match. Returns a score (higher is better) or -1 for no match.
    function fuzzyScore(query, text) {
        var q = query.toLowerCase().replace(/\s+/g, "");
        var t = text.toLowerCase();
        if (!q) return 0;
        var score = 0, ti = 0, prev = -2;
        for (var qi = 0; qi < q.length; qi++) {
            var idx = t.indexOf(q[qi], ti);
            if (idx === -1) return -1;
            score += 1;
            if (idx === prev + 1) score += 5;                          // consecutive chars
            if (idx === 0 || /[\s\-_'(]/.test(t[idx - 1])) score += 3; // word start
            score -= Math.min(idx - ti, 3) * 0.1;                      // small gap penalty
            prev = idx;
            ti = idx + 1;
        }
        if (t.indexOf(query.toLowerCase().trim()) !== -1) score += 10; // exact substring bonus
        return score;
    }

    function attach(input) {
        var list = document.getElementById(input.dataset.fuzzyList);
        var noResults = document.getElementById(input.dataset.fuzzyEmpty);
        if (!list) return;
        var items = Array.prototype.slice.call(list.children);

        input.addEventListener("input", function () {
            var query = input.value.trim();
            var scored = items.map(function (li, i) {
                return { li: li, i: i, s: fuzzyScore(query, li.dataset.search || li.textContent) };
            });
            if (query) {
                scored.sort(function (a, b) { return b.s - a.s || a.i - b.i; });
            }
            var visible = 0;
            scored.forEach(function (x) {
                x.li.hidden = x.s < 0;
                if (x.s >= 0) visible++;
                list.appendChild(x.li);
            });
            if (noResults) noResults.hidden = visible > 0;
        });
    }

    document.querySelectorAll("input[data-fuzzy-list]").forEach(attach);
})();
