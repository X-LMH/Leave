<template>
    <view class="page">
        <view class="status-bar bg-primary"></view>

        <scroll-view class="content" scroll-y="true">
            <view class="list-wrap">
                <view
                    v-for="(item, index) in items"
                    :key="item.id"
                    class="swipe-container"
                    :class="{ opened: swipeStatus[index] === 'opened' }"
                    @touchstart="handleTouchStart($event, index)"
                    @touchend="handleTouchEnd($event, index)"
                >
                    <!-- 卡片内容 -->
                    <view class="card" @click="onListClick(item, index)">
                        <view class="card-body">
                            <view class="card-header">
                                <span class="card-title">{{ item.name }}-{{ item.date }}</span>
                            </view>
                            <view class="meta-row">
                                <view class="meta-left">
                                    <i class="fa fa-calendar-o icon"></i>
                                    <span class="meta-text">本科生外出请假</span>
                                </view>
                                <view class="meta-right">
                                    <i class="fa fa-clock-o icon"></i>
                                    <span class="meta-text">{{ item.time }}</span>
                                </view>
                            </view>
                        </view>
                    </view>

                    <!-- 删除按钮 -->
                    <view class="delete-btn" @click.stop="confirmDelete(index)">
                        <i class="fa fa-trash-o"></i>
                        <span>删除</span>
                    </view>
                </view>

                <view v-if="items.length === 0" class="empty">
                    <i class="fa fa-history empty-icon"></i>
                    <span class="empty-text">暂无历史记录</span>
                </view>

                <view v-if="items.length > 0" class="end-hint">
                    <i class="fa fa-history end-icon"></i>
                    <span class="end-text">以上是所有历史记录</span>
                </view>
            </view>
        </scroll-view>
    </view>
</template>

<script setup>
import { ref } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import dayjs from 'dayjs';
import { GetRecordListAPI,DeleteRecordAPI,GetRecordDetailAPI } from '../../services/user';

const items = ref([]);
// 简化状态管理，只记录 'closed' 或 'opened'
const swipeStatus = ref([]);
const touchStartX = ref(0);
const touchStartY = ref(0);

onLoad(async () => {
    try {
        const res = await GetRecordListAPI();
        const list = res.data;
        items.value = Array.isArray(list)
            ? list.map(item => ({
                ...item,
                date: dayjs(item.nigao).format('YYYYMMDD'),
                time: dayjs(item.nigao).format('YYYY-MM-DD HH:mm:ss')
            }))
            : [];

        // 初始化所有状态为关闭
        swipeStatus.value = Array(items.value.length).fill('closed');
    } catch (error) {
        uni.showToast({
            title: '加载失败',
            icon: 'none'
        });
    }
});

const handleTouchStart = (e, index) => {
    // 记录初始触摸点
    touchStartX.value = e.changedTouches[0].clientX;
    touchStartY.value = e.changedTouches[0].clientY;
};

const handleTouchEnd = (e, index) => {
    // 计算滑动距离
    const touchEndX = e.changedTouches[0].clientX;
    const touchEndY = e.changedTouches[0].clientY;
    const deltaX = touchStartX.value - touchEndX; // 左滑为正
    const deltaY = Math.abs(touchStartY.value - touchEndY);

    // 判断是否为有效的左滑操作
    if (deltaX > 50 && deltaX > deltaY) {
        // 关闭其他所有项
        swipeStatus.value = swipeStatus.value.map(() => 'closed');
        // 打开当前项
        swipeStatus.value[index] = 'opened';
    } else if (deltaX < -20) { // 如果是向右滑动，并且距离足够
        // 关闭当前项
        swipeStatus.value[index] = 'closed';
    }
    // 如果是点击或者滑动不明显，则不做任何操作
};

