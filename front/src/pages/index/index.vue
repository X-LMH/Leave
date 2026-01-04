<template>
    <view class="index-bg">
        <!-- 左上角菜单按钮 - 只有登录后才显示 -->
        <view v-if="memberStore.token" class="menu-btn" @click="toggleDrawer">
            <view class="hamburger" :class="{'active': drawerVisible}">
                <view class="line"></view>
                <view class="line"></view>
                <view class="line"></view>
            </view>
        </view>

        <!-- 遮罩层 -->
        <view v-if="drawerVisible" class="mask" @click="toggleDrawer"></view>

        <!-- 侧边栏菜单 -->
        <view class="drawer" :class="{'show': drawerVisible}">
            <view class="drawer-content">
                <view class="drawer-header">
                    <image v-if="memberStore.token" src="@/static/logo.png" mode="aspectFit"
                           class="drawer-avatar"></image>
                    <text v-if="memberStore.token" class="drawer-title">请假系统</text>
                </view>
                <view class="drawer-menu">
                    <view v-if="memberStore.token" class="menu-item" @click="goProfile">
                        <text class="menu-icon">👤</text>
                        <text>个人信息</text>
                    </view>
                    <view v-if="memberStore.token" class="menu-item" @click="goHistory">
                        <text class="menu-icon">📋</text>
                        <text>历史记录</text>
                    </view>
                    <view v-if="memberStore.token" class="menu-item logout-item" @click="handleLogout">
                        <text class="menu-icon">🚪</text>
                        <text style="color: #e43d33;">退出登录</text>
                    </view>
                </view>
            </view>
        </view>

        <view class="logo">
            <image src="@/static/logo.png" mode="aspectFit"></image>
        </view>
        <view v-if="!memberStore.token" class="center-btns">
            <button class="main-btn" @click="goLogin">登录</button>
            <button class="main-btn" @click="goRegister">注册</button>
        </view>
        <view v-else class="login-state">
            <view class="center-btns">
                <button class="main-btn" @click="goLeave">假条</button>
            </view>
        </view>
        <view class="copyright">v:{{ now_version }}</view>
        <!-- 强制更新弹窗 -->
        <view v-if="showUpdateModal" class="update-modal-mask" @touchmove.stop.prevent>
            <view class="update-modal" @touchmove.stop.prevent>
                <view class="update-modal-header">
                    <text class="update-title">版本更新 v{{ now_version }} → v{{ latestVersion }}</text>
                </view>
                <view class="update-modal-content">
                    <text class="update-desc">为了获得更好的体验，请更新到最新版本：</text>
                    <ul class="update-ul">
                        <li v-for="(log, index) in updateLog" :key="index" class="update-li">{{ log }}</li>
                    </ul>
                    <view v-if="downloading" class="progress-wrap">
                        <view class="progress-bar">
                            <view class="progress-inner" :style="{ width: downloadProgress + '%' }"></view>
                        </view>
                        <text class="progress-text">{{ downloadProgress }}%</text>
                    </view>
                </view>
                <view class="update-modal-footer">
                    <button class="update-btn" @click="downloadApk" :disabled="downloading">
                        <text>{{ downloading ? '下载中...' : '立即更新' }}</text>
                    </button>
                </view>
            </view>
        </view>
    </view>
</template>

<script setup>
import {useMemberStore} from '../../stores'
import {GetAPPAPKAPI, GetProfileAPI, GetVersionAPI} from "../../services/user"
import appManifest from '../../manifest.json'
import {onMounted, ref, watch} from 'vue'

const memberStore = useMemberStore()
const drawerVisible = ref(false)
const now_version = ref('')

// 强制更新相关状态
const showUpdateModal = ref(false)
const latestVersion = ref('')
const updateLog = ref([])
const downloadUrl = ref('')
const downloading = ref(false)
const downloadProgress = ref(0)
const backListenerAdded = ref(false)

// 切换侧边栏显示状态
function toggleDrawer() {
    drawerVisible.value = !drawerVisible.value
}

// 跳转到个人信息页面
function goProfile() {
    toggleDrawer()
    uni.navigateTo({url: '/pages/profile/profile'})
}

