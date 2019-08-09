import * as React from "react"

const Attribute: React.FunctionComponent<{
  name: string
  value: React.ReactNode
}> = ({ name, value }) => (
  <div className="flotilla-attribute-container">
    <div className="flotilla-attribute-name">{name}</div>
    <div className="flotilla-attribute-value">{value}</div>
  </div>
)

export default Attribute
