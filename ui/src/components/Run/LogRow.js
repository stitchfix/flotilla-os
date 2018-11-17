import React, { PureComponent } from "react"
import PropTypes from "prop-types"
import Ansi from "ansi-to-react"
import Pre from "../styled/Pre"

class LogRow extends PureComponent {
  render() {
    const { data, index, style } = this.props

    return (
      <Pre style={style}>
        <Ansi>{data[index]}</Ansi>
      </Pre>
    )
  }
}

LogRow.propTypes = {
  data: PropTypes.array,
  index: PropTypes.number,
  style: PropTypes.object,
}

export default LogRow
