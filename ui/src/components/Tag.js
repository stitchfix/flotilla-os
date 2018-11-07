import React from "react"
import PropTypes from "prop-types"
import cn from "classnames"
import intentTypes from "../constants/intentTypes"

const Tag = props => (
  <div
    className={cn({
      "pl-tag": true,
      [`pl-${props.intent}`]: !!props.intent,
    })}
  >
    {props.children}
  </div>
)

Tag.displayName = "Tag"
Tag.propTypes = {
  children: PropTypes.node,
  intent: PropTypes.oneOf(Object.values(intentTypes)),
}

export default Tag
