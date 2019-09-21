import * as React from "react"
import { Formik, Form, FastField } from "formik"
import * as Yup from "yup"
import { RouteComponentProps } from "react-router-dom"
import { FormGroup, Button, Intent } from "@blueprintjs/core"
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

const validationSchema = Yup.object().shape({
  cluster: Yup.string().required("Required"),
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
    initialValues={initialValues}
    validationSchema={validationSchema}
    onSubmit={data => {
      request({ definitionID, data })
    }}
  >
    {formik => {
      return (
        <Form className="flotilla-form-container">
          {requestStatus === RequestStatus.ERROR && error && (
            <ErrorCallout error={error} />
          )}
          <FormGroup
            label="Cluster"
            helperText="Select a cluster for this task to execute on."
          >
            <FastField
              name="cluster"
              component={ClusterSelect}
              value={formik.values.cluster}
              onChange={(value: string) => {
                formik.setFieldValue("cluster", value)
              }}
            />
            {formik.errors.cluster && (
              <FieldError>{formik.errors.cluster}</FieldError>
            )}
          </FormGroup>
          <EnvFieldArray />
          <Button
            intent={Intent.PRIMARY}
            type="submit"
            disabled={isLoading || formik.isValid === false}
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
          if (ctx.requestStatus === RequestStatus.READY && ctx.data) {
            const initialValues: RunTaskPayload = getInitialValuesForTaskRun({
              task: ctx.data,
              routerState: location.state,
            })
            return (
              <RunForm
                definitionID={ctx.definitionID}
                initialValues={initialValues}
                {...requestProps}
              />
            )
          }

          if (ctx.requestStatus === RequestStatus.ERROR) {
            return "error"
          }

          return "spinner"
        }}
      </TaskContext.Consumer>
    )}
  </Request>
)

export default Connected
