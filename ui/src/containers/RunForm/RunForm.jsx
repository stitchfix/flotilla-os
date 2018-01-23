import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Field, FieldArray, reduxForm } from 'redux-form'
import { isEmpty } from 'lodash'
import { updateRunFormQuery } from '../../actions/'
import { ReduxFormGroup, EnvFormSection } from '../../components/'
import { QueryUpdateTypes } from '../../constants/'

class RunForm extends Component {
  handleClusterChange({ value }) {
    this.props.dispatch(updateRunFormQuery({
      key: 'cluster',
      value,
      updateType: QueryUpdateTypes.SHALLOW
    }))
  }
  handleEnvChange({ envName, envValue, index, updateType }) {
    this.props.dispatch(updateRunFormQuery({
      key: 'env',
      value: `${envName}|${envValue}`,
      index,
      updateType,
    }))
  }
  render() {
    const { handleSubmit, clusterOpts } = this.props
    return (
      <div className="form-container">
        <form onSubmit={handleSubmit}>
          <div className="section-container">
            <div className="section-header">
              <div className="section-header-text">Run Config</div>
            </div>
            <Field
              name="cluster"
              component={ReduxFormGroup}
              props={{ custom: { label: 'Cluster', inputType: 'select', selectOpts: clusterOpts } }}
            />
          </div>
          <FieldArray
            name="env"
            component={EnvFormSection}
            props={{
              shouldSyncWithUrl: true,
              onChange: (q) => { this.handleEnvChange(q) }
            }}
          />
          <div className="section-container">
            <div className="section-content flex j-fs a-c">
              <Field
                name="saveConfig"
                component="input"
                type="checkbox"
              />
              <div style={{ marginLeft: 12 }}>
                Save run configuration for this task.
              </div>
            </div>
          </div>
        </form>
      </div>
    )
  }
}

const envStringToObject = envString => ({
  name: decodeURIComponent(envString).split('|')[0],
  value: decodeURIComponent(envString).split('|')[1]
})

function generateInitialValues(initialValuesObj) {
  const taDa = { saveConfig: true }

  // Set cluster
  if (!!initialValuesObj.cluster) {
    taDa.cluster = initialValuesObj.cluster
  } else {
    taDa.cluster = 'flotilla-adhoc'
  }

  // Set env
  if (!!initialValuesObj.env) {
    // Differentiate between one / multiple values. If only one,
    // will be string. Else will be array.
    if (Array.isArray(initialValuesObj.env)) {
      taDa.env = initialValuesObj.env.map(envStringToObject)
    } else {
      taDa.env = [envStringToObject(initialValuesObj.env)]
    }
  }

  return taDa
}

function mapStateToProps(state) {
  const { query } = state.routing.locationBeforeTransitions
  const props = {
    clusterOpts: state.dropdownOpts.cluster,
  }

  if (!isEmpty(query)) {
    props.initialValues = generateInitialValues(query)
  } /*else if (!isEmpty(state.task.runConfig)) {
    props.initialValues = generateInitialValues(state.task.runConfig)
  }*/

  return props
}

RunForm = reduxForm({ form: 'run' })(RunForm)
export default connect(mapStateToProps)(RunForm)
