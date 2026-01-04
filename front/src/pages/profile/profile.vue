        <template>
    <view class="profile-bg">
        <scroll-view scroll-y class="profile-scroll">
            <view class="logo">
                <image src="@/static/logo.png" mode="aspectFit"></image>
            </view>
            <view class="title">个人信息</view>
            <uni-forms class="profile-card custom-form" ref="formRef" :rules="rules" :modelValue="profile">
                <uni-forms-item name="name" label="学生姓名">
                    <uni-easyinput v-model="profile.name" class="input input-first" type="text"
                                   placeholder="请输入学生姓名"/>
                </uni-forms-item>
                <uni-forms-item name="phone" label="学生电话">
                    <uni-easyinput v-model="profile.phone" class="input" type="text" placeholder="请输入学生电话"/>
                </uni-forms-item>
                <uni-forms-item name="gender" label="性别">
                    <picker mode="selector" :range="genderOptions.slice(1)"
                            :value="selectedGenderIndex > 0 ? selectedGenderIndex - 1 : ''"
                            @change="e => updateGender({detail: {value: Number(e.detail.value) + 1}})">
                        <view class="input input-picker">
                            <span :class="selectedGenderIndex === 0 ? 'placeholder-text' : ''">
                                {{ selectedGenderIndex === 0 ? '请选择性别' : genderOptions[selectedGenderIndex] }}
                            </span>
                        </view>
                    </picker>
                </uni-forms-item>
                <uni-forms-item name="parent_name" label="家长姓名">
                    <uni-easyinput v-model="profile.parent_name" class="input" type="text" placeholder="请输入家长姓名"/>
                </uni-forms-item>
                <uni-forms-item name="parent_phone" label="家长电话">
                    <uni-easyinput v-model="profile.parent_phone" class="input" type="text" placeholder="请输入家长电话"/>
                </uni-forms-item>
                <uni-forms-item name="apartment" label="公寓名称">
                    <picker mode="selector" :range="apartmentOptions.slice(1)"
                            :value="selectedApartmentIndex > 0 ? selectedApartmentIndex - 1 : ''"
                            @change="e => updateApartment({detail: {value: Number(e.detail.value) + 1}})">
                        <view class="input input-picker">
                            <span :class="selectedApartmentIndex === 0 ? 'placeholder-text' : ''">
                                {{
                                    selectedApartmentIndex === 0 ? '请选择公寓名称' : apartmentOptions[selectedApartmentIndex]
                            }}
                            </span>
                        </view>
                    </picker>
                </uni-forms-item>
                <uni-forms-item name="apartment_id" label="宿舍号">
                    <uni-easyinput v-model="profile.apartment_id" class="input" type="text" placeholder="请输入宿舍号"/>
                </uni-forms-item>
                <uni-forms-item name="teacher_name" label="老师姓名">
                    <uni-easyinput v-model="profile.teacher_name" class="input" type="text" placeholder="请输入老师姓名"/>
                </uni-forms-item>
                <button class="profile-btn" @click="handleSave">保 存</button>
            </uni-forms>
            <view class="copyright">© 2024 YourAppName</view>
        </scroll-view>
    </view>
</template>

<script setup>
import {ref, watch} from 'vue'
import {onLoad} from "@dcloudio/uni-app";
import {GetProfileAPI, UpdateProfileAPI} from "../../services/user";
import {useMemberStore} from '../../stores'
const memberStore = useMemberStore()

const formRef = ref(null)

const apartmentOptions = ['', '金川2号楼', '金川A号楼'] // 第一个为空字符串
const genderOptions = ['', '男', '女'] // 第一个为空字符串
const selectedApartmentIndex = ref(0) // 初始为0，表示未选
const selectedGenderIndex = ref(0) // 初始为0，表示未选

const profile = ref({
    name: '',
    phone: '',
    gender: '',
    parent_name: '',
    parent_phone: '',
    apartment: '',
    apartment_id: '',
    teacher_name: '', // 新增老师姓名
})
const getProfile = async () => {
    const res = await GetProfileAPI()
    if (res.data == null) {
        return
    }
    profile.value = res.data
}
onLoad(() => {
    getProfile()
})
// 用watch实现选项和profile的双向同步
{
    watch(() => profile.value.apartment, (val) => {
        selectedApartmentIndex.value = apartmentOptions.indexOf(val)
    })
    watch(selectedApartmentIndex, (val) => {
        profile.value.apartment = apartmentOptions[val] || ''
    })

    watch(() => profile.value.gender, (val) => {
        selectedGenderIndex.value = genderOptions.indexOf(val)
    })
    watch(selectedGenderIndex, (val) => {
        profile.value.gender = genderOptions[val] || ''
    })
}

const updateApartment = (event) => {
    selectedApartmentIndex.value = event.detail.value
}
const updateGender = (event) => {
    selectedGenderIndex.value = event.detail.value
}

const handleSave = async () => {
    await formRef.value?.validate()
    // apartment_id 转为数字类型
    profile.value.apartment_id = Number(profile.value.apartment_id)
    await UpdateProfileAPI(profile.value)
    uni.showToast({
        title: '保存成功',
        icon: 'success',
        duration: 1500,
        success: () => {
            setTimeout(() => {
                // 如果 /pages/index/index 不是 tabBar 页面，改用 navigateTo
                uni.navigateTo({url: '/pages/index/index'})
            }, 1000)
        },
    })
}

