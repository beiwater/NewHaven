import React from 'react'

const IMAGE_EXT_RE = /\.(?:jpg|jpeg|png|gif|webp|svg)(\?[^\s]*)?$/i
const IMGUR_RE = /https?:\/\/i?\.imgur\.com\/[^\s]+/i


export function renderMessageBody(body: string): React.ReactNode {
  const trimmed = body.trim()

  // If the whole message is an image URL
  if (IMAGE_EXT_RE.test(trimmed) || IMGUR_RE.test(trimmed)) {
    return (
      <img
        src={trimmed}
        alt=""
        className="max-w-[200px] max-h-[200px] rounded-lg border border-amber-200/60 my-1"
        loading="lazy"
      />
    )
  }

  // Check if an image URL is embedded in text (split by space and render)
  const parts = trimmed.split(/(\s+)/)
  return (
    <>
      {parts.map((part, i) =>
        (IMAGE_EXT_RE.test(part) || IMGUR_RE.test(part))
          ? <img key={i} src={part} alt="" className="max-w-[200px] max-h-[200px] rounded-lg border border-amber-200/60 my-1 block" loading="lazy" />
          : <span key={i}>{part}</span>
      )}
    </>
  )
}
