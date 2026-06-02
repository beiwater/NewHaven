import type { ImgHTMLAttributes } from 'react'

interface IconProps extends ImgHTMLAttributes<HTMLImageElement> {
  name: string
  alt?: string
}

const ICON_BASE = '/assets/icons'
const ITEM_BASE = '/assets/items'

/**
 * Game UI icon from the actual PNG spritesheet assets.
 * Usage:
 *   <Icon name="icon_coin_v1" className="w-5 h-5" />
 *   <Icon name="item_wheat_v1" fromItems />
 */
export function Icon({ name, alt = '', fromItems, ...props }: IconProps & { fromItems?: boolean }) {
  const base = fromItems ? ITEM_BASE : ICON_BASE
  return (
    <img
      src={`${base}/${name}.png`}
      alt={alt}
      {...props}
    />
  )
}
