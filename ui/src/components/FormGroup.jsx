import React from 'react'
import PropTypes from 'prop-types'

export default function FormGroup(props) {
  const {
    label, input, description, hasError, errorText, style, horizontal,
    isStatic, children, largeStaticFont
  } = props

  return (
    <div className={`form-group ${horizontal ? 'horizontal' : 'vertical'}`} style={style}>
      <div className="form-group-content">
        {!!label && <div className="form-group-label">{label}</div>}
        {
          !!isStatic ?
            <div className={`form-group-static ${largeStaticFont && 'text-large'}`}>{children}</div> :
            <div className={`form-group-input ${hasError ? 'has-error' : ''}`}>
              {input}
            </div>
        }
      </div>
      {!!description && <div className="form-group-description">{description}</div>}
      {!!hasError && <div className="form-group-error">{errorText}</div>}
    </div>
  )
}

FormGroup.propTypes = {
  label: PropTypes.node,
  input: PropTypes.node,
  description: PropTypes.node,
  hasError: PropTypes.bool,
  errorText: PropTypes.node,
  style: PropTypes.object,
  horizontal: PropTypes.bool,
  isStatic: PropTypes.bool,
  children: PropTypes.node,
  largeStaticFont: PropTypes.bool
}
FormGroup.defaultProps = {
  hasError: false,
  horizontal: false,
  isStatic: false,
}
