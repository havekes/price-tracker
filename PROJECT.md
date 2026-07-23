# Receipt Parsing & Price Tracking App — High-Level Implementation Plan

## Phase 1: Foundation & Project Workspace Setup
*Goal: Establish the base repository structure, runtime environments, and communication boundaries between services.*

* **Task 1.1: Monorepo & Backend Workspace Setup**
  * **Objective:** Scaffold a Go backend service and local workspace.
  * **Scope & Requirements:** Must support configuration management (environment variables for tokens/endpoints) and an HTTP framework.
  * **Agent Output:** Executable Go service with routing scaffolding and environment loading.

* **Task 1.2: Frontend Workspace & Design System Setup**
  * **Objective:** Scaffold a SvelteKit application using Tailwind CSS and `shadcn-svelte`.
  * **Scope & Requirements:** Ensure development proxying to the backend API to prevent CORS issues. Configure `shadcn-svelte` with base themes and component distribution setup.
  * **Agent Output:** Bootstrapped SvelteKit app with a working `shadcn-svelte` initialization and API dev proxy.

---

## Phase 2: Core Data Architecture & Persistence
*Goal: Design and instantiate the relational schema and data access layers.*

* **Task 2.1: Relational Schema Design**
  * **Objective:** Model the relational SQLite database to handle normalized receipts, product variants, price history, and cross-marketplace linking.
  * **Scope & Requirements:** Must support `Correspondent`, `Receipt`, `Product` (canonical display item), `RawItem` (verbatim receipt string), `PriceRecord`, and `MarketplaceLink`. Include unit normalization tracking (e.g., price per kg/unit).
  * **Agent Output:** Fully annotated `schema.sql` file.

* **Task 2.2: Persistence & Query Access Layer**
  * **Objective:** Implement code generation or query execution tooling for database operations.
  * **Scope & Requirements:** Generate type-safe database code (e.g., via `sqlc` or a lightweight query builder) and auto-apply migrations on backend start.
  * **Agent Output:** Type-safe database package providing complete CRUD capabilities for all entities.

---

## Phase 3: External Integration Pipelines
*Goal: Ingest raw documents from Paperless and process them through the Vision LLM.*

* **Task 3.1: Paperless Document Ingestion Client**
  * **Objective:** Build a client to query the Paperless API for relevant receipts and retrieve associated metadata and binaries.
  * **Scope & Requirements:** Search by tag/correspondent, extract creation dates, and fetch raw document files (PDF/Image).
  * **Agent Output:** Go module capable of polling/fetching document payloads and correspondent metadata from Paperless.

* **Task 3.2: Vision LLM Ingestion & Extraction Engine**
  * **Objective:** Send receipt image binaries to the OpenAI-compatible endpoint at `https://ai.havek.es/api` and extract structured line items.
  * **Scope & Requirements:** Design system prompts enforcing strict JSON output (extracting item names, total prices, and raw quantity strings).
  * **Agent Output:** Parsing engine that transforms image inputs into validated Go struct representations of receipt line items.

* **Task 3.3: Ingestion Transaction Orchestration & Normalization**
  * **Objective:** Transform raw Vision LLM JSON output into normalized database records.
  * **Scope & Requirements:** Atomic execution. Must automatically register new `RawItem`s and `Product`s if unseen, calculate standardized unit prices based on base units, and write to `PriceRecord`.
  * **Agent Output:** Orchestrator function taking a raw receipt and safely persisting all entities within a single database transaction.

---

## Phase 4: Backend REST API Layer
*Goal: Expose endpoints for the frontend application to consume data, handle direct uploads, and trigger manual actions.*

* **Task 4.1: Product & Price Analytics Endpoints**
  * **Objective:** Build API routes for browsing product catalogs and retrieving historical price trends.
  * **Scope & Requirements:**
    * `GET /api/products`: List products with latest average prices.
    * `GET /api/products/{id}`: Fetch product details, linked store items, and historical price records grouped by correspondent.
  * **Agent Output:** API handlers returning structured JSON for product lists and chart-ready historical data.

