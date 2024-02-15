export const throttled = (fn, delay) => {
    let timer = null
    return (...args) => {
        if (!timer) {
            timer = setTimeout(() => {
                fn.apply(this, args)
                timer = null
            }, delay)
        }
    }
}

const useThrottled = (fn, wait) => {
    return throttled(fn, wait)
}

export default useThrottled