import React, { Component } from 'react'
import { withRouter } from 'react-router'
import { Field } from 'redux-form'
import { X } from 'react-feather'
import { QueryUpdateTypes } from '../constants/'
import { ReduxFormGroup } from '../components/'

class EnvFormSection extends Component {
  createEnv() {
    this.props.onChange({
      envName: '',
      envValue: '',
      updateType: QueryUpdateTypes.NESTED_CREATE
    })
  }
  updateEnv({ which, value, index }) {
    const env = decodeURIComponent(this.props.location.query.env).split(',')
    let envName = env[index].split('|')[0] || ''
    let envValue = env[index].split('|')[1] || ''
    if (which === 'name') {
      envName = value
    } else if (which === 'value') {
      envValue = value
    }

    this.props.onChange({
      envName,
      envValue,
      index,
      updateType: QueryUpdateTypes.NESTED_UPDATE
    })
  }
  removeEnv({ index }) {
    this.props.onChange({
      index,
      updateType: QueryUpdateTypes.NESTED_REMOVE
    })
  }
  render() {
    const {
      fields,
      shouldSyncWithUrl,
    } = this.props
    return (
      <div className="section-container task-definition">
        <div className="section-header">
          <div className="section-header-text">Environment Variables</div>
          <button
            type="button"
            className="button"
            onClick={() => {
              fields.push({})
              if (shouldSyncWithUrl) { this.createEnv() }
            }}
          >
            Add
          </button>
        </div>
        <div>
          {
            fields.map((env, i) => (
              <div className="flex a-c" key={`envvar-${i}`}>
                <Field
                  name={`${env}.name`}
                  component={ReduxFormGroup}
                  props={{ custom: {
                    label: i === 0 ? 'Name' : null,
                    isRequired: true,
                    inputType: 'input',
                  } }}
                  onChange={(evt, newValue) => {
                    if (shouldSyncWithUrl) {
                      this.updateEnv({ index: i, value: newValue, which: 'name' })
                    }
                  }}
                />
                <Field
                  name={`${env}.value`}
                  component={ReduxFormGroup}
                  props={{
                    custom: {
                      label: i === 0 ? 'Value' : null,
                      isRequired: true,
                      inputType: 'input',

                    }
                  }}
                  onChange={(evt, newValue) => {
                    if (shouldSyncWithUrl) {
                      this.updateEnv({ index: i, value: newValue, which: 'value' })
                    }
                  }}
                />
                <button
                  type="button"
                  className="button button-error button-circular"
                  onClick={() => {
                    fields.remove(i)
                    if (shouldSyncWithUrl) { this.removeEnv({ index: i }) }
                  }}
                  style={{ transform: 'translate(-5px, 8px)' }}
                ><X size={14} /></button>
              </div>
            ))
          }
        </div>
      </div>
    )
  }
}

export default withRouter(EnvFormSection)
