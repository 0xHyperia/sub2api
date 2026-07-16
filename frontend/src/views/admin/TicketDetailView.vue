<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1200px]">
      <TicketConversation :number="String(route.params.number)" admin @back="router.push('/admin/tickets')" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import TicketConversation from '@/components/tickets/TicketConversation.vue'

const route = useRoute()
const router = useRouter()
</script>
