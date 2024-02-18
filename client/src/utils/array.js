export const remove = (array, target) => {
    let indexToRemove
    for (let i = 0; i < array.length; i++) {
        if (array[i] === target) {
            indexToRemove = i
            break
        }
    }
    if (indexToRemove !== undefined) {
        array.splice(indexToRemove, 1)
    }
    return array
}