# Projecta 🚀

**Projecta** is a modern project budgeting, asset composition, and financial resource management web application. It helps individuals and teams track investments, organize composite assets, manage budgets, and calculate real-time financial standing across multiple currencies.

---

## 🌟 Main Features

- **Project & Budget Management**:
  - Create and manage projects easily with customizable base currencies.
  - View real-time project summaries, budget status, and total invested amounts.
  - "Total Invested" accurately aggregates all project investments without double-counting hierarchical assets.

- **Hierarchical Asset Composition (DAG)**:
  - Organize project deliverables, milestones, and expenses as structured **Assets**.
  - Compose assets into parent-child hierarchies (Directed Acyclic Graph) with customizable share percentages (e.g. 100%, 50%).
  - Group existing assets into overarching parent assets directly from the UI.
  - Dynamic valuation rollups: **Direct Cost** (direct investments) + **Rolled-Up Total Cost** (direct + sub-asset share contributions).
  - Built-in cycle detection preventing circular dependencies.

- **Target Budgets, Multi-Currency Progress & Overbudget Alerts**:
  - Set target budgets or expected purchase prices (`target_price`) in any currency.
  - Real-time exchange rate conversion to the project's base currency via the National Bank of Ukraine (NBU) provider.
  - Uncapped budget progress percentage (e.g., `125%`, `1000%`).
  - Dedicated visual overbudget indicators: progress bars and labels switch to **Orange** (`Overbudget` / `Перевитрата`) when exceeding 100%.

- **Investments & Resource Tracking**:
  - Log resource contributions directly or shared across assets.
  - Support for multiple resource types: `MONEY`, `TIME` (labor/sweat equity), and `GOODS` (materials/parts).
  - Flexible **Tagging Taxonomy** (`#rent`, `#hardware`, `#labor`) replacing rigid categories.

- **Project Sharing & Collaboration**:
  - Share projects with team members via unique shareable links.
  - Seamlessly collaborate on budgets, assets, and investments.

- **Multi-Currency Support**:
  - Comprehensive international currency support (USD, EUR, UAH, etc.).
  - Automatic exchange rate updates integrated with the National Bank of Ukraine (NBU) provider.

- **Multi-Language Support (i18n)**:
  - Full localization in **English** (`en`) and **Ukrainian** (`uk`).
  - Seamless language switching powered by Intlayer.

- **Secure Authentication**:
  - Email & password registration and login using JWT tokens.
  - Google OAuth 2.0 single sign-on (SSO) integration.

- **Error Resilience & Modern UI**:
  - React Error Boundaries with localized fallback screens to prevent app crashes.
  - Responsive, dark-mode ready UI built with Tailwind CSS, Lucide icons, and Shadcn UI components.

- **Containerized Architecture**:
  - Ready-to-use Docker & Docker Compose setup with Nginx reverse proxy, PostgreSQL database, and RabbitMQ message broker.

---

## 🛠️ Technology Stack

| Layer | Technologies |
| :--- | :--- |
| **Frontend** | React 19, Vite, React Router v7, Intlayer (i18n), Tailwind CSS, Lucide Icons, Shadcn UI |
| **Backend** | Go 1.26, Go-Kit, Gorilla Mux, PostgreSQL (`pgx`), RabbitMQ, JWT, Google OAuth |
| **Infrastructure** | Docker, Docker Compose, Nginx, GitLab CI/CD |

---

## 📚 Documentation

- [Asset-Centric Resource Model & Composition](docs/assets-and-investments.md): Comprehensive guide to asset composition, DAG rules, cost rollup formulas, multi-currency progress, and investment tracking.
- [Frontend Design Principles & Architecture Guide](docs/frontend-design-principals.md): Architectural guide to layered frontend design, humble views, state machines, and domain models.
- [Web UI Development Guide](web-ui/README.md): Frontend project structure, npm scripts, and development setup.

---

## 🚀 Quick Start

### Prerequisites
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [Node.js](https://nodejs.org/) (v22 LTS or newer) - *optional for local frontend development*
- [Go](https://go.dev/) (v1.26 or newer) - *optional for local backend development*

### Running with Docker Compose (Recommended)

1. Clone the repository:
   ```bash
   git clone https://gitlab.com/massimo-ua/projecta.git
   cd projecta
   ```

2. Start all services using Docker Compose:
   ```bash
   docker compose up -d --build
   ```

3. Open your browser and navigate to:
   - **Web Application**: [http://localhost](http://localhost) (port 80)
   - **Direct Web UI Dev Server**: [http://localhost:5173](http://localhost:5173)
   - **Backend API**: [http://localhost:8000](http://localhost:8000)

---

## 🧪 Testing & Building

### Backend Tests (Go)
To run backend unit tests with code coverage:
```bash
go test -v -coverprofile=coverage.out ./internal/... ./pkg/...
```

### Frontend Build (React/Vite)
To install dependencies and build the web frontend for production:
```bash
cd web-ui
npm install
npm run build
```

---

## 📜 License
This project is licensed under the MIT License.
