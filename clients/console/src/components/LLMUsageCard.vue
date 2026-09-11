<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { RefreshCw } from '@lucide/vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

import { llmAPI } from '@/api/llm'
import type { LLMUsageSummary } from '@/types/llm'

const usage = ref<LLMUsageSummary | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

// Default window: the last 7 days including today.
function isoDaysAgo(days: number): string {
  const d = new Date()
  d.setDate(d.getDate() - days)
  return d.toISOString().slice(0, 10)
}

const fromDate = ref(isoDaysAgo(6))
const toDate = ref(isoDaysAgo(0))

const statBlocks = computed(() => {
  const t = usage.value?.totals
  return [
    { label: 'Total tokens', value: t?.total_tokens ?? 0 },
    { label: 'Calls', value: t?.calls ?? 0 },
    { label: 'Prompt', value: t?.prompt_tokens ?? 0 },
    { label: 'Completion', value: t?.completion_tokens ?? 0 },
  ]
})

// Trend bars: one per day in the window, height normalized to the max.
const trendDays = computed(() => {
  if (!usage.value) return []
  const byDate = new Map(usage.value.by_day.map((d) => [d.date, d]))
  const days: { date: string; tokens: number; pct: number }[] = []
  const start = new Date(fromDate.value)
  const end = new Date(toDate.value)
  let max = 0
  for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
    const key = d.toISOString().slice(0, 10)
    const tokens = byDate.get(key)?.total_tokens ?? 0
    max = Math.max(max, tokens)
    days.push({ date: key, tokens, pct: 0 })
  }
  for (const day of days) day.pct = max > 0 ? Math.max((day.tokens / max) * 100, day.tokens > 0 ? 4 : 1) : 1
  return days
})

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`
  return String(n)
}

async function fetchUsage() {
  loading.value = true
  error.value = null
  try {
    usage.value = await llmAPI.getUsageSummary(fromDate.value, toDate.value)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load usage'
  } finally {
    loading.value = false
  }
}

onMounted(fetchUsage)
</script>

<template>
  <Card>
    <CardHeader>
      <div class="flex items-center justify-between">
        <div>
          <CardTitle>Token Usage</CardTitle>
          <CardDescription>LLM consumption recorded by the extraction pipeline</CardDescription>
        </div>
        <Button variant="outline" size="sm" :disabled="loading" @click="fetchUsage">
          <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" />
          Refresh
        </Button>
      </div>
    </CardHeader>
    <CardContent class="space-y-6">
      <div v-if="error" class="text-sm text-destructive">{{ error }}</div>

      <div class="flex flex-wrap items-end gap-3">
        <div>
          <label class="mb-1 block text-xs text-muted-foreground">From</label>
          <Input v-model="fromDate" type="date" class="w-40" @change="fetchUsage" />
        </div>
        <div>
          <label class="mb-1 block text-xs text-muted-foreground">To</label>
          <Input v-model="toDate" type="date" class="w-40" @change="fetchUsage" />
        </div>
      </div>

      <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <div v-for="stat in statBlocks" :key="stat.label" class="rounded-lg border p-4">
          <div class="text-2xl font-semibold tabular-nums">{{ formatTokens(stat.value) }}</div>
          <div class="mt-1 text-xs text-muted-foreground">{{ stat.label }}</div>
        </div>
      </div>

      <div v-if="trendDays.length" class="space-y-2">
        <div class="text-sm font-medium">Daily trend</div>
        <div class="flex h-24 items-end gap-1.5">
          <div
            v-for="day in trendDays"
            :key="day.date"
            class="flex-1 rounded-t bg-primary/70 transition-all hover:bg-primary"
            :style="{ height: `${day.pct}%` }"
            :title="`${day.date}: ${day.tokens.toLocaleString()} tokens`"
          />
        </div>
        <div class="flex justify-between text-xs text-muted-foreground">
          <span>{{ trendDays[0]?.date }}</span>
          <span>{{ trendDays[trendDays.length - 1]?.date }}</span>
        </div>
      </div>

      <div v-if="usage?.by_provider?.length" class="space-y-2">
        <div class="text-sm font-medium">By provider</div>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Provider</TableHead>
              <TableHead>Model</TableHead>
              <TableHead class="text-right">Calls</TableHead>
              <TableHead class="text-right">Prompt</TableHead>
              <TableHead class="text-right">Completion</TableHead>
              <TableHead class="text-right">Total</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="p in usage.by_provider" :key="`${p.provider_id}-${p.model}`">
              <TableCell class="font-mono text-xs">{{ p.provider_id }}</TableCell>
              <TableCell class="font-mono text-xs">{{ p.model }}</TableCell>
              <TableCell class="text-right tabular-nums">{{ p.calls }}</TableCell>
              <TableCell class="text-right tabular-nums">{{ p.prompt_tokens.toLocaleString() }}</TableCell>
              <TableCell class="text-right tabular-nums">{{ p.completion_tokens.toLocaleString() }}</TableCell>
              <TableCell class="text-right font-medium tabular-nums">{{ p.total_tokens.toLocaleString() }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <div v-else-if="!loading && usage" class="text-sm text-muted-foreground">
        No usage recorded in this window yet.
      </div>
    </CardContent>
  </Card>
</template>
