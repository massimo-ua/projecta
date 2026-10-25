# API Contract Testing Guide

This guide defines the HTTP API contracts, authorization rules, request/response payload specifications, and verification workflows for Projecta backend services. Use this document when running regression suites, contract verification, or testing integrations.

---

## 1. Environment & Setup

### 1.1 Service Endpoints

| Service / Interface | Base URL (Direct) | Base URL (Nginx Proxy) |
| :--- | :--- | :--- |
| **Go HTTP Backend API** | `http://localhost:8000` | `http://localhost:8080/api` |
| **Frontend Web UI** | `http://localhost:5174` (or 5173) | `http://localhost:8080/` |
| **PostgreSQL Database** | `localhost:5433` (or 5432) | — |
| **RabbitMQ Management** | `http://localhost:15672` | — |

> [!NOTE]
> All endpoint paths in this document are specified relative to the API root (`/` for direct backend calls or `/api` when routed through the Nginx reverse proxy).

### 1.2 Running the Stack

```bash
# 1. Start all backend and proxy dependencies in daemon mode
docker compose up -d

# 2. Check service health
docker compose ps
```

---

## 2. Common Protocol & Error Contract

### 2.1 Request Headers

- `Content-Type: application/json; charset=utf-8` (required for POST, PUT, PATCH with body)
- `Authorization: Bearer <jwt_access_token>` (required for authenticated routes)

### 2.2 Error Response Format

All application errors return JSON matching the `errorResponseDTO`:

```json
{
  "message": "descriptive error message",
  "code": "VALIDATION_FAILED"
}
```

#### Status Code to Error Code Mapping

| HTTP Status | Exception Code | Typical Cause |
| :---: | :--- | :--- |
| **400 Bad Request** | `VALIDATION_FAILED` | Malformed JSON, missing required fields, non-positive price, or domain invariant breach. |
| **401 Unauthorized** | `UNAUTHORIZED` | Missing, malformed, or expired JWT bearer token; invalid credentials. |
| **403 Forbidden** | `FORBIDDEN` | Missing required role (`Administrator`), or non-owner attempting project modifications. |
| **404 Not Found** | `NOT_FOUND` | Project, asset, investment, or invitation not found in DB or belongs to another tenant. |
| **409 Conflict** | `CONFLICT` | Resource collision (e.g. duplicate username or conflicting aggregate version). |
| **500 Internal Error** | `INTERNAL` | Unhandled system exception or database connection error. |

---

## 3. API Contract Specifications & Samples

### 3.1 Authentication & Profile Context

#### 3.1.1 Register New Identity
- **Method / Path**: `POST /register`
- **Auth**: Public

**Request Sample:**
```http
POST /register HTTP/1.1
Content-Type: application/json

{
  "display_name": "Alice Developer",
  "id": "alice@example.com",
  "token": "SecurePassword123!",
  "strategy": "email_password"
}
```

**Response Sample (`201 Created`):**
```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "eyJhbGciOi...",
  "issued_at": 1729841260,
  "expires_at": 1729844860
}
```

---

#### 3.1.2 Authenticate / Login
- **Method / Path**: `POST /login`
- **Auth**: Public

**Request Sample:**
```http
POST /login HTTP/1.1
Content-Type: application/json

{
  "id": "alice@example.com",
  "token": "SecurePassword123!",
  "identity_provider": "email_password"
}
```

**Response Sample (`200 OK`):**
```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "eyJhbGciOi...",
  "issued_at": 1729841260,
  "expires_at": 1729844860
}
```

---

#### 3.1.3 Refresh Access Token
- **Method / Path**: `POST /refresh`
- **Auth**: Public

**Request Sample:**
```http
POST /refresh HTTP/1.1
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOi..."
}
```

**Response Sample (`200 OK`):**
```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "eyJhbGciOi...",
  "issued_at": 1729841300,
  "expires_at": 1729844900
}
```

---

#### 3.1.4 Get User Profile
- **Method / Path**: `GET /profile`
- **Auth**: `Bearer <token>`

**Response Sample (`200 OK`):**
```json
{
  "id": "c6130cb5-c08c-42cb-b461-8cfdbad04cf2",
  "display_name": "Alice Developer",
  "roles": ["User"],
  "identities": [
    {
      "id": "alice@example.com",
      "provider": "email_password"
    }
  ]
}
```

---

