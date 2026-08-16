<script setup lang="ts">
defineProps<{ status: string }>()
</script>

<template>
  <span class="status-pill" :data-status="status"><span aria-hidden="true" />{{ status }}</span>
</template>