* **Task 4.2: Product Management & Marketplace Linking Endpoints**
  * **Objective:** Provide API routes to mutate product display attributes and associate identical items across stores.
  * **Scope & Requirements:**
    * `PUT /api/products/{id}`: Update `display_name` and standard `base_unit`.
    * `POST /api/products/link`: Create bidirectional relationships between products in `MarketplaceLink`.
    * `POST /api/sync`: Trigger the Paperless & Vision processing pipeline asynchronously.
  * **Agent Output:** Validated mutation endpoints returning updated entity states.

* **Task 4.3: Direct Receipt Upload Endpoint**
  * **Objective:** Provide a multipart endpoint for direct image ingestion bypassing Paperless.
  * **Scope & Requirements:**
    * `POST /api/receipts/upload`: Accept raw image payloads (JPEG/PNG/WebP), user-selected correspondent metadata, and purchase date. Pass payload directly through the Vision LLM processing pipeline.
  * **Agent Output:** Endpoint returning immediate processing status and extracted line items.

---

## Phase 5: Frontend UI & Interactive Analytics
*Goal: Deliver a user interface for tracking prices, editing item metadata, and cross-store analysis.*

* **Task 5.1: Global Navigation & Product Dashboard Page**
  * **Objective:** Build the primary overview page listing canonical products with filtering and search capabilities.
  * **Scope & Requirements:** Utilize `shadcn-svelte` components (e.g., DataTable or Card layouts) to present item names, latest recorded unit prices, and store indicators.
  * **Agent Output:** Responsive dashboard view with client-side/server-side search and filtering.

* **Task 5.2: Interactive Price History & Analytics Detail Page**
  * **Objective:** Construct a detailed view for an individual product featuring a historical price trend chart.
  * **Scope & Requirements:** Render a multi-series line chart (using a Svelte charting library) charting price changes over time, color-coded by marketplace/correspondent. Show historical raw receipt item entries below.
  * **Agent Output:** Interactive visualization page tracking normalized price changes across time and vendors.

* **Task 5.3: Product Metadata & Marketplace Linking UI Components**
  * **Objective:** Build modals/drawers allowing users to edit display details and establish store-to-store product linkages.
  * **Scope & Requirements:** 
    * Modal/Sheet for updating `display_name` and defining quantity ratios (e.g., converting item quantity to base unit kg/liter).
    * Searchable Combobox interface to search other marketplace items and link them together.
  * **Agent Output:** Integrated `shadcn-svelte` Form and Combobox components wired to API mutation endpoints.

---

## Phase 6: PWA Capabilities & Direct Capture Ingestion
*Goal: Transform the SvelteKit frontend into an installable Web App capable of capturing receipts directly via device camera or share target.*

* **Task 6.1: Service Worker & Web App Manifest Configuration**
  * **Objective:** Configure Web App Manifest (`manifest.webmanifest`) and service worker offline caching for mobile/desktop PWA installation.
  * **Scope & Requirements:** 
    * Set app metadata, icons, theme colors, display modes (`standalone`), and asset caching strategies via `@vite-pwa/sveltekit` or standard Workbox implementation.
    * Implement OS Share Target registration to allow sharing images directly into the PWA from native gallery or camera applications.
  * **Agent Output:** Installable PWA setup with manifest and active service worker handling offline routing and share targets.

* **Task 6.2: In-App Camera & Quick-Upload Interface**
  * **Objective:** Create a dedicated mobile-friendly receipt upload workflow using camera input or native file access.
  * **Scope & Requirements:** 
    * Build an interface featuring camera capture using HTML5 MediaDevices / `<input type="file" capture="environment">`.
    * Include a quick form to pick/add the correspondent (e.g., Costco) and purchase date before submitting.
    * Provide real-time upload progress and a preview sheet showing extracted items as soon as the backend processes the image.
  * **Agent Output:** Native-feeling receipt scanner view integrated into the frontend layout and wired to `POST /api/receipts/upload`.
