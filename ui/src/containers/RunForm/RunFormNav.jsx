import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import { submit } from 'redux-form'
import { AppHeader } from '../../components/'
import { allowedLocations } from '../../constants/'

function RunFormNav({ id, dispatch }) {
  return (
    <AppHeader
      currentLocation={allowedLocations.runTask}
      buttons={[
        <Link className="button" to={`/tasks/${id}`}>
          Cancel
        </Link>,
        <button className="button button-primary" onClick={() => { dispatch(submit('run')) }}>Run</button>
      ]}
    />
  )
}

const mapStateToProps = state => ({ id: state.task.task.definition_id || '' })
export default connect(mapStateToProps)(RunFormNav)
