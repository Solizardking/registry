import type { Metadata } from "next"
import { Plus_Jakarta_Sans } from "next/font/google"
import { Navigation } from "@/components/navigation"
import { Footer } from "@/components/footer"
import { ThemeProvider } from "@/components/theme-provider"
import { Toaster } from "@/components/ui/sonner"
import "./globals.css"

const jakarta = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-jakarta",
})

export const metadata: Metadata = {
  title: "CLAWD Registry",
  description: "Clawd catalog for Solana agents, MCP servers, skills, and prompts",
  icons: {
    icon: "/icon.svg",
  },
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning>
      <body className={`${jakarta.variable} font-sans bg-[#08080c] text-zinc-100`}>
        <ThemeProvider>
          <div className="min-h-screen flex flex-col">
            <Navigation />
            <div className="flex-1">
              {children}
            </div>
            <Footer />
          </div>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  )
}
