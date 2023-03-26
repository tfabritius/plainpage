/**
 * Compares two objects for deep equality.
 */
export function deepEqual<T>(o1: T, o2: T): boolean {
  if (typeof o1 !== typeof o2) {
    return false
  }

  if (o1 === null || typeof o1 !== 'object') {
    return o1 === o2
  }

  // Handle Date
  if (o1 instanceof Date && o2 instanceof Date) {
    return o1.getTime() === o2.getTime()
  }

  // Handle Array
  if (Array.isArray(o1) && Array.isArray(o2)) {
    if (o1.length !== o2.length) {
      return false
    }

    return o1.every((v1, i) => deepEqual(v1, o2[i]))
  }

  // Handle Object
  if (o1 instanceof Object && o2 instanceof Object) {
    const keys = [...new Set([...Object.keys(o1), ...Object.keys(o2)])] as Extract<keyof T, string>[]

    return keys.every((key) => {
      if (Object.prototype.hasOwnProperty.call(o1, key) && Object.prototype.hasOwnProperty.call(o2, key)) {
        return deepEqual(o1[key], o2[key])
      }
      return false
    })
  }

  throw new Error('Unable to compare object. Its type isn\'t supported.')
}
