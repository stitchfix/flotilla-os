import * as React from "react"
import { Formik, Form, FastField, Field } from "formik"
import * as Yup from "yup"
import { RouteComponentProps } from "react-router-dom"
import {
  FormGroup,
  Button,
  Intent,
  Spinner,
  Classes,
  RadioGroup,
  Radio,
  Collapse,
} from "@blueprintjs/core"
import api from "../api"
import { LaunchRequestV2, Run, ExecutionEngine } from "../types"
import getInitialValuesForTaskRun from "../helpers/getInitialValuesForTaskRun"
import Request, {
  ChildProps as RequestChildProps,
  RequestStatus,
} from "./Request"
import EnvFieldArray from "./EnvFieldArray"
import ClusterSelect from "./ClusterSelect"
import { TaskContext, TaskCtx } from "./Task"
import Toaster from "./Toaster"
import ErrorCallout from "./ErrorCallout"
import FieldError from "./FieldError"
import NodeLifecycleSelect from "./NodeLifecycleSelect"
import * as helpers from "../helpers/runFormHelpers"
import { commandFieldSpec } from "../helpers/taskFormHelpers"

const validationSchema = Yup.object().shape({
  owner_id: Yup.string(),
  cluster: Yup.string().required("Required"),
  memory: Yup.number()
    .required("Required")
    .min(0),
  cpu: Yup.number()
    .required("Required")
    .min(512),
  env: Yup.array().of(
    Yup.object().shape({
      name: Yup.string().required(),
      value: Yup.string().required(),
    })
  ),
  engine: Yup.string()
    .matches(/(eks|ecs)/)
    .required("A valid engine type of ecs or eks must be set."),
  node_lifecycle: Yup.string().matches(/(spot|ondemand)/),
  command: Yup.string()
    .min(1)
    .nullable(),
})

type Props = RequestChildProps<
  Run,
  { definitionID: string; data: LaunchRequestV2 }
> & {
  definitionID: string
  initialValues: LaunchRequestV2
}

type State = {
  areAdvancedOptionsVisible: boolean
}