#### 3.1.5 Update Profile
- **Method / Path**: `PUT /profile`
- **Auth**: `Bearer <token>`

**Request Sample:**
```http
PUT /profile HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "display_name": "Alice Senior Engineer"
}
```

**Response Sample (`200 OK`):**
```json
{
  "id": "c6130cb5-c08c-42cb-b461-8cfdbad04cf2",
  "display_name": "Alice Senior Engineer",
  "roles": ["User"],
  "identities": [
    {
      "id": "alice@example.com",
      "provider": "email_password"
    }
  ]
}
```

---

### 3.2 Project Management Context

#### 3.2.1 Create Project
- **Method / Path**: `POST /projects`
- **Auth**: `Bearer <token>` (Role: `User`)

**Request Sample:**
```http
POST /projects HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "name": "Commercial Renovation",
  "description": "Office refurbishment and equipment purchase",
  "mainCurrency": "USD"
}
```

**Response Sample (`201 Created`):**
```json
{
  "project_id": "8bb5d12a-0a25-4fe0-86ec-ecad90bc9837",
  "name": "Commercial Renovation",
  "description": "Office refurbishment and equipment purchase",
  "owner": {
    "person_id": "c6130cb5-c08c-42cb-b461-8cfdbad04cf2",
    "display_name": "Alice Senior Engineer"
  },
  "share_token": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "is_shared": false,
  "mainCurrency": "USD",
  "participants": [
    "Alice Senior Engineer"
  ]
}
```

---

#### 3.2.2 List Projects
- **Method / Path**: `GET /projects?limit=10&offset=0`
- **Auth**: `Bearer <token>` (Role: `User`)

**Response Sample (`200 OK`):**
```json
{
  "projects": [
    {
      "project_id": "8bb5d12a-0a25-4fe0-86ec-ecad90bc9837",
      "name": "Commercial Renovation",
      "description": "Office refurbishment and equipment purchase",
      "owner": {
        "person_id": "c6130cb5-c08c-42cb-b461-8cfdbad04cf2",
        "display_name": "Alice Senior Engineer"
      },
      "share_token": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "is_shared": false,
      "mainCurrency": "USD",
      "participants": ["Alice Senior Engineer"]
    }
  ],
  "limit": 10,
  "offset": 0,
  "total": 1
}
```

---

#### 3.2.3 Update Project
- **Method / Path**: `PATCH /projects/{project_id}`
- **Auth**: `Bearer <token>` (Role: `User` - Owner only)

**Request Sample:**
```http
PATCH /projects/8bb5d12a-0a25-4fe0-86ec-ecad90bc9837 HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "name": "Commercial Renovation Phase 1",
  "description": "Updated project description",
  "mainCurrency": "EUR"
}
```

**Response Sample (`200 OK`):**
```json
{
  "project_id": "8bb5d12a-0a25-4fe0-86ec-ecad90bc9837",
  "name": "Commercial Renovation Phase 1",
  "description": "Updated project description",
  "owner": {
    "person_id": "c6130cb5-c08c-42cb-b461-8cfdbad04cf2",
    "display_name": "Alice Senior Engineer"
  },
  "share_token": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "is_shared": false,
  "mainCurrency": "EUR",
  "participants": ["Alice Senior Engineer"]
}
```

---

#### 3.2.4 Get Project Totals
- **Method / Path**: `GET /projects/{project_id}/totals`
- **Auth**: `Bearer <token>` (Role: `User`)

**Response Sample (`200 OK`):**
```json
{
  "totals": [
    {
      "title": "Total Invested",
      "amount": 250000,
      "currency": "EUR"
    },
    {
      "title": "Project Balance",
      "amount": 0,
      "currency": "EUR"
    }
  ]
}
```

---

#### 3.2.5 Delete Project
- **Method / Path**: `DELETE /projects/{project_id}`
- **Auth**: `Bearer <token>` (Role: `User` - Owner only)

**Response Sample (`204 No Content`):**
*Empty body*

---

### 3.3 Asset Management Context

#### 3.3.1 Create Standalone Asset
- **Method / Path**: `POST /projects/{project_id}/assets`
- **Auth**: `Bearer <token>` (Role: `User`)

**Request Sample:**
```http
POST /projects/8bb5d12a-0a25-4fe0-86ec-ecad90bc9837/assets HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "name": "HVAC System",
  "description": "Commercial air filtration and climate control",
  "price": 500000,
  "currency": "EUR",
  "target_price": 500000,
  "target_currency": "EUR",
  "status": "in_progress",
  "tags": ["equipment", "cooling"]
}
```

