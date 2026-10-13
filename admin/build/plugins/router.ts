import type { RouteMeta } from 'vue-router';
import ElegantVueRouter from '@elegant-router/vue/vite';
import type { RouteKey } from '@elegant-router/types';

export function setupElegantRouter() {
  return ElegantVueRouter({
    layouts: {
      base: 'src/layouts/base-layout/index.vue',
      blank: 'src/layouts/blank-layout/index.vue'
    },
    routePathTransformer(routeName, routePath) {
      const key = routeName as RouteKey;

      if (key === 'login') {
        return '/login';
      }

      return routePath;
    },
    onRouteMetaGen(routeName) {
      const key = routeName as RouteKey;

      const constantRoutes: RouteKey[] = ['login', '403', '404', '500'];

      const meta: Partial<RouteMeta> = {
        title: key,
        i18nKey: `route.${key}` as App.I18n.I18nKey
      };

      if (key === 'home') {
        meta.icon = 'mdi:monitor-dashboard';
        meta.order = 1;
      }

      if (key === 'account-settings') {
        meta.icon = 'mdi:account-cog-outline';
        meta.hideInMenu = true;
      }

      if (key === 'feedback' || key === 'apartment') {
        meta.icon = key === 'feedback' ? 'mdi:message-text-outline' : 'mdi:office-building-outline';
        meta.order = key === 'feedback' ? 7 : 5;
      }

      if (key === 'app-version') {
        meta.icon = 'mdi:package-variant-closed';
        meta.order = 8;
      }

      if (key === 'leave-type' || key === 'class') {
        meta.icon = key === 'leave-type' ? 'mdi:format-list-bulleted' : 'mdi:google-classroom';
        meta.order = key === 'class' ? 4 : 6;
      }

      if (key === 'student' || key === 'leave-record') {
        meta.icon = key === 'student' ? 'mdi:account-school' : 'mdi:clipboard-text-clock-outline';
        meta.order = key === 'student' ? 3 : 2;
      }

      if (constantRoutes.includes(key)) {
        meta.constant = true;
      }

      return meta;
    }
  });
}
