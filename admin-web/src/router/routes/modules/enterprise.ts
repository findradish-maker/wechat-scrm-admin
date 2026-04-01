import type { RouteRecordRaw } from 'vue-router';

import { IFrameView } from '#/layouts';

const routes: RouteRecordRaw[] = [
  {
    name: 'EnterpriseOverview',
    path: '/enterprise/overview',
    meta: {
      icon: 'lucide:layout-dashboard',
      order: -2,
      title: '账号管理',
    },
    alias: ['/dashboard/home', '/dashboard/overview'],
    component: () => import('#/views/enterprise/overview/index.vue'),
  },
  {
    name: 'EnterpriseMessages',
    path: '/enterprise/messages',
    alias: '/dashboard/messages',
    component: () => import('#/views/enterprise/messages/index.vue'),
    meta: {
      fullPathKey: false,
      icon: 'lucide:messages-square',
      order: -1,
      title: '消息中心',
    },
  },
  {
    name: 'EnterpriseContacts',
    path: '/enterprise/contacts',
    alias: '/dashboard/contacts',
    component: () => import('#/views/enterprise/contacts/index.vue'),
    meta: {
      icon: 'lucide:contact-round',
      order: 0,
      title: '联系人管理',
    },
  },
  {
    name: 'EnterpriseGroups',
    path: '/enterprise/groups',
    alias: '/dashboard/groups',
    component: () => import('#/views/enterprise/groups/index.vue'),
    meta: {
      icon: 'lucide:users-round',
      order: 1,
      title: '群管理',
    },
  },
  {
    name: 'EnterpriseOfficial',
    path: '/enterprise/official-accounts',
    alias: '/dashboard/official-accounts',
    component: () => import('#/views/enterprise/official-accounts/index.vue'),
    meta: {
      icon: 'lucide:megaphone',
      order: 2,
      title: '公众号操作',
    },
  },
  {
    name: 'EnterpriseMiniPrograms',
    path: '/enterprise/mini-programs',
    alias: '/dashboard/mini-programs',
    component: () => import('#/views/enterprise/mini-programs/index.vue'),
    meta: {
      icon: 'lucide:app-window',
      order: 3,
      title: '小程序操作',
    },
  },
  {
    name: 'EnterpriseMoments',
    path: '/enterprise/moments',
    alias: '/dashboard/moments',
    component: () => import('#/views/enterprise/moments/index.vue'),
    meta: {
      icon: 'lucide:images',
      order: 4,
      title: '朋友圈管理',
    },
  },
  {
    name: 'EnterpriseFinder',
    path: '/enterprise/finder',
    alias: '/dashboard/finder',
    component: () => import('#/views/enterprise/finder/index.vue'),
    meta: {
      icon: 'lucide:clapperboard',
      order: 5,
      title: '视频号管理',
    },
  },
  {
    name: 'EnterpriseProfile',
    path: '/enterprise/profile',
    alias: '/dashboard/profile',
    component: () => import('#/views/enterprise/profile/index.vue'),
    meta: {
      icon: 'lucide:user-round-cog',
      order: 6,
      title: '个人信息管理',
    },
  },
  {
    name: 'EnterpriseFavorites',
    path: '/enterprise/favorites',
    alias: '/dashboard/favorites',
    component: () => import('#/views/enterprise/favorites/index.vue'),
    meta: {
      icon: 'lucide:bookmark',
      order: 7,
      title: '收藏管理',
    },
  },
  {
    name: 'EnterpriseApiOverview',
    path: '/enterprise/interface-overview',
    component: IFrameView,
    meta: {
      icon: 'lucide:waypoints',
      link: 'https://www.baidu.com',
      openInNewWindow: true,
      order: 8,
      title: '微信接口全览',
    },
  },
];

export default routes;
