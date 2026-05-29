<script setup lang="ts">
import Layout from '../components/Layout.vue'
import { Network, Link2, Users, Activity, ChevronRight } from '@lucide/vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

const stats = [
  { name: 'Total Nodes', value: '—', icon: Network },
  { name: 'Total Edges', value: '—', icon: Link2 },
  { name: 'Communities', value: '—', icon: Users },
  { name: 'Active Users', value: '1', icon: Activity },
]

const quickActions = [
  { name: 'Manage Nodes', description: 'View and edit knowledge graph nodes', path: '/nodes' },
  { name: 'Manage Edges', description: 'View and edit relationships', path: '/edges' },
]
</script>

<template>
  <Layout>
    <div class="space-y-6">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">Dashboard</h2>
        <p class="text-muted-foreground">
          Welcome to AI-Graph Console. Manage your knowledge graph here.
        </p>
      </div>

      <!-- Stats Grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card v-for="stat in stats" :key="stat.name">
          <CardContent class="flex items-center gap-4 pt-6">
            <div class="p-2 rounded-md bg-primary/10 text-primary">
              <component :is="stat.icon" class="h-5 w-5" />
            </div>
            <div>
              <p class="text-sm font-medium text-muted-foreground">{{ stat.name }}</p>
              <p class="text-2xl font-bold">{{ stat.value }}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Quick Actions -->
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <CardContent class="p-0">
          <div class="divide-y divide-border">
            <router-link
              v-for="action in quickActions"
              :key="action.name"
              :to="action.path"
              class="flex items-center justify-between p-4 hover:bg-accent/50 transition-colors"
            >
              <div>
                <h4 class="font-medium">{{ action.name }}</h4>
                <p class="text-sm text-muted-foreground">{{ action.description }}</p>
              </div>
              <ChevronRight class="h-5 w-5 text-muted-foreground" />
            </router-link>
          </div>
        </CardContent>
      </Card>

      <!-- System Status -->
      <Card>
        <CardHeader>
          <CardTitle>System Status</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-muted-foreground">Console Server</span>
            <Badge variant="default" class="gap-1.5">
              <span class="h-1.5 w-1.5 rounded-full bg-emerald-500" />
              Operational
            </Badge>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-muted-foreground">Graph API</span>
            <Badge variant="default" class="gap-1.5">
              <span class="h-1.5 w-1.5 rounded-full bg-emerald-500" />
              Operational
            </Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  </Layout>
</template>
