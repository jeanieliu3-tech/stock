<script setup lang="ts">
import AppLayout from '@/components/AppLayout.vue'
</script>

<template>
  <AppLayout>
    <router-view v-slot="{ Component }">
      <Transition name="page" mode="out-in">
        <component :is="Component" />
      </Transition>
    </router-view>
  </AppLayout>
</template>