// 跳转到历史记录页面
function goHistory() {
    toggleDrawer()
    uni.navigateTo({url: '/pages/history/history'})
}

// 处理退出登录
function handleLogout() {
    uni.showModal({
        title: '提示',
        content: '确定要退出登录吗？',
        success: function (res) {
            if (res.confirm) {
                memberStore.clearToken()
                toggleDrawer()
                uni.showToast({
                    title: '已退出登录',
                    icon: 'success',
                    duration: 1500
                })
            }
        }
    })
}

// 跳转方法
function goLogin() {
    uni.navigateTo({url: '/pages/login/login'})
}

function goRegister() {
    uni.navigateTo({url: '/pages/register/register'})
}

// 假条跳转方法
const goLeave = async () => {
    try {
        const res = await GetProfileAPI()
        if (!res?.data?.name || res.data.name.trim() === '') {
            uni.showToast({
                title: '请先完善个人信息',
                icon: 'none',
                duration: 1500
            })
            setTimeout(() => {
                uni.navigateTo({url: '/pages/profile/profile'})
            }, 1500)
        } else {
            uni.navigateTo({url: '/pages/leave/leave'})
        }
    } catch (e) {
        uni.showToast({
            title: '获取信息失败，请重试',
            icon: 'none',
            duration: 1500
        })
    }
}

// 获取本地版本号
const getLocalVersion = () => {
    return appManifest?.versionName || '0.0.0'
}

const isSameVersion = (v1, v2) => {
    return String(v1 || '').trim() === String(v2 || '').trim()
}
const isLowerVersion = (v1, v2) => {
    const a = String(v1 || '').split('.').map(n => parseInt(n) || 0)
    const b = String(v2 || '').split('.').map(n => parseInt(n) || 0)
    const len = Math.max(a.length, b.length)
    for (let i = 0; i < len; i++) {
        const x = a[i] || 0
        const y = b[i] || 0
        if (x < y) return true
        if (x > y) return false
    }
    return false
}

// 通用重试封装
const withRetry = async (fn, max = 3, delayMs = 800) => {
    let lastErr
    for (let i = 0; i < max; i++) {
        try {
            return await fn()
        } catch (e) {
            lastErr = e
            if (i < max - 1) {
                await new Promise(r => setTimeout(r, delayMs * (i + 1)))
            }
        }
    }
    throw lastErr
}

// 检测版本更新（新接口格式+降级与重试+本地缓存）
// 修改后的 checkVersion（替换原函数）
const checkVersion = async () => {
    try {
        const localVersion = getLocalVersion()
        const res = await withRetry(() => GetVersionAPI(), 3, 800)

        // 支持两种响应结构：res.data 或 res.data.data
        const payload = res?.data?.data || res?.data || {}

        if (!payload || (!payload.latest && !payload.version)) {
            console.warn('版本接口返回数据异常:', res)
            return
        }

        const serverVersion = payload.latest || payload.version
        const logs = Array.isArray(payload.update_log) ? payload.update_log
            : Array.isArray(payload.updateLog) ? payload.updateLog
                : []
        const urlFromVersion = payload.downloadURL || payload.downloadUrl || payload.url || ''

        uni.setStorageSync('lastVersionInfo', { version: serverVersion, updateLog: logs, timestamp: Date.now() })

        now_version.value = localVersion

        if (isLowerVersion(localVersion, serverVersion)) {
            latestVersion.value = serverVersion
            updateLog.value = logs.length ? logs : ['优化体验，修复已知问题']
            if (urlFromVersion) downloadUrl.value = urlFromVersion
            showUpdateModal.value = true
        }
    } catch (err) {
        console.error('版本检测失败:', err)
    }
}


