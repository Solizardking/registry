"use client"

import { ThemeProvider as NextThemesProvider } from "next-themes"

/** Dark-only Clawd theme (Solana purple / green). Light mode is disabled. */
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <NextThemesProvider
      attribute="class"
      defaultTheme="dark"
      forcedTheme="dark"
      enableSystem={false}
    >
      {children}
    </NextThemesProvider>
  )
}
