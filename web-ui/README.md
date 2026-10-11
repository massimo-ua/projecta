# Projecta Web UI 💻

The web frontend client for **Projecta**, built with **React 19**, **Vite**, **Tailwind CSS**, and **Intlayer**.

---

## 🏗️ Architecture & Philosophy

The web application follows the layered architecture documented in [Frontend Design Principles & Architecture Guide](../docs/frontend-design-principals.md):

1. **Presentation Layer (`src/components/`, `src/pages/`)**:
   - Pure, humble UI components responsible only for rendering and forwarding user interactions.
   - Built with Tailwind CSS and Radix UI / Shadcn UI components.
2. **State & Orchestration Layer (`src/hooks/`)**:
   - Custom React hooks acting as UI state machines (e.g. `useAssets`, `useInvestments`, `useProjects`).
3. **Domain Layer (`src/models/`)**:
   - Pure JavaScript domain models (`Asset`, `Investment`, `Project`, `User`) encapsulating calculations, statuses, and validations (e.g. `targetProgress`, `isOverBudget`, `totalCost`).
4. **Gateway & Anti-Corruption Layer (`src/api/`)**:
   - API client gateways normalizing backend responses (`toDomain`) and mapping UI payloads (`toPayload`).

---

## 🌟 Key Features & Components

- **Asset Management & Composition**:
  - Tree & flat list view for hierarchical asset structures.
  - Grouping multiple assets into parent assets via modal dialog.
  - Linking and unlinking sub-assets with proportional share weights.
- **Budget & Progress Tracking**:
  - Live budget comparison with automatic currency conversion.
  - Overbudget indicators (orange progress bar and badges when exceeding 100%).
- **Investment Tracking**:
  - Resource investment logs supporting monetary funds, labor/time, and materials/goods.
  - Dynamic tag filtering and tag management.
- **Internationalization (i18n)**:
  - English (`en`) and Ukrainian (`uk`) locales managed through Intlayer content dictionaries (`*.content.js`).

---

## 🛠️ Available Scripts

In the `web-ui` directory, you can run:

### `npm run dev`
Runs the Vite development server with Hot Module Replacement (HMR).
Default address: [http://localhost:5173](http://localhost:5173).

### `npm run build`
Compiles Intlayer dictionary files (`intlayer build`) and builds the production bundle with Vite into the `dist/` directory:
```bash
npm run build
```

### `npm run lint`
Runs ESLint against project source files.

---

## 🐳 Docker Setup

When running within Docker Compose, `web-ui` is mounted with live reloading:
```bash
docker compose up -d web-ui
```
The application is accessible through the Nginx reverse proxy at `http://localhost` (or `http://localhost:8080`) and directly at `http://localhost:5173`.
