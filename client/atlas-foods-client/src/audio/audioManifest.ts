// Audio manifest — all sound keys mapped to their asset paths.
// Keys are used programmatically: AudioManager.playSfx(key), playMusic(key), etc.
// Missing files degrade silently (warning + no-op).
// Asset root: /assets/audio/

const SFX = '/assets/audio/sfx'
const AMB = '/assets/audio/ambience'
const BGM = '/assets/audio/bgm'

/** Every sound effect key in the game. */
export type SfxKey =
  // ── UI ──
  | 'ui_button_click'
  | 'ui_button_hover'
  | 'ui_confirm'
  | 'ui_cancel'
  | 'ui_error'
  | 'ui_disabled'
  | 'ui_panel_open'
  | 'ui_panel_close'
  | 'ui_tab_switch'
  | 'ui_popup'
  | 'ui_drag_start'
  | 'ui_drag_drop'
  | 'ui_scroll_tick'
  // ── Money / Economy ──
  | 'money_coin_gain'
  | 'money_coin_spend'
  | 'money_big_profit'
  | 'money_loss'
  // ── Market / Contracts ──
  | 'market_buy'
  | 'market_sell'
  | 'market_order_created'
  | 'market_order_filled'
  | 'market_price_up'
  | 'market_price_down'
  | 'market_volatility_alert'
  | 'contract_signed'
  | 'debt_issued'
  | 'futures_open'
  | 'futures_close'
  // ── Building ──
  | 'build_place'
  | 'build_preview_move'
  | 'build_confirm'
  | 'build_construction_start'
  | 'build_construction_complete'
  | 'build_upgrade'
  | 'build_repair'
  | 'build_demolish'
  | 'land_unlock'
  | 'road_place'
  // ── Farm / Barn / Resource ──
  | 'farm_plant_seed'
  | 'farm_water_crop'
  | 'farm_harvest'
  | 'farm_crop_ready'
  | 'barn_animal_feed'
  | 'barn_collect_milk'
  | 'barn_collect_eggs'
  | 'resource_pickup'
  // ── Processing / Kitchen ──
  | 'mill_start'
  | 'mill_complete'
  | 'kitchen_chop'
  | 'kitchen_pan_sizzle'
  | 'kitchen_recipe_complete'
  | 'bakery_oven_start'
  | 'bakery_bread_ready'
  | 'cafe_coffee_pour'
  | 'cafe_espresso'
  | 'food_packaged'
  // ── Inventory / Logistics ──
  | 'inventory_open'
  | 'inventory_sort'
  | 'warehouse_store'
  | 'warehouse_takeout'
  | 'truck_depart'
  | 'truck_arrive'
  | 'ship_dock'
  | 'delivery_complete'
  // ── Restaurant ──
  | 'restaurant_open'
  | 'customer_enter'
  | 'menu_set'
  | 'dish_served'
  | 'dish_sold'
  | 'restaurant_revenue_collect'
  | 'occupancy_up'
  | 'occupancy_down'
  | 'restaurant_level_up'
  // ── Research / Executive ──
  | 'research_start'
  | 'research_progress_tick'
  | 'research_complete'
  | 'tech_unlock'
  | 'executive_hire'
  | 'executive_level_up'
  | 'skill_assign'
  | 'buff_activate'
  | 'buff_expire'
  // ── Quest / Social ──
  | 'quest_accept'
  | 'quest_complete'
  | 'achievement_unlock'
  | 'leaderboard_rank_up'
  | 'chat_send'
  | 'chat_receive'
  | 'player_join'
  | 'player_leave'
  | 'trade_request'
  | 'trade_accepted'
  | 'trade_rejected'
  // ── System ──
  | 'day_start'
  | 'day_end'
  | 'season_change'
  | 'save_success'
  | 'autosave'
  | 'notification_general'
  | 'warning_soft'
  | 'critical_alert'

/** Every BGM/music key. */
export type MusicKey =
  | 'bgm_main_menu'
  | 'bgm_harbor_town'
  | 'bgm_market'
  | 'bgm_restaurant'
  | 'bgm_farm'
  | 'bgm_night'

/** Every ambience key. */
export type AmbienceKey =
  | 'amb_harbor_day'
  | 'amb_harbor_night'
  | 'amb_market'
  | 'amb_restaurant'
  | 'amb_kitchen'
  | 'amb_farm'
  | 'amb_barn'
  | 'amb_warehouse'
  | 'amb_cafe'
  | 'amb_rain_town'

