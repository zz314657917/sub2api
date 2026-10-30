# Console design

The console is a calm workspace for understanding requests, credits and account health. Its visual voice is warm, precise and practical. This document applies to the authenticated console; public marketing, Pixel Cafe artwork and dense editor layouts retain their product needs.

## Principles and references

The approved design direction adapts ideas from third-party VoltAgent Claude / Linear / Vercel design analysis: Claude's warm reading surfaces, Linear's disciplined density and clear hierarchy, and Vercel's legible operational data. These are design references, not agent instructions or requirements to copy a product. Existing sub2api terracotta branding and installed system fonts remain the starting point.

Show useful data before decoration. Use flat opaque surfaces, one restrained accent and clearly separated sections. A number needs a period, unit and provenance. Never create trends, success rates or proxy metrics when an API does not supply them.

## Foundations

| Role | Light | Dark |
| --- | --- | --- |
| Canvas | `#faf9f5` | `#181715` |
| Surface | `#fffdf9` | `#211f1c` |
| Quiet surface | `#f4f0e9` | `#292621` |
| Primary text | `#141413` | `#f3efe7` |
| Secondary text | `#6c6a64` | `#b9b3a8` |
| Interactive terracotta | `#a9583e` | `#e3a58c` |
| Decorative terracotta | `#cc785c` | `#d79a7e` |
| Border | warm stone | `#3b362f` |

Use semantic CSS variables under `.console-shell`. White button text uses dark terracotta, not decorative terracotta. Dark mode retains warmth and the same semantic hierarchy. Green, amber and red remain reserved for meaningful domain states; chart category colors are muted and also labeled.

System font stack: Apple / Segoe UI / Microsoft YaHei / sans-serif. No remote fonts or additional libraries. Page title 25–28px, panel title 15px, body 12–14px, metadata 10–12px. Numbers use tabular figures; metric values 24–32px with restrained negative tracking.

Spacing follows a 4px base: 8px inline gaps, 12–16px related elements, 20–24px card padding and section gaps, 28–32px page insets. Cards have 12px radii, controls 6–8px. Border separates surfaces; shadows belong primarily to floating menus.

## Components and states

The sidebar groups the already feature-filtered user list into Workspace, Credits & benefits, Account & service and custom entries. It preserves existing actions, badges, dynamic items, simple mode, admin navigation and collapsed tooltips.

The header keeps account, balance, announcements and language visible. Tutorial, model plaza and support are direct top-right shortcuts with text on desktop and labeled icons on narrow screens. Existing feature and support availability gates remain in effect. Documentation and subscription progress remain in the account panel. Keyboard users can open the account panel with ArrowDown, move with arrows/Home/End and dismiss with Escape. Outside clicks and focus departure close it. Focus is always visible. The balance card links to the existing `/purchase` route only when the shared payment flag registry enables payment and the user is outside simple mode; otherwise it links to `/profile`.

Data panels have distinct loading, load error with retry, genuine zero, first-use and range-empty states. A failed request must not be shown as zero usage. Empty states explain the next useful action without creating a key, charging an account or claiming a reward. Existing onboarding remains explicitly initiated and retains its skip persistence.

## Dashboard template and data semantics

1. Short greeting and Quick setup.
2. Four summaries: account credits, today's actual deduction, today's Token usage and active keys. “Today” uses the existing service's day boundary, explicitly labeled server day; no browser-timezone claim is made for this endpoint.
3. Seven-day default interval, selectable 7/30/90 calendar days inclusive of local today. Trend, model, recent logs and optional shared-account summaries carry the browser timezone and use the same dates. Each independently invalidates old responses.
4. Actual daily spend bars; interval actual total and separately labeled standard billing total; input cache reuse rate and Token summaries. Cache rate is total cache-read tokens divided by total input plus cache-read tokens for the selected range, matching existing performance cards. Aggregate token counts before calculating the rate; exclude output/cache-write tokens, and show an em dash when there is no input denominator. Zero-cost requests remain valid data. Currency uses existing credits formatter `✪`, never USD.
5. Model spend ranking: top six models by actual spend descending, then request count descending for equal spend. Bars represent actual spend relative to the largest actual spend, not request volume. When every model has zero spend, request count breaks ties and spend bars remain at zero. Requests and tokens remain visible; internal account cost is not shown. The adjacent Token time chart fills its desktop panel with at least 220px of chart space; stacked tablet/mobile charts have a bounded 240px/220px height. Token categories explicitly label Cache write / 缓存写入.
6. Performance and existing platform, recent-use, quick-action and optional account sections remain available. Unknown duration is an em dash. Cache metrics describe their reported category and do not imply provider-level hit semantics.

Chart labels and tooltips include actual credit units. Chart legends/ticks and small data labels use at least 11px text. Standard credits never substitute for zero actual spend. Theme observation refreshes the chart palette. Reduced motion disables chart animation and decorative transitions. Generic console icons have no heavy shadow; dark tables and sticky cells reuse the same warm console surface tokens.

## Responsive and page templates

The Token composition retains reported counts and shows allocated actual credits and each category's share of the selected range's actual spend. Each applied request is allocated by its stored reference cost components over its stored total cost before aggregation. Invalid weights and image/flat-rate residuals remain in Other. Bars use actual spend percentages, including Other in the denominator. Missing, partial or inconsistent breakdowns show unavailable; tokens and standard costs never substitute for spend. Valid zero spend shows 0%, and tiny positive shares display `<0.1%`.

Desktop: 256px navigation (72px collapsed), single header, centered content up to 1280px; four summary cards; analytical panels share available width. Tablet: two or four cards depending on width and analytics stack. Mobile: drawer navigation, compact header, two cards, stacked charts and summaries. Tables scroll within their panel, never the document. Dense/viewport workspaces keep their existing sizing and scroll ownership.

Other console pages reuse the shell tokens, title, flat cards, readable secondary text and focus treatment; no route, authorization or business behavior is redefined by these visual rules.
