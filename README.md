# Projecta 🚀

**Projecta** is a modern project budgeting and financial management web application. It helps individuals and teams track project expenses, organize payments, manage assets, and calculate financial standings across multiple currencies.

---

## 🌟 Main Features

- **Project & Budget Management**:
  - Create and manage projects easily.
  - Set primary project currencies and view live budget summaries.

- **Payments & Expense Tracking**:
  - Log income and expenses with detailed descriptions.
  - Group payments into customizable **Categories** and **Cost Types**.
  - Distinguish between standard expenses, compensatory payments, and incomes.

- **Asset Management**:
  - Track physical and financial assets attached to projects.
  - Monitor asset values and historical records.

- **Multi-Currency Support**:
  - Support for international currencies (USD, EUR, UAH, etc.).
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
| **Frontend** | React 18, Vite, React Router v7, Intlayer (i18n), Tailwind CSS, Lucide Icons, Shadcn UI |
| **Backend** | Go 1.26, Go-Kit, Gorilla Mux, PostgreSQL (`pgx`), RabbitMQ, JWT, Google OAuth |
| **Infrastructure** | Docker, Docker Compose, Nginx, GitLab CI/CD |

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
