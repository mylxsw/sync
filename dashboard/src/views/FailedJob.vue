<template>
    <b-row class="mb-5">
        <b-col>
            <b-table :items="jobs" :fields="fields" :busy="isBusy" show-empty hover>
                <template slot="id" slot-scope="row">
                    {{ row.item.name }} <br/>
                    <b>{{ row.item.id }}</b>
                </template>
                <template slot="actions" slot-scope="row">
                    <b-button-group>
                        <b-button size="sm" variant="primary" @click="retryJob(row.item.id)">Retry</b-button>
                        <b-button size="sm" variant="danger" @click="deleteJob(row.item.id)">Delete</b-button>
                    </b-button-group>
                </template>
                <template slot="empty" slot-scope="scope">
                    {{ scope.emptyText }}
                </template>
                <template slot="files" slot-scope="row">
                    <b-list-group>
                        <b-list-group-item v-for="(file, index) in row.item.payload.files" :key="index">
                            {{ file.src }} <b class="text-success">=></b> {{ file.dest }}
                        </b-list-group-item>
                    </b-list-group>
                </template>
                <div slot="table-busy" class="text-center text-danger my-2">
                    <b-spinner class="align-middle"></b-spinner>
                    <strong> Loading...</strong>
                </div>
            </b-table>
        </b-col>
    </b-row>
</template>

<script>
    import axios from 'axios';
    import moment from 'moment';

    export default {
        name: 'Queue',
        data() {
            return {
                jobs: [],
                isBusy: true,
                fields: [
                    {key: "created_at", label: "Time", formatter: (value) => moment(value).format('YYYY-MM-DD HH:mm:ss')},
                    {key: "id", label: "Name/ID"},
                    {key: "payload.name", label: "Job"},
                    {key: "payload.from", label: "From"},
                    {key: "files", label: "Files"},
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
                    this.$bvToast.toast('Successful', {
                        title: 'OK',
                        variant: 'success',
                    });

                    this.updatePage();
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            deleteJob(id) {
                if (!confirm('Are you sure to delete this job?')) {
                    return false;
                }

                axios.delete('/api/failed-jobs/' + id + "/").then(response => {
                    this.$bvToast.toast('Successful', {
                        title: 'OK',
                        variant: 'success',
                    });

                    this.updatePage();
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            updatePage() {
                axios.get('/api/failed-jobs/').then(response => {
                    this.jobs = response.data;
                    this.isBusy = false;
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            }
        },
        mounted() {
            this.updatePage();
        }
    }
</script>
