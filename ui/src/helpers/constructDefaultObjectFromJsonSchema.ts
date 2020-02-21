import { get, isObject } from "lodash"

const DEFAULT_ARRAY: any[] = []
const DEFAULT_STRING = ""
const DEFAULT_NUM = 0
const DEFAULT_BOOL = false

export default function constructDefaultObjectFromJsonSchema(
  schema: object
): object {
  let root: { [k: string]: any } = {}
  const properties = get(schema, "properties", {})

  if (isObject(properties)) {
    try {
      helper(properties, root)
    } catch (e) {
      console.error(
        "Unable to convert JSONSchema to default object, defaulting to `{}`."
      )
    }
  }

  return root
}

function helper(properties: object, root: { [k: string]: any }): void {
  Object.entries(properties).forEach(([k, v]) => {
    if (v.type) {
      switch (v.type) {
        case "object":
          root[k] = {}
          if (v.properties) helper(v.properties, root[k])
          break
        case "array":
          root[k] = v.default ? v.default : DEFAULT_ARRAY
          break
        case "boolean":
          root[k] = v.default ? v.default : DEFAULT_BOOL
          break
        case "string":
          root[k] = v.default ? v.default : DEFAULT_STRING
          break
        case "number":
          root[k] = v.default ? v.default : DEFAULT_NUM
          break
        default:
          root[k] = v.default ? v.default : null
      }
    }
  })
}
