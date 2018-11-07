import React from "react"
import PropTypes from "prop-types"

const FormGroup = props => {
  const {
    children,
    description,
    errorText,
    hasError,
    horizontal,
    input,
    isRequired,
    isStatic,
    label,
    largeStaticFont,
    style,
    labelStyle,
    inputStyle,
    staticTextStyle,
    descriptionStyle,
    errorStyle,
  } = props

  return (
    <div
      className={`pl-form-group ${horizontal ? "horizontal" : "vertical"}`}
      style={style}
    >
      <div className="pl-form-group-content">
        {!!label && (
          <div
            className={`pl-form-group-label ${
              isRequired ? " is-required" : ""
            }`}
            style={labelStyle}
          >
            {label}
          </div>
        )}
        {!!isStatic ? (
          <div
            className={`pl-form-group-static ${largeStaticFont &&
              "text-large"}`}
            style={staticTextStyle}
          >
            {children}
          </div>
        ) : (
          <div
            className={`pl-form-group-input ${hasError ? "has-error" : ""}`}
            style={inputStyle}
          >
            {input}
          </div>
        )}
      </div>
      {!!description && (
        <div className="pl-form-group-description" style={descriptionStyle}>
          {description}
        </div>
      )}
      {!!hasError && (
        <div className="pl-form-group-error" style={errorStyle}>
          {errorText}
        </div>
      )}
    </div>
  )
}

FormGroup.displayName = "FormGroup"
FormGroup.propTypes = {
  children: PropTypes.node,
  description: PropTypes.node,
  descriptionStyle: PropTypes.object,
  errorStyle: PropTypes.object,
  errorText: PropTypes.node,
  hasError: PropTypes.bool.isRequired,
  horizontal: PropTypes.bool.isRequired,
  input: PropTypes.node,
  inputStyle: PropTypes.object,
  isRequired: PropTypes.bool.isRequired,
  isStatic: PropTypes.bool.isRequired,
  label: PropTypes.node,
  labelStyle: PropTypes.object,
  largeStaticFont: PropTypes.bool,
  staticTextStyle: PropTypes.object,
  style: PropTypes.object,
}
FormGroup.defaultProps = {
  hasError: false,
  horizontal: false,
  isRequired: false,
  isStatic: false,
}

export default FormGroup
