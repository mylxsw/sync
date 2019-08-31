<template>
    <b-row class="mb-5">
        <b-col>
            <b-table :items="definitions" :fields="fields" v-if="definitions.length > 0">
                <template slot="files" slot-scope="row">
                    <li v-for="(file, index) in row.item.files" :key="index">{{ file.src }} => {{ file.dest }}</li>
                </template>
                <template slot="actions" slot-scope="row">
                    <b-button-group>
                        <b-button size="sm" variant="primary" @click="trigger_job(row.item.name)">Execute</b-button>
                    </b-button-group>
                </template>
            </b-table>
            <div v-if="definitions.length === 0">暂无记录</div>
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
            };
        },
        methods: {
            trigger_job(name) {
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
