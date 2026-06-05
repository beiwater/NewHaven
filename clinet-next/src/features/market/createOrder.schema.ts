import { z } from 'zod'

export const createOrderSchema = z.object({
  resourceId: z.number({ invalid_type_error: 'Select a resource' }).int().positive(),
  kind: z.enum(['buy', 'sell']),
  quality: z.coerce.number().int().min(0),
  quantity: z.coerce.number().int().positive('Quantity must be positive'),
  price: z.coerce.number().positive('Price must be positive'),
})

export type CreateOrderFormValues = z.infer<typeof createOrderSchema>
