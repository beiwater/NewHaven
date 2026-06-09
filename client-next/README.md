# New Haven — Frontend (Next Client)

Food industry economy simulation game frontend, built on the **Base44** low-code platform. The next-generation client for the New Haven MVP.

> **Previous client**: `client/atlas-foods-client/` (React + PixiJS + Go backend)
> **This client**: `client-next/` (React + Base44 SDK, standalone frontend-consumes-Base44 backend)

---

## Tech Stack

| Layer | Technology |
|---|---|
| Framework | React 18 + Vite 6 |
| Language | JavaScript (JSX), TypeScript (`.ts`) |
| Routing | React Router v6 (hashless, nested layouts) |
| Styling | Tailwind CSS v4 + shadcn/ui components |
| Motion | framer-motion |
| Charts | recharts |
| Backend | Base44 Platform (`@base44/sdk`) |
| Auth | Base44 Auth (JWT, OTP, password reset) |
| Data Fetching | TanStack React Query v5 |
| Forms | react-hook-form + zod |
| Payments | Stripe (react-stripe-js) |
| Drag & Drop | @hello-pangea/dnd |
| Rich Text | react-quill, react-markdown |
| PDF | jspdf + html2canvas |

---

## Project Structure

```
client-next/
├── index.html                  # SPA entry point
├── vite.config.js              # Vite + Base44 plugin config
├── tailwind.config.js          # Tailwind theme (amber-based food-game palette)
├── postcss.config.js
├── eslint.config.js
├── jsconfig.json               # @/ → src/ alias, TS strict check
├── components.json             # shadcn/ui component registry
├── package.json
└── src/
    ├── main.jsx                # React root mount
    ├── App.jsx                 # Router setup, auth gate, 17 page routes
    ├── index.css               # CSS vars (light/dark), animations, scrollbar
    │
    ├── api/
    │   └── base44Client.js     # Base44 DB client proxy (auth, entities, integrations)
    │
    ├── lib/
    │   ├── AuthContext.jsx     # Auth provider: login, register, token, session
    │   ├── app-params.js       # URL params → localStorage → env fallback
    │   ├── gameData.js         # Mock game data (resources, buildings, market, etc.)
    │   ├── query-client.js     # TanStack Query client singleton
    │   ├── utils.js            # cn() classname merge helper
    │   └── PageNotFound.jsx    # 404 page
    │
    ├── components/
    │   ├── AuthLayout.jsx           # Auth page wrapper (login/register layout)
    │   ├── ProtectedRoute.jsx       # Auth guard wrapper
    │   ├── UserNotRegisteredError.jsx
    │   ├── GoogleIcon.jsx
    │   ├── ScrollToTop.jsx          # Scroll reset on route change
    │   │
    │   ├── game/                    # Game UI shell
    │   │   ├── GameLayout.jsx       # Shell: TopBar + Sidebar + Outlet + PriceTicker
    │   │   ├── TopBar.jsx           # Player info, XP bar, cash, notifications, settings
    │   │   ├── Sidebar.jsx          # Nav groups: Town, Trade, Growth, Social, Help
    │   │   ├── PriceTicker.jsx      # Scrolling market price ticker bar
    │   │   ├── SettingsDrawer.jsx   # Account/company/lang/audio/danger zone
    │   │   ├── MobileBottomSheet.jsx # Mobile plot/building detail drawer
    │   │   ├── MobileSecondaryNav.jsx
    │   │   ├── BuildingDetailPanel.jsx # Per-building production UI (Farm, Barn, etc.)
    │   │   └── ...
    │   │
    │   └── ui/                      # shadcn/ui primitives (40+ components)
    │       ├── button.jsx, card.jsx, input.jsx, badge.jsx, tabs.jsx, ...
    │       ├── dialog.jsx, sheet.jsx, drawer.jsx, dropdown-menu.jsx, ...
    │       ├── table.jsx, chart.jsx, calendar.jsx, command.jsx, ...
    │       └── ...
    │
    ├── pages/                       # Route pages (~21 pages)
    │   ├── MapPage.jsx              # / — game map with regions & plots
    │   ├── BuildPage.jsx            # /build — building shop with categories
    │   ├── MarketPage.jsx           # /market — price charts, order book, trading
    │   ├── OrdersPage.jsx           # /orders — active order management
    │   ├── CollectionPage.jsx       # /collection — claim produced goods
    │   ├── WarehousePage.jsx        # /warehouse — inventory management
    │   ├── FinancePage.jsx          # /finance — P&L, cash flow, asset breakdown
    │   ├── ResearchPage.jsx         # /research — tech tree upgrades
    │   ├── ChatPage.jsx             # /chat — channels + DMs
    │   ├── MessagesPage.jsx         # /messages — message inbox
    │   ├── ContractPage.jsx         # /contracts — trade agreements
    │   ├── ExecutivesPage.jsx       # /executives — hire & manage staff
    │   ├── LeaderboardPage.jsx      # /leaderboard — player rankings
    │   ├── AchievementsPage.jsx     # /achievements — milestone tracking
    │   ├── WikiPage.jsx             # /wiki — game guides, resource DB, FAQ
    │   ├── SettingsPage.jsx         # /settings — preferences wrapper
    │   ├── PlayerProfilePage.jsx    # /profile — player stats & company
    │   ├── Login.jsx                # /login — email + password + OTP
    │   ├── Register.jsx             # /register — account creation
    │   ├── ForgotPassword.jsx       # password reset request
    │   └── ResetPassword.jsx        # password reset confirmation
    │
    ├── hooks/
    │   └── use-mobile.jsx           # Mobile detection hook
    │
    └── utils/
        └── index.ts                 # createPageUrl() helper
```

