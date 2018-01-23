import React, { Component, createElement } from 'react'
import { connect } from 'react-redux'
import { withRouter } from 'react-router'
import { Field, FieldArray, reduxForm } from 'redux-form'
import { EnvFormSection, ReduxFormGroup } from '../../components/'
import { imageTagsEndpoint, TaskFormTypes, invalidEnv } from '../../constants/'
import validate from './validate'

class TaskForm extends Component {
  state = {
    imageHasBeenSelected: false,
    image: undefined,
    imageTagOpts: [],
  }
  componentDidMount() {
    this.toFocus.focus()
    if (!!this.props.initialValues && !!this.props.initialValues.image) {
      this.handleImageChange(this.props.initialValues.image)
    }
  }
  componentWillReceiveProps(nextProps) {
    if (!!!this.props.initialValues && !!nextProps.initialValues) {
      this.handleImageChange(nextProps.initialValues.image)
    }
  }
  handleImageChange(val) {
    if (!!val && val !== this.state.image) {
      fetch(imageTagsEndpoint(val))
        .then(res => res.json())
        .then((res) => {
          this.setState({
            imageHasBeenSelected: true,
            imageTagOpts: res.tags.map(t => ({ label: t, value: t }))
          })
        })
    }
  }
  render() {
    const {
      handleSubmit,
      groupOpts,
      imageOpts,
      tagOpts,
      formType,
    } = this.props
    const {
      imageHasBeenSelected,
      imageTagOpts,
    } = this.state
    const shouldDisableImageTagSelect = formType === TaskFormTypes.new && !imageHasBeenSelected

    const fieldsConfig = [
      {
        name: 'alias',
        component: ReduxFormGroup,
        props: {
          custom: {
            label: 'Alias',
            isRequired: true,
            inputType: 'input',
            ref: (toFocus) => { this.toFocus = toFocus }
          }
        }
      },
      {
        name: 'group',
        component: ReduxFormGroup,
        props: {
          custom: {
            label: 'Group',
            isRequired: true,
            selectOpts: groupOpts,
            inputType: 'select',
            allowCreate: true
          }
        }
      },
      {
        name: "image",
        component: ReduxFormGroup,
        props: {
          custom: {
            label: 'Image',
            isRequired: true,
            selectOpts: imageOpts,
            inputType: 'select'
          }
        },
        onChange: (evt, newVal) => {
          this.handleImageChange(newVal)
        }
      },
      {
        name: "imageTag",
        component: ReduxFormGroup,
        props: {
          custom: {
            label: 'Image Tag',
            isRequired: true,
            selectOpts: imageTagOpts,
            disabled: shouldDisableImageTagSelect,
            inputType: 'select'
          }
        }
      },
      {
        name: "command",
        component: ReduxFormGroup,
        props: {
          custom: {
            label: 'Command',
            inputType: 'textarea',
            style: { fontSize: 12 },
            isRequired: true
          }
        }
      },
      {
        name: "memory",
        component: ReduxFormGroup,
        props: {
          custom: {
            label: 'Memory (MB)',
            inputType: 'input',
            isRequired: true,
            type: 'number'
          }
        }
      },
      {
        name: "tags",
        component: ReduxFormGroup,
        props: {
          custom: {
            label: 'Tags',
            inputType: 'select',
            selectOpts: tagOpts,
            multi: true,
            allowCreate: true
          }
        }
      }
    ]

    return (
      <div className="form-container">
        <form onSubmit={handleSubmit}>
          <div className="section-container">
            <div className="section-header">
              <div className="section-header-text">Task Definition</div>
            </div>
            {fieldsConfig.map((f, i) => createElement(Field, { ...f, key: `task-form-field-${i}` }))}
          </div>
          <FieldArray
            name="env"
            component={EnvFormSection}
            props={{ shouldSyncWithUrl: false }}
          />
        </form>
      </div>
    )
  }
}

function mapStateToProps(state, ownProps) {
  const props = {
    groupOpts: state.dropdownOpts.group,
    imageOpts: state.dropdownOpts.image,
    tagOpts: state.dropdownOpts.tag,
    _form: !!state.form.task ? state.form.task : undefined
  }

  if (Object.keys(state.task.task).length > 0 &&
      (ownProps.formType === TaskFormTypes.edit || ownProps.formType === TaskFormTypes.copy)) {
    const task = state.task.task
    const image = !!task.image ? task.image.split('/')[1].split(':') : []
    props.initialValues = {
      group: task.group_name,
      image: image[0],
      imageTag: image[1],
      command: task.command,
      memory: task.memory,
      tags: task.tags,
      env: task.env.filter(e => !invalidEnv.includes(e.name)),
    }

    // Note: don't set the alias' initial value for copying tasks,
    // since tasks can't have the same alias.
    if (ownProps.formType === TaskFormTypes.edit) {
      props.initialValues.alias = task.alias
    }
  }

  return props
}

TaskForm = reduxForm({ form: 'task', validate })(TaskForm)
export default withRouter(connect(mapStateToProps)(TaskForm))
