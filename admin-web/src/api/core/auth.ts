import { requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
  }

  export interface BackendLoginResult {
    expiresAt: string;
    token: string;
    user: {
      id: number;
      name: string;
      status: string;
      username: string;
    };
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  const result = await requestClient.post<AuthApi.BackendLoginResult>(
    '/auth/login',
    data,
  );
  return {
    accessToken: result.token,
  } satisfies AuthApi.LoginResult;
}

/**
 * 刷新accessToken
 */
export async function refreshTokenApi() {
  // 当前企业后台后端未提供 refresh token 接口，刷新策略关闭。
  return { data: '' };
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return Promise.resolve();
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi() {
  // 前端路由权限模式，先提供基础权限集。
  return ['super'];
}
