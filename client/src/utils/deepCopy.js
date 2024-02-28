export const deepCopy = (obj) => {
    if (typeof obj !== "object" || obj === null) {
        return obj
    }

    let copy = Array.isArray(obj) ? [] : {}
    for (let key in obj) {
        if (Object.hasOwnProperty.call(obj, key)) {
            copy[key] = deepCopy(obj[key])
        }
    }
    return copy
}