import {http} from '../utils/request'

export const LoginAPI = (data) => {
    return http({
        method: 'POST',
        url: 'login',
        data
    })
}
export const RegisterAPI = (data) => {
    return http({
        method: 'POST',
        url: 'register',
        data
    })
}
export const GetProfileAPI = (data) => {
    return http({
        method: 'GET',
        url: 'profile',
        data
    })
}
export const UpdateProfileAPI = (data) => {
    return http({
        method: 'POST',
        url: 'profile',
        data
    })
}
export const RecodeAPI = (data) => {
    return http({
        method: 'POST',
        url: 'record',
        data
    })
}

export const GetRecordListAPI = (data) => {
    return http({
        method: 'GET',
        url: 'records',
        data
    })
}

export const GetRecordDetailAPI = (record_id) => {
    return http({
        method: 'GET',
        url: `record/${record_id}`
    })
}

export const DeleteRecordAPI = (record_id) => {
    return http({
        method: 'DELETE',
        url: `record/${record_id}`
    })
}

export const GetVersionAPI = () => {
    return http({
        method: 'GET',
        url: 'version'
    })
}

export const GetAPPAPKAPI = () => {
    return http({
        method: 'GET',
        url: 'update'
    })
}