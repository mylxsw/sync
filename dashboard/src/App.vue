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
                        <b-nav-item to="/errors">Error</b-nav-item>
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
            axios.get('/api/').then(response => {
                this.version = response.data.version;
            });
        },
        beforeMount() {
            axios.defaults.baseURL = this.$store.getters.serverUrl;
            let token = this.$store.getters.token;
            if (token !== "") {
                axios.defaults.headers.common['Authorization'] = "Bearer " + token;
            }
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
