<template>
    <b-row class="mb-5">
        <b-col>
            <b-table :items="jobs" :fields="fields" :busy="isBusy" show-empty hover>
                <template v-slot:cell(id)="row">
                    {{ row.item.name }} <br/>
                    <b>{{ row.item.id }}</b>
                </template>
                <template v-slot:cell(actions)="row">
                    <b-button-group>
                        <b-button size="sm" variant="primary" @click="retryJob(row.item.id)">Retry</b-button>
                        <b-button size="sm" variant="danger" @click="deleteJob(row.item.id)">Delete</b-button>
                    </b-button-group>
                </template>
                <template v-slot:empty="scope">
                    {{ scope.emptyText }}
                </template>
                <template v-slot:cell(files)="row">
                    <b-list-group>
                        <b-list-group-item v-for="(file, index) in row.item.payload.files" :key="index">
                            {{ file.src }} <b class="text-success">=></b> {{ file.dest }}
                        </b-list-group-item>
                    </b-list-group>
                </template>
                <template v-slot:table-busy class="text-center text-danger my-2">
                    <b-spinner class="align-middle"></b-spinner>
                    <strong> Loading...</strong>
                </template>
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
                this.$bvModal.msgBoxConfirm('Are you sure to retry this job?').then((value) => {
                    if (value !== true) {return;}
                    axios.put('/api/failed-jobs/' + id + "/").then(() => {
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
                });
            },
            deleteJob(id) {
                this.$bvModal.msgBoxConfirm('Are you sure to delete this job?').then((value) => {
                    if (value !== true) {return;}
                    axios.delete('/api/failed-jobs/' + id + "/").then(() => {
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
