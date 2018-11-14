// jshint esversion: 6
Vue.use(Vuex);

const HomePage = httpVueLoader('/pages/home.vue');

const routes = [
  { path: '/', component: HomePage }
];

var router = new VueRouter({
  history: true,
  routes
});

const app = new Vue({
  router
}).$mount('#root');
