<script setup lang="ts">
import Layout from '../../components/Layout.vue'
import { GitBranch } from '@lucide/vue'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
</script>

<template>
  <Layout>
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold tracking-tight">Edges</h2>
          <p class="text-muted-foreground">
            Manage knowledge graph relationships
          </p>
        </div>
        <Button>Add Edge</Button>
      </div>

      <Card>
        <CardContent class="flex flex-col items-center justify-center py-16 text-center">
          <GitBranch class="h-12 w-12 text-muted-foreground/50 mb-4" />
          <h3 class="text-lg font-medium">Edge management coming soon</h3>
          <p class="text-sm text-muted-foreground mt-1">
            This feature is under development.
          </p>
        </CardContent>
      </Card>
    </div>
  </Layout>
</template>
