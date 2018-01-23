import React, { cloneElement } from 'react'
import PropTypes from 'prop-types'
import Breadcrumbs from './Breadcrumbs'
import { allowedLocations } from '../constants/'

export default function AppHeader(props) {
  const {
    buttons,
    currentLocation,
  } = props
  return (
    <div className="app-header-container">
      <div className="app-header">
        <div className="app-header-section">
          <Breadcrumbs currentLocation={currentLocation} />
        </div>
        <div className="app-header-section">
          <div className="flex">
            {
              !!buttons &&
              buttons.map((button, i) => !!button ? cloneElement(button, { key: i }) : null)
            }
          </div>
        </div>
      </div>
    </div>
  )
}

AppHeader.propTypes = {
  buttons: PropTypes.arrayOf(PropTypes.node),
  currentLocation: PropTypes.oneOf(Object.values(allowedLocations))
}
