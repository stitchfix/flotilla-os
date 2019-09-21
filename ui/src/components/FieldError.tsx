import * as React from "react"
import { Colors } from "@blueprintjs/core"

const FieldError: React.FunctionComponent = ({ children }) => (
  <div style={{ color: Colors.RED3 }}>{children}</div>
)

export default FieldError
