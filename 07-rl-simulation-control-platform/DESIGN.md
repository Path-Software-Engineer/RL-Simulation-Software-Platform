# RL Simulation Control Platform design system

## Direction

The Sprint 1 interface is a technical observability console: dark, restrained and legible. It uses
green only for confirmed success, cyan for observed state, amber for warnings and magenta for
policy intent. Color never carries meaning alone.

## Tokens

- Canvas: `#050914`.
- Surface: `#0b1220`.
- Elevated surface: `#111b2e`.
- Border: `#26334a`.
- Primary text: `#f8fafc`.
- Muted text: `#9aa8bd`.
- Observed state: `#38bdf8`.
- Policy: `#c084fc`.
- Success: `#22c55e`.
- Warning: `#f59e0b`.
- Failure: `#fb7185`.
- Typography: Inter-compatible system sans for UI and JetBrains Mono-compatible monospace for IDs,
  coordinates and metrics. No external font request is required.

## Interaction

- Native buttons and form controls with visible `:focus-visible` outlines.
- Hover feedback changes border/color without layout-shifting scale.
- Motion stays between 150 and 240 ms and is removed under `prefers-reduced-motion`.
- Grid cells expose textual labels and coordinates; reward is never communicated by color alone.
- Live updates always retain a manual resync action and a visible connection state.

## Responsive behavior

- Desktop: grid and evidence panel share the main canvas.
- Tablet: evidence moves below the grid while controls remain above the fold.
- Mobile: navigation becomes horizontal, metrics stack and the grid remains scroll-free at 375 px.

## Explicit anti-patterns

- No decorative 3D, glass blur or neon glow that obscures evidence.
- No emoji icons, fake live indicators, fabricated charts or controls without backend support.
- No infinite telemetry lists; every collection is bounded or paginated.

## Sprint 2 training dashboard

- Reward and moving average share one time-series plot; epsilon and loss remain separate because
  their units and interpretation differ.
- Every SVG chart includes a programmatic label, non-color legend and expandable exact data table.
- Summary cards use persisted samples only. Missing series render an explicit empty state.
- Action distribution pairs bar length with label, count and percentage.
- Training progress is a semantic progressbar backed by persisted episode count.
- Responsive acceptance targets 375 px, 768 px, 1024 px and 1440 px with no horizontal page scroll.
- Reduced-motion preferences disable progress and control transitions.

## Sprint 3 world-model rollout viewer

- Expected state uses green and predicted state uses amber, always paired with `E`/`P` labels,
  coordinates and exact table values.
- The rollout is a horizontally scrollable semantic list with stable button geometry and a table
  fallback; selection changes border and background without scale transforms.
- Step error and accumulated error share a series chart because both use Manhattan cell distance.
- Risk cards name the omitted obstacle representation and autoregressive drift instead of implying
  a safe or optimal plan.
- The two comparison grids stay visible side by side at 375 px; higher-level cards stack at the
  680 px and 980 px breakpoints.
- Keyboard focus, reduced motion and non-color meaning remain mandatory.