/** Path lookup: SFX key → expected asset file. */
export const SFX_PATHS: Record<SfxKey, string> = {
  // UI
  ui_button_click:       `${SFX}/ui/ui_button_click.wav`,
  ui_button_hover:       `${SFX}/ui/ui_button_hover.wav`,
  ui_confirm:            `${SFX}/ui/ui_confirm.wav`,
  ui_cancel:             `${SFX}/ui/ui_cancel.wav`,
  ui_error:              `${SFX}/ui/ui_error.wav`,
  ui_disabled:           `${SFX}/ui/ui_disabled.wav`,
  ui_panel_open:         `${SFX}/ui/ui_panel_open.wav`,
  ui_panel_close:        `${SFX}/ui/ui_panel_close.wav`,
  ui_tab_switch:         `${SFX}/ui/ui_tab_switch.wav`,
  ui_popup:              `${SFX}/ui/ui_popup.wav`,
  ui_drag_start:         `${SFX}/ui/ui_drag_start.wav`,
  ui_drag_drop:          `${SFX}/ui/ui_drag_drop.wav`,
  ui_scroll_tick:        `${SFX}/ui/ui_scroll_tick.wav`,
  // Money
  money_coin_gain:       `${SFX}/money/money_coin_gain.wav`,
  money_coin_spend:      `${SFX}/money/money_coin_spend.wav`,
  money_big_profit:      `${SFX}/money/money_big_profit.wav`,
  money_loss:            `${SFX}/money/money_loss.wav`,
  // Market
  market_buy:            `${SFX}/market/market_buy.wav`,
  market_sell:           `${SFX}/market/market_sell.wav`,
  market_order_created:  `${SFX}/market/market_order_created.wav`,
  market_order_filled:   `${SFX}/market/market_order_filled.wav`,
  market_price_up:       `${SFX}/market/market_price_up.wav`,
  market_price_down:     `${SFX}/market/market_price_down.wav`,
  market_volatility_alert: `${SFX}/market/market_volatility_alert.wav`,
  contract_signed:       `${SFX}/market/contract_signed.wav`,
  debt_issued:           `${SFX}/market/debt_issued.wav`,
  futures_open:          `${SFX}/market/futures_open.wav`,
  futures_close:         `${SFX}/market/futures_close.wav`,
  // Building
  build_place:               `${SFX}/building/build_place.wav`,
  build_preview_move:        `${SFX}/building/build_preview_move.wav`,
  build_confirm:             `${SFX}/building/build_confirm.wav`,
  build_construction_start:  `${SFX}/building/build_construction_start.wav`,
  build_construction_complete: `${SFX}/building/build_construction_complete.wav`,
  build_upgrade:             `${SFX}/building/build_upgrade.wav`,
  build_repair:              `${SFX}/building/build_repair.wav`,
  build_demolish:            `${SFX}/building/build_demolish.wav`,
  land_unlock:               `${SFX}/building/land_unlock.wav`,
  road_place:                `${SFX}/building/road_place.wav`,
  // Farm / Barn
  farm_plant_seed:    `${SFX}/farm/farm_plant_seed.wav`,
  farm_water_crop:    `${SFX}/farm/farm_water_crop.wav`,
  farm_harvest:       `${SFX}/farm/farm_harvest.wav`,
  farm_crop_ready:    `${SFX}/farm/farm_crop_ready.wav`,
  barn_animal_feed:   `${SFX}/farm/barn_animal_feed.wav`,
  barn_collect_milk:  `${SFX}/farm/barn_collect_milk.wav`,
  barn_collect_eggs:  `${SFX}/farm/barn_collect_eggs.wav`,
  resource_pickup:    `${SFX}/farm/resource_pickup.wav`,
  // Processing
  mill_start:             `${SFX}/production/mill_start.wav`,
  mill_complete:          `${SFX}/production/mill_complete.wav`,
  kitchen_chop:           `${SFX}/production/kitchen_chop.wav`,
  kitchen_pan_sizzle:     `${SFX}/production/kitchen_pan_sizzle.wav`,
  kitchen_recipe_complete: `${SFX}/production/kitchen_recipe_complete.wav`,
  bakery_oven_start:      `${SFX}/production/bakery_oven_start.wav`,
  bakery_bread_ready:     `${SFX}/production/bakery_bread_ready.wav`,
  cafe_coffee_pour:       `${SFX}/production/cafe_coffee_pour.wav`,
  cafe_espresso:          `${SFX}/production/cafe_espresso.wav`,
  food_packaged:          `${SFX}/production/food_packaged.wav`,
  // Inventory / Logistics
  inventory_open:   `${SFX}/inventory/inventory_open.wav`,
  inventory_sort:   `${SFX}/inventory/inventory_sort.wav`,
  warehouse_store:  `${SFX}/inventory/warehouse_store.wav`,
  warehouse_takeout: `${SFX}/inventory/warehouse_takeout.wav`,
  truck_depart:     `${SFX}/inventory/truck_depart.wav`,
  truck_arrive:     `${SFX}/inventory/truck_arrive.wav`,
  ship_dock:        `${SFX}/inventory/ship_dock.wav`,
  delivery_complete: `${SFX}/inventory/delivery_complete.wav`,
  // Restaurant
  restaurant_open:          `${SFX}/restaurant/restaurant_open.wav`,
  customer_enter:           `${SFX}/restaurant/customer_enter.wav`,
  menu_set:                 `${SFX}/restaurant/menu_set.wav`,
  dish_served:              `${SFX}/restaurant/dish_served.wav`,
  dish_sold:                `${SFX}/restaurant/dish_sold.wav`,
  restaurant_revenue_collect: `${SFX}/restaurant/restaurant_revenue_collect.wav`,
  occupancy_up:             `${SFX}/restaurant/occupancy_up.wav`,
  occupancy_down:           `${SFX}/restaurant/occupancy_down.wav`,
  restaurant_level_up:      `${SFX}/restaurant/restaurant_level_up.wav`,
  // Research / Executive
  research_start:         `${SFX}/research/research_start.wav`,
  research_progress_tick: `${SFX}/research/research_progress_tick.wav`,
  research_complete:      `${SFX}/research/research_complete.wav`,
  tech_unlock:            `${SFX}/research/tech_unlock.wav`,
  executive_hire:         `${SFX}/research/executive_hire.wav`,
  executive_level_up:     `${SFX}/research/executive_level_up.wav`,
  skill_assign:           `${SFX}/research/skill_assign.wav`,
  buff_activate:          `${SFX}/research/buff_activate.wav`,
  buff_expire:            `${SFX}/research/buff_expire.wav`,
  // Quest / Social
  quest_accept:      `${SFX}/system/quest_accept.wav`,
  quest_complete:    `${SFX}/system/quest_complete.wav`,
  achievement_unlock: `${SFX}/system/achievement_unlock.wav`,
  leaderboard_rank_up: `${SFX}/system/leaderboard_rank_up.wav`,
  chat_send:         `${SFX}/system/chat_send.wav`,
  chat_receive:      `${SFX}/system/chat_receive.wav`,
  player_join:       `${SFX}/system/player_join.wav`,
  player_leave:      `${SFX}/system/player_leave.wav`,
  trade_request:     `${SFX}/system/trade_request.wav`,
  trade_accepted:    `${SFX}/system/trade_accepted.wav`,
  trade_rejected:    `${SFX}/system/trade_rejected.wav`,
  // System
  day_start:      `${SFX}/system/day_start.wav`,
  day_end:        `${SFX}/system/day_end.wav`,
  season_change:  `${SFX}/system/season_change.wav`,
  save_success:   `${SFX}/system/save_success.wav`,
  autosave:       `${SFX}/system/autosave.wav`,
  notification_general: `${SFX}/system/notification_general.wav`,
  warning_soft:   `${SFX}/system/warning_soft.wav`,
  critical_alert: `${SFX}/system/critical_alert.wav`,
}

