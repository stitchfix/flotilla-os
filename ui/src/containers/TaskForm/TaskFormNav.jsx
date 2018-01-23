import React from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import { submit } from 'redux-form'
import { has } from 'lodash'
import { Save, X } from 'react-feather'
import { allowedLocations, TaskFormTypes } from '../../constants/'
import { AppHeader } from '../../components/'

function TaskFormNav({ formType, dispatch, onCancel, isValid }) {
  let location

  if (formType === TaskFormTypes.new) {
    location = allowedLocations.createTask
  } else if (formType === TaskFormTypes.edit) {
    location = allowedLocations.editTask
  } else if (formType === TaskFormTypes.copy) {
    location = allowedLocations.copyTask
  }

  return (
    <AppHeader
      currentLocation={location}
      buttons={[
        <button className="button" onClick={() => { onCancel() }}>
          <X size={14} />&nbsp;Cancel
        </button>,
        <button
          className="button button-primary"
          onClick={() => { dispatch(submit('task')) }}
          disabled={!isValid}
        >
          <Save size={14} />&nbsp;Save
        </button>
      ]}
    />
  )
}

TaskFormNav.propTypes = {
  onCancel: PropTypes.func,
  formType: PropTypes.oneOf(Object.values(TaskFormTypes)),
  dispatch: PropTypes.func,
  isValid: PropTypes.bool,
}

const mapStateToProps = (state) => {
  return ({
    isValid: !has(state.form.task, 'syncErrors')
  })
}

export default connect(mapStateToProps)(TaskFormNav)
