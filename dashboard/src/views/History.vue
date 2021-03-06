<template>
    <b-row class="mb-5">
        <b-col>
            <b-table :items="histories" :fields="fields" :busy="isBusy" show-empty hover>
                <template v-slot:cell(name)="row">
                    {{ row.item.name }} <br/>
                    <b>{{ row.item.id }}</b>
                </template>
                <template v-slot:cell(status)="row">
                    <b-badge v-if="row.item.status === 'ok'" variant="success">OK</b-badge>
                    <b-badge v-if="row.item.status === 'unstable'" variant="warning">UNSTABLE</b-badge>
                    <b-badge v-if="row.item.status === 'running'" variant="dark">RUNNING</b-badge>
                    <b-badge v-if="row.item.status !== 'ok' && row.item.status !== 'unstable' && row.item.status !== 'running'" variant="danger">FAIL</b-badge>
                </template>
                <template v-slot:empty="scope">
                    {{ scope.emptyText }}
                </template>
                <template v-slot:cell(operations)="row">
                    <b-button-group>
                        <b-button size="sm" @click="row.toggleDetails">{{ row.detailsShowing ? 'Hide' : 'Show'}} Details</b-button>
                        <b-button size="sm" variant="info" @click="console_output(row.item.id, $event.target)" class="mr-2" v-if="row.item.status !== 'running'">Console</b-button>
                        <b-button size="sm" variant="dark" :to="'/jobs/' + row.item.job_id + '/'" class="mr-2" v-if="row.item.status === 'running'">Console</b-button>
                    </b-button-group>
                </template>
                <template v-slot:row-details="row">
                    <b-card bg-variant="dark" text-variant="white">
                        <b-card-text>{{ row.item.status }}</b-card-text>
                    </b-card>
                </template>
                <template v-slot:table-busy class="text-center text-danger my-2">
                    <b-spinner class="align-middle"></b-spinner>
                    <strong> Loading...</strong>
                </template>
            </b-table>
            <b-modal :id="infoModal.id" size="xl" scrollable :title="infoModal.title" ok-only @hide="resetInfoModal">
                <div role="tablist">
                    <b-card no-body class="mb-1" v-for="(stage, index) in infoModal.content" :key="index" v-if="stage.messages.length > 0">
                        <b-card-header header-tag="header" class="p-1" role="tab">
                            <b-button block href="#" v-b-toggle="'accordion-' + index" :variant="stage.messages.filter(item => item.level === 'ERROR').length > 0 ? 'warning': 'success'">
                                {{ stage.name }}
                                <span v-if="stage.max > 0"> ({{ stage.total }}/{{ stage.max }} {{ stage.percentage * 100 }}%)</span>
                            </b-button>
                        </b-card-header>
                        <b-collapse :id="'accordion-' + index" visible accordion="my-accordion" role="tabpanel">
                            <b-card-body body-bg-variant="dark" body-text-variant="white">
                                <b-card-text v-for="(line, index2) in stage.messages" :key="index2">
                                  <b-badge class="mr-2">{{ index2 + 1 }}</b-badge> <b class="text-success">{{ line.timestamp }}</b>
                                  <b-badge class="ml-2" :variant="line.level === 'ERROR' ? 'danger' : 'info'">{{ line.level }}</b-badge> <br/>
                                  {{ line.message }}
                                </b-card-text>
                            </b-card-body>
                        </b-collapse>
                    </b-card>
                </div>
            </b-modal>
        </b-col>
    </b-row>
</template>

<script>
    import axios from 'axios';
    import moment from 'moment';

    export default {
        name: 'History',
        data() {
            return {
                histories: [],
                isBusy: true,
                fields: [
                    {
                        key: "created_at",
                        label: "Time",
                        formatter: (value) => moment(value).format('YYYY-MM-DD HH:mm:ss')
                    },
                    {key: "name", label: "Name/ID"},
                    {key: "job", label: "Job", formatter: (value) => value.payload.name},
                    {key: "status", label: "Status"},
                    {key: "operations", label: "Operations"},
                ],
                infoModal: {
                    id: 'info-modal',
                    title: '',
                    content: ''
                }
            };
        },
        methods: {
            console_output(id, button) {
                axios.get('/api/histories/' + id + '/').then(response => {
                    this.infoModal.title = 'Console';
                    this.infoModal.content = response.data.output.stages;
                    this.$root.$emit('bv::show::modal', this.infoModal.id, button);
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            resetInfoModal() {
                this.infoModal.title = '';
                this.infoModal.content = '';
            },
        },
        mounted() {
            axios.get('/api/histories/').then(response => {
                this.histories = response.data;
                this.isBusy = false;
            }).catch(error => {
                this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                    title: 'ERROR',
                    variant: 'danger'
                });
            });
        }
    }
</script>
