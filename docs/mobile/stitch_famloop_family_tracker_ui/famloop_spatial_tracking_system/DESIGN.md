---
name: FamLoop Spatial Tracking System
colors:
  surface: '#f8f9ff'
  surface-dim: '#cbdbf5'
  surface-bright: '#f8f9ff'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#eff4ff'
  surface-container: '#e5eeff'
  surface-container-high: '#dce9ff'
  surface-container-highest: '#d3e4fe'
  on-surface: '#0b1c30'
  on-surface-variant: '#434655'
  inverse-surface: '#213145'
  inverse-on-surface: '#eaf1ff'
  outline: '#737686'
  outline-variant: '#c3c6d7'
  surface-tint: '#0053db'
  primary: '#004ac6'
  on-primary: '#ffffff'
  primary-container: '#2563eb'
  on-primary-container: '#eeefff'
  inverse-primary: '#b4c5ff'
  secondary: '#00687a'
  on-secondary: '#ffffff'
  secondary-container: '#57dffe'
  on-secondary-container: '#006172'
  tertiary: '#006242'
  on-tertiary: '#ffffff'
  tertiary-container: '#007d55'
  on-tertiary-container: '#bdffdb'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#dbe1ff'
  primary-fixed-dim: '#b4c5ff'
  on-primary-fixed: '#00174b'
  on-primary-fixed-variant: '#003ea8'
  secondary-fixed: '#acedff'
  secondary-fixed-dim: '#4cd7f6'
  on-secondary-fixed: '#001f26'
  on-secondary-fixed-variant: '#004e5c'
  tertiary-fixed: '#6ffbbe'
  tertiary-fixed-dim: '#4edea3'
  on-tertiary-fixed: '#002113'
  on-tertiary-fixed-variant: '#005236'
  background: '#f8f9ff'
  on-background: '#0b1c30'
  surface-variant: '#d3e4fe'
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 44px
    fontWeight: '700'
    lineHeight: 52px
    letterSpacing: -0.02em
  display-lg-mobile:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '700'
    lineHeight: 40px
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '600'
    lineHeight: 40px
    letterSpacing: -0.015em
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 26px
    fontWeight: '600'
    lineHeight: 34px
    letterSpacing: -0.015em
  headline-md:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
    letterSpacing: -0.01em
  headline-sm:
    fontFamily: Inter
    fontSize: 20px
    fontWeight: '600'
    lineHeight: 28px
  title-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '600'
    lineHeight: 24px
  title-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '600'
    lineHeight: 24px
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  body-sm:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '400'
    lineHeight: 16px
  label-lg:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '600'
    lineHeight: 20px
    letterSpacing: 0.01em
  label-md:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '600'
    lineHeight: 16px
    letterSpacing: 0.02em
  label-sm:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: '700'
    lineHeight: 16px
    letterSpacing: 0.04em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  gutter: 1rem
  gutter-tablet: 1.5rem
  gutter-desktop: 2rem
  margin: 1rem
  margin-tablet: 1.5rem
  margin-desktop: 2rem
  space-xs: 0.25rem
  space-sm: 0.5rem
  space-md: 1rem
  space-lg: 1.5rem
  space-xl: 2rem
---

## Brand & Style

The design system establishes a high-reliability, protective interface tailored for Android 14+ family spatial awareness. The visual direction merges Material 3 foundational engineering with modern atmospheric glass and high-precision telemetry accents. 

### Brand Personality & Core Principles
- **Vigilant & Reassuring:** Prioritizes immediate legibility, zero latency perception, and psychological peace of mind for guardians and family circles.
- **Precision Telemetry:** Utilizes technical scanlines, micro-radar pings, and refined geographic indicators rather than generic social mapping iconography.
- **Battery-Conscious Utility:** Interfaces default to cool slate foundations with high-efficiency contrast ratios, minimizing GPU draw while maintaining state clarity.

