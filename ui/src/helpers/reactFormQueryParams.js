import { isString } from "lodash"

const ENV_DELIMITER = "|"

export const transformReactFormValuesToQueryParams = values =>
  Object.keys(values).reduce((acc, key) => {
    const value = values[key]

    // Handle special case for environment variables.
    if (key === "env") {
      acc.env = value.map(env => `${env.name}${ENV_DELIMITER}${env.value}`)
    } else {
      acc[key] = values[key]
    }

    return acc
  }, {})

export const transformQueryParamsToReactFormValues = query =>
  Object.keys(query).reduce((acc, key) => {
    const value = query[key]

    if (key === "env") {
      if (isString(value)) {
        const split = value.split(ENV_DELIMITER)
        acc[key] = [{ name: split[0], value: split[1] }]
      } else {
        acc[key] = value.map(e => {
          const split = e.split(ENV_DELIMITER)
          return { name: split[0], value: split[1] }
        })
      }
    } else {
      acc[key] = value
    }

    return acc
  }, {})
