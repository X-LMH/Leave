<template>
    <view class="nav-container" :style="containerStyle">
        <!-- 状态栏占位区域 -->
        <view :style="{ height: statusBarHeight + 'px' }"></view>

        <!-- 自定义导航栏 -->
        <view class="custom-nav-bar">
            <!-- 左侧返回按钮 -->
            <view class="nav-left" @click="onClickLeft">
                <image src="/static/images/back.png" mode="widthFix" class="back-img"></image>
            </view>

            <!-- 标题 -->
            <view class="nav-title">{{ name }}-{{ date }}</view>

            <!-- 右侧按钮 -->
            <view class="nav-right">
                <image src="/static/images/a.png" mode="widthFix" class="right-img"></image>
            </view>
        </view>
    </view>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'

const statusBarHeight = ref(0)

const props = defineProps({
    name: {
        type: String,
    },
    date: {
        type: String,
        default: ''
    }
})
// 去除 floating prop、emit 与高度计算相关逻辑

onMounted(() => {
    const info = uni.getSystemInfoSync()
    if (info.platform === 'android') {
        statusBarHeight.value = info.statusBarHeight
        uni.setNavigationBarColor({
            frontColor: '#000000',
            backgroundColor: '#ffffff'
        })
    } else {
        statusBarHeight.value = info.statusBarHeight || 0
    }
    // 不再计算 nav 高度或 emit 给父组件
})

const containerStyle = computed(() => {
    return {
        backgroundColor: '#ffffff',
        width: '100%'
    }
})

const onClickLeft = () => {
    uni.reLaunch({
        url: '/pages/index/index'
    });
};
</script>

<style scoped>
.nav-container {
    /* background-color: #ffffff; 由 containerStyle 管理 */
    /* width: 100%; */
}

.custom-nav-bar {
    display: flex;
    align-items: center;
    height: 90rpx;
    box-sizing: border-box;
    padding-left: env(safe-area-inset-left);
    padding-right: env(safe-area-inset-right);
}

.nav-left {
    width: 76rpx;
    height: 100%;
    display: flex;
    align-items: center;
    padding-left: 23rpx; /* 右挪2rpx */
    box-sizing: border-box;
}

.back-img {
    width: 36rpx;
    height: 36rpx;
    object-fit: contain;
    /* 变为黑色 */
    filter: brightness(0) saturate(100%);
}

/* 标题字体大小从36rpx减��到32rpx，且加粗 */
.nav-title {
    color: #000000;
    font-size: 32rpx;
    font-weight: 500; /* 加粗少一点 */
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    margin-left: 14rpx;
}

.nav-right {
    width: 76rpx;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding-right: 14rpx;
    box-sizing: border-box;
}

.right-img {
    width: 36rpx;
    height: 36rpx;
    object-fit: contain;
    /* 变为黑色 */
    filter: brightness(0) saturate(100%);
}

</style>
