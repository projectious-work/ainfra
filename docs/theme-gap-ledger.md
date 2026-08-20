# Theme integration gap ledger

Record only gaps confirmed while integrating the public API of
`brand-theme-hugo-vanilla`. Each entry must describe the attempted idiomatic
route, the local fallback, and a reusable upstream improvement.

The confirmed gaps are reported upstream in
[`brand-theme-hugo-vanilla#51`](https://github.com/projectious-work/brand-theme-hugo-vanilla/issues/51).

## Data-driven status badges

## Consumer project identity

- **Need:** Replace the reference site's projectious.work mark and wordmark
  with the consuming project's ainfra identity.
- **Public API attempted:** Site title, description, and static logo mounts.
- **Gap:** The header partial hard-codes both the theme icon paths and the
  `projectious.work` wordmark; no documented parameter selects a consumer
  mark, dark-mode mark, or display name.
- **Local fallback:** Override the public `brand.html` partial and shadow its
  two static icon paths with the existing ainfra light and dark marks.
- **Suggested upstream improvement:** Add documented `brand.name`,
  `brand.logoLight`, and `brand.logoDark` parameters while retaining the
  current reference-site defaults.

## Data-driven status badges

- **Need:** Render status badges inside a shortcode-generated roadmap sourced
  from `hugo.Data`.
- **Public API attempted:** The theme's `badge` content shortcode.
- **Gap:** Content shortcodes are available to Markdown authors, but no public
  badge partial or documented template interface accepts a label and variant
  from a project layout.
- **Local fallback:** The roadmap emits a semantic status span and styles it
  only with the theme's public semantic tokens in `assets/css/site.css`.
- **Suggested upstream improvement:** Publish a `badge.html` partial with a
  documented argument dictionary and have the shortcode delegate to it.

## Compact data-driven roadmap

- **Need:** Render grouped, reverse-chronological roadmap data with compact
  cards, accessible status text, and development-note links.
- **Public API attempted:** Cards and badges are content shortcodes; the theme
  has no documented data-driven timeline or roadmap template interface.
- **Gap:** A site layout cannot compose those content shortcodes directly from
  structured Hugo data without rendering private component markup.
- **Local fallback:** A project shortcode emits semantic sections and lists;
  its CSS uses only the public spacing, color, type, radius, and surface tokens.
- **Suggested upstream improvement:** Provide a public roadmap/timeline partial
  or document a generic data-driven card-list template interface.

## Tailwind executable allowlist

- **Need:** Build the documented Tailwind-enabled configuration with Hugo's
  default security policy.
- **Public API attempted:** The published installation and configuration
  examples with `params.build.tailwind = true`.
- **Gap:** Hugo 0.165 refuses the `tailwindcss` transform unless the consuming
  site explicitly adds it to `security.exec.allow`; the v0.3.3 consumer example
  does not document this requirement.
- **Local fallback:** Add the narrowly anchored `^tailwindcss$` executable to
  the site security allowlist.
- **Suggested upstream improvement:** Include the minimum security block in the
  getting-started and configuration examples.

## Compact mobile header

- **Need:** Keep the documentation header within a 412 px mobile viewport.
- **Public API attempted:** The theme's standard header, brand, search, and
  tool configuration without overriding its partials.
- **Gap:** At the v0.3.3 documentation breakpoint, the burger, full wordmark,
  and three 44 px tools exceed the viewport and create horizontal scrolling.
- **Local fallback:** Hide only the wordmark text below 30 rem while retaining
  the linked brand mark and its accessible link name.
- **Suggested upstream improvement:** Add an idiomatic compact-brand header
  mode, or make the wordmark collapse automatically when the available inline
  space cannot accommodate all header controls.

## Single active item for section-owned header links

- **Need:** Keep Quick start, Templates, and Roadmap as convenient top-level
  links while marking only Documentation active for every page in `docs`.
- **Public API attempted:** Hugo menu `pageRef` entries and the theme's public
  `menu.html` partial.
- **Gap:** The partial combines `IsMenuCurrent` and `HasMenuCurrent`, so a
  documentation child can mark both its direct shortcut and Documentation.
  No parameter controls which top-level entry owns the section's active state.
- **Local fallback:** Override only `menu.html` and resolve every `docs` page to
  the Documentation item while retaining the theme's URL helper and markup.
- **Suggested upstream improvement:** Support an `activeSection` menu parameter
  or a site-level map from section to its owning top-level menu entry.

## Release-selector label

- **Need:** Present the native version-selector interaction as a Releases
  control while preserving the actual current version in its entries and the
  site footer.
- **Public API attempted:** `params.version`, `params.versions`, and the public
  `version-menu.html` partial.
- **Gap:** The trigger text is always the current version; changing it also
  changes the site's version identity. There is no independent control label.
- **Local fallback:** Override the public partial with the same structure and
  URL behavior, adding `params.releaseMenuLabel` only for the trigger and menu
  heading.
- **Suggested upstream improvement:** Add an optional `versionMenuLabel` or
  `releaseMenuLabel` parameter that defaults to the current version.

## Changelog entry metadata badges

- **Need:** Show phase and release badges for every generated changelog entry.
- **Public API attempted:** The native changelog list layout and public badge
  classes.
- **Gap:** The list layout renders only its hard-coded Latest status and does
  not expose frontmatter-driven badges.
- **Local fallback:** Extend the public list layout with a `badges` frontmatter
  array while reusing the theme's unchanged changelog and badge markup.
- **Suggested upstream improvement:** Render an optional, documented `badges`
  array in the native changelog list and single-page layouts.