class RunForm extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.toggleAdvancedOptionsVisibility = this.toggleAdvancedOptionsVisibility.bind(
      this
    )
  }
  state = {
    areAdvancedOptionsVisible: false,
  }

  toggleAdvancedOptionsVisibility() {
    this.setState(prev => ({
      areAdvancedOptionsVisible: !prev.areAdvancedOptionsVisible,
    }))
  }

  render() {
    const {
      initialValues,
      request,
      requestStatus,
      isLoading,
      error,
      definitionID,
    } = this.props
    const { areAdvancedOptionsVisible } = this.state
    return (
      <Formik
        isInitialValid={(values: any) =>
          validationSchema.isValidSync(values.initialValues)
        }
        initialValues={initialValues}
        validationSchema={validationSchema}
        onSubmit={data => {
          request({ definitionID, data })
        }}
      >
        {({ errors, values, setFieldValue, isValid, ...rest }) => {
          const getEngine = (): ExecutionEngine => values.engine
          return (
            <Form className="flotilla-form-container">
              {requestStatus === RequestStatus.ERROR && error && (
                <ErrorCallout error={error} />
              )}
              {/* Owner ID Field */}
              <FormGroup
                label={helpers.ownerIdFieldSpec.label}
                helperText={helpers.ownerIdFieldSpec.description}
              >
                <FastField
                  name={helpers.ownerIdFieldSpec.name}
                  value={values.owner_id}
                  className={Classes.INPUT}
                />
                {errors.owner_id && <FieldError>{errors.owner_id}</FieldError>}
              </FormGroup>
              <div className="flotilla-form-section-divider" />
              {/* Engine Type Field */}
              <RadioGroup
                inline
                label="Engine Type"
                onChange={(evt: React.FormEvent<HTMLInputElement>) => {
                  setFieldValue("engine", evt.currentTarget.value)

                  if (evt.currentTarget.value === ExecutionEngine.EKS) {
                    setFieldValue(
                      "cluster",
                      process.env.REACT_APP_EKS_CLUSTER_NAME || ""
                    )
                  } else if (getEngine() === ExecutionEngine.EKS) {
                    setFieldValue("cluster", "")
                  }
                }}
                selectedValue={values.engine}
              >
                <Radio label="EKS" value={ExecutionEngine.EKS} />
                <Radio label="ECS" value={ExecutionEngine.ECS} />
              </RadioGroup>
              <div className="flotilla-form-section-divider" />

              {/*
                Cluster Field. Note: this is a "Field" rather than a
                "FastField" as it needs to re-render when value.engine is
                updated.
              */}
              {getEngine() !== ExecutionEngine.EKS && (
                <FormGroup
                  label="Cluster"
                  helperText="Select a cluster for this task to execute on."
                >
                  <Field
                    name="cluster"
                    component={ClusterSelect}
                    value={values.cluster}
                    onChange={(value: string) => {
                      setFieldValue("cluster", value)
                    }}
                  />
                  {errors.cluster && <FieldError>{errors.cluster}</FieldError>}
                </FormGroup>
              )}

              {/* CPU Field */}
              <FormGroup
                label={helpers.cpuFieldSpec.label}
                helperText={helpers.cpuFieldSpec.description}
              >
                <FastField
                  type="number"
                  name={helpers.cpuFieldSpec.name}
                  className={Classes.INPUT}
                  min="512"
                />
                {errors.cpu && <FieldError>{errors.cpu}</FieldError>}
              </FormGroup>

              {/* Memory Field */}
              <FormGroup
                label={helpers.memoryFieldSpec.label}
                helperText={helpers.memoryFieldSpec.description}
              >
                <FastField
                  type="number"
                  name={helpers.memoryFieldSpec.name}
                  className={Classes.INPUT}
                />
                {errors.memory && <FieldError>{errors.memory}</FieldError>}
              </FormGroup>
              <div className="flotilla-form-section-divider" />

              {/* Advanced Options */}
              <div className="flotilla-form-section-header-container">
                <div>Advanced Options</div>
                <Button onClick={this.toggleAdvancedOptionsVisibility}>
                  {areAdvancedOptionsVisible ? "Hide" : "Show"}
                </Button>
              </div>
              <Collapse isOpen={areAdvancedOptionsVisible} keepChildrenMounted>
                {/* Node Lifecycle Field */}
                <FormGroup
                  label="Node Lifecycle"
                  helperText="This field is only applicable to tasks running on EKS. For more information, please view this document: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html"
                >
                  <Field
                    name="node_lifecycle"
                    component={NodeLifecycleSelect}
                    value={values.node_lifecycle}
                    onChange={(value: string) => {
                      setFieldValue("node_lifecycle", value)
                    }}
                    isDisabled={getEngine() !== ExecutionEngine.EKS}
                  />
                  {errors.node_lifecycle && (
                    <FieldError>{errors.node_lifecycle}</FieldError>
                  )}
                </FormGroup>

                <FormGroup
                  label={commandFieldSpec.label}
                  helperText="Override your task definition command."
                >
                  <FastField
                    className={`${Classes.INPUT} ${Classes.CODE}`}
                    component="textarea"
                    name={commandFieldSpec.name}
                    rows={14}
                    style={{ fontSize: "0.8rem" }}
                  />
                  {errors.command && <FieldError>{errors.command}</FieldError>}
                </FormGroup>
              </Collapse>
              <div className="flotilla-form-section-divider" />
              <EnvFieldArray />
              <Button
                intent={Intent.PRIMARY}
                type="submit"
                disabled={isLoading || isValid === false}
                style={{ marginTop: 24 }}
                large
              >
                Submit
              </Button>
            </Form>
          )
        }}
      </Formik>
    )
  }
}

const Connected: React.FunctionComponent<RouteComponentProps> = ({
  location,
  history,
}) => (
  <Request<Run, { definitionID: string; data: LaunchRequestV2 }>
    requestFn={api.runTask}
    shouldRequestOnMount={false}
    onSuccess={(data: Run) => {
      Toaster.show({
        message: `Run ${data.run_id} submitted successfully!`,
        intent: Intent.SUCCESS,
      })
      history.push(`/runs/${data.run_id}`)
    }}
    onFailure={() => {
      Toaster.show({
        message: "An error occurred.",
        intent: Intent.DANGER,
      })
    }}
  >
    {requestProps => (
      <TaskContext.Consumer>
        {(ctx: TaskCtx) => {
          switch (ctx.requestStatus) {
            case RequestStatus.ERROR:
              return <ErrorCallout error={ctx.error} />
            case RequestStatus.READY:
              if (ctx.data) {
                const initialValues: LaunchRequestV2 = getInitialValuesForTaskRun(
                  {
                    task: ctx.data,
                    routerState: location.state,
                  }
                )
                return (
                  <RunForm
                    definitionID={ctx.definitionID}
                    initialValues={initialValues}
                    {...requestProps}
                  />
                )
              }
              break
            case RequestStatus.NOT_READY:
            default:
              return <Spinner />
          }
        }}
      </TaskContext.Consumer>
    )}
  </Request>
)

export default Connected
