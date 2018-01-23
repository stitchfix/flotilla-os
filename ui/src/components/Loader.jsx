import React from 'react'
import PropTypes from 'prop-types'
import { runStatusTypes } from '../constants/'

export default function Loader({
  containerStyle = {},
  spinnerStyle = {},
  mini = false,
  status
}) {
  let containerClassName = 'loader-container'
  let loaderClassName = 'loader'
  if (mini) {
    containerClassName += ' mini'
    loaderClassName += ' mini'
  }
  if (status) {
    containerClassName += ` ${status}`
    loaderClassName += ` ${status}`
  }
  return (
    <div className={containerClassName} style={containerStyle}>
      <div className={loaderClassName} style={spinnerStyle} />
    </div>
  )
}

Loader.propTypes = {
  containerStyle: PropTypes.object,
  spinnerStyle: PropTypes.object,
  mini: PropTypes.bool,
  status: PropTypes.oneOf(Object.values(runStatusTypes))
}
