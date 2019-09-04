<template>
    <div id="app">
        <b-container fluid>
            <b-navbar type="dark" toggleable="md" variant="primary" class="mb-3" sticky>
                <b-navbar-brand href="/">Sync</b-navbar-brand>
                <b-collapse is-nav id="nav_dropdown_collapse">
                    <b-navbar-nav>
                        <b-nav-item to="/" exact>History</b-nav-item>
                        <b-nav-item to="/queue" exact>Queue</b-nav-item>
                        <b-nav-item to="/sync/definitions" exact>Sync</b-nav-item>
                        <b-nav-item to="/failed-job">Failed Job</b-nav-item>
                        <b-nav-item to="/setting">Setting</b-nav-item>
                    </b-navbar-nav>
                    <ul class="navbar-nav flex-row ml-md-auto d-none d-md-flex">
                        <li class="nav-item">
                            <a href="https://github.com/mylxsw/sync" class="text-white">{{ version }}</a>
                        </li>
                    </ul>
                </b-collapse>
            </b-navbar>
            <div class="main-view">
                <router-view/>
            </div>
        </b-container>

    </div>
</template>

<script>
    import axios from 'axios';
    export default {
        data() {
            return {
                version: 'v-0',
            }
        },
        mounted() {
            this.version = 'v-1000101';
        },
        beforeMount() {
            axios.defaults.baseURL = this.$store.getters.serverUrl;
            axios.interceptors.response.use(function (response) {return response;}, function (error) {
                return Promise.reject(error);
            });
        }
    }
</script>

<style>
    .container-fluid {
        padding: 0;
    }

    .main-view {
        padding: 15px;
    }
</style>
