<template>
  <EmailVerifyView embedded @back="emit('back')" />
</template>

<script setup lang="ts">
import EmailVerifyView from '@/views/auth/EmailVerifyView.vue'

const emit = defineEmits<{
  back: []
}>()
</script>
