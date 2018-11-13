import React from "react"
import PropTypes from "prop-types"
import Field from "./Field"

const KeyValues = props => (
  <div>
    {Object.keys(props.items).map(key => (
      <Field label={key}>{props.items[key]}</Field>
    ))}
  </div>
)

KeyValues.propTypes = {
  items: PropTypes.objectOf(
    PropTypes.shape({
      key: PropTypes.string.isRequired,
      value: PropTypes.node,
      renderValue: PropTypes.func,
    })
  ),
}

KeyValues.defaultProps = {
  items: {},
}

export default KeyValues