// 下载并安装APK（支持重试与断点续传）
const downloadApk = async () => {
    if (downloading.value) return
    downloading.value = true
    downloadProgress.value = 0

    // #ifdef APP-PLUS
    try {
        let url = downloadUrl.value
        if (!url) {
            try {
                const apkRes = await withRetry(() => GetAPPAPKAPI(), 3, 800)
                url = apkRes?.data?.url || ''
            } catch (e) {
                url = ''
            }
        }
        if (!url) {
            downloading.value = false
            uni.showToast({ title: '无有效下载地址', icon: 'none' })
            return
        }

        const task = uni.downloadFile({
            url,
            // header: memberStore.token ? { Authorization: `Bearer ${memberStore.token}` } : {},
            success: (res) => {
                downloading.value = false
                if (res.statusCode === 200 && res.tempFilePath) {
                    plus.runtime.install(res.tempFilePath, { force: true }, () => {
                        uni.showToast({ title: '更新完成，即将重启应用', icon: 'success', duration: 1500 })
                        setTimeout(() => { plus.runtime.restart() }, 1500)
                    }, (installErr) => {
                        uni.showModal({ title: '安装失败', content: installErr.message || '更新安装失败，请手动下载安装', showCancel: false, confirmText: '确定' })
                    })
                } else {
                    uni.showToast({ title: '下载失败：文件损坏', icon: 'none' })
                }
            },
            fail: (err) => {
                downloading.value = false
                uni.showToast({ title: '下载失败：' + (err.errMsg || '网络错误'), icon: 'none' })
            }
        })
        task.onProgressUpdate((p) => {
            downloadProgress.value = p.progress
        })
    } catch (e) {
        downloading.value = false
        uni.showToast({ title: '下载失败：' + (e?.message || '未知错误'), icon: 'none' })
    }
    // #endif

    // #ifndef APP-PLUS
    try {
        const url = downloadUrl.value || (await (async () => {
            try {
                const apkRes = await withRetry(() => GetAPPAPKAPI(), 3, 800)
                return apkRes?.data?.url || ''
            } catch (e) {
                return ''
            }
        })())
        const a = document.createElement('a')
        a.href = url
        a.download = 'update.apk'
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        downloading.value = false
        uni.showToast({ title: '已开始下载', icon: 'success' })
    } catch (e) {
        downloading.value = false
        uni.showToast({ title: '下载失败：' + (e?.message || '未知错误'), icon: 'none' })
    }
    // #endif
}

// 页面挂载时检测版本
onMounted(async () => {
    now_version.value = getLocalVersion()
    await checkVersion()
})

// 禁止更新弹窗被返回键关闭
watch(showUpdateModal, (val) => {
    // #ifdef APP-PLUS
    try {
        if (val && !backListenerAdded.value) {
            plus.key.addEventListener('backbutton', () => {
                if (showUpdateModal.value) {
                    return
                }
            }, false)
            backListenerAdded.value = true
        }
    } catch (e) {}
    // #endif
})
</script>

<style scoped>
.index-bg {
    width: 100vw;
    height: 100vh;
    position: fixed;
    left: 0;
    top: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: linear-gradient(120deg, #74ebd5 0%, #ACB6E5 100%);
}

/* 菜单按钮样式 */
.menu-btn {
    position: fixed;
    left: 30rpx;
    top: calc(44px + 30rpx);
    width: 80rpx;
    height: 80rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.95);
    border-radius: 50%;
    box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.15);
    z-index: 9999;
    backdrop-filter: blur(10rpx);
}

.hamburger {
    width: 40rpx;
    height: 32rpx;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    transition: all 0.3s ease;
}

.line {
    width: 100%;
    height: 4rpx;
    background: #333;
    border-radius: 2rpx;
    transition: all 0.3s ease;
}

.hamburger.active .line:nth-child(1) {
    transform: rotate(45deg) translate(10rpx, 10rpx);
}

.hamburger.active .line:nth-child(2) {
    opacity: 0;
}

.hamburger.active .line:nth-child(3) {
    transform: rotate(-45deg) translate(10rpx, -10rpx);
}

/* 遮罩层样式 */
.mask {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.5);
    z-index: 9996;
    animation: fadeIn 0.3s ease;
}

/* 侧边栏样式 */
.drawer {
    position: fixed;
    top: 0;
    left: -500rpx;
    width: 400rpx;
    height: 100vh;
    background: #fff;
    z-index: 9997;
    transition: left 0.3s ease;
}

