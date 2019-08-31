import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    serverUrl: localStorage.getItem('server_url') || 'http://localhost:8819',
  },
  mutations: {
    updateServerUrl: (state, url) => {
      state.serverUrl = url;
      localStorage.setItem('server_url', url);
    }
  },
  getters: {
    serverUrl: (state) => state.serverUrl
  },
  actions: {

  }
})
