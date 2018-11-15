import React from "react"
import PropTypes from "prop-types"
import Field from "./Field"

const KeyValues = props => (
  <div>
    {Object.keys(props.items).map(key => (
      <Field label={key} key={key}>
        {props.items[key]}
      </Field>
    ))}
  </div>
)

KeyValues.propTypes = {
  items: PropTypes.objectOf(PropTypes.node),
}

KeyValues.defaultProps = {
  items: {},
}

export default KeyValues
