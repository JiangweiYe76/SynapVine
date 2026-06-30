<script setup lang="ts">
import { useAuthStore } from '../stores/auth'
import {
  LayoutDashboard,
  CircleDot,
  GitBranch,
  Layers,
  Network,
  FileText,
  ClipboardCheck,
  Brain,
  Waves,
  Cpu,
  ChevronDown,
  LogOut,
} from '@lucide/vue'
import { computed, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

async function logout() {
  await authStore.logout()
  router.push('/login')
}

interface NavItem {
  name: string
  path: string
  icon: typeof LayoutDashboard
}

interface NavGroup {
  name: string
  icon: typeof LayoutDashboard
  children: NavItem[]
}

type NavEntry = NavItem | NavGroup

const navItems: NavEntry[] = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard },
  {
    name: 'Knowledge Graph',
    icon: Network,
    children: [
      { name: 'Nodes', path: '/nodes', icon: CircleDot },
      { name: 'Edges', path: '/edges', icon: GitBranch },
      { name: 'Communities', path: '/communities', icon: Layers },
    ],
  },
  { name: 'Papers', path: '/papers', icon: FileText },
  { name: 'Review', path: '/review', icon: ClipboardCheck },
  {
    name: 'AI Models',
    icon: Cpu,
    children: [
      { name: 'LLM', path: '/llm', icon: Brain },
      { name: 'Embedding', path: '/embedding', icon: Waves },
    ],
  },
]

function isGroup(entry: NavEntry): entry is NavGroup {
  return 'children' in entry
}

const currentPath = computed(() => route.path)

function isActive(path: string) {
  return currentPath.value === path
}

function isGroupActive(group: NavGroup): boolean {
  return group.children.some((c) => currentPath.value === c.path)
}

const openGroups = reactive<Record<string, boolean>>(
  Object.fromEntries(
    navItems.filter(isGroup).map((g) => [
      g.name,
      g.children.some((c) => c.path === route.path),
    ]),
  ),
)

function toggleGroup(name: string) {
  openGroups[name] = !openGroups[name]
}
</script>

<template>
  <nav class="flex-1 space-y-1 p-3">
    <template v-for="entry in navItems" :key="isGroup(entry) ? entry.name : entry.path">
      <div v-if="isGroup(entry)">
        <Button
          variant="ghost"
          :class="[
            'w-full justify-between gap-3',
            isGroupActive(entry) && 'bg-accent text-accent-foreground',
          ]"
          @click="toggleGroup(entry.name)"
        >
          <span class="flex items-center gap-3">
            <component :is="entry.icon" class="h-4 w-4" />
            {{ entry.name }}
          </span>
          <ChevronDown
            class="h-4 w-4 shrink-0 transition-transform duration-200"
            :class="openGroups[entry.name] && 'rotate-180'"
          />
        </Button>
        <div v-show="openGroups[entry.name]" class="space-y-1 pl-4 pt-1">
          <Button
            v-for="child in entry.children"
            :key="child.path"
            variant="ghost"
            as-child
            :class="[
              'w-full justify-start gap-3',
              isActive(child.path) && 'bg-accent text-accent-foreground',
            ]"
          >
            <router-link :to="child.path">
              <component :is="child.icon" class="h-4 w-4" />
              {{ child.name }}
            </router-link>
          </Button>
        </div>
      </div>
      <Button
        v-else
        variant="ghost"
        as-child
        :class="[
          'w-full justify-start gap-3',
          isActive(entry.path) && 'bg-accent text-accent-foreground',
        ]"
      >
        <router-link :to="entry.path">
          <component :is="entry.icon" class="h-4 w-4" />
          {{ entry.name }}
        </router-link>
      </Button>
    </template>
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
</template>
