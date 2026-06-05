import { handlers as marketHandlers } from './market'
import { handlers as warehouseHandlers } from './warehouse'
import { handlers as productionHandlers } from './production'
import { handlers as buildingsHandlers } from './buildings'
import { handlers as researchHandlers } from './research'
import { handlers as financialHandlers } from './financial'
import { handlers as executivesHandlers } from './executives'
import { handlers as chatHandlers } from './chat'
import { handlers as powerupHandlers } from './powerup'
import { handlers as leaderboardHandlers } from './leaderboard'

export const handlers = [
  ...marketHandlers,
  ...warehouseHandlers,
  ...productionHandlers,
  ...buildingsHandlers,
  ...researchHandlers,
  ...financialHandlers,
  ...executivesHandlers,
  ...chatHandlers,
  ...powerupHandlers,
  ...leaderboardHandlers,
]
