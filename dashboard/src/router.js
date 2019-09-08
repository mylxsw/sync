import Vue from 'vue';
import Router from 'vue-router';
import History from './views/History';
import Queue from './views/Queue';
import FailedJob from "./views/FailedJob";
import Sync from "./views/Sync";
import Setting from "./views/Setting";
import Job from "./views/Job";

Vue.use(Router);

export default new Router({
    routes: [
        {path: '/', component: History},
        {path: '/queue', component: Queue},
        {path: '/failed-job', component: FailedJob},
        {path: '/sync/definitions', component: Sync},
        {path: '/setting', component: Setting},
        {path: '/jobs/:id/', component: Job},
    ]
});
