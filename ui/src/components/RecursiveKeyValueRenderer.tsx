import * as React from "react"
import { isArray, isObject } from "lodash"
import Attribute from "./Attribute"
import { Tag } from "@blueprintjs/core"

type Props = {
  data: object
}

class RecursiveKeyValueRenderer extends React.Component<Props> {
  render() {
    const { data } = this.props
    return (
      <div className="flotilla-attributes-container flotilla-attributes-container-vertical">
        {Object.entries(data).map(([k, v]) => {
          let value: React.ReactNode = v

          if (isArray(value)) {
            value = (
              <div>
                {value.map(x => (
                  <Tag style={{ marginRight: 8 }}>{JSON.stringify(x)}</Tag>
                ))}
              </div>
            )
          } else if (isObject(value)) {
            value = <RecursiveKeyValueRenderer data={v} />
          }

          return <Attribute key={k} name={k} value={value} />
        })}
      </div>
    )
  }
}

export default RecursiveKeyValueRenderer
