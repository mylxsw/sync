<template>
    <b-row class="mb-5">
        <b-col>
            <b-table :items="jobs" :fields="fields" v-if="jobs.length > 0">
                <template slot="actions" slot-scope="row">
                    <b-button-group>
                        <b-button size="sm" variant="primary" @click="retryJob(row.item.id)">Retry</b-button>
                        <b-button size="sm" variant="danger" @click="deleteJob(row.item.id)">Delete</b-button>
                    </b-button-group>
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
                    {key: "payload.name", label: "Sync"},
                    {key: "payload.from", label: "From"},
                    {key: "actions", label: "Operations"},
                ],
            };
        },
        methods: {
            retryJob(id) {
                if (!confirm('Are you sure to retry this job?')) {
                    return false;
                }

                axios.put('/api/failed-jobs/' + id + "/").then(response => {
                    if (response.status !== 200) {
                        this.$bvToast.toast('Retry job failed', {
                            title: 'ERROR',
                            variant: 'danger',
                        });
                        return false;
                    }

                    this.$bvToast.toast('Successful', {
                        title: 'OK',
                        variant: 'success',
                    });

                    this.updatePage();
                });
            },
            deleteJob(id) {
                if (!confirm('Are you sure to delete this job?')) {
                    return false;
                }

                axios.delete('/api/failed-jobs/' + id + "/").then(response => {
                    if (response.status !== 200) {
                        this.$bvToast.toast('Delete job failed', {
                            title: 'ERROR',
                            variant: 'danger',
                        });
                        return false;
                    }

                    this.$bvToast.toast('Successful', {
                        title: 'OK',
                        variant: 'success',
                    });

                    this.updatePage();
                });
            },
            updatePage() {
                axios.get('/api/failed-jobs/').then(response => {
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
        },
        mounted() {
            this.updatePage();
        }
    }
</script>
