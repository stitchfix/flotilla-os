import * as React from "react"
import { Formik, Form, FastField } from "formik"
import * as Yup from "yup"
import { RouteComponentProps } from "react-router-dom"
import { FormGroup, Button, Intent, Spinner, Classes } from "@blueprintjs/core"
import api from "../api"
import { RunTaskPayload, Run } from "../types"
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
import * as helpers from "../helpers/runFormHelpers"

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
})

type Props = RequestChildProps<
  Run,
  { definitionID: string; data: RunTaskPayload }
> & {
  definitionID: string
  initialValues: RunTaskPayload
}

const RunForm: React.FunctionComponent<Props> = ({
  initialValues,
  request,
  requestStatus,
  isLoading,
  error,
  definitionID,
}) => (
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
      return (
        <Form className="flotilla-form-container">
          {requestStatus === RequestStatus.ERROR && error && (
            <ErrorCallout error={error} />
          )}
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
          <FormGroup
            label="Cluster"
            helperText="Select a cluster for this task to execute on."
          >
            <FastField
              name="cluster"
              component={ClusterSelect}
              value={values.cluster}
              onChange={(value: string) => {
                setFieldValue("cluster", value)
              }}
            />
            {errors.cluster && <FieldError>{errors.cluster}</FieldError>}
          </FormGroup>
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
          <EnvFieldArray />
          <Button
            intent={Intent.PRIMARY}
            type="submit"
            disabled={isLoading || isValid === false}
          >
            Submit
          </Button>
        </Form>
      )
    }}
  </Formik>
)

const Connected: React.FunctionComponent<RouteComponentProps> = ({
  location,
  history,
}) => (
  <Request<Run, { definitionID: string; data: RunTaskPayload }>
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
                const initialValues: RunTaskPayload = getInitialValuesForTaskRun(
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
