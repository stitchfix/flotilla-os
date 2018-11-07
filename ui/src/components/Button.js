import React from "react"
import PropTypes from "prop-types"
import cn from "classnames"
import colors from "../constants/colors"
import intentTypes from "../constants/intentTypes"
import Loader from "./Loader"

const Button = props => {
  const className = cn({
    "pl-button": true,
    [`pl-intent-${props.intent}`]: !!props.intent,
    "pl-small": !!props.small,
    "pl-invert": !!props.invert,
  })

  return (
    <button
      className={className}
      onClick={props.onClick}
      disabled={props.disabled || props.isLoading}
      type={props.type}
    >
      {props.isLoading ? (
        <Loader
          mini
          spinnerStyle={{
            borderColor: colors.gray.gray_2,
            borderLeftColor: colors.light_gray.light_gray_3,
          }}
        />
      ) : (
        props.children
      )}
    </button>
  )
}

Button.displayName = "Button"
Button.propTypes = {
  children: PropTypes.node,
  disabled: PropTypes.bool.isRequired,
  intent: PropTypes.oneOf(Object.values(intentTypes)),
  invert: PropTypes.bool.isRequired,
  isLoading: PropTypes.bool.isRequired,
  onClick: PropTypes.func,
  small: PropTypes.bool.isRequired,
  type: PropTypes.string.isRequired,
}
Button.defaultProps = {
  isLoading: false,
  small: false,
  invert: false,
  disabled: false,
  type: "button",
}

export default Button
