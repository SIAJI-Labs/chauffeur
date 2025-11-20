import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
});

const jetbrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Chauffeur - Linux PHP Development Environment",
  description: "Linux's missing Valet/Herd equivalent. Zero-config PHP environments with isolated runtimes.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="scroll-smooth dark">
      <body
        className={`${inter.variable} ${jetbrainsMono.variable} font-sans antialiased bg-background text-slate-100 selection:bg-primary selection:text-slate-900`}
      >
        {children}
      </body>
    </html>
  );
}
