<template>
    <b-row class="mb-5">
        <b-col>
            <b-button-toolbar class="mb-2">
                <b-button variant="primary" v-b-modal.newSyncModel>New Sync Definition</b-button>
                <b-button variant="info" @click="batch_edit($event.target)" class="ml-2">Bulk Edit</b-button>
            </b-button-toolbar>
            <b-table ref="selectableTable" :items="definitions" :busy="isBusy" :fields="fields" show-empty hover
                     selectable select-mode="multi" selected-variant="success" @row-selected="onRowSelected">
                <template v-slot:cell(selected)="{ rowSelected }">
                    <template v-if="rowSelected">
                        <span aria-hidden="true">&check;</span>
                        <span class="sr-only">Selected</span>
                    </template>
                    <template v-else>
                        <span aria-hidden="true">&nbsp;</span>
                        <span class="sr-only">Not selected</span>
                    </template>
                </template>
                <template v-slot:cell(files)="row">
                    <b-list-group>
                        <b-list-group-item v-for="(file, index) in row.item.files" :key="index">
                            {{ file.src }} <b class="text-success">=></b> {{ file.dest }}
                        </b-list-group-item>
                    </b-list-group>
                </template>
                <template v-slot:empty="scope">
                    {{ scope.emptyText }}
                </template>
                <template v-slot:cell(actions)="row">
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
                <template v-slot:table-busy class="text-center text-danger my-2">
                    <b-spinner class="align-middle"></b-spinner>
                    <strong> Loading...</strong>
                </template>
            </b-table>

            <b-button-toolbar class="mt-2">
                <b-button variant="dark" @click="$refs.selectableTable.selectAllRows()">Select All</b-button>
                <b-button class="ml-2" @click="$refs.selectableTable.clearSelected()" v-if="selected.length > 0">Reset
                </b-button>

                <b-button variant="primary" class="ml-2" @click="trigger_jobs()" v-if="selected.length > 0">Execute
                </b-button>
                <b-button variant="danger" class="ml-2" @click="delete_syncs()" v-if="selected.length > 0">Delete
                </b-button>
            </b-button-toolbar>

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
                    {key: "selected", label: "-"},
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
                },
                selected: [],
            };
        },
        methods: {
            onRowSelected(items) {
                this.selected = items
            },
            /**
             * insert a template to textarea when add new sync definition
             */
            insert_template() {
                this.newSyncModel.content = '- name: sync-logs\n' +
                    '  from: 127.0.0.1:8818\n' +
                    '  files:\n' +
                    '  - src: /var/log\n' +
                    '    dest: /tmp/logs\n' +
                    '    ignores:\n' +
                    '    - .git/\n' +
                    '    - .DS_Store\n' +
                    '  rules: []\n' +
                    '  before:\n' +
                    '  - action: command\n' +
                    '    command: pwd\n' +
                    '  after:\n' +
                    '  - action: command\n' +
                    '    command: curl -i https://www.baidu.com\n' +
                    '  errors:\n' +
                    '  - action: dingding\n' +
                    '    body: "## Server {{ sysinfo \\"hostname\\" }} : {{ .FileSyncGroup.Name }} Has errors\\n\\n**IP:**\n' +
                    '      {{ sysinfo \\"ip\\" }}\\n\\n**ERR:** \\n\\n    {{ .Err }}\\n"\n' +
                    '    token: YOUR_DINGDING_GROUP_TOKEN\n';
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
                axios.post('/api/sync/', this.newSyncModel.content, {headers: {'Content-Type': 'application/yaml'}}).then(() => {
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
             * bulk trigger jobs
             */
            trigger_jobs() {
                this.$bvModal.msgBoxConfirm('Are you sure to execute all this definitions as jobs ?').then((value) => {
                    if (value !== true) {return;}
                    axios.post("/api/jobs-bulk/", {defs: this.selected.map(item => item.name)}).then(() => {
                        this.$bvToast.toast('Successful', {
                            title: 'OK',
                            variant: 'success',
                        })
                    }).catch(error => {
                        this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                            title: 'ERROR',
                            variant: 'danger'
                        });
                    });
                });
            },
            /**
             * trigger a job, add this job to queue
             *
             * @param name
             */
            trigger_job(name) {
                this.$bvModal.msgBoxConfirm('Are you sure to execute this definition as a job ?').then((value) => {
                    if (value !== true) {return;}
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
                })
            },
            /**
             * Bulk delete sync definitions
             */
            delete_syncs() {
                this.$bvModal.msgBoxConfirm('Are you sure to delete all this definitions ?').then((value) => {
                    if (value !== true) {return;}
                    axios.delete("/api/sync-bulk/", {data: {names: this.selected.map(item => item.name)}}).then(() => {
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
                })
            },
            /**
             * Delete a sync definition
             */
            delete_sync(name) {
                this.$bvModal.msgBoxConfirm('Are you sure to delete this definition ?').then((value) => {
                    if (value !== true) {return;}
                    axios.delete("/api/sync/" + name + "/").then(() => {
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
                })
            },
            /**
             * batch edit sync definitions
             */
            batch_edit(button) {
                axios.get('/api/sync/', {params: {format: 'yaml'}}).then(response => {
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
