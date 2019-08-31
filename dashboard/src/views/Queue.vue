<template>
    <b-row class="mb-5">
        <b-col>
            <b-table :items="jobs" :fields="fields" v-if="jobs.length > 0">
                <template slot="payload" slot-scope="row">
                    {{ row.item.payload.name }}
                </template>
                <template slot="payload" slot-scope="row">
                    {{ row.item.payload.from }}
                </template>
            </b-table>
            <div v-if="jobs.length === 0">Nothing</div>
        </b-col>
    </b-row>
</template>

<script>
    import axios from 'axios';

    export default {
        name: 'Queue',
        data() {
            return {
                jobs: [],
                fields: [
                    {key: "id", label: "ID"},
                    {key: "created_at", label: "Time"},
                    {key: "name", label: "Name"},
                    {key: "payload", label: "Sync"},
                    {key: "payload", label: "From"},
                ],
            };
        },
        mounted() {
            axios.get('/api/jobs/').then(response => {
                if (response.status !== 200) {
                    this.$bvToast.toast('Load data failed', {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                    return false;
                }

                this.jobs = response.data;
            });
        }
    }
</script>
