import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { createOrderSchema, type CreateOrderFormValues } from './createOrder.schema'
import { useCreateOrder } from '@/api/hooks/market.hooks'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface CreateOrderFormProps {
  resourceId: number
  onSuccess?: () => void
}

export function CreateOrderForm({ resourceId, onSuccess }: CreateOrderFormProps) {
  const createOrder = useCreateOrder()
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<CreateOrderFormValues>({
    resolver: zodResolver(createOrderSchema),
    defaultValues: { resourceId, kind: 'buy', quality: 0, quantity: 1, price: 1 },
  })

  // eslint-disable-next-line react-hooks/incompatible-library
  const kind = watch('kind')

  const onSubmit = (data: CreateOrderFormValues) => {
    createOrder.mutate(
      {
        resourceId: data.resourceId,
        kind: data.kind === 'buy' ? 0 : 1,
        quality: data.quality,
        quantity: data.quantity,
        price: data.price,
      },
      { onSuccess },
    )
  }

  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle>New Order</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
          <div className="flex gap-2">
            <Label className="flex-1">
              <input
                type="radio"
                value="buy"
                {...register('kind')}
                className="mr-1"
              />
              Buy
            </Label>
            <Label className="flex-1">
              <input
                type="radio"
                value="sell"
                {...register('kind')}
                className="mr-1"
              />
              Sell
            </Label>
          </div>

          <div>
            <Label htmlFor="quantity">Quantity</Label>
            <Input id="quantity" type="number" {...register('quantity')} />
            {errors.quantity && <p className="text-xs text-red-500 mt-1">{errors.quantity.message}</p>}
          </div>

          <div>
            <Label htmlFor="price">Price</Label>
            <Input id="price" type="number" step="0.01" {...register('price')} />
            {errors.price && <p className="text-xs text-red-500 mt-1">{errors.price.message}</p>}
          </div>

          <Button type="submit" disabled={createOrder.isPending}>
            {kind === 'buy' ? 'Place Buy Order' : 'Place Sell Order'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
