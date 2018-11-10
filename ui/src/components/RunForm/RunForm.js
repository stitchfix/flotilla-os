import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { Form as ReactForm } from "react-form"
import { get, isEmpty } from "lodash"

import Button from "../Button"
import Loader from "../Loader"
import View from "../View"
import ViewHeader from "../ViewHeader"

import Form from "../Form/Form"
import FieldSelect from "../Form/FieldSelect"
import FieldKeyValue from "../Form/FieldKeyValue"
import api from "../../api"
import config from "../../config"

import * as requestStateTypes from "../../constants/requestStateTypes"

import TaskContext from "../Task/TaskContext"

class RunForm extends Component {
  handleSubmit = values => {
    // api.runTask()
  }

  getDefaultValues = () => {
    const { data } = this.props

    return {
      cluster: get(config, "DEFAULT_CLUSTER", ""),
      env: get(data, ["env"], []).map(e => ({ key: e.name, value: e.value })),
    }
  }

  render() {
    const { clusterOptions, requestState } = this.props

    if (
      isEmpty(clusterOptions) ||
      requestState === requestStateTypes.NOT_READY
    ) {
      return <Loader />
    }

    return (
      <ReactForm
        defaultValues={this.getDefaultValues()}
        onSubmit={this.handleSubmit}
      >
        {formAPI => {
          return (
            <form onSubmit={formAPI.submitForm}>
              <View>
                <ViewHeader
                  title="fill me out"
                  // title={this.renderTitle()}
                  actions={
                    <Button type="submit" intent="primary">
                      submit
                    </Button>
                  }
                />
                <Form>
                  <FieldSelect
                    label="Cluster"
                    field="cluster"
                    options={clusterOptions}
                  />
                  <FieldKeyValue
                    label="Environment Variables"
                    field="env"
                    addValue={formAPI.addValue}
                    removeValue={formAPI.removeValue}
                    values={get(formAPI, ["values", "env"], [])}
                  />
                </Form>
              </View>
            </form>
          )
        }}
      </ReactForm>
    )
  }
}

RunForm.propTypes = {
  clusterOptions: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string,
      value: PropTypes.string,
    })
  ),
}

const mapStateToProps = state => ({
  clusterOptions: get(state, ["selectOpts", "cluster"], []),
})

export default connect(mapStateToProps)(props => (
  <TaskContext.Consumer>
    {ctx => <RunForm {...props} {...ctx} />}
  </TaskContext.Consumer>
))
