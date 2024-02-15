export const genShortUUID = () => {
    const timestamp = Date.now().toString(16) // 使用当前时间戳，并转换为16进制字符串
    const randomPart = Math.floor(Math.random() * 1000).toString(16) // 添加随机数部分
    return timestamp + randomPart
}