import * as React from "react"
import { withRouter } from "react-router-dom"
import { Formik, FormikProps, Form, Field } from "formik"
import { get, omit } from "lodash"
import Loader from "../styled/Loader"
import View from "../styled/View"
import Navigation from "../Navigation/Navigation"
import StyledForm from "../styled/Form"
import api from "../../api"
import config from "../../config"
import TaskContext from "../Task/TaskContext"
import filterInvalidRunEnv from "../../helpers/filterInvalidRunEnv"
import { FormikFieldSelect } from "../Field/FieldSelect"
import FormikKVField from "../Field/FormikKVField"
import {
  IFlotillaEnv,
  IFlotillaRunTaskPayload,
  IFlotillaUITaskContext,
  flotillaUIRequestStates,
  IFlotillaUINavigationLink,
  flotillaUIIntents,
  IFlotillaUIBreadcrumb,
  IFlotillaAPIError,
  IFlotillaUIPopupProps,
} from "../../.."
import PopupContext from "../Popup/PopupContext"

interface IRunFormProps extends IFlotillaUITaskContext {
  push: (opt: any) => void
  previousRunState?: IFlotillaRunTaskPayload
  goBack: () => void
  renderPopup: (p: IFlotillaUIPopupProps) => void
}

class RunForm extends React.PureComponent<IRunFormProps> {
  handleSubmit = (values: IFlotillaRunTaskPayload): void => {
    const { data, push } = this.props

    if (data && data.definition_id) {
      api
        .runTask({
          values,
          definitionID: data.definition_id,
        })
        .then(res => {
          push(`/runs/${res.run_id}`)
        })
        .catch(this.handleSubmitError)
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

  getDefaultValues = (): IFlotillaRunTaskPayload => {
    const { data, previousRunState } = this.props

    const cluster = get(
      previousRunState,
      "cluster",
      get(config, "DEFAULT_CLUSTER", "")
    )

    let env

    if (previousRunState && previousRunState.env) {
      env = filterInvalidRunEnv(previousRunState.env)
    } else {
      env = get(data, "env", [])
    }

    let runTags: IFlotillaEnv[]

    if (previousRunState && previousRunState.run_tags) {
      runTags = previousRunState.run_tags
    } else {
      const requiredRunTags: string[] = get(config, "REQUIRED_RUN_TAGS", [])
      runTags = requiredRunTags.map((name: string) => ({ name, value: "" }))
    }

    return {
      cluster,
      env,
      run_tags: runTags,
    }
  }

  getRunTagsDescription = (): string => {
    const requiredRunTags = get(config, "REQUIRED_RUN_TAGS", [])

    if (requiredRunTags.length < 1) {
      return ""
    }

    return `The following run tags must be filled out in order for this task to run: ${requiredRunTags}.`
  }

  getActions = ({
    shouldDisableSubmitButton,
  }: {
    shouldDisableSubmitButton: boolean
  }): IFlotillaUINavigationLink[] => {
    return [
      {
        isLink: false,
        text: "Cancel",
        buttonProps: {
          onClick: this.props.goBack,
          type: "button",
        },
      },
      {
        isLink: false,
        text: "Run",
        buttonProps: {
          type: "submit",
          intent: flotillaUIIntents.PRIMARY,
          isDisabled: shouldDisableSubmitButton,
        },
      },
    ]
  }

  getBreadcrumbs = (): IFlotillaUIBreadcrumb[] => {
    const { definitionID, data } = this.props

    return [
      { text: "Tasks", href: "/tasks" },
      {
        text: get(data, "alias", definitionID),
        href: `/tasks/${definitionID}`,
      },
      { text: "Run", href: `/tasks/${definitionID}/run` },
    ]
  }

  render() {
    const { requestState, definitionID, data } = this.props

    if (requestState === flotillaUIRequestStates.NOT_READY) {
      return <Loader />
    }

    return (
      <Formik
        initialValues={this.getDefaultValues()}
        onSubmit={this.handleSubmit}
        validateOnChange={false}
      >
        {(formikProps: FormikProps<IFlotillaRunTaskPayload>) => {
          return (
            <Form>
              <View>
                <Navigation
                  breadcrumbs={this.getBreadcrumbs()}
                  actions={this.getActions({
                    shouldDisableSubmitButton: formikProps.isValid !== true,
                  })}
                />
                <StyledForm title={`Run ${get(data, "alias", definitionID)}`}>
                  <Field
                    name="cluster"
                    value={formikProps.values.cluster}
                    onChange={formikProps.handleChange}
                    component={FormikFieldSelect}
                    label="Cluster"
                    description="Select a cluster for this task to be executed on."
                    requestOptionsFn={api.getClusters}
                    shouldRequestOptions
                    isCreatable
                    isRequired
                  />
                  <FormikKVField
                    name="run_tags"
                    value={formikProps.values.run_tags}
                    description={this.getRunTagsDescription()}
                    isKeyRequired
                    isValueRequired={false}
                    label="Environment Variables"
                    setFieldValue={formikProps.setFieldValue}
                  />
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
          )
        }}
      </Formik>
    )
  }
}

const WrappedRunForm = withRouter(props => {
  return (
    <PopupContext.Consumer>
      {popupContext => (
        <TaskContext.Consumer>
          {ctx => (
            <RunForm
              push={props.history.push}
              previousRunState={get(props, ["location", "state"])}
              goBack={props.history.goBack}
              renderPopup={popupContext.renderPopup}
              {...omit(props, [
                "history",
                "location",
                "match",
                "staticContext",
              ])}
              {...ctx}
            />
          )}
        </TaskContext.Consumer>
      )}
    </PopupContext.Consumer>
  )
}) as React.ComponentType<any>

export default WrappedRunForm
