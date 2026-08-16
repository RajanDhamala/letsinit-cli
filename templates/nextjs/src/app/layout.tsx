import type { Metadata } from "next"

import { QueryProvider } from "@/providers/query-provider"

import "./globals.css"

export const metadata: Metadata = {
  title: "LetsInit",
  description: "Bootstrap modern projects from pre-configured templates.",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <body>
        <QueryProvider>{children}</QueryProvider>
      </body>
    </html>
  )
}
