<template>
    <b-row class="mb-5">
        <b-col>
            <b-table :items="definitions" :fields="fields" v-if="definitions.length > 0">
                <template slot="files" slot-scope="row">
                    <b-list-group>
                        <b-list-group-item v-for="(file, index) in row.item.files" :key="index">
                            {{ file.src }} <b class="text-success">=></b> {{ file.dest }}
                        </b-list-group-item>
                    </b-list-group>
                </template>
                <template slot="actions" slot-scope="row">
                    <b-button-group>
                        <b-button size="sm" variant="primary" @click="trigger_job(row.item.name)">Execute</b-button>
                        <b-button size="sm" @click="show_yaml(row.item.name, $event.target)" class="mr-2">Details
                        </b-button>
                    </b-button-group>
                </template>
            </b-table>
            <div v-if="definitions.length === 0">Nothing</div>

            <b-modal :id="infoModal.id" :title="infoModal.title" ok-only @hide="resetInfoModal">
                <b-card bg-variant="dark" text-variant="white">
                    <b-card-text><pre class="text-white">{{ infoModal.content }}</pre></b-card-text>
                </b-card>
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
                fields: [
                    {key: "name", label: "Name"},
                    {key: "from", label: "From"},
                    {key: "files", label: "Files"},
                    {key: "actions", label: "Operation"},
                ],
                infoModal: {
                    id: 'info-modal',
                    title: '',
                    content: ''
                }
            };
        },
        methods: {
            trigger_job(name) {
                if (!confirm('Are you sure to execute this job?')) {
                    return;
                }

                axios.post("/api/jobs/", {def: name}).then(response => {
                    if (response.status !== 200) {
                        this.$bvToast.toast('Trigger sync job failed', {
                            title: 'ERROR',
                            variant: 'danger'
                        });
                        return false;
                    }

                    this.$bvToast.toast('Successful，Job ID = ' + response.data.id, {
                        title: 'OK',
                        variant: 'success',
                    })
                });
            },
            show_yaml(name, button) {
                axios.get('/api/sync/' + name + '/', {params: {format: 'yaml'}}).then(response => {
                    if (response.status !== 200) {
                        this.$bvToast.toast('Get sync definition failed', {
                            title: 'ERROR',
                            variant: 'danger',
                        });
                        return false;
                    }

                    this.infoModal.title = 'Settings In YAML';
                    this.infoModal.content = response.data;
                    this.$root.$emit('bv::show::modal', this.infoModal.id, button);
                });
            },
            resetInfoModal() {
                this.infoModal.title = '';
                this.infoModal.content = '';
            },
        },
        mounted() {
            axios.get('/api/sync/').then(response => {
                if (response.status !== 200) {
                    this.$bvToast.toast('Load data failed', {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                    return false;
                }

                this.definitions = response.data;
            });
        }
    }
</script>