const rules = {
    name: {
        rules: [{required: true, errorMessage: '学生姓名不能为空'}],
    },
    phone: {
        rules: [
            {required: true, errorMessage: '学生电话不能为空'},
            {pattern: /^1[3-9]\d{9}$/, errorMessage: '请输入有效的手机号'},
        ],
    },
    gender: {
        rules: [
            {required: true, errorMessage: '性别不能为空'},
        ],
    },
    parent_name: {
        rules: [{required: true, errorMessage: '家长姓名不能为空'}],
    },
    parent_phone: {
        rules: [
            {required: true, errorMessage: '家长电话不能为空'},
            {pattern: /^1[3-9]\d{9}$/, errorMessage: '请输入有效的手机号'},
        ],
    },
    apartment: {
        rules: [{required: true, errorMessage: '请选择公寓名称'}],
    },
    apartment_id: {
        rules: [
            {required: true, errorMessage: '宿舍号不能为空'},
            {pattern: /^\d+$/, errorMessage: '宿舍号必须为数字'},
        ],
    },
    teacher_name: {
        rules: [{required: true, errorMessage: '老师姓名不能为空'}], // 新增校验
    },
}
</script>

<style scoped>
.profile-bg {
    position: fixed;
    left: 0;
    top: 0;
    width: 100vw;
    height: 100vh;
    background: linear-gradient(120deg, #74ebd5 0%, #ACB6E5 100%);
    /* 移除 display:flex 相关属性 */
}

.profile-scroll {
    width: 100vw;
    min-height: 100vh;
    height: 100vh;
    box-sizing: border-box;
    padding-top: 120rpx;
    display: flex;
    flex-direction: column;
    align-items: center;
    /* 内容居中且整体可滑动 */
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
    align-self: center;
    margin-left: auto;
    margin-right: auto;
}

.title {
    font-size: 44rpx;
    color: #333;
    font-weight: 600;
    margin-bottom: 48rpx;
    letter-spacing: 2rpx;
    text-shadow: 0 2rpx 8rpx rgba(172, 182, 229, 0.12);
    align-self: center;
    width: 100%;
    text-align: center;
}

.profile-card {
    width: 92vw;
    min-width: 280rpx;
    max-width: 540rpx;
    padding: 40rpx 16rpx 40rpx 16rpx;
    background: rgba(255, 255, 255, 0.90);
    border-radius: 40rpx;
    box-shadow: 0 8rpx 32rpx rgba(172, 182, 229, 0.10);
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 24rpx;
    margin-bottom: 32rpx;
    margin-left: auto;
    margin-right: auto;
}

.uni-forms-item__label {
    font-size: 28rpx;
    color: #666;
    font-weight: 500;
    margin-bottom: 8rpx;
    letter-spacing: 1rpx;
}

.input {
    width: 100%;
    height: 80rpx;
    margin: 0;
    border: none;
    border-radius: 32rpx;
    background: #f4f8fb;
    font-size: 34rpx;
    padding: 0 24rpx;
    box-sizing: border-box;
    color: #333;
    box-shadow: 0 2rpx 8rpx rgba(172, 182, 229, 0.08);
    display: flex;
    align-items: center;
    text-align: right;
}

.input-picker {
    justify-content: flex-end;
}

.input-first {
    margin-top: 0 !important;
}

.input :deep(.uni-easyinput__content-input) {
    background: transparent;
    border: none;
    font-size: 34rpx;
    color: #333;
    padding: 0;
    height: 80rpx;
    line-height: 80rpx;
    border-radius: 32rpx;
}

/* 放大输入框默认文字（placeholder） */
.input :deep(.uni-easyinput__content-input)::placeholder {
    font-size: 40rpx;
    color: #bbb;
}

.input :deep(.uni-easyinput__content) {
    background: transparent;
    border: none;
    box-shadow: none;
    padding: 0;
    height: 80rpx;
    border-radius: 32rpx;
}

.input:focus,
.input :deep(.uni-easyinput__content-input:focus) {
    background: #eaf6ff;
}

.profile-btn {
    width: 90%;
    height: 80rpx;
    background: linear-gradient(90deg, #74ebd5 0%, #ACB6E5 100%);
    color: #fff;
    font-size: 34rpx;
    font-weight: bold;
    border: none;
    border-radius: 32rpx;
    box-shadow: 0 4rpx 16rpx rgba(172, 182, 229, 0.12);
    margin: 32rpx auto 0 auto;
    transition: background 0.3s;
    letter-spacing: 4rpx;
}

.profile-btn:active {
    background: linear-gradient(90deg, #ACB6E5 0%, #74ebd5 100%);
}

.copyright {
    width: 100%;
    text-align: center;
    color: #888;
    font-size: 24rpx;
    letter-spacing: 2rpx;
    margin-bottom: 24rpx;
    position: static;
    align-self: center;
}

.input-picker .placeholder-text {
    font-size: 28rpx;
    color: #bbb;
}

.custom-form ::v-deep .uni-forms-item__error {
    text-align: right !important;
    padding-right: 0;
    width: 90%;
}
</style>
