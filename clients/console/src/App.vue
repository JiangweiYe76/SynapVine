<script setup lang="ts">
import { useRoute } from 'vue-router'
import Layout from './components/Layout.vue'

// Session restoration on reload is handled by the router guard in
// router/index.ts (silent /api/auth/refresh via httpOnly cookie), not
// by an onMounted hook here, so App.vue has no auth bootstrap logic.
const route = useRoute()
</script>

<template>
  <Layout v-if="!route.meta.public">
    <router-view />
  </Layout>
  <router-view v-else />
</template>
