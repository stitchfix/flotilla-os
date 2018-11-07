import getValueOfDeepKeyInCurrentQuery from "./getValueOfDeepKeyInCurrentQuery"
import queryUpdateTypes from "./queryUpdateTypes"

export default function getNextQuery({
  currQuery,
  key,
  value,
  index,
  updateType,
}) {
  let currentValue
  switch (updateType) {
    case queryUpdateTypes.SHALLOW:
      return {
        ...currQuery,
        [key]: value,
      }
    case queryUpdateTypes.DEEP_CREATE:
      currentValue = getValueOfDeepKeyInCurrentQuery(currQuery, key)
      return {
        ...currQuery,
        [key]: [...currentValue, value],
      }
    case queryUpdateTypes.DEEP_UPDATE:
      if (index === undefined || index === null) {
        console.error(
          `You attempted to update the value of a nested key without passing an index.`
        )
        return currQuery
      }
      if (!key) {
        console.error(
          `You attempted to update the value of a nested key without passing a key.`
        )
        return currQuery
      }
      currentValue = getValueOfDeepKeyInCurrentQuery(currQuery, key)
      return {
        ...currQuery,
        [key]: [
          ...currentValue.slice(0, index),
          value,
          ...currentValue.slice(index + 1),
        ],
      }
    case queryUpdateTypes.DEEP_REMOVE:
      if (index === undefined || index === null) {
        console.error(
          `You attempted to update the value of a nested key without passing an index.`
        )
        return currQuery
      }
      if (!key) {
        console.error(
          `You attempted to update the value of a nested key without passing a key.`
        )
        return currQuery
      }
      currentValue = getValueOfDeepKeyInCurrentQuery(currQuery, key)
      return {
        ...currQuery,
        [key]: [
          ...currentValue.slice(0, index),
          ...currentValue.slice(index + 1),
        ],
      }
    default:
      return currQuery
  }
}
