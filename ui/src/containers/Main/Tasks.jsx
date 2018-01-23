import React from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import DebounceInput from 'react-debounce-input'
import { ReactSelectWrapper, serverTableConnect, Loader, ServerTable } from '../../components/'
import { getApiRoot } from '../../constants/'

const TaskRow = ({ task, path, index }) => (
  <Link
    to={path}
    className="tr"
    key={`task-link-${index}`}
  >
    <div className="td flex j-c" style={{ flex: 1 }}>
      <Link
        to={`${path}/run`}
        onClick={(evt) => { evt.stopPropagation() }}
        className="button"
      >
        Run
      </Link>
    </div>
    <div className="td" style={{ flex: 8 }}>
      {task.alias || task.definition_id}
    </div>
    <div className="td" style={{ flex: 2 }}>
      {task.group_name}
    </div>
    <div className="td" style={{ flex: 1 }}>
      {task.memory} MB
    </div>
  </Link>
)

const Tasks = (props) => {
  const {
    data,
    onQueryChange,
    query,
    groupOpts,
    imageOpts,
    isFetching,
  } = props

  const headers = [
    { displayName: 'Actions', key: 'actions', sortable: false, style: { flex: 1, justifyContent: 'center' } },
    { displayName: 'Alias', key: 'alias', sortable: true, style: { flex: 8 } },
    { displayName: 'Group Name', key: 'group_name', sortable: true, style: { flex: 2 } },
    { displayName: 'Memory', key: 'memory', sortable: true, style: { flex: 1 } },
  ]

  const queryInputs = [
    {
      style: { flex: 2 },
      label: 'Alias',
      input: (
        <DebounceInput
          minLength={1}
          debounceTimeout={250}
          onChange={(evt) => { onQueryChange('alias', evt.target.value) }}
          value={query.alias || ''}
          className="input"
        />
      )
    },
    {
      style: { flex: 1 },
      label: 'Group',
      input: (
        <ReactSelectWrapper
          value={query.group_name}
          onChange={(o) => { onQueryChange('group_name', o) }}
          options={groupOpts}
        />
      )
    },
    {
      style: { flex: 1 },
      label: 'Image',
      input: (
        <ReactSelectWrapper
          value={query.image}
          onChange={(o) => { onQueryChange('image', o) }}
          options={imageOpts}
        />
      )
    }
  ]

  return (
    <div className="section-container">
      <ServerTable
        headers={headers}
        queryInputs={queryInputs}
        {...props}
      >
        {
          isFetching ?
            <Loader containerStyle={{ height: 960 }} /> :
          data && data.definitions ? data.definitions.map((d, i) => (
            <TaskRow
              path={`/tasks/${d.definition_id}`}
              index={i}
              task={d}
              key={`tasks-row-${i}`}
            />
          )) : null
        }
      </ServerTable>
    </div>
  )
}

const mapStateToProps = state => ({
  groupOpts: state.dropdownOpts.group,
  imageOpts: state.dropdownOpts.image,
})

export default connect(mapStateToProps)(
  serverTableConnect({
    urlRoot: () => `${getApiRoot()}/task?`,
    initialQuery: {
      sort_by: 'alias',
      order: 'asc',
    }
  })(Tasks)
)
