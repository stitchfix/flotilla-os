import React from 'react'
import { Link } from 'react-router'
import { Edit } from 'react-feather'
import { allowedLocations } from '../../constants/'
import { AppHeader } from '../../components/'

export default function MainNav(props) {
  const { location } = props
  const { tasks, runs } = allowedLocations
  return (
    <AppHeader
      currentLocation={location.pathname.endsWith('tasks') ? tasks : runs}
      buttons={[
        <Link to={`/create-task`} className="button button-primary">
          <Edit size={14} />&nbsp;Create Task
        </Link>
      ]}
    />
  )
}
