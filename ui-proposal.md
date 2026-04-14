# Allele UI Re-Architecture: Pro Financial Terminal

## Problem Statement
The current UI for the Allele trading engine is considered a "lazy effort" and too "cyberpunk/cheesy". It fails to provide a professional, highly-dense, utilitarian interface expected of a quantitative trading platform. Furthermore, the UI is statically laid out, with plugin configurations haphazardly dumped at the bottom of the screen. It does not reflect the highly modular nature of the Allele backend kernel.

## Requirements
1.  **Aesthetic Shift**: Pivot to a "Pro Financial Terminal" aesthetic (e.g., Bloomberg Terminal, TradingView). Dense, tabular data, high information density, focus on charts, numbers, and logs without excessive padding or neon styling.
2.  **True Modularity**: The UI must be as modular as the kernel. Users should be able to adjust screen real estate dynamically, opening, closing, and docking the components they want to see.
3.  **Plugin Configuration Handling**: Plugin configuration must not be a persistent block on the main screen.
    *   There must be an "in-your-face setup wizard" if critical plugin config is missing.
    *   Otherwise, it should be a pull-up or modal settings menu that gets out of the way of the data.
4.  **Backend Synergy**: If the backend Plugin API or Kernel needs to change to support a better UI (e.g., exposing configuration schemas dynamically), this must be planned.

## Open Questions for the Panel
1.  **Layout Engine**: What is the recommended React-based layout framework to achieve a true docking/windowing system (like FlexLayout-React, GoldenLayout, or React-Grid-Layout) that feels like a professional terminal?
2.  **Plugin Schema API**: How should the backend kernel communicate plugin configuration requirements (schemas) to the UI so the UI can auto-generate the "setup wizard" forms?
3.  **State Management**: How should we persist the user's custom layout and plugin configurations locally or via the engine?
4.  **Information Theory & UX**: How do we organize the core components (Radar/Firehose, Causality Trace, Charts) to maximize usability and information density without overwhelming the user?