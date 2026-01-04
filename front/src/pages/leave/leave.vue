<template>
    <view class="html-body" style="overflow:hidden;margin:0px;height:100%;width:100%">
        <!-- 恢复：top 固定为 0 -->
        <view :style="scrollStyle">
            <NavigationBar
                :name="profile.name"
                :date="times.date"
                style="position:fixed;top:0;left:0;width:100%;z-index:100;"
            />
            <view
                :style="{
      height: navBarHeight + 'px',
      background: '#fff',
      width: '100%'
    }"
            ></view>
            <view id="appContent" class="appContent" style="overflow:hidden;-webkit-transform:translateZ(0)">
                <view style="width:100%;height:100%;overflow:auto;background-color:rgb(255,255,255)">
                    <view
                        style="min-height:100%;font-size:14px;overflow:hidden;background-color:rgb(240,240,240);user-select:text">
                        <view class="css css8d83b6318c3b4252b4039b0c006de157" mwftype="form"
                              id="8d83b631-8c3b-4252-b403-9b0c006de157" style="background-color:rgb(240,240,240)">
                            <!--<NavigationBar :name="profile.name" :date="times.date" />-->
                            <view id="div" mwftype="div" style="margin:auto">
                                <view id="label" mwftype="label"
                                      style="height:40px;background-color:rgb(255,255,255);line-height:40px;padding:0px 15px;color:rgb(60,118,183);font-size:16px;font-weight:bold;border-bottom:0px solid rgb(204,204,204);text-align:center">
                                    本科生外出请假
                                </view>
                                <view id="div_2" mwftype="div" style="min-height:auto;overflow:hidden;height:auto">
                                    <view id="div_1" mwftype="div"
                                          style="height:30px;min-height:30px;margin:5px 0px;float:left;font-size:12px">
                                        <view id="label_1" mwftype="label" style="float:left;font-size:12px">
                                        </view>
                                        <view id="label_2" mwftype="label" style="float:left">拟稿人：{{
                                                profile.name
                                            }}
                                        </view>
                                    </view>
                                    <view id="div_1_1" mwftype="div"
                                          style="height:30px;min-height:30px;margin:5px 0px;float:right;font-size:12px">
                                        <view id="label_1_1" mwftype="label" style="float:left;font-size:12px">

                                        </view>
                                        <view id="label_2_1" mwftype="label" style="float:left">
                                            拟稿时间：{{ times.nigao }}
                                        </view>
                                    </view>
                                </view>
                                <view id="tab" mwftype="tab">
                                    <view
                                        style="height:40px;background:rgb(255,255,255);border-top:0px solid rgb(204,204,204);border-bottom:0px solid rgb(204,204,204)">
                                        <view class="tabNodeContainerRight"
                                              style="width:10px;height:22px;margin-top:3px;border:1px solid rgb(238,238,238);margin-right:3px;float:right;background:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAA4AAAAOCAYAAAAfSC3RAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAA7SURBVDhPYxgFJICUlBRtIDZGw9pQadwgLS3NA4i/A/F/KAaxPaDS+AFIIVQD8ZpgAKqZNE2jAAQYGAAS0R4NS6fivQAAAABJRU5ErkJggg==)center center no-repeat;cursor:pointer;display:none"></view>
                                        <view class="tabNodeContainerLeft" style="margin-right:15px;height:30px">
                                            <view class="tabNodeContainerArea">
                                                <view id="tab_tab$Page" mwftype="tab$Page"
                                                      style="position:relative;margin-left:2px;margin-right:0px;float:left;height:40px;cursor:pointer;min-width:70px;background:rgb(255,255,255);border-width:0px 0px 2px;border-top-style:initial;border-right-style:initial;border-bottom-style:solid;border-left-style:initial;border-top-color:initial;border-right-color:initial;border-bottom-color:rgb(191,99,100);border-left-color:initial;border-image:initial;margin-top:0px">
                                                    <view
                                                        style="color:rgb(251,71,71);line-height:40px;text-align:center;margin-right:4px;margin-left:4px;font-size:14px;margin-top:0px">
                                                        基本信息
                                                    </view>
                                                </view>
                                                <view id="tab_tab$Page_1" mwftype="tab$Page"
                                                      style="position:relative;margin-left:2px;margin-right:0px;float:left;height:40px;cursor:pointer;min-width:70px;background:rgb(255,255,255);border:0px">
                                                    <view
                                                        style="color:rgb(102,102,102);line-height:40px;text-align:center;margin-right:4px;margin-top:0px;margin-left:4px;font-size:14px">
                                                        正文附件
                                                    </view>
                                                </view>
                                                <view id="tab_tab$Page_2" mwftype="tab$Page"
                                                      style="position:relative;margin-left:2px;margin-right:0px;float:left;height:40px;cursor:pointer;min-width:70px;background:rgb(255,255,255);border:0px">
                                                    <view
                                                        style="color:rgb(102,102,102);line-height:40px;text-align:center;margin-right:4px;margin-top:0px;margin-left:4px;font-size:14px">
                                                        流程记录
                                                    </view>
                                                </view>
                                                <view id="tab_tab$Page_3" mwftype="tab$Page"
                                                      style="position:relative;margin-left:2px;margin-right:0px;float:left;height:40px;cursor:pointer;min-width:70px;background:rgb(255,255,255);border:0px">
                                                    <view
                                                        style="color:rgb(102,102,102);line-height:40px;text-align:center;margin-right:4px;margin-top:0px;margin-left:4px;font-size:14px">
                                                        流程监控
                                                    </view>
                                                </view>
                                            </view>
                                        </view>
                                    </view>
                                    <view name="MWFcontentNodeContainer"
                                          style="line-height:normal;clear:both;overflow:hidden">
                                        <view
                                            style="overflow:hidden;display:block;background-color:transparent;opacity:1">
                                            <view id="tab_tab$Content" mwftype="tab$Content" style="padding:10px 0px">
                                                <view id="div_3" mwftype="div"
                                                      style="background-color:rgb(255,255,255);border:1px solid rgb(204,204,204)">
                                                    <view mwftype="label" id="label_5"
                                                          style="height:40px;background-color:rgb(255,255,255);line-height:40px;padding:0px 15px;color:rgb(60,118,183);font-size:18px;font-weight:bold;border-bottom:0px solid rgb(204,204,204)">
                                                        基本信息
                                                    </view>
                                                    <view id="table_3" mwftype="table"
                                                          style="margin-bottom:10px;margin-top:10px">
                                                        <scroll-view scroll-y="true">
                                                            <uni-forms class="custom-form" v-model="form" ref="formRef"
                                                                       :rules="rules">
                                                                <table width="100%" align="center" border="0"
                                                                       cellspacing="0" cellpadding="5"
                                                                       style="border-collapse:separate;border-top:1px solid rgb(0,0,0);border-left:1px solid rgb(0,0,0)">
                                                                    <tbody>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>学生姓名</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border-width:0px;border-bottom-style:solid;border-bottom-color:rgb(204,204,204);border-top-style:initial;border-top-color:initial;border-left-style:initial;border-left-color:initial;border-right-style:initial;border-right-color:initial;line-height:26px">
                                                                                {{ profile.name }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>学号</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                {{ profile.student_id }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>手机号码</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                {{ profile.phone }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:200px;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>性别</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view>{{ profile.gender }}</view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>所在学院</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                智能科学与技术学院（网络空间安全学院）
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>年级专业</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                2024级计算机科学与技术
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>现住公寓名称</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                {{ profile.apartment }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>现住宿舍号</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                {{ profile.apartment_id }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>家长姓名</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                {{ profile.parent_name }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>家长联系电话</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                {{ profile.parent_phone }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>请假类型</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="border:0px solid rgb(204,204,204);margin-top:5px;color:rgb(102,102,102);font-size:18px;height:26px;text-align:right">
                                                                                <uni-forms-item name="leave_type"
                                                                                                v-if="!endShow">
                                                                                    <picker
                                                                                        :value="selectedIndex"
                                                                                        mode=selector
                                                                                        :range="leaveTypes"
                                                                                        @change="handleTypeChange"
                                                                                    >
                                                                                        <view class="picker-wrapper">
                                                                                            {{
                                                                                                leaveTypes[selectedIndex] || '请选择请假类型'
                                                                                            }}
                                                                                        </view>
                                                                                    </picker>
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">{{
                                                                                        form.leave_type
                                                                                    }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>请假开始时间</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                <uni-forms-item name="start_time"
                                                                                                v-if="!endShow">
                                                                                    <Data
                                                                                        type="开始"
                                                                                        :init-time="times.start_time"
                                                                                        @update:time="(timeStr) => times.start_time = timeStr"
                                                                                    />
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">{{
                                                                                        times.start_time
                                                                                    }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>请假结束时间</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                <uni-forms-item name="end_time"
                                                                                                v-if="!endShow">
                                                                                    <Data
                                                                                        type="结束"
                                                                                        :init-time="times.end_time"
                                                                                        @update:time="(timeStr) => times.end_time = timeStr"
                                                                                    />
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">{{
                                                                                        times.end_time
                                                                                    }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view style="float:left;margin-top:8px">
                                                                                共多少天
                                                                            </view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:visible;position:relative;margin-right:0px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;text-align:center;margin-top:10px;color:rgb(102,102,102)">
                                                                                {{ form.leave_days }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>请假详细原因</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="position:relative;padding-right:2px;border:0px solid rgb(204,204,204);color:rgb(102,102,102)">
                                                                                <uni-forms-item name="leave_reason"
                                                                                                v-if="!endShow">
                                                                                    <uni-easyinput
                                                                                        placeholder="请输入详细原因"
                                                                                        v-model="form.leave_reason"
                                                                                        type="text"
                                                                                        style="width:70%;margin-left: auto"
                                                                                    ></uni-easyinput>
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">
                                                                                    {{ form.leave_reason }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>请假期间是否离校</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view>
                                                                                <uni-forms-item name="leave_school"
                                                                                                v-if="!endShow">
                                                                                    <uni-data-checkbox
                                                                                        v-model="form.leave_school"
                                                                                        :localdata="[{ value: '是', text: '是' }, { value: '否', text: '否' }]"
                                                                                        @change="leaveChange"
                                                                                        checkColor="#007aff"
                                                                                        style="float: right"
                                                                                    />
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">
                                                                                    {{ form.leave_school }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>

                                                                    <tr v-show="nextShow">
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>目的地</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="border:0px solid rgb(204,204,204);color:rgb(102,102,102);font-size:18px;margin-top:5px;height:26px;text-align:right">
                                                                                {{ form.province }}
                                                                            </view>
                                                                            <view
                                                                                style="border:0px solid rgb(204,204,204);color:rgb(102,102,102);font-size:18px;margin-top:5px;height:26px;text-align:right">
                                                                                {{ form.city }}
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr v-show="nextShow">
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>详细地址</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                <uni-forms-item name="place_detail"
                                                                                                v-if="!endShow">
                                                                                    <uni-easyinput
                                                                                        placeholder="请输入详细地址"
                                                                                        v-model="form.place_detail"
                                                                                        type="text"
                                                                                        style="width:70%;margin-left: auto"
                                                                                    ></uni-easyinput>
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">
                                                                                    {{ form.place_detail }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr v-show="nextShow">
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>出行方式</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                <uni-forms-item name="leave_type"
                                                                                                v-if="!endShow">
                                                                                    <uni-easyinput
                                                                                        placeholder="请输入出行方式"
                                                                                        v-model="form.leave_way"
                                                                                        type="text"
                                                                                        style="width:70%;margin-left: auto"
                                                                                    ></uni-easyinput>
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">{{
                                                                                        form.leave_way
                                                                                    }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr v-show="nextShow">
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>请假期间是否回校住宿</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view>
                                                                                <uni-forms-item name="back_school"
                                                                                                v-if="!endShow">
                                                                                    <uni-data-checkbox
                                                                                        v-model="form.back_school"
                                                                                        :localdata="[{ value: '是', text: '是' }, { value: '否', text: '否' }]"
                                                                                        @change="backChange"
                                                                                        checkColor="#007aff"
                                                                                        style="float: right"
                                                                                    />
                                                                                </uni-forms-item>
                                                                                <text v-if="endShow">{{
                                                                                        form.back_school
                                                                                    }}
                                                                                </text>
                                                                            </view>
                                                                        </td>
                                                                    </tr>

                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>佐证材料</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view mwftype="attachment"
                                                                                  id="attachment_1">
                                                                                <view
                                                                                    style="border:0 solid rgb(182,182,182);min-height:30px;overflow:hidden;background:transparent;height:auto">

                                                                                    <view
                                                                                        style="float:right;height:26px;margin:2px;border:1px solid rgb(151,151,151);border-radius:3px;background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAMAAAAaCAIAAAD9k3G7AAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyFpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNS1jMDE0IDc5LjE1MTQ4MSwgMjAxMy8wMy8xMy0xMjowOToxNSAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDowRDYxMTZERUY4ODgxMUU0QkZDMkM4QkM5RjQ1QUZEOCIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDowRDYxMTZERkY4ODgxMUU0QkZDMkM4QkM5RjQ1QUZEOCI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOjBENjExNkRDRjg4ODExRTRCRkMyQzhCQzlGNDVBRkQ4IiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOjBENjExNkRDRjg4ODExRTRCRkMyQzhCQzlGNDVBRkQ4Ii8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+UNR7TwAAACtJREFUeNpi+g8DLP/+/WMAAxYgj0QWXC8TAwxQYgo2WQImk+x6hHkAAQYARLxLCgkuDyQAAAAASUVORK5CYII=)">
                                                                                        <view title="上传"
                                                                                              style="float:left;width:26px;height:26px;cursor:default">
                                                                                            <view
                                                                                                style="width:26px;height:26px;background-repeat:no-repeat;background-position:center center;background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAA8AAAAPCAYAAAA71pVKAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyFpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNS1jMDE0IDc5LjE1MTQ4MSwgMjAxMy8wMy8xMy0xMjowOToxNSAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDpEMEM2NTU0NEY5NDYxMUU0ODk0M0QwQUY5RERDNzRENCIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDpEMEM2NTU0NUY5NDYxMUU0ODk0M0QwQUY5RERDNzRENCI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOkQwQzY1NTQyRjk0NjExRTQ4OTQzRDBBRjlEREM3NEQ0IiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOkQwQzY1NTQzRjk0NjExRTQ4OTQzRDBBRjlEREM3NEQ0Ii8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+FttbFgAAAQNJREFUeNpibGpqYkAGgvwCMGY0EIMk+4F4OhD/ff/xA4paJiwamYE4HohrgVgUiAuAOAuIWZEMRtUMlWCDKs4B4glAfAWIu4A4AohrgJgL2QBkm0Eay4E4GIjzgHgLEDMC8SEgzgRieyBuARmATTPIuSBPZQDxcSQ5kPglIE4F4scg58M0sCBp/g7Ekxlwg9vQwMMeYDCAHqrofAybcSnAJ8/EQAFggQU9IZuxJSSQzTw4XMCIx2UgcW4QsQCItbEo+AnEj4D4BxY5JSBeCAowcyAWQE96QPACmjjeYpEDudacBeo8NSD+gsWGf0Asi0VcA+R0kOZ30HT7m5SABlkGEGAAa7Q06pe/xecAAAAASUVORK5CYII=)"></view>
                                                                                        </view>
                                                                                        <view title="删除"
                                                                                              style="float:left;width:26px;height:26px;cursor:default">
                                                                                            <view
                                                                                                style="width:26px;height:26px;background-repeat:no-repeat;background-position:center center;background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyFpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNS1jMDE0IDc5LjE1MTQ4MSwgMjAxMy8wMy8xMy0xMjowOToxNSAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDo0RkE3OUE4QUY5NDYxMUU0OTFCM0IwRTlDNTY0MTIxQSIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDo0RkE3OUE4QkY5NDYxMUU0OTFCM0IwRTlDNTY0MTIxQSI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOjRGQTc5QTg4Rjk0NjExRTQ5MUIzQjBFOUM1NjQxMjFBIiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOjRGQTc5QTg5Rjk0NjExRTQ5MUIzQjBFOUM1NjQxMjFBIi8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+Q7Xj9wAAAYlJREFUeNpsk00vA1EUhqcz/WJoCYkQFtgIFkhY2FlaWdoRG4mfoVsLW0sLP8DKT/AHiNhIJJVIG1Ed7fQDrecm720u7SRP72TuOe99z7mniUKh4NlnPD9mliG4gi1oaisLt3AArUr1o5eTJGmJNYQOJGAE5uEGHhS3ChuwCTE5XVYfakl+TmERZuAN6pCCdVh2HKThTIdNwCs8GoETmIRLOId7CZRgCsxpZZiGhlwcwxFUjI13KMKnat6BfQnswa7eD2HblGCsw4vJ9WXxB9owrNNn9d10Nad34yCpEhrK8azAt1RzcpL1/j6+EquK6RPoyH4ogVCn9W5LTYy1V1NvegLmiaQeKzhw9tJyZRLzuinvv0BDM1Bzrq0rUnLRVJ9MjGcGynemytQ3qqBAAp4zB4Hc2T71Oag7HQ6UZB1kNKVtfR9YQiR7XwrOOHuhhNoqM3K76zqw9d3pap+V1NKENiUcDxKINNIrcA0L8KST5/RtTTEDHZTUoAvnn9nVnn23jSzbpF8BBgC08G0awQQnHQAAAABJRU5ErkJggg==)"></view>
                                                                                        </view>
                                                                                        <view title="替换"
                                                                                              style="float:left;width:26px;height:26px;cursor:default">
                                                                                            <view
                                                                                                style="width:26px;height:26px;background-repeat:no-repeat;background-position:center center;background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABEAAAAPCAYAAAACsSQRAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyFpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNS1jMDE0IDc5LjE1MTQ4MSwgMjAxMy8wMy8xMy0xMjowOToxNSAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSIgeG1wTU06SW5zdGFuY2VJRD0ieG1wLmlpZDpCMzRCNzhDNkY5NDYxMUU0QkZDQUFEQkI3N0M2MzkxMCIgeG1wTU06RG9jdW1lbnRJRD0ieG1wLmRpZDpCMzRCNzhDN0Y5NDYxMUU0QkZDQUFEQkI3N0M2MzkxMCI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOkIzNEI3OEM0Rjk0NjExRTRCRkNBQURCQjc3QzYzOTEwIiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOkIzNEI3OEM1Rjk0NjExRTRCRkNBQURCQjc3QzYzOTEwIi8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+x4XU3QAAAUVJREFUeNqU0zssBFEUxvF9TBQIG49C2ChEK0panYqCVqPyiIIgIRQThUY8CqFQKWULBZ1CoyR0Eq3EWwhiiV3/I98kE5m5mz3JL7u5M/fMuWfOJH3fT5QZHpYxhAes2sIi0kjGbDrCsf5n0IxbvOAC05ZkFru4j0kSJG/AijbbnjfM2HVL8mwl4dJxhCbdk8I8PnGID3x5elLakaAV63jFlPowjD793qVKNLEd2+rBBJ4whhFsKuFfp+OiAxs4wwLyqmQAk2p2BRrjKrH1HrRpcyXm0I/x0Nuy67moSqpV7j5OsIYuFLR+HrrXKslGVdKJQbzjFFuqZvRfgiCKUZX04kdPzWIJB3h0jXA46tCtnlRhD9eaB+d3UNR5Lb41jVcawlJRCI7jqey8Et6gBrWO7ymIFhtUS1CPHRvfRPlhbyfzK8AAz5pHzi3GBCQAAAAASUVORK5CYII=)"></view>
                                                                                        </view>
                                                                                        <view title="下载"
                                                                                              style="float:left;width:26px;height:26px;cursor:default">
                                                                                            <view
                                                                                                style="width:26px;height:26px;background-repeat:no-repeat;background-position:center center;background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAA8AAAAPCAYAAAA71pVKAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyFpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNS1jMDE0IDc5LjE1MTQ4MSwgMjAxMy8wMy8xMy0xMjowOToxNSAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtcE1NOkRvY3VtZW50SUQ9InhtcC5kaWQ6NjgyMDc1QjVGOTQ2MTFFNEJGMTBDMzU2NDIwMkEyMDMiIHhtcE1NOkluc3RhbmNlSUQ9InhtcC5paWQ6NjgyMDc1QjRGOTQ2MTFFNEJGMTBDMzU2NDIwMkEyMDMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIChXaW5kb3dzKSI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ4bXAuaWlkOjY1QUZEODM2RjkzQzExRTQ4NEI4QTFBRDI1Q0YxNTRBIiBzdFJlZjpkb2N1bWVudElEPSJ4bXAuZGlkOjY1QUZEODM3RjkzQzExRTQ4NEI4QTFBRDI1Q0YxNTRBIi8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+UxfsvQAAAOpJREFUeNrc0k8LAUEYx/ElhVK4uHEgUspV3oN34Yq4cHFaiqPYUC6Ug5emyJ+SA1m+j0ZNm1UcPfXZaWefXzs7sx7TNA2paDhiOCqFPqpY6w/2x8Nz9BruFUACfreGT2EpG/dfwx/rD8JhNJB26c2rYwu+C18Rx0w13rQdL2KKiDZv+LSf44w2WphgoY6phDJWGOHyyvgQUkFbjV1s0UQMdfSw1N4qKw7KZY6cY/ljDLBBR63ipvUkZU7eXFDfopc0WpVa1ZIbazhybp6stiBhDzI4OTsI2S7HmZV5Ce/URl2/OGLJnR4CDAAcMDKZgPoLXgAAAABJRU5ErkJggg==)"></view>
                                                                                        </view>
                                                                                        <view title="附件排序"
                                                                                              style="float:left;width:26px;height:26px;cursor:default">
                                                                                            <view
                                                                                                style="width:26px;height:26px;background-repeat:no-repeat;background-position:center center;background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAABmJLR0QA/wD/AP+gvaeTAAABQklEQVQ4jcWQsU4CYRCEvwVDRJJLLIj9JTwAyU9hNNHzBUQqG30FjRRaiYmNFjYW9rT3DCY2Gov7OSusaaioCAXnFazNHTkFIzQ61c7sZjYz8N+QdLDWbgNbmd2LMeY5e2ytPTfG3Ga1XDqo6iqwLiIbwJWIbM55ePNdmBrUarXHYrF4qaquiDw5jnO/SISpQbfbLYzHY19ECo7jHFQqlY+lDKIoOgH2VPV1OByeJp38ipV0mEwmbyLyAJSAUtLJclDVnLX2TFVz8/bWWv0xAkAYhrvAXafT2Vn06RcDVT0C3oHjrB4EQdX3/TyA7/v5IAiqMwbW2jWgoaoh0Eg4ACLScl23DeC6bltEWjMGqroP9ESkD/QSDsBoNDoEygktJ3wmwkBVm8aYC1VtAoN04XleFMdxHbiO47jueV60QD1/hE8qJoEKdHMe5gAAAABJRU5ErkJggg==)"></view>
                                                                                        </view>
                                                                                        <view class="separateNode"
                                                                                              style="width:1px;height:16px;margin-top:5px;margin-left:5px;margin-right:5px;border-left:1px solid rgb(188,188,188);background-color:rgb(241,241,241);float:left"></view>
                                                                                        <view title="简洁模式"
                                                                                              style="float:left;width:26px;height:26px;cursor:default">
                                                                                            <view
                                                                                                style="width:26px;height:26px;background-repeat:no-repeat;background-position:center center;background-image:url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAA8AAAAPCAQAAACR313BAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAAAmJLR0QAAKqNIzIAAAAJcEhZcwAAASwAAAEsAHOI6VIAAADGSURBVBjTdc49SgNRFAbQ84YYrAaUOGDvBgQhWIhrsLF1GTYuQDuLGIuARYilhSBYuQexFhsLEdLZxMyfhYPkTeLX3Xvguzc4dWLdb0YePNpGrVR7DT4MPTf8Yu5MprbpQNc3U7viBJmJuZkZU3st7hnLjTwplrlnIje04U7Z5q0GU4l7VcyZW7krKRIDb4u85lphIP1r2uksFNfenbvw1cyJo7g8RE8e+kyiRd2aypiXkqD8xyqho7Kvu5L7iuDSsdUnSjc/lXY6/pSxbuEAAAAldEVYdGRhdGU6Y3JlYXRlADIwMTQtMDQtMjRUMTM6MjI6MTQrMDg6MDBOsPVLAAAAJXRFWHRkYXRlOm1vZGlmeQAyMDE0LTA0LTI0VDEzOjIyOjE0KzA4OjAwP+1N9wAAAE10RVh0c29mdHdhcmUASW1hZ2VNYWdpY2sgNi44LjgtNyBRMTYgeDg2XzY0IDIwMTQtMDItMjggaHR0cDovL3d3dy5pbWFnZW1hZ2ljay5vcmdZpF9/AAAAY3RFWHRzdmc6Y29tbWVudAAgR2VuZXJhdG9yOiBBZG9iZSBJbGx1c3RyYXRvciAxNi4wLjAsIFNWRyBFeHBvcnQgUGx1Zy1JbiAuIFNWRyBWZXJzaW9uOiA2LjAwIEJ1aWxkIDApICByC3WWAAAAGHRFWHRUaHVtYjo6RG9jdW1lbnQ6OlBhZ2VzADGn/7svAAAAGHRFWHRUaHVtYjo6SW1hZ2U6OkhlaWdodAAzMzUnUtgSAAAAF3RFWHRUaHVtYjo6SW1hZ2U6OldpZHRoADMzM13ALXoAAAAZdEVYdFRodW1iOjpNaW1ldHlwZQBpbWFnZS9wbmc/slZOAAAAF3RFWHRUaHVtYjo6TVRpbWUAMTM5ODMxNjkzNNk2x/cAAAATdEVYdFRodW1iOjpTaXplADMuNjdLQkLxaVyBAAAAYnRFWHRUaHVtYjo6VVJJAGZpbGU6Ly8vaG9tZS9mdHAvMTUyMC9lYXN5aWNvbi5jbi9lYXN5aWNvbi5jbi9jZG4taW1nLmVhc3lpY29uLmNuL3BuZy8xMTQ4Mi8xMTQ4MjIwLnBuZ4hwq9YAAAAASUVORK5CYII=)"></view>
                                                                                        </view>
                                                                                    </view>
                                                                                    <div
                                                                                        style=clear:both;min-height:30px;overflow:hidden>
                                                                                    </div>
                                                                                </view>
                                                                            </view>
                                                                        </td>

                                                                    </tr>
                                                                    <tr>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                            <view>状态</view>
                                                                        </td>
                                                                        <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                            <view
                                                                                style="overflow:hidden;position:relative;margin-right:20px;padding-right:4px;border:0px solid rgb(204,204,204);line-height:26px;color:rgb(102,102,102)">
                                                                                请假
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    <tr>
                                                                        <td colspan="2"
                                                                            style="border-bottom:1px solid rgb(0,0,0);border-right:1px solid rgb(0,0,0)">
                                                                            <view>学生本人承诺：</view>
                                                                            <view>
                                                                                1.以上填报各项信息真实准确；并已向家长或配偶汇报我的外出安排，家长或配偶对此完全知情并同意我外出；
                                                                            </view>
                                                                            <view>
                                                                                2.班主任已向我宣讲减少外出、不参加聚集活动、加强个人防护、尽量不乘坐公共交通工具等疫情防控相关要求，并已告知外出期间有关注意事项及可能存在的风险；
                                                                            </view>
                                                                            <view>
                                                                                3.外出期间我将自觉遵守国家法律法规和校纪校规，切实做好个人防护，确保通讯畅通，自觉注意人身及财产、交通、饮食、消防等安全，外出期间一切安全责任由本人承担；
                                                                            </view>
                                                                            <view>
                                                                                4.按照学校要求及时报告行程和身体健康状况，按时返校并主动销假；身体如有不适及时就医并向老师和学院报告。
                                                                            </view>
                                                                        </td>
                                                                    </tr>
                                                                    </tbody>
                                                                </table>
                                                            </uni-forms>
                                                        </scroll-view>
                                                    </view>
                                                </view>
                                                <view mwftype="label" id="label_16"
                                                      style="height:40px;background-color:rgb(255,255,255);line-height:40px;padding:0px 15px;color:rgb(60,118,183);font-size:18px;font-weight:bold;border-bottom:0px solid rgb(204,204,204)">
                                                    审核信息
                                                </view>
                                                <view mwftype="div" id="div_4">
                                                    <view id="table" mwftype="table"
                                                          style="margin-bottom:10px;margin-top:10px">
                                                        <view scroll-x="true" style="width:100%">
                                                            <table border="0" cellpadding="5" cellspacing="0"
                                                                   width="100%" align="center"
                                                                   style="border-top:1px solid rgb(0,0,0);border-left:1px solid rgb(0,0,0)">
                                                                <tbody>
                                                                <tr>
                                                                    <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:200px;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                        <view>班主任审批</view>
                                                                    </td>
                                                                    <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                        <view style="user-select:text">
                                                                            <view
                                                                                style="overflow:hidden;margin:0px;padding:10px 0px;border:1px solid rgb(204,204,204);background-color:rgb(255,255,255)">
                                                                                <view
                                                                                    style="clear:both;overflow:hidden;background-color:rgb(246,246,246);min-height:30px;line-height:40px">
                                                                                    <view
                                                                                        style="clear:both;overflow:hidden;line-height:28px;font-size:12px;background:transparent">
                                                                                        <view
                                                                                            style="height:28px;width:28px;background-position:center center;background-repeat:no-repeat;float:left"></view>
                                                                                        <view
                                                                                            style="margin-right:10px;overflow:hidden;margin-left:28px">
                                                                                            <text style="color:#ff5400">
                                                                                                {{
                                                                                                    profile.teacher_name
                                                                                                }}
                                                                                            </text>
                                                                                            （计算机系）
                                                                                            选择【同意】，提交于：
                                                                                            {{ times.nigao }}，处理意见：
                                                                                            <text
                                                                                                style="color:#00F"></text>
                                                                                        </view>
                                                                                    </view>
                                                                                </view>
                                                                            </view>
                                                                        </view>
                                                                    </td>
                                                                </tr>
                                                                <tr>
                                                                    <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                        <view>团学办公室审批</view>
                                                                    </td>
                                                                    <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                        <view style="user-select:text">
                                                                            <view
                                                                                style="overflow:hidden;margin:0px;padding:10px 0px;border:1px solid rgb(204,204,204);background-color:rgb(255,255,255)"></view>
                                                                        </view>
                                                                    </td>
                                                                </tr>
                                                                <tr>
                                                                    <td style="background-color:rgb(255,255,255);color:rgb(51,51,51);border-bottom:1px solid rgb(0,0,0);height:36px;font-size:16px;width:30%;line-height:36px;text-align:left;font-weight:bold;border-right:1px solid rgb(0,0,0)">
                                                                        <view>分管学工副院长审批</view>
                                                                    </td>
                                                                    <td style="background-color:rgb(255,255,255);color:rgb(102,102,102);font-size:18px;border-bottom:1px solid rgb(0,0,0);text-align:right;border-right:1px solid rgb(0,0,0)">
                                                                        <view style="user-select:text">
                                                                            <view
                                                                                style="overflow:hidden;margin:0px;padding:10px 0px;border:1px solid rgb(204,204,204);background-color:rgb(255,255,255)"></view>
                                                                        </view>
                                                                    </td>
                                                                </tr>
                                                                </tbody>

                                                            </table>
                                                            <view class="button-group" v-show="!endShow">
                                                                <!-- 提交按钮 -->
                                                                <button class="submit-btn" type="primary"
                                                                        :disabled="!canSubmit" @tap="submitForm">
                                                                    提交
                                                                </button>

                                                                <!-- 修改个人信息按钮 -->
                                                                <button class="info-btn" @click="Profile()">
                                                                    修改个人信息
                                                                </button>
                                                            </view>

                                                        </view>
                                                    </view>
                                                </view>
                                            </view>
                                        </view>
                                    </view>
                                </view>
                            </view>
                        </view>
                    </view>
                </view>
            </view>
        </view>
    </view>
</template>

<script setup>
import NavigationBar from "./components/NavigationBar.vue"
import Data from "./components/Data.vue";
import {computed, ref, watch, onMounted} from "vue"
import {onLoad} from "@dcloudio/uni-app";
import dayjs from "dayjs";
import {GetProfileAPI, RecodeAPI} from "../../services/user";

function Profile() {
    uni.navigateTo({url: '/pages/profile/profile'})
}

const test = () => {
    console.log(form.value)
}

const canSubmit = ref(true)
const profile = ref({
    name: '',
    student_id: '',
    phone: '',
    gender: '',
    apartment: '',
    apartment_id: '',
    parent_name: '',
    parent_phone: '',
    teacher_name: '',
})
const form = ref({
    leave_type: '',
    leave_days: 0,
    leave_reason: '',
    leave_school: '',
    place_detail: '',
    leave_way: '',
    back_school: '',
    city: '呼和浩特市',
    province: '内蒙古自治区',
})

const times = ref({
    start_time: '',
    end_time: '',
    date: '',
    nigao: '',
})

// 获取个人信息
const getProfile = async () => {
    const res = await GetProfileAPI()
    profile.value = res.data
    recode.value.student_id = res.data.student_id
    recode.value.name = res.data.name
}
onLoad(() => {
        getProfile()
        const record = uni.getStorageSync('recordDetail');
        if (record) {
            getHistoryRecord()
        } else {
            const yesterday = dayjs().subtract(1, 'day');
            times.value.nigao = yesterday.format('YYYY-MM-DD HH:mm:ss');
            times.value.date = yesterday.format('YYYYMMDD');
        }
    }
)

const getHistoryRecord = () => {
    const record = uni.getStorageSync('recordDetail');

    if (record.nigao) {
        const recordDayjs = dayjs(record.nigao);
        times.value.nigao = recordDayjs.format('YYYY-MM-DD HH:mm:ss');
        times.value.date = recordDayjs.format('YYYYMMDD');
        times.value.start_time = dayjs(record.start_time).format('YYYY-MM-DD HH:mm:ss')
        times.value.end_time = dayjs(record.end_time).format('YYYY-MM-DD HH:mm:ss')
    }

    form.value = {
        ...form.value,
        leave_type: record.leave_type,
        leave_days: record.leave_days,
        leave_reason: record.leave_reason,
        leave_school: record.leave_school,
        place_detail: record.place_detail,
        leave_way: record.leave_way,
        back_school: record.back_school,
    };

    if (record.leave_school === '否') {
        form.value.city = '',
            form.value.province = ''
    }

    uni.removeStorageSync('recordDetail');

    // 开始隐藏页面内容
    nextShow.value = true;
    endShow.value = true;
};

// 是否标签完善
const backChange = (e) => {
    form.value.back_school = e.detail.value;
}

const nextShow = ref(true)
const leaveChange = (e) => {
    form.value.leave_school = e.detail.value;
    if (form.value.leave_school === '是') {
        nextShow.value = true
    } else {
        nextShow.value = false
        form.value.back_school = ''
    }
}

// 时间的处理
const leaveDays = computed(() => {
    // 验证时间是否完整
    if (!times.value.start_time || !times.value.end_time) {
        return 0;
    }

    // 解析时间
    const start = dayjs(times.value.start_time);
    const end = dayjs(times.value.end_time);

    // 验证时间有效性
    if (!start.isValid() || !end.isValid()) {
        return 0;
    }

    // 确保结束时间不早于开始时间
    if (end.isBefore(start)) {
        uni.showToast({
            title: '结束时间不能早于开始时间',
            icon: 'error',
            duration: 1000,
        })
        canSubmit.value = false
        return 0;
    }

    // 计算时间差（天）
    const diffDays = end.diff(start, 'day', true); // true表示返回浮点数
    canSubmit.value = true
    // 不足1天按0.5天计算，1天及以上向上取整
    return diffDays < 1 ? 0.5 : Math.ceil(diffDays);
});

// 监听计算结果变化，同步到表单字段
watch(
    leaveDays,
    (newVal) => {
        form.value.leave_days = newVal;
    },
    {immediate: true} // 初始化时立即执行一次
);

// 请假类型
const leaveTypes = ['事假', '病假', '实习', '假期离呼', '其他'];
const selectedIndex = ref(null);
const handleTypeChange = (e) => {
    selectedIndex.value = e.detail.value;
    form.value.leave_type = leaveTypes[selectedIndex.value];
};

// 提交表单
const endShow = ref(false);
const formRef = ref(null);
const rules = {
    // 请假类型（必填）
    leave_type: {
        rules: [
            {
                required: true,
                errorMessage: '请选择请假类型'
            }
        ]
    },
    // 请假天数（必填，大于0）
    leave_days: {
        rules: [
            {
                required: true,
                errorMessage: '请填写请假天数'
            },
            {
                validator: (rule, value) => {
                    return value > 0;
                },
                errorMessage: '请假天数必须大于0'
            }
        ]
    },
    // 请假原因（必填）
    leave_reason: {
        rules: [
            {
                required: true,
                errorMessage: '请填写请假原因'
            },
        ]
    },
    // 请假学校（必填）
    leave_school: {
        rules: [
            {
                required: true,
                errorMessage: '请选择是否离开学校'
            }
        ]
    },
    // 详细地点（当leave_school为"是"时必填）
    place_detail: {
        rules: [
            {
                validator: (rule, value) => {
                    // 如果leave_school是"否"，则不验证
                    if (form.value.leave_school === '否') {
                        return true;
                    }
                    // 否则验证必填
                    return !!value;
                },
                errorMessage: '请填写详细地点'
            }
        ]
    },
    // 请假方式（当leave_school为"是"时必填）
    leave_way: {
        rules: [
            {
                validator: (rule, value) => {
                    if (form.value.leave_school === '否') {
                        return true;
                    }
                    return !!value;
                },
                errorMessage: '请选择请假方式'
            }
        ]
    },
    // 返校情况（当leave_school为"是"时必填）
    back_school: {
        rules: [
            {
                validator: (rule, value) => {
                    if (form.value.leave_school === '否') {
                        return true;
                    }
                    return !!value;
                },
                errorMessage: '请选择是否返校'
            }
        ]
    }
};
const recode = ref({
    student_id: profile.value.student_id,
    name: profile.value.name,
    leave_type: form.value.leave_type,
    leave_reason: form.value.leave_reason,
    leave_school: form.value.leave_school,
    place_detail: form.value.place_detail,
    leave_way: form.value.leave_way,
    back_school: form.value.back_school,
    start_time: times.value.start_time,
    end_time: times.value.end_time,
    nigao: times.value.nigao,
})
const submitForm = () => {
    if (!(times.value.end_time && times.value.start_time)) {
        uni.showToast({
            title: '请完善时间信息',
            icon: 'error',
            duration: 1000,
        })
        return
    }
    if (!nextShow.value) {
        form.value.place_detail = '不离校';
        form.value.leave_way = '不离校';
        form.value.back_school = '是';
        form.value.city = '';
        form.value.province = '';
    }
    formRef.value.validate()
        .then(async () => {
            times.value.start_time += ':00';
            times.value.end_time += ':00';

            const recodeToSubmit = {
                leave_type: form.value.leave_type,
                leave_reason: form.value.leave_reason,
                leave_school: form.value.leave_school,
                place_detail: form.value.place_detail,
                leave_way: form.value.leave_way,
                back_school: form.value.back_school,
                start_time: formatTime(times.value.start_time),
                end_time: formatTime(times.value.end_time),
                nigao: formatTime(times.value.nigao),
            };
            await RecodeAPI(recodeToSubmit);

            uni.showToast({
                title: '提交成功',
                icon: 'success',
                duration: 1000,
            })

            nextShow.value = true;
            endShow.value = true;
        })
        .catch((err) => {
            // err.message 是 uni-forms 的错误提示
            uni.showToast({
                title: '请完善表单信息',
                icon: 'error',
                duration: 1500,
            })
            console.log(err)
        });
};

const formatTime = (timeStr) => {
    if (!timeStr) return '';
    return dayjs(timeStr).format('YYYY-MM-DDTHH:mm:ss+08:00');
};

const scrollStyle = computed(() => {
    // 恢复静态 top:0
    return `-webkit-overflow-scrolling:touch;overflow:auto;position:absolute;bottom:4em;top:0px;left:0px;width:100%`
})
const navBarHeight = ref(0)

onMounted(() => {
    const info = uni.getSystemInfoSync()
    // 90rpx 转 px，假设1rpx = windowWidth / 750
    const rpx2px = info.windowWidth / 750
    navBarHeight.value = (info.statusBarHeight || 0) + 90 * rpx2px
})
</script>

<style scoped>
/* 针对整个表单的验证提示设置靠右 */
.custom-form ::v-deep .uni-forms-item__error {
    text-align: right !important; /* 强制靠右对齐 */
    padding-right: 0;
    /* 如果输入框有固定宽度，可统一设置提示宽度与之匹配 */
    width: 90%;
}

.button-group {
    display: flex;
    gap: 20rpx;
    margin-top: 40rpx;
    padding: 0 20rpx;
}

/* 提交按钮样式 */
.submit-btn {
    flex: 1;
    height: 80rpx;
    line-height: 80rpx;
    font-size: 32rpx;
    border-radius: 8rpx;
}

/* 修改个人信息按钮样式 */
.info-btn {
    flex: 1;
    height: 80rpx;
    line-height: 80rpx;
    font-size: 32rpx;
    border: 1px solid #007aff;
    color: #007aff;
    background-color: #fff;
    border-radius: 8rpx;
}

/* 简单的外边框样式 */
.picker-wrapper {
    border: 1px solid #ddd; /* 浅灰色边框 */
    padding: 12rpx 16rpx; /* 内边距 */
    border-radius: 4rpx; /* 轻微圆角 */
    width: 40%;
    margin-left: auto;
    height: 50%;
}

html, body, .html-body {
    width: 100%;
    height: 100%;
    margin: 0;
    padding: 0;
    overflow: auto;
    -webkit-overflow-scrolling: touch;
    font-family: "Microsoft YaHei", "SimSun", "宋体", serif;
    font-size: 14px;
}

.appContent * {
    box-sizing: content-box !important;
}

::-ms-clear, ::-ms-reveal {
    display: none;
}
</style>

