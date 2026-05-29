<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2 } from '@lucide/vue'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')

watch([username, password], () => {
  if (authStore.error) {
    authStore.error = null
  }
})

async function handleSubmit() {
  const success = await authStore.login({
    username: username.value,
    password: password.value,
  })
  if (success) {
    router.push('/')
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-background p-4">
    <Card class="w-full max-w-lg pt-8">
      <CardHeader class="text-center pb-6 px-8">
        <CardTitle class="text-2xl">AI-Graph Console</CardTitle>
      </CardHeader>

      <CardContent class="px-8">
        <form id="login-form" @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <Label for="username" class="mb-1.5 block">Username</Label>
            <Input
              id="username"
              v-model="username"
              type="text"
              placeholder="Enter your username"
              required
              :disabled="authStore.loading"
              class="h-10"
            />
          </div>

          <div>
            <Label for="password" class="mb-1.5 block">Password</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="Enter your password"
              required
              :disabled="authStore.loading"
              :aria-invalid="!!authStore.error"
              class="h-10"
            />
            <p class="mt-1 text-[13px] leading-none text-destructive min-h-[1rem]">
              {{ authStore.error || '\u00A0' }}
            </p>
          </div>
        </form>
      </CardContent>

      <CardFooter class="p-8">
        <Button
          type="submit"
          form="login-form"
          class="w-full h-10"
          :disabled="authStore.loading"
        >
          <Loader2 v-if="authStore.loading" class="mr-2 h-4 w-4 animate-spin" />
          {{ authStore.loading ? 'Signing in...' : 'Sign In' }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>
