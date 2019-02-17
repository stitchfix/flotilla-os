import * as React from "react"
import { withRouter, RouteComponentProps } from "react-router-dom"
import { Formik, FormikProps, Form, FastField } from "formik"
import * as Yup from "yup"
import { get, omit } from "lodash"
import CreatableSelect from "react-select/lib/Creatable"
import Navigation from "../Navigation/Navigation"
import Loader from "../styled/Loader"
import PopupContext from "../Popup/PopupContext"
import View from "../styled/View"
import StyledForm from "../styled/Form"
import TaskContext from "../Task/TaskContext"
import api from "../../api"
import config from "../../config"
import {
  IFlotillaCreateTaskPayload,
  IFlotillaEditTaskPayload,
  IFlotillaAPIError,
  IFlotillaTaskDefinition,
  flotillaUITaskFormTypes,
  flotillaUIIntents,
  IFlotillaUIPopupProps,
  flotillaUIRequestStates,
  IFlotillaUIBreadcrumb,
  IFlotillaUINavigationLink,
  IReactSelectOption,
} from "../../types"
import FormikKVField from "../Field/FormikKVField"
import StyledField from "../styled/Field"
import {
  stringToSelectOpt,
  preprocessSelectValue,
  preprocessMultiSelectValue,
} from "../../helpers/reactSelectHelpers"

// Shared Yup configuration for all task form types. Yup is used for Formik's
// form validation.
const sharedYup = {
  command: Yup.string()
    .min(1, "")
    .required("Required"),
  memory: Yup.number()
    .min(1, "")
    .required("Required"),
  image: Yup.string()
    .min(1, "")
    .required("Required"),
  group_name: Yup.string()
    .min(1, "")
    .required("Required"),
  tags: Yup.array().of(
    Yup.string()
      .min(1, "")
      .required("Required")
  ),
  env: Yup.array().of(
    Yup.object().shape({
      name: Yup.string()
        .min(1, "")
        .required("Required"),
      value: Yup.string(),
    })
  ),
}

const CreateTaskYupSchema = Yup.object().shape({
  alias: Yup.string()
    .min(1, "")
    .required("Required"),
  ...sharedYup,
})

const EditTaskYupSchema = Yup.object().shape(sharedYup)

interface ITaskFormProps {
  type: flotillaUITaskFormTypes
  data?: IFlotillaTaskDefinition
  requestState?: flotillaUIRequestStates
  definitionID?: string
  requestData?: () => void
}

interface IUnwrappedTaskFormProps extends ITaskFormProps {
  push: (opt: any) => void
  renderPopup: (p: IFlotillaUIPopupProps) => void
  goBack: () => void
}

interface ITaskFormState {
  inFlight: boolean
  error: IFlotillaAPIError | undefined
  groupOptions: IReactSelectOption[]
  tagOptions: IReactSelectOption[]
  hasFetchedOptions: boolean
}

export class TaskForm extends React.PureComponent<
  IUnwrappedTaskFormProps,
  ITaskFormState