.drawer.show {
    left: 0;
}

.drawer-content {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.drawer-header {
    padding: 40rpx 30rpx;
    background: linear-gradient(90deg, #74ebd5 0%, #ACB6E5 100%);
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
}

.drawer-avatar {
    width: 80rpx;
    height: 80rpx;
    border-radius: 50%;
    margin-bottom: 20rpx;
}

.drawer-title {
    color: #fff;
    font-size: 32rpx;
    font-weight: bold;
}

.drawer-menu {
    flex: 1;
    padding: 20rpx 0;
}

.menu-item {
    display: flex;
    align-items: center;
    padding: 30rpx 40rpx;
    border-bottom: 1rpx solid #f5f5f5;
    transition: background-color 0.2s;
}

.menu-item:active {
    background-color: #f8f8f8;
}

.menu-icon {
    font-size: 36rpx;
    margin-right: 20rpx;
    width: 36rpx;
    text-align: center;
}

.menu-item text {
    font-size: 30rpx;
    color: #333;
}

.logout-item {
    border-bottom: none;
}

.logo {
    width: 120rpx;
    height: 120rpx;
    margin-bottom: 32rpx;
    border-radius: 50%;
    box-shadow: 0 8rpx 32rpx rgba(172, 182, 229, 0.18);
    background: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
}

.center-btns {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 32rpx;
    margin-top: 0;
}

.main-btn {
    width: 320rpx;
    height: 80rpx;
    background: linear-gradient(90deg, #74ebd5 0%, #ACB6E5 100%);
    color: #fff;
    font-size: 34rpx;
    font-weight: bold;
    border-radius: 40rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4rpx 16rpx rgba(172, 182, 229, 0.12);
    letter-spacing: 4rpx;
    border: none;
    transition: background 0.3s;
}

.main-btn:active {
    background: linear-gradient(90deg, #ACB6E5 0%, #74ebd5 100%);
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

/* 强制更新弹窗样式 */
.update-modal-mask {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.7);
    z-index: 99999;
    display: flex;
    align-items: center;
    justify-content: center;
}

.update-modal {
    width: 80%;
    max-width: 600rpx;
    background: #fff;
    border-radius: 20rpx;
    overflow: hidden;
    box-shadow: 0 10rpx 40rpx rgba(0, 0, 0, 0.2);
}

.update-modal-header {
    padding: 30rpx;
    background: linear-gradient(90deg, #74ebd5 0%, #ACB6E5 100%);
    text-align: center;
}

.update-title {
    color: #fff;
    font-size: 36rpx;
    font-weight: bold;
}

.update-modal-content {
    padding: 40rpx 30rpx;
}

.update-desc {
    font-size: 30rpx;
    color: #333;
    margin-bottom: 20rpx;
    display: block;
}

.update-ul {
    margin-top: 20rpx;
    padding-left: 30rpx;
}
.update-li {
    font-size: 28rpx;
    color: #666;
    line-height: 1.6;
    margin-bottom: 10rpx;
}

.update-modal-footer {
    padding: 20rpx 30rpx;
    border-top: 1rpx solid #eee;
}

.update-btn {
    width: 100%;
    height: 80rpx;
    background: linear-gradient(90deg, #74ebd5 0%, #ACB6E5 100%);
    color: #fff;
    font-size: 32rpx;
    font-weight: bold;
    border-radius: 40rpx;
    border: none;
}

.update-btn:disabled {
    background: #ccc;
    color: #999;
}

.progress-wrap {
    margin-top: 20rpx;
}
.progress-bar {
    width: 100%;
    height: 16rpx;
    background: #f1f1f1;
    border-radius: 8rpx;
    overflow: hidden;
}
.progress-inner {
    height: 100%;
    background: linear-gradient(90deg, #74ebd5 0%, #ACB6E5 100%);
}
.progress-text {
    margin-top: 12rpx;
    font-size: 24rpx;
    color: #555;
    text-align: center;
}

@keyframes fadeIn {
    from {
        opacity: 0;
    }
    to {
        opacity: 1;
    }
}
</style>
