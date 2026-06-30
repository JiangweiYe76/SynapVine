<script setup lang="ts">
import { Menu } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import NavContent from './NavContent.vue'
</script>

<template>
  <div class="min-h-screen bg-background">
    <!-- Desktop Sidebar -->
    <aside class="hidden lg:flex fixed top-0 left-0 z-40 h-full w-64 flex-col border-r bg-card">
      <div class="flex h-14 items-center border-b px-4">
        <h1 class="text-lg font-semibold tracking-tight">AI-Graph</h1>
      </div>
      <NavContent />
    </aside>

    <!-- Mobile Header + Sheet -->
    <header class="lg:hidden flex h-14 items-center justify-between border-b bg-card px-4">
      <Sheet>
        <SheetTrigger as-child>
          <Button variant="ghost" size="icon">
            <Menu class="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" class="w-64 p-0 flex flex-col">
          <SheetHeader class="border-b px-4 py-3">
            <SheetTitle>AI-Graph</SheetTitle>
          </SheetHeader>
          <NavContent />
        </SheetContent>
      </Sheet>

      <h1 class="text-lg font-semibold">AI-Graph Console</h1>
      <div class="w-9" />
    </header>

    <!-- Main content -->
    <div class="lg:ml-64">
      <main class="p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