---

## Routing & Auth

| Route | Page | Auth |
|---|---|---|
| `/` | MapPage | ✓ |
| `/build` | BuildPage | ✓ |
| `/market` | MarketPage | ✓ |
| `/orders` | OrdersPage | ✓ |
| `/warehouse` | WarehousePage | ✓ |
| `/collection` | CollectionPage | ✓ |
| `/finance` | FinancePage | ✓ |
| `/research` | ResearchPage | ✓ |
| `/executives` | ExecutivesPage | ✓ |
| `/messages` | MessagesPage | ✓ |
| `/contracts` | ContractsPage | ✓ |
| `/chat` | ChatPage | ✓ |
| `/leaderboard` | LeaderboardPage | ✓ |
| `/achievements` | AchievementsPage | ✓ |
| `/wiki` | WikiPage | ✓ |
| `/settings` | SettingsPage | ✓ |
| `/profile` | PlayerProfilePage | ✓ |
| `/login` | Login | ✗ |
| `/register` | Register | ✗ |
| `/forgot-password` | ForgotPassword | ✗ |
| `/reset-password` | ResetPassword | ✗ |

**Auth flow**: `AuthProvider` → Base44 SDK `db.auth.isAuthenticated()` → checks token in URL/storage → renders `ProtectedRoute` or redirects to login.

---

## Game Domain

A food-industry tycoon game with these core systems:

### 🗺️ Map & Regions
- 4 regions: **New Harbor** (trading), **Inland Estate** (farming), **Sandy Coast** (tourism), **Mountain Route** (premium trade)
- Grid of plots in each region; build on available (blue dashed) plots
- Each region has a unique bonus and recommended buildings

### 🏗️ Buildings (12 types)
- **Production**: Farm, Barn
- **Processing**: Mill, Kitchen, Bakery
- **Commerce**: Market Stall, Café, Food Truck, Restaurant, Trading Hub, Shop
- **Storage**: Warehouse
- Each building has its own detail panel with production mechanics

### 🌾 Resources (16 types)
- **Raw**: Wheat, Corn, Milk, Egg, Fish, Apple, Honey, Coffee, Steak, Sugar, Vegetable
- **Processed**: Flour, Dough, Butter, Cheese
- **Finished**: Bread, Cake, Pie, Soup, Cookie
- Resources have icons, colors, and tier-based pricing

### 📈 Market
- Real-time price charts (area, line, depth views)
- Buy/sell order book
- Price change tracking & trend indicators (Rising/Stable/Falling)
- Time ranges: 1h–7d

### 💰 Finance
- Cash flow chart (daily income/expenses)
- Asset breakdown (buildings, inventory, cash)
- P&L summary cards

### 🔬 Research
- Tech tree for unlocking upgrades
- Perks tied to building efficiency

### 👔 Executives
- Hire managers with special bonuses
- Staff assignments to buildings

### 💬 Chat & Messages
- Chat channels (General, Sales, Help)
- Direct messages with NPC bots (Atlas Trading Bot, Nova Market Bot)
- Player-to-player messaging for trade negotiation

### 📚 Wiki
- Resource database (raw/processed/finished, tiers, prices, production chains)
- Building catalog (cost, region, mechanics, upgrades)
- FAQ guides (Getting Started, Market Guide)

---

## Styling System

**Theme**: Amber/warm food-game palette (HSL CSS vars in `index.css`)
- Light mode: warm background (35 30% 95%), caramel primary (25 60% 45%)
- Dark mode: dark background (25 15% 10%)
- Game-specific colors: `game-blue`, `game-green`, `game-yellow`, `game-orange`, `game-red`, `game-purple`
- Font: Nunito (Google Fonts) for heading, body, display

**shadcn/ui**: 40+ Radix-based primitives, customized to the amber theme via CSS vars.

---

## Development

```bash
# Install
npm install

# Set environment
cp .env.example .env.local
# Fill in:
#   VITE_BASE44_APP_ID=your_app_id
#   VITE_BASE44_APP_BASE_URL=your_backend_url

# Run dev server (port 5173)
npm run dev

# Build
npm run build

# Lint
npm run lint

# TypeScript check
npm run typecheck
```

---

## Key Patterns

- **`@/` path alias** → `src/` (configured in `jsconfig.json`, resolved by Vite)
- **Components**: Functional components, hooks, `useMemo` for derived data
- **Game data**: Mock data in `src/lib/gameData.js` — 16 resources, 12 buildings, market prices, finance data, chat messages, etc.
- **UI components**: shadcn/ui generated into `src/components/ui/`, customized via `tailwind.config.js`
- **Auth**: Base44 SDK wrapped in `AuthContext` — handles token extraction from URL, storage, and redirect
- **Maps**: Static grid map with emoji decorations (no PixiJS/Leaflet in this client)
