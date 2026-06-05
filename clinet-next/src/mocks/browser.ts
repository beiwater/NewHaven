import { setupWorker } from 'msw/browser'
import { handlers } from './handlers/research'
import { handlers as marketHandlers } from './handlers/market'
import { handlers as financialHandlers } from './handlers/financial'
import { handlers as executivesHandlers } from './handlers/executives'
import { handlers as chatHandlers } from './handlers/chat'

export const worker = setupWorker(
  ...handlers,
  ...marketHandlers,
  ...financialHandlers,
  ...executivesHandlers,
  ...chatHandlers,
)
