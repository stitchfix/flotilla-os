import React from "react"
import PropTypes from "prop-types"
import cn from "classnames"

const Card = props => {
  const className = cn({
    "pl-card-container": true,
    "pl-hoverable": !!props.hoverable,
  })

  return (
    <div
      className={`${props.className} ${className}`}
      style={props.containerStyle}
    >
      {!!props.header && (
        <div className="pl-card-header-container">{props.header}</div>
      )}
      {!props.collapsed && (
        <div className="pl-card-content" style={props.contentStyle}>
          {props.children}
        </div>
      )}
      {!!props.footer && (
        <div className="pl-card-footer-container">{props.footer}</div>
      )}
    </div>
  )
}

Card.displayName = "Card"
Card.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  collapsed: PropTypes.bool,
  containerStyle: PropTypes.object,
  contentStyle: PropTypes.object,
  footer: PropTypes.node,
  header: PropTypes.node,
  hoverable: PropTypes.bool,
}
Card.defaultProps = {
  containerStyle: {},
  contentStyle: {},
  hoverable: false,
  className: "",
  collapsed: false,
}

export default Card
