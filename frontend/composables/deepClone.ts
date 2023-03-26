/**
 * Creates deep clone of object
 */
export function deepClone<T>(obj: T): T {
  // Handle the 3 simple types, and null or undefined
  if (obj == null || typeof obj != 'object') {
    return obj
  }

  // Handle Date
  if (obj instanceof Date) {
    return new Date(obj.getTime()) as T
  }

  // Handle Array
  if (Array.isArray(obj)) {
    function cloneArray<E>(arr: E[]) {
      return arr.map(el => deepClone(el))
    }
    return cloneArray(obj) as T
  }

  // Handle Object
  if (obj instanceof Object) {
    const clone = {} as T
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        clone[key] = deepClone(obj[key])
      }
    }

    return clone
  }

  throw new Error('Unable to clone object. Its type isn\'t supported.')
}
