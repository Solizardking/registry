"use client"

import { BrandLogo } from "@/components/brand-logo"

export function Footer() {
  return (
    <footer className="border-t border-[#9945FF]/15 mt-auto bg-[#08080c]/70">
      <div className="container mx-auto px-6 py-6 flex flex-col sm:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-3 text-sm text-zinc-400">
          <BrandLogo size="sm" />
          <span className="hidden sm:inline text-zinc-500">
            Solana agent catalog · MCP · skills · prompts
          </span>
        </div>

        <div className="flex items-center gap-4">
          <a
            href="https://cheshireterminal.ai"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-zinc-400 hover:text-[#9945FF] transition-colors"
          >
            Cheshire Terminal
          </a>
          <a
            href="https://clawd.xyz"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-zinc-400 hover:text-[#14F195] transition-colors"
          >
            clawd.xyz
          </a>
          <span className="text-xs tabular-nums font-medium tracking-wide text-[#9945FF]/90">
            $CLAWD
          </span>
        </div>
      </div>
    </footer>
  )
}
