import { resourceIcon, buildingIcon, systemIcon } from '@/game/icons'
import type { ImgHTMLAttributes } from 'react'

interface IconProps extends Omit<ImgHTMLAttributes<HTMLImageElement>, 'resource'> {
  name: string
  alt?: string
  /** If true, resolve via resourceIcon(name) */
  resource?: boolean
  /** If true, resolve via buildingIcon(Number(name)) */
  building?: boolean
  /** If true, resolve via systemIcon(name) */
  system?: boolean
}

/**
 * Game UI icon.
 * Legacy: name like "icon_coin_v1" → /assets/icons/{name}.png
 * resource, building, system flags → lookup in icon registry with mapped path.
 */
export function Icon({
  name,
  alt = '',
  resource,
  building,
  system,
  ...props
}: IconProps) {
  let src: string
  if (resource) {
    src = resourceIcon(Number(name) || 0)
  } else if (building) {
    src = buildingIcon(Number(name) || 0)
  } else if (system) {
    src = systemIcon(name)
  } else {
    src = `/assets/icons/${name}.png`
  }
  return <img src={src} alt={alt} {...props} />
}
