<script setup>
import { useToast } from '../composables/useToast'

const { toastMessage, toastVisible } = useToast()
</script>

<template>
  <div :class="['toast', { 'show': toastVisible }]">
    {{ toastMessage }}
  </div>
</template>
