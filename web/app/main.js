import Vue from 'vue'

import App from './components/App'

import truncate from 'vue-truncate'

Vue.use(truncate);
Vue.config.devtools = true;
// mount a root Vue instance
new Vue({
  el: 'body',
  components: {
    app: App
  }
})