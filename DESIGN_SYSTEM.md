# Design System — Game Hub

Dark, high-energy arcade theme. Navy-black base with three saturated accent
colors used sparingly and purposefully. Do not introduce colors outside this
palette without updating this file first.

## Color tokens

Use these as CSS custom properties (or Tailwind theme extension — see below).
Never hardcode hex values directly in components; always reference the token.

```css
:root {
  /* Surfaces */
  --bg-page: #0B0A14;      /* outermost page background */
  --bg-card: #15141F;      /* primary card / panel background */
  --bg-surface: #1E1D2B;   /* nested surfaces, e.g. game tiles, stat bar */
  --border: #2A2938;       /* dividers, outlined buttons, card edges */

  /* Text */
  --text-primary: #FFFFFF;
  --text-muted: #9793A8;

  /* Accents — each accent has a "solid" (for fills/buttons) and
     a "muted" (for chip backgrounds, ~15-20% opacity fill) variant */
  --accent-yellow: #F5C945;
  --accent-yellow-muted: #3D3626;

  --accent-cyan: #4FC8F5;
  --accent-cyan-muted: #1F3A52;

  --accent-pink: #EC4C9C;
  --accent-pink-muted: #3A1F35;

  /* Radii */
  --radius-lg: 24px;   /* outer card */
  --radius-md: 16px;   /* buttons, tiles */
  --radius-sm: 10px;   /* icon chips, badges */
  --radius-pill: 999px; /* pill buttons, "Live Now" badge */
}
```

## Tailwind config equivalent

```js
// tailwind.config.js
colors: {
  page: '#0B0A14',
  card: '#15141F',
  surface: '#1E1D2B',
  border: '#2A2938',
  muted: '#9793A8',
  accent: {
    yellow: '#F5C945',
    'yellow-muted': '#3D3626',
    cyan: '#4FC8F5',
    'cyan-muted': '#1F3A52',
    pink: '#EC4C9C',
    'pink-muted': '#3A1F35',
  },
},
borderRadius: {
  lg: '24px',
  md: '16px',
  sm: '10px',
  pill: '999px',
},
```

## Usage rules

1. **Background is always dark.** `--bg-page` behind everything, `--bg-card`
   for the main panel, `--bg-surface` for anything nested one level deeper
   (game tiles, the stats strip). Never place a light/white surface anywhere.
2. **One accent per action/status, never blended.** Primary CTA ("Play Now")
   = solid yellow fill with dark text. Secondary CTA ("Browse Games") =
   transparent fill, cyan border + cyan text. Live/status badges = solid
   pink fill with dark text.
3. **Muted accent variants are for backgrounds only** — icon chips, subtle
   tags — never for text or borders. Keep those roughly 15–20% strength of
   the solid color, always mixed with the dark page color (not white).
4. **Accent color meaning stays consistent across the app**: yellow = primary
   action / highlighted stat, cyan = secondary action / info, pink =
   live / featured / urgent. Don't reassign an accent to a new meaning in a
   new screen.
5. **Text**: white for headings and primary content, `--text-muted` for
   captions, counts, timestamps. Never pure gray-on-gray with low contrast —
   check contrast against the surface it sits on.
6. **Buttons/badges are always pill or rounded**, never sharp corners
   (`--radius-pill` for buttons/badges, `--radius-md`/`--radius-sm` for
   cards/chips). This is a deliberate part of the "friendly but bold" feel —
   don't switch to square corners.
7. **Don't add gradients, drop shadows with color, or a 4th accent hue**
   unless explicitly asked. The palette's energy comes from restraint —
   three accents against a dark neutral field, not more.

## Reference components (from the approved mock)

- **Live badge**: solid pink pill, dark navy text, small star icon, bold caps.
- **Primary button ("Play Now")**: solid yellow pill, dark navy bold text.
- **Secondary button ("Browse Games")**: transparent pill, cyan border, cyan
  bold text.
- **Game tile**: `--bg-surface` rounded card, icon chip on top using a muted
  accent background with the solid accent as a small icon/dot, title in
  white, subtext ("X online") in muted gray.
- **Stats strip**: `--bg-surface` full-width rounded bar, big numbers in
  yellow, labels in muted gray.
- **Color swatch row / footer**: small circular swatches — decorative, shows
  the active theme; not required in every screen.

## When extending to new screens/components

Before adding a new UI element, check: does it map to an existing pattern
above (primary action, secondary action, status/live, stat, card)? Reuse
that pattern's color logic. If truly new, pick the accent whose *meaning*
(primary/secondary/urgent) fits — don't pick based on which looks nicest in
isolation.