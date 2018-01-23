import React from 'react'
import PropTypes from 'prop-types'

export default function TablePageButton(props) {
  const {
    pageNumber,
    onClick,
    className,
    isActive,
  } = props

  let style = { width: 50 }
  if (pageNumber < 1) { style = { visibility: 'hidden' } }

  return (
    <button
      className={`button ${isActive ? 'button-active' : ''} ${className}`}
      onClick={() => { onClick(pageNumber) }}
      style={style}
    >
      {pageNumber}
    </button>
  )
}

TablePageButton.propTypes = {
  pageNumber: PropTypes.number.isRequired,
  onClick: PropTypes.func.isRequired,
  className: PropTypes.string,
}