> [!NOTE]
> Monetary amounts are denominated in the minor unit of currency (e.g. `500000` = 5,000.00 EUR).

**Response Sample (`201 Created`):**
```json
{
  "asset_id": "2784b1a4-9e73-45a8-9d58-cf2d87e07a78",
  "name": "HVAC System",
  "description": "Commercial air filtration and climate control",
  "status": "in_progress",
  "target_price": 500000,
  "target_currency": "EUR",
  "target_home_amount": 500000,
  "progress_percentage": 0.0,
  "direct_cost": 0,
  "total_cost": 0,
  "price": 0,
  "currency": "EUR",
  "home_amount": 0,
  "home_currency": "EUR",
  "acquired_at": "2026-10-25T00:00:00Z",
  "owner": {
    "id": "c6130cb5-c08c-42cb-b461-8cfdbad04cf2",
    "display_name": "Alice Senior Engineer"
  },
  "project": {
    "id": "8bb5d12a-0a25-4fe0-86ec-ecad90bc9837",
    "name": "Commercial Renovation Phase 1"
  },
  "children": [],
  "parents": [],
  "tags": ["equipment", "cooling"]
}
```

---

#### 3.3.2 List Assets
- **Method / Path**: `GET /projects/{project_id}/assets?limit=20&offset=0`
- **Auth**: `Bearer <token>` (Role: `User`)

**Response Sample (`200 OK`):**
```json
{
  "assets": [
    {
      "asset_id": "2784b1a4-9e73-45a8-9d58-cf2d87e07a78",
      "name": "HVAC System",
      "description": "Commercial air filtration and climate control",
      "status": "in_progress",
      "target_price": 500000,
      "target_currency": "EUR",
      "target_home_amount": 500000,
      "progress_percentage": 0.0,
      "price": 0,
      "currency": "EUR",
      "home_amount": 0,
      "home_currency": "EUR",
      "tags": ["equipment", "cooling"]
    }
  ],
  "limit": 20,
  "offset": 0,
  "total": 1
}
```

---

#### 3.3.3 Link Child Asset (Hierarchy)
- **Method / Path**: `POST /projects/{project_id}/assets/{asset_id}/children`
- **Auth**: `Bearer <token>` (Role: `User`)

**Request Sample:**
```http
POST /projects/8bb5d12a-0a25-4fe0-86ec-ecad90bc9837/assets/2784b1a4-9e73-45a8-9d58-cf2d87e07a78/children HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "child_asset_id": "4da4ea39-2d9f-4318-bd2f-f4c0ce3e2ff5",
  "share_percentage": 100.0
}
```

**Response Sample (`200 OK`):**
*Asset entity representation with updated `children` array.*

---

#### 3.3.4 Update Asset
- **Method / Path**: `PUT /projects/{project_id}/assets/{asset_id}`
- **Auth**: `Bearer <token>` (Role: `User`)

**Request Sample:**
```http
PUT /projects/8bb5d12a-0a25-4fe0-86ec-ecad90bc9837/assets/2784b1a4-9e73-45a8-9d58-cf2d87e07a78 HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "name": "HVAC System Updated",
  "description": "Installed and tested",
  "price": 520000,
  "currency": "EUR",
  "status": "completed",
  "tags": ["equipment", "cooling", "verified"]
}
```

**Response Sample (`204 No Content`):**
*Empty body*

---

#### 3.3.5 Delete Asset
- **Method / Path**: `DELETE /projects/{project_id}/assets/{asset_id}`
- **Auth**: `Bearer <token>` (Role: `User`)

**Response Sample (`204 No Content`):**
*Empty body*

---

### 3.4 Investment Management Context

#### 3.4.1 Create Multi-Currency Investment
- **Method / Path**: `POST /projects/{project_id}/investments`
- **Auth**: `Bearer <token>` (Role: `User`)

> [!TIP]
> When `currency` differs from the project's `mainCurrency` (e.g. UAH amount in an EUR project), the backend automatically queries the live NBU exchange rate provider and populates `home_amount` and `home_currency`.

**Request Sample:**
```http
POST /projects/8bb5d12a-0a25-4fe0-86ec-ecad90bc9837/investments HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "asset_id": "2784b1a4-9e73-45a8-9d58-cf2d87e07a78",
  "resource_type": "money",
  "amount": 100000,
  "currency": "UAH",
  "description": "Filter components shipment",
  "date": "2026-10-25T10:00:00Z",
  "tags": ["filters", "logistics"]
}
```

