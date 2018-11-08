import React from "react"
import PropTypes from "prop-types"
import Card from "../Card"

const Form = ({ children }) => (
  <Card containerStyle={{ maxWidth: 600 }} contentStyle={{ padding: 0 }}>
    <div className="key-value-container vertical full-width">{children}</div>
  </Card>
)

Form.displayName = "Form"

Form.propTypes = {
  children: PropTypes.node,
}

export default Form
