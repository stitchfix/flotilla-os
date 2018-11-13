import React, { PureComponent } from "react"
import Pre from "../styled/Pre"

class LogRow extends PureComponent {
  render() {
    const { data, index, style } = this.props
    return <Pre style={style}>{data[index]}</Pre>
  }
}

export default LogRow
