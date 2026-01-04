import { useMemberStore } from '../stores'

// const baseURL = 'http://localhost:8080/'
const baseURL = 'http://39.104.89.165:8080/'

// 请求拦截器
const requestInterceptor = {
    invoke(options) {
        // 1. 拼接基础地址
        if (!options.url.startsWith('http')) {
            options.url = baseURL + options.url
        }
        // 2. 设置超时时间
        options.timeout = 10000
        // 3. 处理请求头和 token
        if (!options.header || typeof options.header !== 'object') {
            options.header = {}
        }
        const memberStore = useMemberStore()
        const token = memberStore.token
        if (token) {
            options.header.Authorization = `Bearer ${token}`
        }
        return options
    },
}

// 注册拦截器
uni.addInterceptor('request', requestInterceptor)
uni.addInterceptor('uploadFile', requestInterceptor)

export const http = (options) => {
    return new Promise((resolve, reject) => {
        uni.request({
            ...options,
            success(res) {
                if (res.data.code === 0) {
                    resolve(res.data)
                } else if (res.data.code === 7) {
                    // 登录过期处理
                    const memberStore = useMemberStore()
                    memberStore.clearToken()
                    uni.navigateTo({ url: '/pages/login/login' })
                    uni.showToast({
                        icon: 'none',
                        title: '登录已过期，请重新登录',
                    })
                    reject(res)
                } else {
                    // 其他业务错误
                    uni.showToast({
                        icon: 'none',
                        title: res.data.message || '请求失败',
                    })
                    reject(res)
                }
            },
            fail(err) {
                // 网络错误处理
                console.error('网络请求失败:', err)
                uni.showToast({
                    icon: 'none',
                    title: '网络错误，无法连接服务器',
                })
                reject(err)
            },
        })
    })
}