### Visual Style
The design language is an evolution of **Material 3 Adaptive Glass**:
- Core surfaces use crisp architectural whites layered above a subtle Slate-50 substrate.
- Dynamic telemetry details feature bright, directional cyan accents against deep royal blue structural foundations.
- Header bands and floating context cards incorporate translucent frosted-glass physics with hardware-accelerated background blur (`backdrop-filter: blur(16px)`), allowing background live map tiles to show through without interfering with touch-target legibility.

## Colors

The palette balances authoritative navigational anchors with vibrant real-time sensor highlights and explicit status signals.

### Palette Architecture
- **Primary (`#2563eb` Royal Blue / Deep `#1d4ed8`):** The structural anchor. Governs primary navigation shells, active tracking states, circle selectors, and dominant interactive buttons.
- **Secondary (`#06b6d4` Electric Cyan / Glow `#22d3ee`):** Telemetry highlight. Used exclusively for active geofence perimeters, GPS live breadcrumb sweeps, real-time pulse animations, and map radar rings.
- **Tertiary (`#10b981` Emerald Safety):** The safety status token. Denotes safe arrival verification, home zones, optimal device battery levels (>30%), and secure connection pings.
- **Neutral (`#64748b` Slate & `#0f172a` Deep Charcoal):** Provides high-contrast typography, structural boundary outlines, and balanced metadata labeling.
- **Surface Foundations:** Grounded in `#f8fafc` (Slate-50) for outer canvas backgrounds and viewport viewports, with elevated sheet containers cast in pure `#ffffff` (White).

### Semantic Alert Tokens
- **Critical Alert (`#ef4444`):** SOS broadcasts, low battery warnings (<15%), and unauthorized zone exits.
- **Warning (`#f59e0b`):** Degraded GPS signal, network lag, and vehicle speed advisories.

## Typography

Typography is set exclusively in **Inter** to ensure maximum glyph clarity under adverse lighting conditions, small viewports, and moving vehicle context.

### Hierarchical Usage
- **Telemetry Headings (`headline-lg`, `headline-md`):** Used for family member names, active address resolutions, and critical status headlines. High-density letter spacing reduces horizontal sprawl across narrow mobile canvases.
- **Body Telemetry (`body-md`, `body-sm`):** Handles distance metrics ("1.2 mi away"), time stamps, battery percentages, and geofence boundary rules.
- **Status Badges & Labels (`label-lg`, `label-md`, `label-sm`):** Rendered with uppercase styling or increased font weight to communicate real-time states such as "LIVE", "DRIVING", "ARRIVED", and "CHARGING".

## Layout & Spacing

The layout is built upon an **8dp baseline grid** native to Android specifications, ensuring predictable vertical rhythm and crisp edge rendering.

### Touch Target System
- Every interactive element strictly conforms to an accessible touch target of at least **48dp × 48dp** (3rem). Visual representations (such as a 36dp avatar ping button) preserve transparent outer bounds to guarantee hit-testing ergonomics.

### Screen Adaptability
- **Mobile (<600dp):** Fluid edge-to-edge layout with full-viewport interactive map layers. Modals and circle listings reside inside an expandable Material bottom sheet (`margin: 1rem`).
- **Tablet & Foldable (600dp - 1024dp):** Two-pane view. Persistent left rail (360dp width) houses family member rosters, history scrubbers, and live statuses; the remaining canvas hosts map telemetry (`gutter-tablet: 1.5rem`).
- **Desktop / Large Dashboard (>1024dp):** Three-tier layout comprising a slim icon rail, an information pane, and a primary command center viewport (`gutter-desktop: 2rem`).

## Elevation & Depth

Depth in this system merges Material 3 tonal elevation with Android 14 window blur capabilities. Rather than relying on heavy drop shadows, surfaces rely on crisp borders, ambient shadows, and translucent frosted backdrops.

