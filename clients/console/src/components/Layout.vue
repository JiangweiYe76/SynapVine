<script setup lang="ts">
import { useAuthStore } from '../stores/auth'
import {
  LayoutDashboard,
  CircleDot,
  GitBranch,
  Layers,
  FileText,
  ClipboardCheck,
  Brain,
  LogOut,
  Menu,
} from '@lucide/vue'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'

const authStore = useAuthStore()
const router = useRouter()
const mobileOpen = ref(false)

async function logout() {
  // authStore.logout hits POST /api/auth/logout and then clears the
  // local session. We await so the server has invalidated the refresh
  // token before we navigate away.
  await authStore.logout()
  router.push('/login')
}

const navItems = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  { name: 'Nodes', path: '/nodes', icon: CircleDot },
  { name: 'Edges', path: '/edges', icon: GitBranch },
  { name: 'Communities', path: '/communities', icon: Layers },
  { name: 'Papers', path: '/papers', icon: FileText },
  { name: 'Review', path: '/review', icon: ClipboardCheck },
  { name: 'LLM', path: '/llm', icon: Brain },
]

function isActive(path: string) {
  return router.currentRoute.value.path === path
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <!-- Desktop Sidebar -->
    <aside class="hidden lg:flex fixed top-0 left-0 z-40 h-full w-64 flex-col border-r bg-card">
      <div class="flex h-14 items-center border-b px-4">
        <h1 class="text-lg font-semibold tracking-tight">AI-Graph</h1>
      </div>

      <nav class="flex-1 space-y-1 p-3">
        <Button
          v-for="item in navItems"
          :key="item.path"
          variant="ghost"
          as-child
          :class="[
            'w-full justify-start gap-3',
            isActive(item.path) && 'bg-accent text-accent-foreground',
          ]"
        >
          <router-link :to="item.path">
            <component :is="item.icon" class="h-4 w-4" />
            {{ item.name }}
          </router-link>
        </Button>
      </nav>

      <div class="p-3">
        <Separator class="mb-3" />
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <Avatar class="h-8 w-8">
              <AvatarFallback class="text-xs">
                {{ authStore.user?.username?.charAt(0).toUpperCase() ?? 'U' }}
              </AvatarFallback>
            </Avatar>
            <div class="text-sm">
              <p class="font-medium">{{ authStore.user?.username }}</p>
              <p class="text-xs text-muted-foreground capitalize">{{ authStore.user?.role }}</p>
            </div>
          </div>
          <Button variant="ghost" size="icon" @click="logout" title="Logout">
            <LogOut class="h-4 w-4" />
          </Button>
        </div>
      </div>
    </aside>

    <!-- Mobile Header + Sheet -->
    <header class="lg:hidden flex h-14 items-center justify-between border-b bg-card px-4">
      <Sheet v-model:open="mobileOpen">
        <SheetTrigger as-child>
          <Button variant="ghost" size="icon">
            <Menu class="h-5 w-5" />
          </Button>
        </SheetTrigger>
        <SheetContent side="left" class="w-64 p-0 flex flex-col">
          <SheetHeader class="border-b px-4 py-3">
            <SheetTitle>AI-Graph</SheetTitle>
          </SheetHeader>

          <nav class="flex-1 space-y-1 p-3">
            <Button
              v-for="item in navItems"
              :key="item.path"
              variant="ghost"
              as-child
              :class="[
                'w-full justify-start gap-3',
                isActive(item.path) && 'bg-accent text-accent-foreground',
              ]"
              @click="mobileOpen = false"
            >
              <router-link :to="item.path">
                <component :is="item.icon" class="h-4 w-4" />
                {{ item.name }}
              </router-link>
            </Button>
          </nav>

          <div class="p-3">
            <Separator class="mb-3" />
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <Avatar class="h-8 w-8">
                  <AvatarFallback class="text-xs">
                    {{ authStore.user?.username?.charAt(0).toUpperCase() ?? 'U' }}
                  </AvatarFallback>
                </Avatar>
                <div class="text-sm">
                  <p class="font-medium">{{ authStore.user?.username }}</p>
                  <p class="text-xs text-muted-foreground capitalize">{{ authStore.user?.role }}</p>
                </div>
              </div>
              <Button variant="ghost" size="icon" @click="logout" title="Logout">
                <LogOut class="h-4 w-4" />
              </Button>
            </div>
          </div>
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
