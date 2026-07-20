"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { BrandLogo } from "@/components/brand-logo"

export function Navigation() {
  const pathname = usePathname()

  const isActive = (path: string) => {
    if (path === "/") {
      return pathname === "/"
    }
    return pathname.startsWith(path)
  }

  return (
    <nav className="border-b border-[#9945FF]/15 bg-[#08080c]/85 backdrop-blur-md sticky top-0 z-50">
      <div className="container mx-auto px-6">
        <div className="flex items-center gap-10 h-14">
          <Link href="/" className="flex items-center shrink-0 rounded-md px-1 py-1 hover:opacity-90 transition-opacity">
            <BrandLogo size="md" />
          </Link>

          <div className="flex items-center gap-1">
            <Link
              href="/"
              className={`relative px-3 py-1.5 text-[15px] font-medium transition-colors ${
                isActive("/")
                  ? "text-white after:absolute after:bottom-[-13px] after:left-1 after:right-1 after:h-[2px] after:bg-[#9945FF] after:rounded-full after:shadow-[0_0_8px_rgba(153,69,255,0.8)]"
                  : "text-zinc-400 hover:text-zinc-100"
              }`}
            >
              Catalog
            </Link>
            <Link
              href="/deployed"
              className={`relative px-3 py-1.5 text-[15px] font-medium transition-colors ${
                isActive("/deployed")
                  ? "text-white after:absolute after:bottom-[-13px] after:left-1 after:right-1 after:h-[2px] after:bg-[#9945FF] after:rounded-full after:shadow-[0_0_8px_rgba(153,69,255,0.8)]"
                  : "text-zinc-400 hover:text-zinc-100"
              }`}
            >
              Deployed
            </Link>
          </div>

          <div className="ml-auto flex items-center gap-2">
            <span className="hidden sm:inline-flex items-center gap-1.5 rounded-full border border-[#9945FF]/25 bg-[#9945FF]/10 px-2.5 py-1 text-[11px] font-medium tracking-wide text-[#c4a0ff]">
              <span className="h-1.5 w-1.5 rounded-full bg-[#14F195] shadow-[0_0_6px_#14F195]" />
              Solana
            </span>
          </div>
        </div>
      </div>
    </nav>
  )
}
