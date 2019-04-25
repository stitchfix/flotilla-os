import { isObject, size, has, toString } from "lodash"

/** Performs a shallow comparison of two query objects. */
const areObjectsEqualShallow = (
  a: { [key: string]: any },
  b: { [key: string]: any }
): boolean => {
  // Return false if the arguments are not objects.
  if (!isObject(a) || !isObject(b)) {
    return false
  }

  // Return false if the size differs.
  if (size(a) !== size(b)) {
    return false
  }

  // Perform shallow comparison.
  for (let key in a) {
    if (!has(b, key)) {
      return false
    }

    if (toString(a[key]) !== toString(b[key])) {
      return false
    }
  }

  return true
}

export default areObjectsEqualShallow
