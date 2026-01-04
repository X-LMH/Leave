<template>
    <view class="login-bg">
        <view class="logo">
            <image src="@/static/logo.png" mode="aspectFit"></image>
        </view>
        <view class="title">欢迎登录</view>
        <uni-forms ref="formRef" class="login-card custom-form" v-model="form" :rules="rules">
            <uni-forms-item name="student_id">
                <uni-easyinput class="input input-first" type="text" placeholder="学号" v-model="form.student_id"/>
            </uni-forms-item>
            <uni-forms-item name="password">
                <uni-easyinput class="input" type="password" placeholder="密码" v-model="form.password"/>
            </uni-forms-item>
            <button class="login-btn" :disabled="loading" @click="handleLogin">
                {{ loading ? '登录中...' : '登 录' }}
            </button>
            <navigator class="register-link" url="/pages/register/register">没有账号？去注册</navigator>
        </uni-forms>
        <view class="copyright">© 2024 YourAppName</view>
    </view>
</template>

<script setup>
import { ref } from 'vue'
import { useMemberStore } from '../../stores'
import {onLoad} from "@dcloudio/uni-app";
import {LoginAPI} from "../../services/user";

const form = ref({
    student_id: '',
    password: ''
})
const loading = ref(false)
const formRef = ref(null)

const rules = {
    student_id: {
        rules: [
            { required: true, errorMessage: '请输入学号' },
            { minLength: 12, maxLength: 12, errorMessage: '学号必须为12位' }
        ]
    },
    password: {
        rules: [
            { required: true, errorMessage: '请输入密码' },
            { minLength: 6, errorMessage: '密码长度至少6位' }
        ]
    }
}

// 页面加载时自动填入学号
onLoad((options) => {
    if (options.student_id) {
        form.value.student_id = options.student_id
    }
})

const handleLogin = async () => {
    formRef.value.validate().then(async () => {
        loading.value = true
        try {
            const res = await LoginAPI(form.value)
            // 保存用户信息
            const memberStore = useMemberStore()
            memberStore.token = res.data
            await uni.showToast({title: '登录成功', icon: 'success'})
            setTimeout(() => {
                uni.redirectTo({ url: '/pages/index/index' })
            }, 800)
        } catch (e) {
            form.value.password = ''
        } finally {
            loading.value = false
        }
    })
}
</script>

<style scoped>
.custom-form ::v-deep .uni-forms-item__error {
    text-align: right !important; /* 强制靠右对齐 */
    padding-right: 0;
    /* 如果输入框有固定宽度，可统一设置提示宽度与之匹配 */
    width: 60%;
}
.login-bg {
    position: fixed;
    left: 0;
    top: 0;
    width: 100vw;
    height: 100vh;
    background: linear-gradient(120deg, #74ebd5 0%, #ACB6E5 100%);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: flex-start;
    overflow: hidden;
    padding-top: 120rpx;
}

.logo {
    width: 120rpx;
    height: 120rpx;
    margin-bottom: 32rpx;
    border-radius: 50%; /* 圆形logo */
    box-shadow: 0 8rpx 32rpx rgba(172, 182, 229, 0.18);
    background: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
}

.title {
    font-size: 44rpx;
    color: #333;
    font-weight: 600;
    margin-bottom: 48rpx;
    letter-spacing: 2rpx;
    text-shadow: 0 2rpx 8rpx rgba(172, 182, 229, 0.12);
}

.login-card {
    width: 540rpx;
    padding: 0 0 40rpx 0;
    background: rgba(255, 255, 255, 0.85);
    border-radius: 40rpx; /* 更大圆角 */
    box-shadow: 0 8rpx 32rpx rgba(172, 182, 229, 0.10);
    display: flex;
    flex-direction: column;
    align-items: center;
}

.input {
    width: 95%;
    height: 80rpx;
    margin: 10rpx auto 0 auto;
    border: none;
    border-radius: 32rpx; /* 与按钮一致 */
    background: #f4f8fb;
    font-size: 34rpx;     /* 与按钮一致 */
    padding: 0 24rpx;
    box-sizing: border-box;
    color: #333;
    box-shadow: 0 2rpx 8rpx rgba(172, 182, 229, 0.08);
    display: flex;
    align-items: center;
}
.input-first {
    margin: 48rpx auto 0 auto !important;
}

.input :deep(.uni-easyinput__content-input) {
    background: transparent;
    border: none;
    font-size: 34rpx;     /* 与按钮一致 */
    color: #333;
    padding: 0;
    height: 80rpx;
    line-height: 80rpx;
    border-radius: 32rpx; /* 与按钮一致 */
}

.input :deep(.uni-easyinput__content) {
    background: transparent;
    border: none;
    box-shadow: none;
    padding: 0;
    height: 80rpx;
    border-radius: 32rpx; /* 与按钮一致 */
}

.input:focus,
.input :deep(.uni-easyinput__content-input:focus) {
    background: #eaf6ff;
}

.login-btn {
    width: 90%;
    height: 80rpx;
    background: linear-gradient(90deg, #74ebd5 0%, #ACB6E5 100%);
    color: #fff;
    font-size: 34rpx;
    font-weight: bold;
    border: none;
    border-radius: 32rpx; /* 更大圆角 */
    box-shadow: 0 4rpx 16rpx rgba(172, 182, 229, 0.12);
    margin: 32rpx auto 0 auto;
    transition: background 0.3s;
    letter-spacing: 4rpx;
}

.login-btn:active {
    background: linear-gradient(90deg, #ACB6E5 0%, #74ebd5 100%);
}

.register-link {
    margin-top: 24rpx;
    color: #2193b0;
    font-size: 28rpx;
    text-align: center;
    cursor: pointer;
    text-decoration: underline;
    transition: color 0.2s;
}

.register-link:active {
    color: #74ebd5;
}

.copyright {
    position: absolute;
    bottom: 32rpx;
    width: 100%;
    text-align: center;
    color: #888;
    font-size: 24rpx;
    letter-spacing: 2rpx;
}
</style>