> {
  static getCreateTaskPayload = (
    values: IFlotillaCreateTaskPayload
  ): IFlotillaCreateTaskPayload => {
    return {
      alias: get(values, "alias", ""),
      command: values.command,
      env: values.env,
      group_name: values.group_name,
      image: values.image,
      memory: +values.memory,
      tags: values.tags,
    }
  }

  static getEditTaskPayload = (
    values: IFlotillaCreateTaskPayload
  ): IFlotillaEditTaskPayload => {
    return {
      command: values.command,
      env: values.env,
      group_name: values.group_name,
      image: values.image,
      memory: +values.memory,
      tags: values.tags,
    }
  }

  static renderTitle(type: flotillaUITaskFormTypes) {
    switch (type) {
      case flotillaUITaskFormTypes.CREATE:
        return "Create New Task"
      case flotillaUITaskFormTypes.EDIT:
        return "Edit Task"
      case flotillaUITaskFormTypes.COPY:
        return "Copy Task"
      default:
        return "Task Form"
    }
  }

  state = {
    inFlight: false,
    error: undefined,
    hasFetchedOptions: false,
    groupOptions: [],
    tagOptions: [],
  }

  componentDidMount() {
    this.requestSelectOptions()
  }

  /** Requests the groups and tags options. */
  requestSelectOptions(): void {
    Promise.all([api.getGroups(), api.getTags()]).then(
      (values: IReactSelectOption[][]) => {
        this.setState({
          groupOptions: values[0],
          tagOptions: values[1],
          hasFetchedOptions: true,
        })
      }
    )
  }

  handleSubmit = (values: IFlotillaCreateTaskPayload) => {
    const { data, type, push, requestData } = this.props

    this.setState({ inFlight: true })

    switch (type) {
      case flotillaUITaskFormTypes.EDIT:
        api
          .updateTask({
            definitionID: get(data, "definition_id", ""),
            values: TaskForm.getEditTaskPayload(values),
          })
          .then(responseData => {
            this.setState({ inFlight: false })
            if (!!requestData) requestData()
            push(`/tasks/${get(responseData, "definition_id", "")}`)
          })
          .catch(error => {
            this.handleSubmitError(error)
          })
        break
      case flotillaUITaskFormTypes.CREATE:
      case flotillaUITaskFormTypes.COPY:
        api
          .createTask({ values: TaskForm.getCreateTaskPayload(values) })
          .then(responseData => {
            this.setState({ inFlight: false })
            push(`/tasks/${get(responseData, "definition_id", "")}`)
          })
          .catch(error => {
            this.handleSubmitError(error)
          })
        break
      default:
        console.warn("TaskForm's `type` prop was not specified, doing nothing.")
    }
  }

  /**
   * Renders a popup with the error returned by the server.
   */
  handleSubmitError = (error: IFlotillaAPIError) => {
    this.setState({ inFlight: false })
    const { renderPopup } = this.props

    renderPopup({
      body: error.data,
      intent: flotillaUIIntents.ERROR,
      shouldAutohide: false,
      title: `An error occurred (Status Code: ${error.status})`,
    })
  }

  /**
   * For the clone and update forms, the task definition is required to fill
   * out the default values of the form before it can be rendered.
   */
  shouldNotRenderForm = (): boolean => {
    const { type, requestState } = this.props
    const { hasFetchedOptions } = this.state

    if (
      type !== flotillaUITaskFormTypes.CREATE &&
      requestState === flotillaUIRequestStates.NOT_READY
    ) {
      return true
    }

    if (hasFetchedOptions === false) return true

    return false
  }

  /** Returns the default values of the form. */
  getDefaultValues = (): IFlotillaCreateTaskPayload => {
    const { data } = this.props

    return {
      alias: "",
      memory: get(data, "memory", 1024),
      group_name: get(data, "group_name", ""),
      image: get(
        data,
        "image",
        `${get(config, "IMAGE_PREFIX", "")}IMAGE_NAME:IMAGE_TAG`
      ),
      command: get(data, "command", ""),
      env: get(data, "env", []),
      tags: get(data, "tags", []),
    }
  }

  /** Returns a breadcrumbs array. */
  getBreadcrumbs = (): IFlotillaUIBreadcrumb[] => {
    const { type, data, definitionID } = this.props

    if (type === flotillaUITaskFormTypes.CREATE) {
      return [
        { text: "Tasks", href: "/tasks" },
        { text: "Create Task", href: "/tasks/create" },
      ]
    }

    const hrefSuffix = type === flotillaUITaskFormTypes.COPY ? "copy" : "edit"

    return [
      { text: "Tasks", href: "/tasks" },
      {
        text: data ? data.alias : "",
        href: `/tasks/${definitionID}`,
      },
      {
        text: TaskForm.renderTitle(this.props.type),
        href: `/tasks/${definitionID}/${hrefSuffix}`,
      },
    ]
  }

  /** Returns an action array for the view to render. */
  getActions = ({
    shouldDisableSubmitButton,
  }: {
    shouldDisableSubmitButton: boolean
  }): IFlotillaUINavigationLink[] => {
    const { inFlight } = this.state
    const { goBack } = this.props

    return [
      {
        isLink: false,
        text: "Cancel",
        buttonProps: {
          onClick: goBack,
        },
      },
      {
        isLink: false,
        text: "Submit",
        buttonProps: {
          type: "submit",
          intent: flotillaUIIntents.PRIMARY,
          isDisabled: shouldDisableSubmitButton === true,
          isLoading: !!inFlight,
        },
      },
    ]
  }

  render() {
    const { type } = this.props
    const { groupOptions, tagOptions } = this.state

    // Don't render the form if, say, the task definition for updating a task
    // has not been fetched. Wait until the next render call.
    if (this.shouldNotRenderForm()) {
      return <Loader />
    }

    return (
      <Formik
        initialValues={this.getDefaultValues()}
        onSubmit={this.handleSubmit}
        validationSchema={
          type === flotillaUITaskFormTypes.EDIT
            ? EditTaskYupSchema
            : CreateTaskYupSchema
        }
      >
        {(formikProps: FormikProps<IFlotillaCreateTaskPayload>) => (
          <Form>
            <View>
              <Navigation
                breadcrumbs={this.getBreadcrumbs()}
                actions={this.getActions({
                  shouldDisableSubmitButton: formikProps.isValid !== true,
                })}
              />
              <StyledForm title={TaskForm.renderTitle(this.props.type)}>
                {type !== flotillaUITaskFormTypes.EDIT && (
                  <StyledField
                    label="Alias"
                    description="Choose a descriptive alias for this task."
                    isRequired
                  >
                    <FastField name="alias" />
                  </StyledField>
                )}
                <StyledField
                  label="Group Name"
                  description="Create a new group name or select an existing one to help searching for this task in the future."
                  isRequired
                >
                  <FastField
                    name="group_name"
                    onChange={(selected: IReactSelectOption) => {
                      formikProps.setFieldValue(
                        "group_name",
                        preprocessSelectValue(selected)
                      )
                    }}
                    value={stringToSelectOpt(formikProps.values.group_name)}
                    component={CreatableSelect}
                    options={groupOptions}
                  />
                </StyledField>
                <StyledField
                  label="Image"
                  description="The full URL of the Docker image and tag."
                  isRequired
                >
                  <FastField name="image" />
                </StyledField>
                <StyledField
                  label="Command"
                  description="The command for this task to execute."
                  isRequired
                >
                  <FastField name="command" component="textarea" />
                </StyledField>
                <StyledField
                  label="Memory (MB)"
                  description="The amount of memory this task needs."
                  isRequired
                >
                  <FastField name="memory" type="number" />
                </StyledField>
                <StyledField label="Tags">
                  <FastField
                    name="tags"
                    onChange={(selected: IReactSelectOption[]) => {
                      formikProps.setFieldValue(
                        "tags",
                        preprocessMultiSelectValue(selected)
                      )
                    }}
                    value={get(formikProps, ["values", "tags"], []).map(
                      stringToSelectOpt
                    )}
                    component={CreatableSelect}
                    options={tagOptions}
                    isMulti
                  />
                </StyledField>
                <FormikKVField
                  name="env"
                  value={formikProps.values.env}
                  description="Environment variables that can be adjusted during execution."
                  isKeyRequired
                  isValueRequired={false}
                  label="Environment Variables"
                  setFieldValue={formikProps.setFieldValue}
                />
              </StyledForm>
            </View>
          </Form>
        )}
      </Formik>
    )
  }
}

// Connect the TaskForm component to the router to access various methods and
// to the PopupContext in order to render error messages if the POST fails.
const ConnectedTaskForm: React.ComponentType<any> = withRouter(
  (props: ITaskFormProps & RouteComponentProps<any>) => (
    <PopupContext.Consumer>
      {ctx => (
        <TaskForm
          {...omit(props, ["history", "location", "match", "staticContext"])}
          push={props.history.push}
          goBack={props.history.goBack}
          renderPopup={ctx.renderPopup}
          type={props.type}
        />
      )}
    </PopupContext.Consumer>
  )
)

export const CreateTaskForm: React.SFC<{}> = () => (
  <ConnectedTaskForm type={flotillaUITaskFormTypes.CREATE} />
)

export const UpdateTaskForm: React.SFC<{}> = () => (
  <TaskContext.Consumer>
    {ctx => <ConnectedTaskForm type={flotillaUITaskFormTypes.EDIT} {...ctx} />}
  </TaskContext.Consumer>
)

export const CloneTaskForm: React.SFC<{}> = () => (
  <TaskContext.Consumer>
    {ctx => <ConnectedTaskForm type={flotillaUITaskFormTypes.COPY} {...ctx} />}
  </TaskContext.Consumer>
)
