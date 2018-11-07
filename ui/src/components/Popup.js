import React, { Component } from "react"
import PropTypes from "prop-types"
import cn from "classnames"
import Button from "./Button"
import intentTypes from "../constants/intentTypes"

export default class Popup extends Component {
  static displayName = "Popup"
  static propTypes = {
    actions: PropTypes.node,
    autohide: PropTypes.bool.isRequired,
    duration: PropTypes.number,
    hide: PropTypes.func,
    intent: PropTypes.oneOf(Object.values(intentTypes)),
    message: PropTypes.node,
    title: PropTypes.node,
  }

  static defaultProps = {
    autohide: true,
    duration: 5000,
  }

  componentDidMount() {
    const { autohide, duration, hide } = this.props

    if (autohide) {
      window.setTimeout(() => {
        hide()
      }, duration)
    }
  }

  render() {
    const { intent, title, message, actions, hide } = this.props

    const intentClassName = cn({
      "pl-popup-intent": true,
      [`pl-intent-${intent}`]: !!intent,
    })

    return (
      <div className="pl-popup-container">
        <div className="pl-popup">
          <div className={intentClassName} />
          <div className="pl-popup-content">
            {!!title && <h3 className="pl-popup-title">{title}</h3>}
            {!!message && <div className="pl-popup-message">{message}</div>}
            <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
              <Button onClick={hide}>Close</Button>
              {!!actions && <div className="pl-popup-actions">{actions}</div>}
            </div>
          </div>
        </div>
      </div>
    )
  }
}
