import React from 'react'

// Image host allowlist — only these domains can render as inline images.
// Matched by hostname (not substring), protocol must be https.
const IMAGE_HOST_ALLOWLIST = new Set([
  // 国内图床
  'imgse.com', 'www.imgse.com',
  'imgchr.com', 'www.imgchr.com',
  'superbed.cn', 'www.superbed.cn',
  'superbed.cc', 'www.superbed.cc',
  'imgurl.org', 'www.imgurl.org',
  'imgbed.cn', 'www.imgbed.cn',
  'img.st', 'www.img.st',
  // 海外图床
  'postimages.org', 'postimg.cc',
  'i.postimg.cc',
  'imgbb.com', 'ibb.co', 'i.ibb.co',
  'freeimage.host', 'iili.io',
  'imgbox.com', 'images2.imgbox.com',
  'thumbs2.imgbox.com',
  'catbox.moe', 'files.catbox.moe',
])

function isAllowedImageUrl(input: string): boolean {
  try {
    const url = new URL(input)
    if (url.protocol !== 'https:') return false
    return IMAGE_HOST_ALLOWLIST.has(url.hostname)
  } catch {
    return false
  }
}

export function renderMessageBody(body: string): React.ReactNode {
  const trimmed = body.trim()

  // If the whole message is an allowed image URL
  if (isAllowedImageUrl(trimmed)) {
    return (
      <img
        src={trimmed}
        alt=""
        className="max-w-[200px] max-h-[200px] rounded-lg border border-amber-200/60 my-1"
        loading="lazy"
      />
    )
  }

  // Check if an image URL is embedded in text
  const parts = trimmed.split(/(\s+)/)
  return (
    <>
      {parts.map((part, i) =>
        isAllowedImageUrl(part)
          ? <img key={i} src={part} alt="" className="max-w-[200px] max-h-[200px] rounded-lg border border-amber-200/60 my-1 block" loading="lazy" />
          : <span key={i}>{part}</span>
      )}
    </>
  )
}
