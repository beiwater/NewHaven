import { handlers as marketHandlers } from './market'
import { handlers as warehouseHandlers } from './warehouse'
import { handlers as productionHandlers } from './production'
import { handlers as buildingsHandlers } from './buildings'

export const handlers = [
  ...marketHandlers,
  ...warehouseHandlers,
  ...productionHandlers,
  ...buildingsHandlers,
]
