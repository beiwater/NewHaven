// Resource icons and colors
export const RESOURCES = {
  wheat: { icon: "🌾", name: "Wheat", color: "text-amber-600" },
  corn: { icon: "🌽", name: "Corn", color: "text-yellow-500" },
  milk: { icon: "🥛", name: "Milk", color: "text-blue-100" },
  egg: { icon: "🥚", name: "Eggs", color: "text-amber-100" },
  flour: { icon: "🫘", name: "Flour", color: "text-amber-200" },
  bread: { icon: "🍞", name: "Bread", color: "text-amber-500" },
  cake: { icon: "🎂", name: "Cake", color: "text-pink-400" },
  fish: { icon: "🐟", name: "Fish", color: "text-blue-400" },
  cheese: { icon: "🧀", name: "Cheese", color: "text-yellow-400" },
  butter: { icon: "🧈", name: "Butter", color: "text-yellow-300" },
  apple: { icon: "🍎", name: "Apple", color: "text-red-400" },
  honey: { icon: "🍯", name: "Honey", color: "text-amber-400" },
  coffee: { icon: "☕", name: "Coffee", color: "text-amber-800" },
  pie: { icon: "🥧", name: "Pie", color: "text-orange-400" },
  soup: { icon: "🍲", name: "Soup", color: "text-green-500" },
  cookie: { icon: "🍪", name: "Cookies", color: "text-amber-600" },
  sugar: { icon: "🍬", name: "Sugar", color: "text-pink-300" },
  steak: { icon: "🥩", name: "Steak", color: "text-red-600" },
  vegetable: { icon: "🥦", name: "Vegetables", color: "text-green-500" },
  dough: { icon: "🫓", name: "Dough", color: "text-amber-200" },
};

export const BUILDINGS = [
  { id: "farm", name: "Farm", icon: "🌾", desc: "Grows wheat & corn", price: 500, produces: ["wheat", "corn"], category: "production" },
  { id: "barn", name: "Barn", icon: "🐄", desc: "Produces milk & eggs", price: 800, produces: ["milk", "egg"], category: "production" },
  { id: "mill", name: "Mill", icon: "⚙️", desc: "Grinds wheat into flour", price: 1200, produces: ["flour"], category: "processing" },
  { id: "kitchen", name: "Kitchen", icon: "🍳", desc: "Cooks soups & pies", price: 1500, produces: ["soup", "pie"], category: "processing" },
  { id: "bakery", name: "Bakery", icon: "🍞", desc: "Bakes bread & cookies", price: 2000, produces: ["bread", "cookie"], category: "processing" },
  { id: "market_stall", name: "Market Stall", icon: "🏪", desc: "Sells basic goods", price: 1000, produces: [], category: "commerce" },
  { id: "cafe", name: "Café", icon: "☕", desc: "Brews coffee & serves cake", price: 3000, produces: ["coffee", "cake"], category: "commerce" },
  { id: "food_truck", name: "Food Truck", icon: "🚚", desc: "Mobile food sales", price: 2500, produces: [], category: "commerce" },
  { id: "restaurant", name: "Restaurant", icon: "🍽️", desc: "High-end dining", price: 5000, produces: [], category: "commerce" },
  { id: "trading_hub", name: "Trading Hub", icon: "⚓", desc: "Import & export goods", price: 4000, produces: [], category: "commerce" },
  { id: "warehouse", name: "Warehouse", icon: "📦", desc: "+200 storage capacity", price: 1500, produces: [], category: "storage" },
  { id: "shop", name: "Shop", icon: "🛒", desc: "Retail storefront", price: 2000, produces: [], category: "commerce" },
];

export const MARKET_DATA = [
  { resource: "wheat", price: 12.5, change: 2.3, high: 14.0, low: 11.0, buyPrice: 12.8, sellPrice: 12.2, inventory: 340 },
  { resource: "corn", price: 8.7, change: -1.1, high: 10.0, low: 7.5, buyPrice: 9.0, sellPrice: 8.4, inventory: 520 },
  { resource: "milk", price: 15.0, change: 0.5, high: 16.0, low: 13.5, buyPrice: 15.3, sellPrice: 14.7, inventory: 180 },
  { resource: "egg", price: 6.2, change: -0.3, high: 7.0, low: 5.5, buyPrice: 6.5, sellPrice: 5.9, inventory: 890 },
  { resource: "flour", price: 22.0, change: 3.1, high: 24.0, low: 19.0, buyPrice: 22.5, sellPrice: 21.5, inventory: 150 },
  { resource: "bread", price: 35.0, change: 1.8, high: 38.0, low: 32.0, buyPrice: 35.5, sellPrice: 34.5, inventory: 95 },
  { resource: "cake", price: 65.0, change: -2.5, high: 70.0, low: 58.0, buyPrice: 66.0, sellPrice: 64.0, inventory: 30 },
  { resource: "fish", price: 18.0, change: 4.2, high: 20.0, low: 15.0, buyPrice: 18.5, sellPrice: 17.5, inventory: 210 },
  { resource: "cheese", price: 28.0, change: 0.9, high: 30.0, low: 25.0, buyPrice: 28.5, sellPrice: 27.5, inventory: 120 },
  { resource: "butter", price: 20.0, change: -0.8, high: 22.0, low: 18.0, buyPrice: 20.5, sellPrice: 19.5, inventory: 200 },
  { resource: "honey", price: 45.0, change: 5.0, high: 48.0, low: 40.0, buyPrice: 45.5, sellPrice: 44.5, inventory: 60 },
  { resource: "coffee", price: 55.0, change: 2.0, high: 58.0, low: 50.0, buyPrice: 55.5, sellPrice: 54.5, inventory: 45 },
  { resource: "pie", price: 42.0, change: -1.5, high: 45.0, low: 38.0, buyPrice: 42.5, sellPrice: 41.5, inventory: 75 },
  { resource: "soup", price: 30.0, change: 1.2, high: 33.0, low: 27.0, buyPrice: 30.5, sellPrice: 29.5, inventory: 110 },
  { resource: "cookie", price: 25.0, change: 3.5, high: 28.0, low: 22.0, buyPrice: 25.5, sellPrice: 24.5, inventory: 160 },
];

