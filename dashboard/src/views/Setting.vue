<template>
    <b-row class="mb-5 main-box">
        <b-col>
            <b-form @submit="updateSetting">
                <b-card-group class="mb-3">
                    <b-card header="Basic">
                        <b-form-group horizontal id="server_url" label="Server URL*" label-for="server_url_input">
                            <b-form-input id="server_url_input" type="text" v-model="server_url" placeholder="http://localhost:8819"></b-form-input>
                        </b-form-group>
                        <b-form-group horizontal id="token" label="Token" label-for="token_input">
                            <b-form-input id="token_input" type="text" v-model="token"></b-form-input>
                        </b-form-group>
                    </b-card>
                </b-card-group>
                <b-button type="submit" variant="primary">Save</b-button>
            </b-form>
        </b-col>
    </b-row>
</template>

<script>
    export default {
        name: 'Setting',
        data() {
            return {
                server_url: '',
                token: '',
            };
        },
        methods: {
            updateSetting() {
                this.$store.commit('updateServerUrl', this.server_url);
                this.$store.commit('updateToken', this.token);
                this.$bvToast.toast('Successful， Please refresh your web page', {
                    title: 'OK',
                    variant: 'success'
                });
            },
        },
        mounted() {
            this.server_url = this.$store.getters.serverUrl;
            this.token = this.$store.getters.token;
        }
    }
</script>

<style>
    .main-box {
        max-width: 1000px;
    }
</style>