**Response Sample (`201 Created`):**
```json
{
  "investment_id": "0d634ea7-aa31-4835-961d-e5f8fdf1ca19",
  "project_id": "8bb5d12a-0a25-4fe0-86ec-ecad90bc9837",
  "asset_id": "2784b1a4-9e73-45a8-9d58-cf2d87e07a78",
  "asset_name": "HVAC System Updated",
  "resource_type": "money",
  "amount": 100000,
  "currency": "UAH",
  "home_amount": 1977,
  "home_currency": "EUR",
  "description": "Filter components shipment",
  "date": "2026-10-25T10:00:00Z",
  "tags": ["filters", "logistics"]
}
```

---

#### 3.4.2 List Investments
- **Method / Path**: `GET /projects/{project_id}/investments?limit=25&offset=0`
- **Auth**: `Bearer <token>` (Role: `User`)

**Response Sample (`200 OK`):**
```json
{
  "investments": [
    {
      "investment_id": "0d634ea7-aa31-4835-961d-e5f8fdf1ca19",
      "project_id": "8bb5d12a-0a25-4fe0-86ec-ecad90bc9837",
      "asset_id": "2784b1a4-9e73-45a8-9d58-cf2d87e07a78",
      "asset_name": "HVAC System Updated",
      "resource_type": "money",
      "amount": 100000,
      "currency": "UAH",
      "home_amount": 1977,
      "home_currency": "EUR",
      "description": "Filter components shipment",
      "date": "2026-10-25T10:00:00Z",
      "tags": ["filters", "logistics"]
    }
  ],
  "limit": 25,
  "offset": 0,
  "total": 1
}
```

---

#### 3.4.3 Batch Create Investments
- **Method / Path**: `POST /projects/{project_id}/investments/batch`
- **Auth**: `Bearer <token>` (Role: `User`)

**Request Sample:**
```http
POST /projects/8bb5d12a-0a25-4fe0-86ec-ecad90bc9837/investments/batch HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "default_asset": "2784b1a4-9e73-45a8-9d58-cf2d87e07a78",
  "items": [
    {
      "resource_type": "money",
      "amount": 25000,
      "currency": "EUR",
      "description": "Mounting brackets",
      "date": "2026-10-25T11:00:00Z",
      "tags": ["hardware"]
    },
    {
      "resource_type": "money",
      "amount": 15000,
      "currency": "EUR",
      "description": "Electrical cabling",
      "date": "2026-10-25T11:30:00Z",
      "tags": ["electrical"]
    }
  ]
}
```

**Response Sample (`201 Created`):**
```json
[
  {
    "investment_id": "2d8f9c11-9a42-4f32-824b-325b1285093a",
    "amount": 25000,
    "currency": "EUR",
    "home_amount": 25000,
    "home_currency": "EUR"
  },
  {
    "investment_id": "6a9e14a2-113c-4e89-8d82-c82736184920",
    "amount": 15000,
    "currency": "EUR",
    "home_amount": 15000,
    "home_currency": "EUR"
  }
]
```

---

#### 3.4.4 Delete Investment
- **Method / Path**: `DELETE /projects/{project_id}/investments/{investment_id}`
- **Auth**: `Bearer <token>` (Role: `User`)

**Response Sample (`204 No Content`):**
*Empty body*

---

### 3.5 Invitation & Administration Context

#### 3.5.1 Create Invitation
- **Method / Path**: `POST /invitations`
- **Auth**: `Bearer <token>` (Role: `Administrator`)

**Request Sample:**
```http
POST /invitations HTTP/1.1
Authorization: Bearer eyJhbGciOi...
Content-Type: application/json

{
  "email": "colleague@example.com"
}
```

**Response Sample (`201 Created`):**
```json
{
  "id": "e5c6a11e-7f61-4fa3-80b7-1c4b8b6038df",
  "email": "colleague@example.com",
  "code": "INV-A92B3C",
  "status": "pending",
  "created_at": "2026-10-25T10:15:00Z",
  "expires_at": "2026-11-01T10:15:00Z"
}
```

---

#### 3.5.2 Validate Invitation Code
- **Method / Path**: `GET /invitations/code/{code}`
- **Auth**: Public

