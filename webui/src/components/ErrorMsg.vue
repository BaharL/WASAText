<script>
export default {
  name: 'ErrorMsg',
  props: {
    msg: {
      type: String,
      required: true
    }
  }
}
</script>

<template>
  <div class="alert alert-danger" role="alert">
    {{ msg }}
  </div>
</template>

<style>
</style>
