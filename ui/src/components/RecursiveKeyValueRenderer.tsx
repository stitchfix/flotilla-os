import * as React from "react"
import { isArray, isObject } from "lodash"
import Attribute from "./Attribute"
import { Tag, H1, H2, H3, H4, H5, H6 } from "@blueprintjs/core"

type Props = {
  data: object
  depth: number
  title?: string
}

class RecursiveKeyValueRenderer extends React.Component<Props> {
  static defaultProps = {
    depth: 0,
  }

  renderHeader() {
    const { title, depth } = this.props
    if (title === undefined) return null

    switch (depth) {
      case 0:
        return <H3>{title}</H3>
      case 1:
        return <H4>{title}</H4>
      case 2:
        return <H5>{title}</H5>
      case 3:
      default:
        return <H6>{title}</H6>
    }
  }

  render() {
    const { data, depth } = this.props
    return (
      <div>
        {this.renderHeader()}
        <div className="flotilla-attributes-container flotilla-attributes-container-vertical">
          {Object.entries(data).map(([k, v]) => {
            if (isArray(v)) {
              return (
                <Attribute
                  key={k}
                  name={k}
                  value={v.map(x => (
                    <Tag style={{ marginRight: 8 }}>{x}</Tag>
                  ))}
                />
              )
            }

            if (isObject(v)) {
              return (
                <RecursiveKeyValueRenderer
                  data={v}
                  title={k}
                  depth={depth + 1}
                />
              )
            }

            return <Attribute key={k} name={k} value={v} />
          })}
        </div>
      </div>
    )
  }
}

export default RecursiveKeyValueRenderer
