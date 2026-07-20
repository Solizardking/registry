"use client"

import Image from "next/image"
import { cn } from "@/lib/utils"

type BrandLogoProps = {
  className?: string
  showWordmark?: boolean
  size?: "sm" | "md" | "lg"
}

const SIZES = {
  sm: { mark: 28, text: "text-base" },
  md: { mark: 36, text: "text-lg" },
  lg: { mark: 44, text: "text-xl" },
} as const

export function BrandLogo({ className, showWordmark = true, size = "md" }: BrandLogoProps) {
  const s = SIZES[size]
  return (
    <span className={cn("inline-flex items-center gap-2.5", className)}>
      <span className="relative shrink-0 overflow-hidden rounded-xl ring-1 ring-[#9945FF]/50 shadow-[0_0_18px_rgba(153,69,255,0.4)]">
        <Image
          src="/clawd-mascot.jpg"
          alt="Clawd"
          width={s.mark}
          height={s.mark}
          className="object-cover object-top"
          style={{ width: s.mark, height: s.mark }}
          priority
        />
      </span>
      {showWordmark && (
        <span className={cn("font-bold tracking-wide leading-none", s.text)}>
          <span className="text-gradient-clawd">CLAWD</span>
          <span className="ml-1.5 font-medium text-zinc-400">Registry</span>
        </span>
      )}
    </span>
  )
}
