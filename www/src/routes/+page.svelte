<script lang="ts">
  import Terminal from '$lib/Terminal.svelte';
</script>

<svelte:head>
  <title>bontree</title>
  <meta name="description" content="A fast, interactive terminal file explorer for coding agents." />
</svelte:head>

<div class="wrapper">
  <Terminal />
</div>

<style>
  .wrapper {
    width: 100%;
    max-width: 720px;
    padding: 24px;
  }
</style>
