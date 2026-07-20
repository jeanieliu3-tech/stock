import type { RouteRecordRaw } from 'vue-router'

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue'),
    meta: { title: '首页', icon: 'home' },
  },
  {
    path: '/rank',
    name: 'rank',
    component: () => import('@/views/RankView.vue'),
    meta: { title: '排名', icon: 'rank' },
  },
  {
    path: '/position',
    name: 'position',
    component: () => import('@/views/PositionView.vue'),
    meta: { title: '持仓', icon: 'position' },
  },
  {
    path: '/strategies',
    name: 'strategies',
    component: () => import('@/views/StrategiesView.vue'),
    meta: { title: '策略', icon: 'strategies' },
  },
  {
    path: '/settings',
    redirect: '/strategies',
  },
]
