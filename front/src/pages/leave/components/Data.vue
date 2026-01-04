<template>
    <view style="
    display: flex;
    align-items: center;
    gap: 15rpx;
    justify-content: flex-end;
    border: 1px solid #ddd;
    border-radius: 6rpx;
    padding: 15rpx 30rpx;
    width: fit-content;
    margin-left: auto;
  ">
        <!-- 左侧日历图标 -->
        <view style="font-size: 32rpx; color: #666;">📅</view>

        <!-- 日期选择器 -->
        <picker
                mode="date"
                :value="selectedDate"
                start="1970-01-01"
                end="2050-12-31"
                @change="onDateChange"
        >
            <view style="font-size: 32rpx;">{{ selectedDate || `请选择${type}日期` }}</view>
        </picker>

        <!-- 时间选择器 -->
        <picker
                mode="time"
                :value="selectedTime"
                start="00:00"
                end="23:59"
                @change="onTimeChange"
        >
            <view style="font-size: 32rpx;">{{ selectedTime || `请选择${type}时间` }}</view>
        </picker>
    </view>
</template>

<script setup>
import { ref, watch, defineProps, defineEmits } from 'vue';

// 接收父组件参数
const props = defineProps({
    // 类型提示（如"开始"、"结束"）
    type: {
        type: String,
        default: ''
    },
    // 父组件传递的初始完整时间（格式：YYYY-MM-DD HH:mm）
    initTime: {
        type: String,
        default: ''
    }
});

// 定义向父组件传递事件
const emit = defineEmits(['update:time']);

// 选中的日期和时间
const selectedDate = ref('');
const selectedTime = ref('');

// 初始化：拆分父组件传递的初始时间
watch(() => props.initTime, (val) => {
    if (val) {
        const [date, time] = val.split(' ');
        selectedDate.value = date || '';
        selectedTime.value = time || '';
    }
}, { immediate: true });

// 监听日期和时间变化，实时向父组件传递拼接结果
watch([() => selectedDate.value, () => selectedTime.value], () => {
    if (selectedDate.value && selectedTime.value) {
        const fullTime = `${selectedDate.value} ${selectedTime.value}`;
        emit('update:time', fullTime);
    }
});

// 处理日期选择
const onDateChange = (e) => {
    selectedDate.value = e.detail.value;
};

// 处理时间选择
const onTimeChange = (e) => {
    selectedTime.value = e.detail.value;
};
</script>
