<template>
  <div>
    <!--<core-servers-type-generic v-if="this.server" :server="server"></core-servers-type-generic>-->
    <server-render v-if="this.server" :server="server"></server-render>
    <b-row v-else>
      <b-col cols="5"/>
      <b-col cols="2">
        <b-spinner class="align-middle"/>
        <strong v-text="$t('common.Loading')"></strong>
      </b-col>
    </b-row>
  </div>
</template>

<script>
export default {
  data () {
    return {
      server: null,
      recover: null,
      statRequest: null
    }
  },
  mounted () {
    this.server = this.loadServer()
  },
  methods: {
    loadServer () {
      let vue = this
      this.$http.get('/api/servers/' + this.$route.params.id).then(function (response) {
        vue.server = response.data.data
        let base = location.protocol === 'https' ? 'wss://' : 'ws:/' + location.host
        let url = base + '/daemon/server/' + vue.server.id + '/socket'
        vue.$connect(url)
        vue.statRequest = setInterval(vue.callStats, 3000)
      })
    },
    callStats () {
      this.$socket.sendObj({ type: 'stat' })
    }
  },
  beforeDestroy: function () {
    this.$disconnect()
    if (this.statRequest) {
      clearInterval(this.statRequest)
    }
  }
}
</script>