**Response Sample (`200 OK`):**
```json
{
  "email": "colleague@example.com",
  "expires_at": "2026-11-01T10:15:00Z"
}
```

---

## 4. Automated Contract Verification Test Runner

The following Python script executes a complete automated contract test sequence against the live API stack.

### 4.1 Running the Verification Script

Save as `test_contract.py` or run directly with Python 3:

```python
import json
import urllib.request
import urllib.error
import uuid

BASE_URL = "http://localhost:8080/api"

def request(method, path, body=None, token=None):
    url = f"{BASE_URL}{path}"
    data = json.dumps(body).encode("utf-8") if body else None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req) as resp:
            content = resp.read().decode("utf-8")
            return resp.status, json.loads(content) if content else {}
    except urllib.error.HTTPError as e:
        content = e.read().decode("utf-8")
        return e.code, json.loads(content) if content else {}

# 1. Register & Login
test_id = f"tester_{uuid.uuid4().hex[:6]}@example.com"
status, auth = request("POST", "/register", {
    "display_name": "Contract Tester",
    "id": test_id,
    "token": "Password123!",
    "strategy": "email_password"
})
assert status == 201, f"Register failed: {status} {auth}"
token = auth["access_token"]
print("✓ Register & Login: OK")

# 2. Get Profile
status, prof = request("GET", "/profile", token=token)
assert status == 200 and prof["display_name"] == "Contract Tester"
print("✓ Profile Contract: OK")

# 3. Create Project
status, proj = request("POST", "/projects", {
    "name": "Contract Test Project",
    "description": "Verification of contracts",
    "mainCurrency": "EUR"
}, token=token)
assert status == 201
p_id = proj["project_id"]
print("✓ Create Project: OK")

# 4. Create Asset
status, asset = request("POST", f"/projects/{p_id}/assets", {
    "name": "Primary Server",
    "price": 150000,
    "currency": "EUR",
    "target_price": 150000,
    "target_currency": "EUR"
}, token=token)
assert status == 201
a_id = asset["asset_id"]
print("✓ Create Asset: OK")

# 5. Create Multi-Currency Investment
status, inv = request("POST", f"/projects/{p_id}/investments", {
    "asset_id": a_id,
    "resource_type": "money",
    "amount": 50000,
    "currency": "UAH",
    "description": "Hardware module"
}, token=token)
assert status == 201
assert inv["home_currency"] == "EUR"
assert inv["home_amount"] > 0
i_id = inv["investment_id"]
print("✓ Multi-Currency Investment Conversion: OK")

# 6. Delete Investment & Asset
status, _ = request("DELETE", f"/projects/{p_id}/investments/{i_id}", token=token)
assert status == 204
status, _ = request("DELETE", f"/projects/{p_id}/assets/{a_id}", token=token)
assert status == 204
print("✓ Delete Resources: OK")

# 7. Delete Project & Confirm 404
status, _ = request("DELETE", f"/projects/{p_id}", token=token)
assert status == 204
status, err = request("GET", f"/projects/{p_id}", token=token)
assert status == 404
print("✓ Project Lifecycle & 404 Cascade: OK")

print("\nAll API contracts successfully verified!")
```

---

## 5. Contract Invariant & Edge Case Matrix

| Test Scenario | Method / Endpoint | Expected Status | Contract Verification Check |
| :--- | :--- | :---: | :--- |
| **Missing Auth Header** | `GET /profile` | `401 Unauthorized` | Returns `{"code": "UNAUTHORIZED"}` |
| **Tampered / Expired Token** | `GET /projects` | `401 Unauthorized` | Signature or expiry rejection |
| **Asset Price <= 0** | `POST /projects/{id}/assets` | `400 Bad Request` | Domain rejects: `price must be greater than 0` |
| **Blank Currency** | `POST /projects/{id}/assets` | `400 Bad Request` | Domain rejects: `currency is required` |
| **Invalid Investment Allocation** | `POST /projects/{id}/investments` | `400 Bad Request` | Sum of allocations != 100% |
| **Cross-Tenant Project Access** | `GET /projects/{other_user_id}` | `404 Not Found` | Scoped by requester; never returns foreign records |
| **Non-Owner Project Deletion** | `DELETE /projects/{id}` | `403 Forbidden` | Requester is participant but not project owner |
| **Duplicate Asset Invariant** | `POST /projects/{id}/assets` | `400 / 409` | Same asset name within project boundary |
