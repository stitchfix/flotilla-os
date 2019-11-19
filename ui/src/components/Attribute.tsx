import * as React from "react"
import { Tag, Colors } from "@blueprintjs/core"

const Attribute: React.FunctionComponent<{
  name: React.ReactNode
  value: React.ReactNode
  isExperimental?: boolean
  containerStyle?: object
}> = ({ name, value, isExperimental, containerStyle }) => (
  <div
    className="flotilla-attribute-container"
    style={containerStyle ? containerStyle : {}}
  >
    <div className="flotilla-attribute-name">
      {name}{" "}
      {isExperimental && isExperimental === true && (
        <Tag
          style={{
            color: Colors.WHITE,
            fontWeight: 500,
            background: Colors.INDIGO4,
          }}
        >
          BETA
        </Tag>
      )}
    </div>
    <div className="flotilla-attribute-value">{value}</div>
  </div>
)

export default Attribute
