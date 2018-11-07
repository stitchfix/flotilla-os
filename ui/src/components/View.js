import React from "react"
import PropTypes from "prop-types"
import cn from "classnames"

const View = ({ children, noHeader }) => {
  const className = cn({
    "pl-view-container": true,
    "no-header": noHeader,
  })
  return (
    <div className={className}>
      <div className="pl-view-inner">{children}</div>
    </div>
  )
}

View.displayName = "View"
View.propTypes = {
  children: PropTypes.node,
  noHeader: PropTypes.bool.isRequired,
}
View.defaultProps = {
  noHeader: false,
}

export default View
