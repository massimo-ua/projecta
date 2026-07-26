import React from 'react';
import { Logo } from './components/Logo';
import Logout from './components/Logout';
import { AppFooter } from './components/index.js';
import { Toaster } from '@/components/ui/sonner';

export default function HomeLayout({ children }) {
  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground">
      <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-14 items-center justify-between px-4 sm:px-8">
          <Logo />
          <Logout />
        </div>
      </header>
      <main className="flex-1 container px-4 sm:px-8 py-6">
        {children}
      </main>
      <footer className="border-t py-4 px-4 sm:px-8 bg-muted/30">
        <AppFooter />
      </footer>
      <Toaster />
    </div>
  );
}
