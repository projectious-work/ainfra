/**
 * Colour-mode control in the navbar.
 *
 * The documentation site puts its light/dark control in the header, between the
 * search field and the GitHub link. Hextra v0.10.0 renders its theme toggle at
 * the foot of the sidebar instead, and its navbar is built from `menu.main`,
 * which handles exactly one special entry type — `search`. There is no
 * configuration and no `custom/` partial that places anything else in the bar.
 *
 * So the rendered control is moved rather than rebuilt. It keeps Hextra's own
 * markup, its click handler and its `data-theme` state; only its position
 * changes, which is why this does not have to know how the toggle works.
 *
 * The alternative is overriding `_partials/navbar.html`, which would copy ~100
 * lines of theme markup into this example and make the whole navbar ours to
 * maintain across theme upgrades — for one control.
 *
 * An earlier version of this file left the sidebar's now-empty sticky strip
 * behind it, visible as a white bar at the bottom of the sidebar. The container
 * is hidden in custom.css, keyed off the same class this script sets, so the
 * strip never appears — with or without JavaScript.
 *
 * Loaded through Hextra's documented `custom/head-end.html` hook. If a future
 * Hextra grows a navbar toggle, delete this file and the CSS block that pairs
 * with it.
 */
(function () {
  "use strict";

  function move() {
    var nav = document.querySelector(".hextra-nav-container nav");
    if (!nav) return false;

    // The sidebar instance, not the footer one: the footer copy is far down the
    // page and is removed from the layout on small screens.
    var toggle =
      document.querySelector(".hextra-sidebar-container .hextra-theme-toggle") ||
      document.querySelector(".hextra-theme-toggle");
    if (!toggle) return false;

    if (toggle.closest(".pj-navbar-theme")) return true; // already moved

    var slot = document.createElement("div");
    slot.className = "pj-navbar-theme";
    slot.appendChild(toggle);

    // Between the search field and the GitHub link, matching the documentation
    // site's order. Falling back to the end of the bar keeps it reachable if
    // either neighbour is absent.
    var github = nav.querySelector('a[href*="github.com"]');
    var search = nav.querySelector(".hextra-search-wrapper");
    if (github) nav.insertBefore(slot, github);
    else if (search && search.nextSibling) nav.insertBefore(slot, search.nextSibling);
    else nav.appendChild(slot);

    document.documentElement.classList.add("pj-navbar-theme-ready");
    return true;
  }

  if (!move()) {
    document.addEventListener("DOMContentLoaded", move, { once: true });
  }
})();