const onListClick = async (item, index) => {
    if (swipeStatus.value[index] === 'closed') {
        try {
            const res = await GetRecordDetailAPI(item.id);
            uni.setStorageSync('recordDetail', res.data);
            uni.navigateTo({ url: '/pages/leave/leave' });
        } catch (err) {
            console.error("导航失败:", err);
        }
    }
};

const confirmDelete = (index) => {
    // 先关闭滑块
    swipeStatus.value[index] = 'closed';

    uni.showModal({
        title: '确认删除',
        content: '您确定要删除这条记录吗？',
        confirmColor: '#ff4d4f',
        success: async (res) => {
            if (res.confirm) {
                try {
                    await DeleteRecordAPI(items.value[index].id);
                    items.value.splice(index, 1);
                    swipeStatus.value.splice(index, 1); // 同步删除状态
                    await uni.showToast({title: '删除成功'});
                } catch (error) {
                    await uni.showToast({title: '删除失败', icon: 'none'});
                }
            }
        }
    });
};
</script>

<style scoped>
/* 导入 font-awesome */
@import url("https://cdn.jsdelivr.net/npm/font-awesome@4.7.0/css/font-awesome.min.css");

/* 主题色与基础样式 */
:root {
    --primary: #3b82f6;
    --card-radius: 12px;
    --card-padding: 20px;
    --shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    --status-bar-height: env(safe-area-inset-top, 44px);
    font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
}

/* ==================== 恢复原始背景色 ==================== */
.page, .content, .list-wrap {
    background-color: #ffffff;
}
/* ======================================================== */

.page {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.status-bar {
    height: var(--status-bar-height);
    background-color: var(--primary);
}

.content {
    padding: 16px;
    padding-top: 12px;
    box-sizing: border-box;
    flex: 1;
}

.list-wrap {
    display: flex;
    flex-direction: column;
    gap: 12px; /* 列表项之间的间距 */
}

/* ==================== 核心滑动样式 ==================== */
.swipe-container {
    position: relative;
    width: 100%;
    height: auto;
    overflow: hidden;
    border-radius: var(--card-radius);
}

/* 卡片内容 */
.card {
    position: relative;
    z-index: 2;
    background: #fff;
    border-radius: var(--card-radius);
    box-shadow: var(--shadow);
    transition: transform 0.3s ease; /* 关键：平滑过渡 */
    cursor: pointer;
}

/* 当 .swipe-container 有 .opened 类时，移动 .card */
.swipe-container.opened .card {
    transform: translateX(-120px); /* 关键：向左移动 120px，露出删除按钮 */
}

/* 删除按钮 */
.delete-btn {
    position: absolute;
    top: 0;
    right: 0;
    width: 120px;
    height: 100%;
    background-color: #ff4d4f;
    color: white;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    z-index: 1;
    font-size: 14px;
}
/* ===================================================== */

.card-body {
    padding: var(--card-padding);
    padding-bottom: 8px;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
}

.card-title {
    font-size: 16px;
    color: #1f2937;
    font-weight: 600;
}

.status {
    padding: 3px 10px;
    border-radius: 999px;
    font-size: 12px;
    color: #10b981;
    background: #ecfdf5;
}
.status.finished { color: #10b981; }
.status.pending { color: #f59e0b; background: #fef3c7; }
.status.rejected { color: #ff4d4f; background: #fee2e2; }

.meta-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    flex-wrap: wrap;
}

.meta-left, .meta-right {
    display: flex;
    align-items: center;
    color: #6b7280;
    font-size: 14px;
}

.icon {
    font-size: 14px;
    color: #9ca3af;
    margin-right: 8px;
}

/* 空状态与结尾提示 */
.empty, .end-hint {
    text-align: center;
    padding: 28px 12px;
    color: #9ca3af;
    font-size: 14px;
}
.empty-icon, .end-icon {
    font-size: 28px;
    margin-bottom: 8px;
    opacity: 0.6;
    display: block;
}

/* 小屏适配 */
@media (max-width: 360px) {
    .card-body {
        padding: 16px;
    }
}
</style>