/** Path lookup: Music key → expected asset file. */
export const MUSIC_PATHS: Record<MusicKey, string> = {
  bgm_main_menu:    `${BGM}/bgm_main_menu.mp3`,
  bgm_harbor_town:  `${BGM}/bgm_harbor_town.ogg`,
  bgm_market:       `${BGM}/bgm_market.ogg`,
  bgm_restaurant:   `${BGM}/bgm_restaurant.ogg`,
  bgm_farm:         `${BGM}/bgm_farm.ogg`,
  bgm_night:        `${BGM}/bgm_night.ogg`,
}

/** Path lookup: Ambience key → expected asset file. */
export const AMBIENCE_PATHS: Record<AmbienceKey, string> = {
  amb_harbor_day:   `${AMB}/amb_harbor_day.ogg`,
  amb_harbor_night: `${AMB}/amb_harbor_night.ogg`,
  amb_market:       `${AMB}/amb_market.ogg`,
  amb_restaurant:   `${AMB}/amb_restaurant.ogg`,
  amb_kitchen:      `${AMB}/amb_kitchen.ogg`,
  amb_farm:         `${AMB}/amb_farm.ogg`,
  amb_barn:         `${AMB}/amb_barn.ogg`,
  amb_warehouse:    `${AMB}/amb_warehouse.ogg`,
  amb_cafe:         `${AMB}/amb_cafe.ogg`,
  amb_rain_town:    `${AMB}/amb_rain_town.ogg`,
}

