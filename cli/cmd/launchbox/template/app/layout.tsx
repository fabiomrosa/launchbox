import type { Metadata } from "next"
import {
  Geist,
  Geist_Mono,
  Special_Gothic_Expanded_One,
} from "next/font/google"

import { BASE_URL, SITE_DESCRIPTION, SITE_TITLE } from "@/config"
import { getMetadata } from "@/config/metadata"

import {
  GoogleTagManagerHead,
  GoogleTagManagerNoScript,
} from "@/components/GoogleTagManager"

import "./globals.css"

const geistSans = Geist({
  variable: "--font-geist",
  subsets: ["latin"],
  display: "swap",
})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
  display: "swap",
})

const specialGothic = Special_Gothic_Expanded_One({
  variable: "--font-special-gothic",
  subsets: ["latin"],
  display: "swap",
  weight: "400",
  adjustFontFallback: false,
  fallback: ["system-ui", "Arial", "sans-serif"],
})

export const metadata: Metadata = getMetadata({
  baseURL: BASE_URL,
  siteTitle: SITE_TITLE,
  siteDescription: SITE_DESCRIPTION,
})

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en">
      <head>
        <GoogleTagManagerHead />
      </head>
      <body
        className={`
          antialiased
          ${geistSans.variable}
          ${geistMono.variable}
          ${specialGothic.variable}
        `}
      >
        <GoogleTagManagerNoScript />
        {children}
      </body>
    </html>
  )
}