### Elevation Hierarchy
- **Level 0 (Map Base / Canvas Floor):** Background Slate-50 (`#f8fafc`) or dynamic map rendering plane. Zero elevation.
- **Level 1 (Surface Cards & Lists):** Pure white (`#ffffff`) surfaces with a subtle border (`1px solid rgba(226, 232, 240, 0.8)`) and an ambient shadow: `box-shadow: 0 1px 3px 0 rgba(15, 23, 42, 0.05), 0 1px 2px -1px rgba(15, 23, 42, 0.05)`.
- **Level 2 (Hovering Controls & Floating Action Badges):** Pure white elements hovering directly over dynamic maps. `box-shadow: 0 4px 6px -1px rgba(15, 23, 42, 0.07), 0 2px 4px -2px rgba(15, 23, 42, 0.05)`.
- **Level 3 (BottomSheet & Context Drawers):** Elevated sheet components casting deep soft diffusion: `box-shadow: 0 10px 15px -3px rgba(15, 23, 42, 0.08), 0 4px 6px -4px rgba(15, 23, 42, 0.04)`.
- **Level 4 (Frosted Glass Headers & Sticky Overlays):** Semi-transparent background (`rgba(255, 255, 255, 0.82)`) paired with `backdrop-filter: blur(16px)` and an ambient bottom hairline (`1px solid rgba(226, 232, 240, 0.6)`). Creates contextual focus while maintaining visual continuity with underlying map markers.

## Shapes

The design system incorporates geometric roundedness calibrated between **16px and 24px**, aligning with modern Android hardware displays and Material You container styling.

### Component Shape Rules
- **Standard Cards, Bottom Sheets & Modals:** Built with `rounded-xl` (24px / 1.5rem) corner curvature, complementing physical handset bezel curves.
- **Interactive Buttons & Chips:** Standardized at `rounded-lg` (16px / 1rem) to maintain distinct visual weight against background cards.
- **Status Badges, Map Bubbles & Avatars:** Fully rounded pill-shapes (`border-radius: 9999px`) to distinguish micro-telemetry tokens from structural container components.

## Components

### Buttons
- **Primary Action (SOS, Share Location):** Deep Royal Blue (`#2563eb`) fill, white text, height 48px, horizontal padding 24px, radius 16px. Active states shift to `#1d4ed8` with a subtle transform scale (`0.98`).
- **Scanner/Ping Action:** White surface with an active Cyan glow border (`1.5px solid #06b6d4`), Cyan text (`#0891b2`), incorporating a 12px pulsing micro-indicator icon.
- **Touch Target:** All button layouts enforce a 48dp minimum vertical and horizontal bounding box regardless of inner content sizing.

### Chips & Badges
- **Status Chips:** Height 32px, pill-shaped (`9999px`), internal horizontal padding 12px.
- **Battery Status Indicator:** Displays an Emerald Green (`#10b981`) outline and matching icon when safe (>30%), switching to Warning Amber (`#f59e0b`) or Critical Red (`#ef4444`) with accompanying bold percentage labels (`label-md`).

### Cards & Telemetry Pods
- **Member Location Card:** Elevated white container (`radius: 20px`), padded at 16px. Left side displays a 48px family avatar wrapped in an online status ring; center displays member identity (`title-md`) and geocoded location (`body-sm`); right side houses direct navigation and ping actions.
- **Geofence Zone Card:** Inactive zones display muted slate outlines; active entry triggers a 2px inner border highlight in Electric Cyan (`#06b6d4`).

### Input Fields
- Built with a subtle Slate-100 fill (`#f1f5f9`), height 52px, corner radius 16px, and 1px border (`#cbd5e1`).
- Focus states trigger a 2px ring in Royal Blue (`#2563eb`) with an ambient zero-spread cyan glow.

### Map Telemetry & Radar Pings
- **User Pin:** 40px circular container with a high-contrast white rim (3px), centered profile thumbnail, and an animated concentric pulse ring expanding via `#06b6d4` opacity falloff.
- **Breadcrumb Tracker:** 4px polyline rendered in `#2563eb` with neon cyan dots spaced at chronological sync points.