import React from "react"
import PropTypes from "prop-types"
import cn from "classnames"

const Loader = ({ containerStyle = {}, spinnerStyle, mini = false }) => {
  const containerClassName = cn({
    "loader-container": true,
    mini: !!mini,
  })
  const loaderClassName = cn({
    loader: true,
    mini: !!mini,
  })
  return (
    <div className={containerClassName} style={containerStyle}>
      <div className={loaderClassName} style={spinnerStyle} />
    </div>
  )
}

Loader.displayName = "Loader"
Loader.propTypes = {
  containerStyle: PropTypes.object,
  mini: PropTypes.bool,
  spinnerStyle: PropTypes.object,
}
Loader.defaultProps = {
  mini: false,
}

export default Loader
