import {createRouter, createWebHistory} from 'vue-router'
import Home from '@/views/Home'
import EndpointDetails from "@/views/EndpointDetails";
import SuiteDetails from '@/views/SuiteDetails';

const singleEndpoint = window.config?.singleEndpoint
const singleEndpointMode = singleEndpoint && singleEndpoint !== '{{ .UI.SingleEndpoint }}'

const routes = [
    {
        path: '/',
        name: 'Home',
        component: Home,
        beforeEnter: () => singleEndpointMode ? {name: 'EndpointDetails', params: {key: singleEndpoint}} : true
    },
    {
        path: '/endpoints/:key',
        name: 'EndpointDetails',
        component: EndpointDetails,
    },
    {
        path: '/suites/:key',
        name: 'SuiteDetails',
        component: SuiteDetails
    }
];

const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes
});

export default router;
