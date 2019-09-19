<template>
    <b-row class="mb-5 main-box">
        <b-col>
            <b-card-group class="mb-3">
                <b-card header="Browser">
                    <b-form @submit="updateBrowserSetting">
                        <b-form-group horizontal id="server_url" label="Server URL*" label-for="server_url_input">
                            <b-form-input id="server_url_input" type="text" v-model="server_url" placeholder="http://localhost:8819"></b-form-input>
                        </b-form-group>
                        <b-form-group horizontal id="token" label="Token" label-for="token_input">
                            <b-form-input id="token_input" type="text" v-model="token"></b-form-input>
                        </b-form-group>

                        <b-button type="submit" variant="primary">Save</b-button>
                    </b-form>
                </b-card>
            </b-card-group>

            <b-card-group class="mb-3">
                <b-card header="Global Sync Actions">
                    <b-form @submit="updateGlobalSyncActions">

                        <b-form-textarea v-model="globalSyncActions" rows="15"
                                         class="bg-dark text-white monospaced-font"></b-form-textarea>

                        <b-button type="submit" variant="primary" class="mt-2">Save</b-button>
                        <b-button class="mt-2 ml-2" @click="globalSyncActionTemplate()">Insert Template</b-button>
                    </b-form>
                </b-card>
            </b-card-group>

        </b-col>
    </b-row>
</template>

<script>
    import axios from 'axios';
    export default {
        name: 'Setting',
        data() {
            return {
                server_url: '',
                token: '',
                globalSyncActions: '',
            };
        },
        methods: {
            updateBrowserSetting() {
                this.$store.commit('updateServerUrl', this.server_url);
                this.$store.commit('updateToken', this.token);
                this.$bvToast.toast('Successful， Please refresh your web page', {
                    title: 'OK',
                    variant: 'success'
                });

                this.refreshBrowserSetting();
            },
            updateGlobalSyncActions() {
                axios.post('/api/setting/global-sync/', this.globalSyncActions, {headers: {'Content-Type': 'application/yaml'}}).then(() => {
                    this.$bvToast.toast('Successful', {
                        title: "OK",
                        variant: "success",
                    });
                    this.refreshSyncActionSetting();
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            globalSyncActionTemplate() {
                let self = this;
                let update = function () {
                    self.globalSyncActions = 'from: "127.0.0.1:8819"\n' +
                        'token: ""\n' +
                        'before:\n' +
                        '- action: command\n' +
                        '  command: pwd\n' +
                        'after:\n' +
                        '- action: command\n' +
                        '  command: curl -i https://www.baidu.com\n' +
                        'errors:\n' +
                        '- action: dingding\n' +
                        '  body: "## Server {{ sysinfo \\"hostname\\" }} : {{ .FileSyncGroup.Name }} Has errors\\n\\n**IP:**\n' +
                        '    {{ sysinfo \\"ip\\" }}\\n\\n**ERR:** \\n\\n    {{ .Err }}\\n"\n' +
                        '  token: YOUR_DINGDING_GROUP_TOKEN'
                };

                if (this.globalSyncActions !== '') {
                    this.$bvModal.msgBoxConfirm('Do you want to replace your settings with template?').then((value) => {
                        if (value !== true) {return;}
                        update()
                    });
                } else {
                    update();
                }
            },
            refreshBrowserSetting() {
                this.server_url = this.$store.getters.serverUrl;
                this.token = this.$store.getters.token;
            },
            refreshSyncActionSetting() {
                axios.get('/api/setting/global-sync/', {responseType: 'text', params: {format: 'yaml'}}).then(response => {
                    this.globalSyncActions = typeof response.data === 'string' ? response.data : '';
                }).catch(error => {
                    this.$bvToast.toast(error.response !== undefined ? error.response.data.error : error.toString(), {
                        title: 'ERROR',
                        variant: 'danger'
                    });
                });
            },
            refreshPage() {
                this.refreshBrowserSetting();
                this.refreshSyncActionSetting();
            }
        },
        mounted() {
            this.refreshPage();
        }
    }
</script>

<style>
    .main-box {
        max-width: 1000px;
    }
</style>