export const PLAYER = {
  name: "Captain Mochi",
  avatar: "🧑🍳",
  level: 14,
  xp: 2450,
  xpMax: 3000,
  cash: 24580,
  boosts: ["⚡ 2x Speed", "🍀 Lucky"],
  notifications: 3,
};

export const EXECUTIVES = [
  { id: 1, name: "Chef Bruno", avatar: "👨🍳", role: "Head Chef", salary: 500, prodBonus: 15, salesBonus: 5, mgmtDiscount: 0, level: 3, maxLevel: 10 },
  { id: 2, name: "Trader Lin", avatar: "👩💼", role: "Trade Master", salary: 700, prodBonus: 0, salesBonus: 20, mgmtDiscount: 10, level: 5, maxLevel: 10 },
  { id: 3, name: "Farmer Joe", avatar: "👨🌾", role: "Farm Manager", salary: 400, prodBonus: 25, salesBonus: 0, mgmtDiscount: 5, level: 2, maxLevel: 10 },
  { id: 4, name: "Baker Yuki", avatar: "👩🍳", role: "Pastry Expert", salary: 600, prodBonus: 20, salesBonus: 10, mgmtDiscount: 0, level: 4, maxLevel: 10 },
];

export const RESEARCH_NODES = [
  { id: 1, name: "Crop Rotation", icon: "🌱", desc: "+20% farm yield", cost: { wheat: 50, corn: 30 }, duration: "2h", progress: 100, status: "completed" },
  { id: 2, name: "Cold Storage", icon: "❄️", desc: "+50 warehouse capacity", cost: { flour: 20 }, duration: "4h", progress: 65, status: "in_progress" },
  { id: 3, name: "Artisan Baking", icon: "🥐", desc: "Unlock croissants", cost: { flour: 40, butter: 20, egg: 30 }, duration: "6h", progress: 0, status: "available" },
  { id: 4, name: "Deep Sea Fishing", icon: "🎣", desc: "+30% fish catch rate", cost: { fish: 60 }, duration: "8h", progress: 0, status: "available" },
  { id: 5, name: "Gourmet Recipes", icon: "📖", desc: "Unlock premium dishes", cost: { bread: 30, cheese: 20, honey: 10 }, duration: "12h", progress: 0, status: "locked" },
  { id: 6, name: "Trade Routes", icon: "🗺️", desc: "-15% trade fees", cost: { coffee: 15 }, duration: "10h", progress: 0, status: "locked" },
];

export const ORDERS = [
  { id: 1, resource: "wheat", type: "buy", quantity: 100, remaining: 65, price: 12.0, count: 3 },
  { id: 2, resource: "wheat", type: "sell", quantity: 50, remaining: 50, price: 13.5, count: 1 },
  { id: 3, resource: "bread", type: "buy", quantity: 20, remaining: 12, price: 33.0, count: 2 },
  { id: 4, resource: "fish", type: "sell", quantity: 80, remaining: 30, price: 19.0, count: 4 },
  { id: 5, resource: "honey", type: "buy", quantity: 15, remaining: 15, price: 43.0, count: 1 },
];

export const CHAT_MESSAGES = [
  { id: 1, user: "FarmKing", avatar: "👨🌾", msg: "Anyone selling wheat below 12?", time: "2m ago", channel: "general" },
  { id: 2, user: "BakeQueen", avatar: "👩🍳", msg: "Fresh bread available! 50 units @ 34 each", time: "5m ago", channel: "sales" },
  { id: 3, user: "Captain Mochi", avatar: "🧑🍳", msg: "Looking for a trade partner for fish!", time: "8m ago", channel: "general", isMe: true },
  { id: 4, user: "TradeMaster", avatar: "🏴☠️", msg: "Honey prices going up! Buy now!", time: "12m ago", channel: "sales" },
  { id: 5, user: "NewPlayer42", avatar: "🐣", msg: "How do I build a bakery?", time: "15m ago", channel: "help" },
  { id: 6, user: "FarmKing", avatar: "👨🌾", msg: "You need a mill first, then unlock bakery", time: "14m ago", channel: "help" },
  { id: 7, user: "SeaWolf", avatar: "🐺", msg: "Just caught a rare golden fish! 🎉", time: "20m ago", channel: "general" },
  { id: 8, user: "BakeQueen", avatar: "👩🍳", msg: "[image: cake_showcase.jpg]", time: "22m ago", channel: "general", isImage: true },
];

