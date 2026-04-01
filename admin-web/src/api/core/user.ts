import type { UserInfo } from '@vben/types';

import { requestClient } from '#/api/request';

type BackendUserInfo = {
  id: number;
  name: string;
  status: string;
  username: string;
};

/**
 * 获取用户信息
 */
export async function getUserInfoApi() {
  const user = await requestClient.get<BackendUserInfo>('/auth/me');
  return {
    avatar: '',
    desc: user.status === 'active' ? '企业后台管理员' : '管理员账号受限',
    homePath: '/enterprise/overview',
    realName: user.name || user.username,
    roles: ['super'],
    token: '',
    userId: String(user.id),
    username: user.username,
  } satisfies UserInfo;
}
