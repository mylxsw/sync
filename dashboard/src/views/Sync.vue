<template>
    <b-row class="mb-5">
        <b-col>
            <b-button-group class="mb-2">
                <b-button variant="primary" v-b-modal.newSyncModel>New Sync Definition</b-button>
            </b-button-group>
            <b-table :items="definitions" :busy="isBusy" :fields="fields" show-empty>
                <template slot="files" slot-scope="row">
                    <b-list-group>
                        <b-list-group-item v-for="(file, index) in row.item.files" :key="index">
                            {{ file.src }} <b class="text-success">=></b> {{ file.dest }}
                        </b-list-group-item>
                    </b-list-group>
                </template>
                <template slot="empty" slot-scope="scope">
                    {{ scope.emptyText }}
                </template>
                <template slot="actions" slot-scope="row">
                    <b-button-group>
                        <b-button size="sm" variant="primary" @click="trigger_job(row.item.name)" class="mr-2">Execute
                        </b-button>
                        <b-button size="sm" variant="info" @click="edit_sync(row.item.name, $event.target)">Edit
                        </b-button>
                        <b-button size="sm" variant="danger" @click="delete_sync(row.item.name)">Delete</b-button>

                        <b-button size="sm" @click="show_yaml(row.item.name, $event.target)" class="ml-2">Details
                        </b-button>
                    </b-button-group>
                </template>
                <div slot="table-busy" class="text-center text-danger my-2">
                    <b-spinner class="align-middle"></b-spinner>
                    <strong> Loading...</strong>
                </div>
            </b-table>

            <b-modal :id="infoModal.id" :title="infoModal.title" ok-only @hide="resetInfoModal" size="xl">
                <b-card bg-variant="dark" text-variant="white">
                    <b-card-text>
                        <pre class="text-white">{{ infoModal.content }}</pre>
                    </b-card-text>
                </b-card>
            </b-modal>

            <b-modal id="newSyncModel" title="Sync Definition" ok-title="Save" @ok="new_sync" size="xl">
                <b-button variant="info" size="sm" class="mb-2" @click="insert_template()">Insert Template</b-button>
                <b-form-textarea v-model="newSyncModel.content" rows="15"
                                 class="bg-dark text-white monospaced-font"></b-form-textarea>
            </b-modal>
        </b-col>
    </b-row>
</template>

<script>
    import axios from 'axios';

    export default {
        name: 'Queue',
        data() {
            return {
                definitions: [],
                isBusy: true,
                fields: [
                    {key: "name", label: "Name"},
                    {key: "from", label: "From"},
                    {key: "files", label: "Files"},
                    {key: "actions", label: "Operation"},
                ],
                newSyncModel: {
                    content: "",
                },
                infoModal: {
                    id: 'info-modal',
                    title: '',
                    content: ''
                }
            };
        },
        methods: {
            /**
             * insert a template to textarea when add new sync definition
             */
            insert_template() {
                this.newSyncModel.content = '- name: sync-logs\n' +
                    '  from: 127.0.0.1:8818\n' +
                    '  files:\n' +
                    '  - src: /data/logs\n' +
                    '    dest: /tmp\n' +
                    '    ignores:\n' +
                    '    - .git/\n' +
                    '    - .DS_Store\n' +
                    '  rules: []\n' +
                    '  after:\n' +
                    '  - action: command\n' +
                    '    command: curl -i https://www.baidu.com';
            },
            /**
             * create a new sync definition
             *
             * @param bvModalEvt
             */
            new_sync(bvModalEvt) {
                bvModalEvt.preventDefault();

                if (this.newSyncModel.content.trim() === "") {
                    this.$bvToast.toast('Definition content must not be empty!', {
                        title: "ERROR",
                        variant: "danger",
                    });

                    return;
                }

                this.isBusy = true;
                axios.post('/api/sync/', this.newSyncModel.content, {headers: {'Content-Type': 'application/yaml'}}).then(response => {
                    this.$bvToast.toast('Successful', {
                        title: "OK",
                        variant: "success",
                    });

                    this.$nextTick(() => {
                        this.refreshPage();
                        bvModalEvt.vueTarget.hide();
                    });
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            /**
             * trigger a job, add this job to queue
             *
             * @param name
             */
            trigger_job(name) {
                if (!confirm('Are you sure to execute this definition as a job ?')) {
                    return;
                }

                axios.post("/api/jobs/", {def: name}).then(response => {
                    this.$bvToast.toast('Successful，Job ID = ' + response.data.id, {
                        title: 'OK',
                        variant: 'success',
                    })
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            /**
             * Delete a sync definition
             */
            delete_sync(name) {
                if (!confirm('Are you sure to delete this definition ?')) {
                    return;
                }

                axios.delete("/api/sync/" + name + "/").then(response => {
                    this.$bvToast.toast('Successful', {
                        title: 'OK',
                        variant: 'success',
                    });
                    this.isBusy = true;
                    this.refreshPage();
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            /**
             * Edit a sync definition
             */
            edit_sync(name, button) {
                axios.get('/api/sync/' + name + '/', {params: {format: 'yaml'}}).then(response => {
                    this.newSyncModel.content = response.data;
                    this.$root.$emit('bv::show::modal', "newSyncModel", button);
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            /**
             * show yaml definition details
             *
             * @param name
             * @param button
             */
            show_yaml(name, button) {
                axios.get('/api/sync/' + name + '/', {params: {format: 'yaml'}}).then(response => {
                    this.infoModal.title = 'Settings In YAML';
                    this.infoModal.content = response.data;
                    this.$root.$emit('bv::show::modal', this.infoModal.id, button);
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            /**
             * reset detail model box
             */
            resetInfoModal() {
                this.infoModal.title = '';
                this.infoModal.content = '';
            },
            /**
             * refresh table contents
             */
            refreshPage() {
                axios.get('/api/sync/').then(response => {
                    this.definitions = response.data;
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
            this.refreshPage();
        }
    }
</script>

<style>
    .monospaced-font {
        font-family: 'Courier New', Courier, monospace;
    }
</style>