/** All SFX keys as an array (useful for dev panels, preloading). */
export const ALL_SFX_KEYS: SfxKey[] = Object.keys(SFX_PATHS) as SfxKey[]

/** All Music keys as an array. */
export const ALL_MUSIC_KEYS: MusicKey[] = Object.keys(MUSIC_PATHS) as MusicKey[]

/** All Ambience keys as an array. */
export const ALL_AMBIENCE_KEYS: AmbienceKey[] = Object.keys(AMBIENCE_PATHS) as AmbienceKey[]

/** Categorised groups for dev panel. */
export const SFX_CATEGORIES: Record<string, SfxKey[]> = {
  UI: [
    'ui_button_click', 'ui_button_hover', 'ui_confirm', 'ui_cancel',
    'ui_error', 'ui_disabled', 'ui_panel_open', 'ui_panel_close',
    'ui_tab_switch', 'ui_popup', 'ui_drag_start', 'ui_drag_drop',
    'ui_scroll_tick',
  ],
  Money: [
    'money_coin_gain', 'money_coin_spend', 'money_big_profit', 'money_loss',
  ],
  Market: [
    'market_buy', 'market_sell', 'market_order_created', 'market_order_filled',
    'market_price_up', 'market_price_down', 'market_volatility_alert',
    'contract_signed', 'debt_issued', 'futures_open', 'futures_close',
  ],
  Building: [
    'build_place', 'build_preview_move', 'build_confirm',
    'build_construction_start', 'build_construction_complete',
    'build_upgrade', 'build_repair', 'build_demolish', 'land_unlock', 'road_place',
  ],
  Farm: [
    'farm_plant_seed', 'farm_water_crop', 'farm_harvest', 'farm_crop_ready',
    'barn_animal_feed', 'barn_collect_milk', 'barn_collect_eggs', 'resource_pickup',
  ],
  Production: [
    'mill_start', 'mill_complete', 'kitchen_chop', 'kitchen_pan_sizzle',
    'kitchen_recipe_complete', 'bakery_oven_start', 'bakery_bread_ready',
    'cafe_coffee_pour', 'cafe_espresso', 'food_packaged',
  ],
  Logistics: [
    'inventory_open', 'inventory_sort', 'warehouse_store', 'warehouse_takeout',
    'truck_depart', 'truck_arrive', 'ship_dock', 'delivery_complete',
  ],
  Restaurant: [
    'restaurant_open', 'customer_enter', 'menu_set', 'dish_served', 'dish_sold',
    'restaurant_revenue_collect', 'occupancy_up', 'occupancy_down', 'restaurant_level_up',
  ],
  Research: [
    'research_start', 'research_progress_tick', 'research_complete', 'tech_unlock',
    'executive_hire', 'executive_level_up', 'skill_assign', 'buff_activate', 'buff_expire',
  ],
  Social: [
    'quest_accept', 'quest_complete', 'achievement_unlock', 'leaderboard_rank_up',
    'chat_send', 'chat_receive', 'player_join', 'player_leave',
    'trade_request', 'trade_accepted', 'trade_rejected',
  ],
  System: [
    'day_start', 'day_end', 'season_change', 'save_success', 'autosave',
    'notification_general', 'warning_soft', 'critical_alert',
  ],
}

/** Map music keys to their display names (for dev panel / settings). */
export const MUSIC_LABELS: Record<MusicKey, string> = {
  bgm_main_menu:   'Main Menu',
  bgm_harbor_town: 'Harbor Town',
  bgm_market:      'Market',
  bgm_restaurant:  'Restaurant',
  bgm_farm:        'Farm',
  bgm_night:       'Night',
}

/** Map ambience keys to display names. */
export const AMBIENCE_LABELS: Record<AmbienceKey, string> = {
  amb_harbor_day:   'Harbor (Day)',
  amb_harbor_night: 'Harbor (Night)',
  amb_market:       'Market',
  amb_restaurant:   'Restaurant',
  amb_kitchen:      'Kitchen',
  amb_farm:         'Farm',
  amb_barn:         'Barn',
  amb_warehouse:    'Warehouse',
  amb_cafe:         'Cafe',
  amb_rain_town:    'Rain',
}
