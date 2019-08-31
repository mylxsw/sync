<template>
  <b-row class="mb-5">
    <b-col>
      <b-table :items="histories" :fields="fields" v-if="histories.length > 0">
        <template slot="job" slot-scope="row">
          {{ row.item.job.payload.name }}
        </template>
      </b-table>
      <div v-if="histories.length === 0">Nothing</div>
    </b-col>
  </b-row>
</template>

<script>
import axios from 'axios';

export default {
  name: 'History',
  data() {
    return {
      histories: [],
      fields: [
        {key: "id", label: "ID"},
        {key: "created_at", label: "Time"},
        {key: "name", label: "Name"},
        {key: "job", label: "Job"},
        {key: "status", label: "Status"},
      ],
    };
  },
  mounted() {
    axios.get('/api/histories/').then(response => {
      if (response.status !== 200) {
        this.$bvToast.toast('Load data failed', {
          title: 'ERROR',
          variant: 'danger'
        });
        return false;
      }

      this.histories = response.data;
    });
  }
}
</script>
