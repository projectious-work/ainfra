/**
 * Table-of-contents active marker.
 *
 * Hextra v0.10.0 tracks the active heading with an IntersectionObserver whose
 * root margin is "-20px 0px -80% 0px" — it only reacts while a heading is
 * inside a narrow band near the top of the viewport. That leaves two gaps on a
 * reference page:
 *
 *   - Following a table-of-contents link scrolls the heading to the very top,
 *     above the band, so the observer never fires and the marker stays on
 *     whichever section it was on.
 *   - Scrolling through a section taller than the viewport — a long code block,
 *     a wide table — means no heading is in the band at all, so the marker can
 *     be left pointing at the wrong section or at none.
 *
 * This derives the active entry from scroll position instead: the current
 * section is the last heading that has passed the navbar. There is therefore
 * always exactly one marker, whatever is between the headings.
 *
 * Additive, and loaded through Hextra's documented `custom/head-end.html` hook
 * — no theme file is overridden. Remove it once upstream widens the band.
 */
(function () {
  "use strict";

  // Navbar height (4rem) plus enough clearance that a heading counts as
  // "current" once it is comfortably visible rather than exactly at the edge.
  var OFFSET = 88;

  function init() {
    var toc = document.querySelector(".hextra-toc");
    if (!toc) return;

    var links = Array.prototype.slice.call(toc.querySelectorAll('a[href^="#"]'));
    if (!links.length) return;

    // Document order matters: the scan below stops at the first heading that
    // has not yet passed the navbar.
    var targets = links
      .map(function (link) {
        var id = decodeURIComponent(link.getAttribute("href").slice(1));
        return { link: link, el: document.getElementById(id) };
      })
      .filter(function (t) {
        return t.el;
      });
    if (!targets.length) return;

    var queued = false;

    function apply() {
      queued = false;

      var current = targets[0];
      for (var i = 0; i < targets.length; i++) {
        if (targets[i].el.getBoundingClientRect().top - OFFSET <= 0) {
          current = targets[i];
        } else {
          break;
        }
      }

      // At the end of the page the last section is the one being read, even
      // though its heading may be far above the navbar.
      var atBottom =
        window.innerHeight + window.scrollY >=
        document.documentElement.scrollHeight - 2;
      if (atBottom) current = targets[targets.length - 1];

      // Only write when the state actually differs. The theme's observer is
      // still running and also toggles this class, so this function is
      // re-entered from the MutationObserver below; writing unconditionally
      // would mutate the DOM on every pass and loop.
      for (var j = 0; j < targets.length; j++) {
        var wanted = targets[j] === current;
        if (targets[j].link.classList.contains("hextra-toc-active") !== wanted) {
          targets[j].link.classList.toggle("hextra-toc-active", wanted);
        }
      }
    }

    function schedule() {
      if (queued) return;
      queued = true;
      window.requestAnimationFrame(apply);
    }

    window.addEventListener("scroll", schedule, { passive: true });
    window.addEventListener("resize", schedule, { passive: true });
    window.addEventListener("hashchange", schedule);

    // The theme's observer keeps running and marks its own entry, which briefly
    // leaves two markers lit. Correcting on class mutations makes this script
    // the single authority without having to disable the theme's code.
    if (window.MutationObserver) {
      new MutationObserver(schedule).observe(toc, {
        subtree: true,
        attributes: true,
        attributeFilter: ["class"]
      });
    }

    apply();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
