# Design System — Warm Coin / Bronze-Gold

The locked visual language for a-couple-of-coins. Every page, component, and chart should pull from these tokens — never hardcode hex values, named Tailwind colors (`bg-blue-500`), or font names.

## Brand voice

Warm, characterful, a bit editorial. Closer to "your money" than "the bank's website." Serif headings give the wordmark and key numbers a bit of personality; sans body keeps the data legible.

## Color tokens

All colors live in `src/index.css` as OKLch CSS variables. Use them via shadcn-style Tailwind utilities (`bg-primary`, `text-foreground`, `border-border`) or directly as `var(--token)`.

### Light mode (`:root`)

| Token | OKLch | Role |
|---|---|---|
| `--background` | `0.98 0.01 85` | Cream page background |
| `--foreground` | `0.20 0.02 60` | Espresso body text |
| `--card` | `1 0 0` | Cards / elevated surfaces |
| `--primary` | `0.72 0.15 75` | **Amber** — brand, CTAs, links |
| `--primary-foreground` | `0.20 0.02 60` | Text on amber |
| `--accent` | `0.55 0.10 200` | **Teal** — secondary highlights, income |
| `--destructive` | `0.55 0.18 35` | **Rust** — destructive actions, expenses |
| `--muted` / `--muted-foreground` | `0.95 / 0.50` | Subtle surfaces, secondary text |
| `--border` / `--input` | `0.90 0.01 85` | Hairlines, input borders |
| `--ring` | `0.72 0.15 75` | Focus ring (amber) |
| `--income` | `0.55 0.10 200` | Positive money (= teal) |
| `--expense` | `0.55 0.18 35` | Negative money (= rust) |

### Dark mode (`.dark`)

Backgrounds shift to espresso (`0.20 0.02 60`); foreground inverts to cream. Primary/accent/destructive bump up ~0.06 lightness for contrast on dark surfaces. Border switches to translucent white at 12% opacity.

## Typography

| Family | Use | Tailwind |
|---|---|---|
| **Inter** | Body, UI, labels, table cells | `font-sans` (default) |
| **Fraunces Variable** | Headings, wordmark, hero numerals | `font-serif` |
| System mono | Aligned monetary figures in tables | `font-mono` |

Both fonts are self-hosted via `@fontsource` (see `src/main.tsx`). No external CDN.

Headings (`h1`–`h4`) get Fraunces automatically via `@layer base` in `index.css`. For one-off display use (e.g. a big net-worth number on the dashboard), reach for `font-serif` explicitly.

Letter-spacing is tightened by `-0.01em` on display type, and Fraunces' `ss01` stylistic set is enabled site-wide for slightly friendlier digits.

## Radius

`--radius: 0.5rem` is the canonical value. Tailwind utilities map automatically:
- `rounded-sm` → `0.25rem`
- `rounded-md` → `0.375rem`
- `rounded-lg` → `0.5rem`
- `rounded-xl` → `0.75rem`

Full pills (`rounded-full`) are reserved for avatars and tag chips.

## Semantic color rules

This is the most important section. Misusing these breaks the design system fast.

### Do

- Use `bg-primary` / `text-primary` for **brand and CTAs only**: primary buttons, the wordmark, the active nav link.
- Use `text-income` for positive amounts (deposits, paychecks, gains).
- Use `text-expense` for negative amounts (purchases, fees, losses).
- Use `bg-destructive` / `text-destructive` for **destructive UI** (delete buttons, error states). It happens to be the same hue as `--expense` — that's intentional but they're separate tokens semantically.
- Use `bg-card` for any elevated surface in dashboards / settings.

### Don't

- Don't use amber (`primary`) to mean "income" or "positive." Amber is brand, not data.
- Don't use teal (`accent`) for destructive UI. It's a positive/info color.
- Don't introduce a new green for income, even though it's the finance convention. We chose teal on purpose.
- Don't reach for raw Tailwind palette colors (`text-emerald-500`, `bg-amber-400`, etc.) in app code — they bypass the theme and won't dark-mode-flip.

## Density & spacing

Default content max-width is `max-w-6xl` (~72rem) with `px-6 py-3` on the navbar and `px-6 py-8` on page bodies. Stay on Tailwind's 4px grid (`gap-2`, `gap-4`, `gap-6`, `gap-8`). Reserve `gap-12` and larger for marketing-style pages (Welcome).

## Charts (Recharts)

Chart series should pull from CSS variables, not hardcoded:

```tsx
<Line stroke="var(--primary)" />
<Bar fill="var(--accent)" />
<ReferenceLine stroke="var(--muted-foreground)" />
```

For income/expense pairs, use `var(--income)` and `var(--expense)` directly. Multi-series charts beyond two categories should rotate through `primary → accent → muted-foreground` — if a third hue is genuinely needed, propose it in this doc first rather than inlining a one-off color.

## Adding new tokens

If a new semantic concept appears (e.g. "warning" for budget-over-threshold), add the variable to **both** `:root` and `.dark` blocks in `index.css`, register it in `@theme inline` as `--color-<name>`, and document it here. Never add a new color anywhere else.
