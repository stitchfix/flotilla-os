import { has } from "lodash"
import queryUpdateTypes from "./queryUpdateTypes"

export default function validateUpdateOptions(opts) {
  // Log an error if no key is passed.
  if (!has(opts, "key")) {
    console.error("The options passed to props.updateQuery needs a `key`.")
    return false
  }

  // Log an error if no updateType is passed.
  if (!has(opts, "updateType")) {
    console.error(
      "The options passed to props.updateQuery needs a `updateType`."
    )
    return false
  }

  // Logs an error if updateType is invalid.
  if (!Object.values(queryUpdateTypes).includes(opts.updateType)) {
    console.error(
      `The updateType ${opts.updateType} is invalid. Please use one of ` +
        `${queryUpdateTypes.join(", ")}.`
    )
    return false
  }

  // Logs an error if no value is passed for non-DEEP_REMOVE update types.
  if (!has(opts, "value") && opts.updateType !== queryUpdateTypes.DEEP_REMOVE) {
    console.error("The options passed to props.updateQuery needs a `value`.")
    return false
  }

  // Logs an error if no `index` in DEEP_UPDATE or DEEP_REMOVE.
  if (
    (opts.updateType === queryUpdateTypes.DEEP_UPDATE ||
      opts.updateType === queryUpdateTypes.DEEP_REMOVE) &&
    !has(opts, "index")
  ) {
    console.error(
      "For 'DEEP_UPDATE' or 'DEEP_REMOVE' query updates, an index is required."
    )
    return false
  }

  return true
}