export const WAREHOUSE_ITEMS = [
  { resource: "wheat", quantity: 340, quality: 92, capacity: 500 },
  { resource: "corn", quantity: 520, quality: 88, capacity: 600 },
  { resource: "milk", quantity: 180, quality: 95, capacity: 300 },
  { resource: "egg", quantity: 890, quality: 90, capacity: 1000 },
  { resource: "flour", quantity: 150, quality: 85, capacity: 400 },
  { resource: "bread", quantity: 95, quality: 97, capacity: 200 },
  { resource: "fish", quantity: 210, quality: 80, capacity: 400 },
  { resource: "honey", quantity: 60, quality: 99, capacity: 100 },
  { resource: "coffee", quantity: 45, quality: 93, capacity: 100 },
  { resource: "cookie", quantity: 160, quality: 91, capacity: 300 },
];

export const FINANCE_DATA = {
  cash: 24580,
  income: 8450,
  expenses: 5230,
  profit: 3220,
  assets: 42000,
  liabilities: 8000,
  equity: 34000,
  transactions: [
    { id: 1, desc: "Sold 50 Bread", amount: 1750, type: "income", time: "1h ago" },
    { id: 2, desc: "Bought 100 Wheat", amount: -1200, type: "expense", time: "2h ago" },
    { id: 3, desc: "Chef Bruno Salary", amount: -500, type: "expense", time: "3h ago" },
    { id: 4, desc: "Sold 30 Cookies", amount: 750, type: "income", time: "4h ago" },
    { id: 5, desc: "Warehouse Rent", amount: -200, type: "expense", time: "5h ago" },
    { id: 6, desc: "Fish Export", amount: 3600, type: "income", time: "6h ago" },
    { id: 7, desc: "Milk Purchase", amount: -900, type: "expense", time: "8h ago" },
    { id: 8, desc: "Cake Sales", amount: 1950, type: "income", time: "12h ago" },
  ],
};

export const MAP_PLOTS = [
  { id: 1, x: 1, y: 1, state: "occupied", building: "farm", level: 3 },
  { id: 2, x: 2, y: 1, state: "occupied", building: "barn", level: 2 },
  { id: 3, x: 3, y: 1, state: "available", building: null, level: 0 },
  { id: 4, x: 4, y: 1, state: "locked", building: null, level: 0 },
  { id: 5, x: 1, y: 2, state: "occupied", building: "mill", level: 1 },
  { id: 6, x: 2, y: 2, state: "occupied", building: "bakery", level: 2 },
  { id: 7, x: 3, y: 2, state: "available", building: null, level: 0 },
  { id: 8, x: 4, y: 2, state: "available", building: null, level: 0 },
  { id: 9, x: 1, y: 3, state: "occupied", building: "warehouse", level: 1 },
  { id: 10, x: 2, y: 3, state: "occupied", building: "market_stall", level: 1 },
  { id: 11, x: 3, y: 3, state: "occupied", building: "cafe", level: 1 },
  { id: 12, x: 4, y: 3, state: "locked", building: null, level: 0 },
  { id: 13, x: 1, y: 4, state: "available", building: null, level: 0 },
  { id: 14, x: 2, y: 4, state: "occupied", building: "trading_hub", level: 1 },
  { id: 15, x: 3, y: 4, state: "locked", building: null, level: 0 },
  { id: 16, x: 4, y: 4, state: "locked", building: null, level: 0 },
];

export const LEADERBOARD = [
  { rank: 1, name: "GoldHarvest", avatar: "👑", level: 28, netWorth: 125000, trend: "up" },
  { rank: 2, name: "SeaTrader", avatar: "⚓", level: 25, netWorth: 98000, trend: "up" },
  { rank: 3, name: "BakeQueen", avatar: "👩🍳", level: 22, netWorth: 87000, trend: "down" },
  { rank: 4, name: "FarmKing", avatar: "👨🌾", level: 20, netWorth: 72000, trend: "up" },
  { rank: 5, name: "Captain Mochi", avatar: "🧑🍳", level: 14, netWorth: 42000, trend: "up", isMe: true },
  { rank: 6, name: "TradeMaster", avatar: "🏴☠️", level: 18, netWorth: 38000, trend: "down" },
  { rank: 7, name: "CookieMonster", avatar: "🍪", level: 16, netWorth: 31000, trend: "up" },
  { rank: 8, name: "SeaWolf", avatar: "🐺", level: 15, netWorth: 28000, trend: "down" },
  { rank: 9, name: "NewPlayer42", avatar: "🐣", level: 5, netWorth: 5000, trend: "up" },
  { rank: 10, name: "HoneyBear", avatar: "🐻", level: 12, netWorth: 22000, trend: "up" },
];