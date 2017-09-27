import React from 'react'
import { Link } from 'react-router'
import { Trash2, Zap, Copy, Edit } from 'react-feather'
import ReactTooltip from 'react-tooltip'
import { AppHeader } from '../../components/'
import { allowedLocations } from '../../constants/'

export default function TaskNav({ onDeleteButtonClick, hasError }) {
  const buttons = hasError ? null : [
    <div>
      <ReactTooltip id="deleteButton" effect="solid">
        Delete Task
      </ReactTooltip>
      <button
        data-tip
        data-for="deleteButton"
        className="button button-error"
        onClick={onDeleteButtonClick}
      >
        <Trash2 size={14} />
      </button>
    </div>,
    <div data-tip data-for="copyButton">
      <ReactTooltip id="copyButton" effect="solid">
        Copy Task
      </ReactTooltip>
      <Link className="button" to={location => `${location.pathname}/copy`}>
        <Copy size={14} />
      </Link>
    </div>,
    <div data-tip data-for="editButton">
      <ReactTooltip id="editButton" effect="solid">
        Edit Task
      </ReactTooltip>
      <Link className="button" to={location => `${location.pathname}/edit`}>
        <Edit size={14} />
      </Link>
    </div>,
    <Link className="button button-primary" to={location => `${location.pathname}/run`}>
      <Zap size={14} />&nbsp;Run
    </Link>
  ]
  return (
    <AppHeader
      currentLocation={allowedLocations.task}
      buttons={buttons}
    />
  )
}
