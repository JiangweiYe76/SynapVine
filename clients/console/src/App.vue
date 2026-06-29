<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useRoute } from 'vue-router'
import Layout from './components/Layout.vue'

const authStore = useAuthStore()
const route = useRoute()

onMounted(() => {
  authStore.init()
})
</script>

<template>
  <Layout v-if="!route.meta.public">
    <router-view />
  </Layout>
  <router-view v-else />
</template>
