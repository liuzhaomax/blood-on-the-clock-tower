export const debounce = (func, wait) => {
    let timeout
    return (...args) => {
        clearTimeout(timeout)
        timeout = setTimeout(() => {
            func.apply(this, args)
        }, wait)
    }
}

const useDebounce = (fn, wait) => {
    return debounce(fn, wait)
}

export default useDebounce