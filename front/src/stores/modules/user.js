import {defineStore} from 'pinia'
import {ref} from 'vue'

// 定义 Store
export const useMemberStore = defineStore(
    'user',
    () => {

        const token = ref('')

        const clearToken = () => {
            token.value = ''
        }
        return {
            token,

            clearToken,
        }
    },
    // TODO: 持久化
    {
        // 网页端
        // persist: true,
        // 小程序需要修改
        persist: {
            storage: {
                getItem(key) {
                    return uni.getStorageSync(key)
                },
                setItem(key, value) {
                    uni.setStorageSync(key, value)
                },
            },
        },
    },
)
